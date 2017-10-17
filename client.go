package gremlin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	maxMessages    = 1024
	maxMessageSize = 0
	writeWait      = 5 * time.Second
	pongTimeout    = 5 * time.Second
)

var (
	ErrConnectionClosed = errors.New("Connection closed")
	ErrUnknownError     = errors.New("An unknown error occured")
)

// ClientRequest describes the message to be sent
// to the client for issuing a request
type clientRequest struct {
	request  *Request
	dispatch chan clientResponse
}

// ClientResponse describe the message returned by the
// client after a request is made
type clientResponse struct {
	response Response
	err      error
}

// Client describe a connection to the gremlin server
type Client struct {
	endpoint            string
	dialer              websocket.Dialer
	conn                *websocket.Conn
	dispatch            map[string]chan clientResponse
	send                chan clientRequest
	read                chan []byte
	quit                chan error
	connected           atomic.Value
	running             atomic.Value
	connectedEvent      chan bool
	disconnectedEvent   chan bool
	pingTicker          *time.Ticker
	connectedHandler    func()
	disconnectedHandler func(error)
}

// NewClient creates a new gremlin client
func NewClient(serverURL string) *Client {
	d := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 3 * time.Second,
		ReadBufferSize:   8192,
		WriteBufferSize:  32768,
	}
	c := &Client{
		endpoint:            serverURL,
		dialer:              d,
		dispatch:            make(map[string]chan clientResponse),
		send:                make(chan clientRequest, maxMessages),
		read:                make(chan []byte, maxMessages),
		quit:                make(chan error, 2),
		connectedEvent:      make(chan bool, 1),
		disconnectedEvent:   make(chan bool, 1),
		connectedHandler:    func() { return },
		disconnectedHandler: func(err error) { return },
	}
	c.connected.Store(false)
	c.running.Store(true)
	return c
}

// AddConnectedHandler runs handler when client is connected
func (c *Client) AddConnectedHandler(h func()) {
	c.connectedHandler = h
}

// AddDisconnectedHandler runs handler when connection is closed
func (c *Client) AddDisconnectedHandler(h func(error)) {
	c.disconnectedHandler = h
}

func (c *Client) connect() (err error) {
	c.conn, _, err = c.dialer.Dial(c.endpoint, http.Header{})
	if err != nil {
		return
	}

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(100 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	c.connected.Store(true)
	c.connectedEvent <- true
	c.pingTicker = time.NewTicker(pongTimeout * 8 / 10)

	defer c.connected.Store(false)
	go c.connectedHandler()
	return c.run()
}

func (c *Client) run() error {
	go func() {
		for c.running.Load() == true {
			_, m, err := c.conn.ReadMessage()
			if err != nil {
				if c.running.Load() == true {
					c.quit <- err
				}
				break
			}
			c.read <- m
		}
	}()

	for {
		select {
		case err := <-c.quit:
			// Send errors to clients waiting for a response
			// before closing
			for id, dispatch := range c.dispatch {
				dispatch <- clientResponse{err: ErrConnectionClosed}
				close(dispatch)
				delete(c.dispatch, id)
			}
			return err
		case cr := <-c.send:
			c.dispatch[cr.request.RequestId] = cr.dispatch
			err := c.sendRequest(cr.request)
			if err != nil {
				cr.dispatch <- clientResponse{err: err}
			}
		case m := <-c.read:
			c.dispatchMsg(m)
		case <-c.pingTicker.C:
			c.sendPing()
		}
	}
}

func (c *Client) sendPing() {
	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
		fmt.Println(err)
	}
}

func (c *Client) dispatchMsg(msg []byte) {
	var res Response
	err := json.Unmarshal(msg, &res)
	if ch, ok := c.dispatch[res.RequestId]; ok {
		if err != nil {
			ch <- clientResponse{err: err}
		} else {
			ch <- clientResponse{response: res}
		}
		if res.Status.Code != StatusPartialContent {
			close(c.dispatch[res.RequestId])
			delete(c.dispatch, res.RequestId)
		}
	}
}

func (c *Client) sendRequest(r *Request) (err error) {
	// Prepare the Data
	message, err := json.Marshal(r)
	if err != nil {
		return
	}
	// Prepare the request message
	var requestMessage []byte
	mimeType := []byte("application/json")
	mimeTypeLen := byte(len(mimeType))
	requestMessage = append(requestMessage, mimeTypeLen)
	requestMessage = append(requestMessage, mimeType...)
	requestMessage = append(requestMessage, message...)

	c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := c.conn.WriteMessage(websocket.BinaryMessage, requestMessage); err != nil {
		fmt.Println(err)
	}
	return
}

// ConnectAsync connects to the server - and reconnect if necessary
func (c *Client) ConnectAsync() {
	go func() {
		for c.running.Load() == true {
			err := c.connect()
			c.connectedEvent = make(chan bool, 1)
			c.disconnectedHandler(err)
			time.Sleep(1 * time.Second)
		}
		c.disconnectedEvent <- true
	}()
}

// Connect connects to the server and block until connection is successful
func (c *Client) Connect() {
	c.ConnectAsync()
	<-c.connectedEvent
}

// Disconnect the client without waiting for termination
func (c *Client) Disconnect() {
	c.running.Store(false)
	if c.connected.Load() == true {
		c.quit <- nil
		c.conn.Close()
		close(c.send)
		close(c.read)
	}
	<-c.disconnectedEvent
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	return c.connected.Load() == true
}

// Send sends a request to the gremlin server and wait for results
func (c *Client) Send(r *Request) (data []byte, err error) {
	if !c.IsConnected() {
		err = errors.New("Not connected")
		return
	}
	ch := make(chan clientResponse)
	c.send <- clientRequest{request: r, dispatch: ch}
	inBatchMode := false
	var dataItems []json.RawMessage
	for clientRes := range ch {
		if clientRes.err != nil {
			err = clientRes.err
			return
		}
		res := clientRes.response
		var items []json.RawMessage

		switch clientRes.response.Status.Code {
		case StatusNoContent:
			return

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

		default:
			if msg, exists := ErrorMsg[res.Status.Code]; exists {
				err = errors.New(msg)
			} else {
				err = ErrUnknownError
			}
		}
	}
	return
}
