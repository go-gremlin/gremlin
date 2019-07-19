package gremlin

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Clients include the necessary info to connect to the server and the underlying socket
type Client struct {
	Remote   *url.URL
	Ws       *websocket.Conn
	Auth     []OptAuth
	Host     string
	sendCh   chan *Request
	requests map[string]*Request
	quit     chan bool
	closeCh  chan bool
	lock     sync.Mutex
}

func NewClient(urlStr string, options ...OptAuth) (*Client, error) {
	r, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	client := Client{Remote: r, Auth: options, Host: urlStr}
	client.quit = make(chan bool, 1)
	client.requests = make(map[string]*Request)
	client.sendCh = make(chan *Request, 10)

	go client.loop()
	return &client, nil
}

// Client executes the provided request
func (c *Client) ExecQuery(query string) ([]byte, error) {
	req := Query(query)
	//return c.Exec(req)
	responseCh, err := c.queueRequest(req)
	if err != nil {
		return nil, err
	}

	response := <-responseCh
	if response.Err != nil {
		return nil, response.Err
	}

	return response.Result.Data, nil

}
func (c *Client) queueRequest(req *Request) (<-chan *Response, error) {
	requestMessage, err := GraphSONSerializer(req)
	if err != nil {
		return nil, err
	}
	req.Msg = requestMessage
	req.responseCh = make(chan *Response, 1)
	req.inBatchMode = false
	req.dataItems = make([]json.RawMessage, 0)
	select {
	case <-c.closeCh:
		return nil, ErrConnectionClosed
	default:
	}
	c.sendCh <- req
	return req.responseCh, nil
}

func (c *Client) loop() {
	for {
		if err := c.createConnection(); err != nil {
			return
		}
		c.closeCh = make(chan bool)
		var wg sync.WaitGroup

		wg.Add(1)
		go func() {
			err := c.sendLoop()
			if err != nil {
				log.Println(err)
			}
			c.closeConnection()
			wg.Done()
		}()

		wg.Add(1)
		go func() {
			err := c.recvLoop()
			log.Println(err)
			if err == nil {
				panic("recvloop not get nil err")
			}
			close(c.closeCh)
			wg.Done()
		}()

		wg.Wait()

		select {
		case <-c.quit:
			c.flushRequest()
			return
		default:
		}

		// server is not close,should flush request
		c.flushRequest()
	}
}
func (c *Client) flushRequest() {
	c.lock.Lock()
	defer c.lock.Unlock()

	for requestId, request := range c.requests {
		response := Response{Err: ErrClosing}
		request.responseCh <- &response
		delete(c.requests, requestId)
	}

	c.requests = make(map[string]*Request)
}
func (c *Client) sendLoop() error {
	for {
		select {
		case request := <-c.sendCh:
			err := c.Ws.WriteMessage(websocket.BinaryMessage, request.Msg)
			// if fail, direct return
			// send responseCh error
			if err != nil {
				response := Response{Err: ErrConnectionClosed}
				responseCh := request.responseCh
				responseCh <- &response
				return err
			}
			// if success, put request in requests map
			if request.Op != "authentication" {
				c.lock.Lock()
				c.requests[request.RequestId] = request
				c.lock.Unlock()
			}

		case <-c.closeCh:
			return nil
		case <-c.quit:
			return nil
		}
	}
}

/**
func (c *Client) Exec(req *Request) ([]byte, error) {
	requestMessage, err := GraphSONSerializer(req)
	if err != nil {
		return nil, err
	}

	// Open a TCP connection
	if err = c.Ws.WriteMessage(websocket.BinaryMessage, requestMessage); err != nil {
		print("error", err)
		return nil, err
	}
	return c.ReadResponse()
}
**/

func (c *Client) createConnection() error {
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(c.Host, http.Header{})
	if err != nil {
		return err
	}
	c.Ws = ws
	return nil
}
func (c *Client) closeConnection() {
	if c.Ws != nil {
		c.Ws.Close()
	}
}
func (c *Client) Close() {
	close(c.quit)
}
func (c *Client) recvLoop() (err error) {
	// Receive data
	for {
		// Data buffer
		var message []byte

		if _, message, err = c.Ws.ReadMessage(); err != nil {
			return
		}
		var res *Response
		if err = json.Unmarshal(message, &res); err != nil {
			return
		}
		var items []json.RawMessage
		switch res.Status.Code {
		case StatusNoContent:
			c.lock.Lock()
			if request, ok := c.requests[res.RequestId]; ok {
				delete(c.requests, res.RequestId)
				request.responseCh <- res
			}
			c.lock.Unlock()

		case StatusAuthenticate:
			if err = c.Authenticate(res.RequestId); err != nil {
				return
			}
		case StatusPartialContent:
			if err = json.Unmarshal(res.Result.Data, &items); err != nil {
				c.lock.Lock()
				if request, ok := c.requests[res.RequestId]; ok {
					delete(c.requests, res.RequestId)
					res.Err = err
					request.responseCh <- res
				}
				c.lock.Unlock()
				return
			}

			c.lock.Lock()
			if request, ok := c.requests[res.RequestId]; ok {
				request.inBatchMode = true
				request.dataItems = append(request.dataItems, items...)
			}
			c.lock.Unlock()

		case StatusSuccess:
			c.lock.Lock()
			request, ok := c.requests[res.RequestId]
			// not find request
			if !ok {
				c.lock.Unlock()
				continue
			}
			delete(c.requests, res.RequestId)
			c.lock.Unlock()

			if request.inBatchMode {
				if err = json.Unmarshal(res.Result.Data, &items); err != nil {
					res.Err = err
					request.responseCh <- res
					return
				}
				request.dataItems = append(request.dataItems, items...)

				res.Result.Data, _ = json.Marshal(request.dataItems)
				request.responseCh <- res
			} else {
				request.responseCh <- res
			}

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
func (c *Client) Authenticate(requestId string) error {
	auth, err := NewAuthInfo(c.Auth...)
	if err != nil {
		return err
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
		// responseCh: make(chan *Response, nil),
	}
	return c.queueAuthRequest(authReq)
}
func (c *Client) queueAuthRequest(req *Request) error {
	requestMessage, err := GraphSONSerializer(req)
	if err != nil {
		return err
	}
	req.Msg = requestMessage
	c.sendCh <- req

	return nil
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
