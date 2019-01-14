package gremlin

import (
	"errors"
	"fmt"
	"sync"
)

var ErrClosed = errors.New("pool is closed")

// gremlinPool maintains a pool of GoGremlin clients
type gremlinPool struct {
	mu        sync.RWMutex
	conns     chan GoGremlin
	maxConn   int
	newClient newClientFn
}

type newClientFn func() (GoGremlin, error)

func newGremlinPool(maxCap int, newClient newClientFn) (*gremlinPool, error) {
	if maxCap <= 0 {
		return nil, errors.New("Invalid capacity settings")
	}

	pool := &gremlinPool{
		conns:     make(chan GoGremlin, maxCap),
		maxConn:   maxCap,
		newClient: newClient,
	}

	for i := 0; i < maxCap; i++ {
		conn, err := pool.newClient()
		if err != nil {
			_ = closeClient(conn) // TODO: this is good, but check for other places we may be leaking conns
			return nil, fmt.Errorf("newClient() is not able to fill the pool: %s", err)
		}
		pool.conns <- conn
	}

	return pool, nil
}

func (c *gremlinPool) getConns() chan GoGremlin {
	c.mu.RLock()
	conns := c.conns
	c.mu.RUnlock()
	return conns
}

// Get implements the Pool interfaces Get() method. If there is no new
// connection available in the pool, a new connection will be created via the
// newClient() method.
func (c *gremlinPool) Get() (GoGremlin, error) {
	conns := c.getConns()
	if conns == nil {
		return nil, ErrClosed
	}

	// wrap our connections with out custom net.Conn implementation (wrapConn
	// method) that puts the connection back to the pool if it's closed.
	select {
	case conn := <-conns:
		if conn == nil {
			return nil, ErrClosed
		}

		return conn, nil
	default:
		conn, err := c.newClient()
		if err != nil {
			return nil, err
		}

		// commenting because if there is a race condition between the Len and Put,
		// i.e. if there is a another request from another goroutine,
		// the chan might be full and Put would close this newly created connection
		// and then execWithRetry would be using a closed conn
		// if c.Len() < c.maxConn {
		// 	c.Put(conn)
		// }

		return conn, nil
	}
}

func (c *gremlinPool) Len() int {
	conns := c.getConns()
	return len(conns)
}

// put puts the connection back to the pool. If the pool is full or closed,
// conn is simply closed. A nil conn will be rejected.
func (c *gremlinPool) Put(conn GoGremlin) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conns == nil {
		// pool is closed, close passed connection
		return conn.Close()
	}

	// put the resource back into the pool. If the pool is full, this will
	// block and the default case will be executed.
	select {
	case c.conns <- conn:
		return nil
	default:
		// pool is full, close passed connection
		return conn.Close()
	}
}

func (c *gremlinPool) Close() error {
	c.mu.Lock()
	conns := c.conns
	c.conns = nil
	c.mu.Unlock()

	if conns == nil {
		return nil
	}

	close(conns)
	for client := range conns {
		err := client.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Function that loops through connections in the pool
// For existing connections, maintains
// If the pool has dropped below full (i.e. a connection has been broken by an error), re-fill the pool
func (c *gremlinPool) MaintainPool(urlStr string) error {
	conns := c.getConns()
	if conns == nil {
		return ErrClosed
	}
	nConns := len(conns)

	// Ping existing connections
	for i := 0; i < nConns; i++ { // Limit your connections here to the number of conns in the pool, rather than a range over the channel.
		// Otherwise, the Put() call will cause an infinite loop
		client := <-conns
		err := client.MaintainConnection(urlStr)
		if err != nil {
			_ = closeClient(client)
			return err
		}
		_ = c.Put(client)
	}

	// Refill pool with new connections
	for j := nConns; j < c.maxConn; j++ {
		conn, err := c.newClient()
		if err != nil {
			_ = closeClient(conn) // TODO: this is good, but check for other places we may be leaking conns
			return fmt.Errorf("newClient() is not able to fill the pool: %s", err)
		}
		c.conns <- conn
	}

	return nil
}
