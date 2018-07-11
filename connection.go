package gremlin

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/websocket"
)

// Clients include the necessary info to connect to the server and the underlying socket
type Client struct {
	Remote *url.URL
	Ws     *websocket.Conn
	Auth   []OptAuth
}

func NewClient(urlStr string, origin string, options ...OptAuth) (*Client, error) {
	r, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	ws, err := websocket.Dial(urlStr, "", origin)
	if err != nil {
		return nil, err
	}
	return &Client{Remote: r, Ws: ws, Auth: options}, nil
}

// Client executes the provided request
func (c *Client) ExecQuery(query string) ([]byte, error) {
	req := Query(query)
	return c.Exec(req)
}

func (c *Client) Exec(req *Request) ([]byte, error) {
	requestMessage, err := GraphSONSerializer(req)
	if err != nil {
		return nil, err
	}

	// Open a TCP connection
	if err := websocket.Message.Send(c.Ws, requestMessage); err != nil {
		print("error", err)
		return nil, err
	}
	return c.ReadResponse()
}

func (c *Client) ReadResponse() (data []byte, err error) {
	// Data buffer
	var dataItems []json.RawMessage
	inBatchMode := false
	// Receive data
	for {
		var res *Response

		if err = websocket.JSON.Receive(c.Ws, &res); err != nil {
			return nil, err
		}

		var items []json.RawMessage
		switch res.Status.Code {
		case StatusNoContent:
			return

		case StatusAuthenticate:
			return c.Authenticate(res.RequestId)
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
func (c *Client) Authenticate(requestId string) ([]byte, error) {
	auth, err := NewAuthInfo(c.Auth...)
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
	return c.Exec(authReq)
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
