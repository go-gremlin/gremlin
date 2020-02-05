package gremlin

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/go-kit/kit/log"
	"golang.org/x/net/websocket"
)

const TheBigestMaxGremlinConnectionsLimit = 5

// Clients include the necessary info to connect to the server and the underlying socket
type Pool struct {
	urlStr         string
	origin         string
	MaxConnections int
	Auth           []OptAuth
	Connections    chan Connection
	logger         log.Logger
}

type Connection struct {
	id   int
	ws   *websocket.Conn
	busy bool
	pool *Pool
}

func NewClient(logger log.Logger, urlStr string, origin string, maxConn []int, options ...OptAuth) (*Pool, error) {
	pool := &Pool{
		urlStr: urlStr,
		origin: origin,
		Auth:   options,
		logger: logger,
	}

	// Can make maxConn as an optional since we already have optional arguments here
	if len(maxConn) > 0 {
		pool.MaxConnections = maxConn[0]
	} else {
		pool.MaxConnections = TheBigestMaxGremlinConnectionsLimit
	}
	pool.Connections = make(chan Connection, pool.MaxConnections)
	for i := 0; i < pool.MaxConnections; i++ {
		ws, err := pool.createSocket()
		if err != nil {
			return nil, err
		}
		conn := Connection{
			id:   i,
			ws:   ws,
			pool: pool,
		}
		pool.Connections <- conn
	}
	return pool, nil
}

// createSocket() holds backoff retry logic
func (p *Pool) createSocket() (ws *websocket.Conn, err error) {
	backoff.Retry(func() error {
		ws, err = websocket.Dial(p.urlStr, "", p.origin)
		return err
	}, exponentialBackOff())
	if err != nil {
		p.logger.Log("socketCreateErr", err)
		return ws, err
	}
	return ws, nil
}

func (p *Pool) Get() Connection {
	return <-p.Connections
}

// Put - put used connection back to poll and get positive status if it's gone well
func (p *Pool) Put(conn Connection) {
	p.Connections <- conn
}

func (p *Pool) ExecQuery(query string) ([]byte, error) {
	req := Query(query)
	return p.Exec(req)
}

func (p *Pool) Exec(req *Request) ([]byte, error) {
	requestMessage, err := GraphSONSerializer(req)
	if err != nil {
		p.logger.Log("serializeRequestErr", err)
		return nil, err
	}
	conn := p.Get()
	defer p.Put(conn)
start:
	if err := websocket.Message.Send(conn.ws, requestMessage); err != nil {
		p.logger.Log("sendMessageErr", err)
		return nil, err
	}
	data, err := conn.ReadResponse()
	if err != nil {
		if err == io.EOF { // EOF err we are getting on connection loss
			p.logger.Log("connectionLossErr", err)
			conn.ws, err = p.createSocket()
			if err != nil {
				return nil, err
			}
			p.logger.Log("socketRecovered", conn.id)
			goto start
		} else {
			p.logger.Log("readResponseErr", err)
			return nil, err
		}
	}
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

func exponentialBackOff() *backoff.ExponentialBackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     128 * time.Millisecond,
		RandomizationFactor: 0.5,
		Multiplier:          2,
		MaxInterval:         512 * time.Millisecond,
		MaxElapsedTime:      10 * time.Second,
		Clock:               SystemClock,
	}
	b.Reset()
	return b
}

type systemClock struct{}

func (t systemClock) Now() time.Time {
	return time.Now()
}

// SystemClock implements Clock interface that uses time.Now().
var SystemClock = systemClock{}
