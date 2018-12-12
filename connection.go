package gremlin

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

// Clients include the necessary info to connect to the server and the underlying socket
type Pool struct {
	urlStr         string
	origin         string
	ws             *websocket.Conn
	MaxConnections int
	Auth           []OptAuth
	Connections    map[int]*Connection
	active         int
	m              sync.Mutex
	cond           *sync.Cond
	maxBlocker     chan int
}

type Connection struct {
	id   int
	ws   *websocket.Conn
	busy bool
	pool *Pool
}

func NewClient(urlStr string, origin string, maxConn int, options ...OptAuth) (*Pool, error) {
	ws, err := websocket.Dial(urlStr, "", origin)
	if err != nil {
		return nil, err
	}
	pool := &Pool{
		urlStr:         urlStr,
		origin:         origin,
		ws:             ws,
		Auth:           options,
		MaxConnections: maxConn,
		Connections:    make(map[int]*Connection),
		m:              sync.Mutex{},
		active:         0,
		maxBlocker:     make(chan int),
	}
	return pool, nil
}

func (p *Pool) createSocket() (*websocket.Conn, error) {
	ws, err := websocket.Dial(p.urlStr, "", p.origin)
	if err != nil {
		return nil, err
	}
	return ws, nil
}

func (p *Pool) Get() (*Connection, error) {
	for {
		p.m.Lock()
		// If there are available connection
		for i, _ := range p.Connections {
			if !p.Connections[i].busy {
				// TBD: find how to check the connection for being still valid
				p.Connections[i].busy = true
				p.active++
				p.m.Unlock()
				return p.Connections[i], nil
			}
		}
		// In case all connection in use - create new one
		if len(p.Connections) < p.MaxConnections {
			newConnId := len(p.Connections)
			ws, err := p.createSocket()
			if err != nil {
				p.m.Unlock()
				return nil, err
			}
			conn := Connection{
				id:   newConnId,
				ws:   ws,
				pool: p,
				busy: true,
			}
			p.Connections[newConnId] = &conn
			p.active++
			p.m.Unlock()
			return &conn, nil
		}
		p.m.Unlock()
		// Pushing data to `maxBlocker` chan means that it's more concurrent clients than connections and next connection will be taken from already created, so this chan will be unbloked when someone will put back some connnection
		p.maxBlocker <- 1
	}
	return nil, nil
}

// Put - put used connection back to poll and get positive status if it's gone well
func (p *Pool) Put(connId int) bool {
	p.m.Lock()
	for id, _ := range p.Connections {
		if id == connId {
			p.Connections[id].busy = false
			if p.active == p.MaxConnections {
				// This moment is controversial. Here we unblock `maxBlocker` in case someone wait because maximum amount of connections limit exceed so for the next cycle caller has to wait until mutex will be unblocked. It means once that will be that much callers to take all possible connections, one connection, the last one taken after using will be pending here until someone will exceed limit again and will be blocked my mutex. This connection has no difference from others and beghave like all others
				<-p.maxBlocker
			}
			p.active--
			p.m.Unlock()
			return true
		}
	}
	p.m.Unlock()
	return false
}

func (p *Pool) ExecQuery(query string) ([]byte, error) {
	req := Query(query)
	return p.Exec(req)
}

func (p *Pool) Exec(req *Request) ([]byte, error) {
	conn, err := p.Get()
	if err != nil {
		return nil, err
	}
	requestMessage, err := GraphSONSerializer(req)
	if err != nil {
		return nil, err
	}
	// Open a TCP connection
	if err := websocket.Message.Send(conn.ws, requestMessage); err != nil {
		return nil, err
	}

	data, err := conn.ReadResponse()
	// Put connection back concurrently because caller doesn't have to wait
	go p.Put(conn.id)
	return data, err
}

func (c *Connection) ReadResponse() (data []byte, err error) {
	// Data buffer
	var dataItems []json.RawMessage
	inBatchMode := false
	// Receive data
	for {
		var res *Response

		if err = websocket.JSON.Receive(c.ws, &res); err != nil {
			return nil, err
		}

		var items []json.RawMessage
		switch res.Status.Code {
		case StatusNoContent:
			return

		case StatusAuthenticate:
			return c.pool.Authenticate(res.RequestId)
		case StatusPartialContent:
			inBatchMode = true
			if err = json.Unmarshal(res.Result.Data, &items); err != nil {
				return
			}
			dataItems = append(dataItems, items...)

		case StatusSuccess:
			if inBatchMode {
				if err = json.Unmarshal(res.Result.Data, &items); err != nil {
					return
				}
				dataItems = append(dataItems, items...)
				data, err = json.Marshal(dataItems)
			} else {
				data = res.Result.Data
			}
			return

		default:
			if msg, exists := ErrorMsg[res.Status.Code]; exists {
				err = errors.New(msg)
			} else {
				err = errors.New("An unknown error occured")
			}
			return
		}
	}
	return
}

// Authenticates the connection
func (p *Pool) Authenticate(requestId string) ([]byte, error) {
	auth, err := NewAuthInfo(p.Auth...)
	if err != nil {
		return nil, err
	}
	var sasl []byte
	sasl = append(sasl, 0)
	sasl = append(sasl, []byte(auth.User)...)
	sasl = append(sasl, 0)
	sasl = append(sasl, []byte(auth.Pass)...)
	saslEnc := base64.StdEncoding.EncodeToString(sasl)
	args := &RequestArgs{Sasl: saslEnc}
	authReq := &Request{
		RequestId: requestId,
		Processor: "trasversal",
		Op:        "authentication",
		Args:      args,
	}
	return p.Exec(authReq)
}

// AuthInfo includes all info related with SASL authentication with the Gremlin server
// ChallengeId is the  requestID in the 407 status (AUTHENTICATE) response given by the server.
// We have to send an authentication request with that same RequestID in order to solve the challenge.
type AuthInfo struct {
	ChallengeId string
	User        string
	Pass        string
}

type OptAuth func(*AuthInfo) error

// Constructor for different authentication possibilities
func NewAuthInfo(options ...OptAuth) (*AuthInfo, error) {
	auth := &AuthInfo{}
	for _, op := range options {
		err := op(auth)
		if err != nil {
			return nil, err
		}
	}
	return auth, nil
}

// Sets authentication info from environment variables GREMLIN_USER and GREMLIN_PASS
func OptAuthEnv() OptAuth {
	return func(auth *AuthInfo) error {
		user, ok := os.LookupEnv("GREMLIN_USER")
		if !ok {
			return errors.New("Variable GREMLIN_USER is not set")
		}
		pass, ok := os.LookupEnv("GREMLIN_PASS")
		if !ok {
			return errors.New("Variable GREMLIN_PASS is not set")
		}
		auth.User = user
		auth.Pass = pass
		return nil
	}
}

// Sets authentication information from username and password
func OptAuthUserPass(user, pass string) OptAuth {
	return func(auth *AuthInfo) error {
		auth.User = user
		auth.Pass = pass
		return nil
	}
}

var servers []*url.URL

func NewCluster(s ...string) (err error) {
	servers = nil
	// If no arguments use environment variable
	if len(s) == 0 {
		connString := strings.TrimSpace(os.Getenv("GREMLIN_SERVERS"))
		if connString == "" {
			err = errors.New("No servers set. Configure servers to connect to using the GREMLIN_SERVERS environment variable.")
			return
		}
		servers, err = SplitServers(connString)
		return
	}
	// Else use the supplied servers
	for _, v := range s {
		var u *url.URL
		if u, err = url.Parse(v); err != nil {
			return
		}
		servers = append(servers, u)
	}
	return
}

func SplitServers(connString string) (servers []*url.URL, err error) {
	serverStrings := strings.Split(connString, ",")
	if len(serverStrings) < 1 {
		err = errors.New("Connection string is not in expected format. An example of the expected format is 'ws://server1:8182, ws://server2:8182'.")
		return
	}
	for _, serverString := range serverStrings {
		var u *url.URL
		if u, err = url.Parse(strings.TrimSpace(serverString)); err != nil {
			return
		}
		servers = append(servers, u)
	}
	return
}

func CreateConnection() (conn net.Conn, server *url.URL, err error) {
	connEstablished := false
	for _, s := range servers {
		c, err := net.DialTimeout("tcp", s.Host, 1*time.Second)
		if err != nil {
			continue
		}
		connEstablished = true
		conn = c
		server = s
		break
	}
	if !connEstablished {
		err = errors.New("Could not establish connection. Please check your connection string and ensure at least one server is up.")
	}
	return
}
