package gremlin

import (
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

// GoGremlin is an interface based on the go-gremlin *Client
type GoGremlin interface {
	ExecQuery(query string) ([]byte, error)
	Close() error
	Reconnect(urlStr string) error
	MaintainConnection(urlStr string) error
}

// gogremlinClient is a private wrapper for a go-gremlin *Client
// it augments the *Client with Close and Reconnect methods
type gogremlinClient struct {
	Client *Client
}

func newGoGremlinClient(urlStr string, verboseLogging bool, options ...OptAuth) (GoGremlin, error) {
	client, err := NewVerboseClient(urlStr, verboseLogging, options...)
	if err != nil {
		return nil, err
	}
	return &gogremlinClient{Client: client}, nil
}

func (g *gogremlinClient) ExecQuery(query string) ([]byte, error) {
	return g.Client.ExecQuery(query)
}

func (g *gogremlinClient) Close() error {
	return g.Client.Ws.Close()
}

func (g *gogremlinClient) Reconnect(urlStr string) error {
	dialer := websocket.Dialer{}
	ws, _, err := dialer.Dial(urlStr, http.Header{})
	g.Client.Ws = ws
	return err
}

// Send a dummy query to neptune
// If there is a network error, attempt to reconnect
func (g *gogremlinClient) MaintainConnection(urlStr string) error {
	simpleQuery := `g.V().limit(0)`

	_, err := g.ExecQuery(simpleQuery)
	if err == nil {
		return nil
	}

	_, isNetErr := err.(*net.OpError) // check if err is a network error
	if err != nil && !isNetErr {      // if it's not network error, so something else went wrong, no point in retrying
		return err
	}
	// if it is a network error, attempt to reconnect
	err = g.Reconnect(urlStr)
	if err != nil {
		return err
	}
	return nil
}
