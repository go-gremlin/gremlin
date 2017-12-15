package gremlin

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Client struct {
	Remote *url.URL
	Ws     *websocket.Conn
}

func NewClient(urlStr string) (*Client, error) {
	r, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(urlStr, http.Header{})
	if err != nil {
		return nil, err
	}
	return &Client{Remote: r, Ws: ws}, nil

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
	fmt.Println(string(requestMessage))
	// Open a TCP connection
	if err = c.Ws.WriteMessage(websocket.BinaryMessage, requestMessage); err != nil {
		print("error", err)
		return nil, err
	}
	return c.ReadResponse()
}

func (c *Client) ReadResponse() (data []byte, err error) {
	// Data buffer
	var message []byte
	var dataItems []json.RawMessage
	inBatchMode := false
	// Receive data
	for {
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
			fmt.Println(res)
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
func (c *Client) Authenticate(requestId string) ([]byte, error) {
	user := os.Getenv("GREMLIN_USER")
	pass := os.Getenv("GREMLIN_PASS")
	var sasl []byte
	sasl = append(sasl, 0)
	sasl = append(sasl, []byte(user)...)
	sasl = append(sasl, 0)
	sasl = append(sasl, []byte(pass)...)
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
