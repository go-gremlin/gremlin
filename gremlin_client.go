package gremlin

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"sync"
)

type GremlinClient struct {
	pool       *gremlinPool
	urlStr     string
	argRegexp  *regexp.Regexp
	mutex      *sync.Mutex
	maxRetries int
}

func NewGremlinClient(urlStr string, maxCap int, maxRetries int, verboseLogging bool, options ...OptAuth) (*GremlinClient, error) {
	newClientFn := func() (GoGremlin, error) {
		return NewVerboseGremlinConnection(urlStr, verboseLogging, options...)
	}

	pool, err := newGremlinPool(maxCap, newClientFn)
	if err != nil {
		return nil, err
	}
	argRegexp, err := regexp.Compile(ARG_REGEX)
	if err != nil {
		return nil, err
	}
	c := &GremlinClient{
		urlStr:     urlStr,
		pool:       pool,
		argRegexp:  argRegexp,
		mutex:      &sync.Mutex{},
		maxRetries: maxRetries,
	}

	return c, nil
}

func (c *GremlinClient) ExecQueryF(ctx context.Context, query string, args ...interface{}) (string, error) {
	args = EscapeArgs(args, EscapeGremlin)
	for _, arg := range args {
		// if the argument is not a string (i.e. an int) or matches the regex string, then we're good
		if InterfaceToString(arg) != "" && !c.argRegexp.MatchString(InterfaceToString(arg)) {
			return "", fmt.Errorf("Invalid character in your query argument: %s", InterfaceToString(arg))
		}
	}
	query = fmt.Sprintf(query, args...)
	rawResponse, err := c.execWithRetry(ctx, query)
	if err != nil {
		return "", err
	}

	_ = Vertex{
		Type: "g:Vertex",
		Value: VertexValue{
			ID:    "test-id",
			Label: "label",
			Properties: map[string][]VertexProperty{
				"health": []VertexProperty{
					VertexProperty{
						Type: "g:VertexProperty",
						Value: VertexPropertyValue{
							ID: GenericValue{
								Type:  "Type",
								Value: 1,
							},
							Value: "1",
							Label: "health",
						},
					},
				},
			},
		},
	}
	return string(rawResponse), nil
}

func (c *GremlinClient) execWithRetry(ctx context.Context, query string) (rawResponse []byte, err error) {
	var (
		client GoGremlin
		tryNum = 1
	)

	for {
		select {
		case <-ctx.Done():
			_ = closeClient(client)
			return nil, ctx.Err()
		default:
		}

		if tryNum > c.maxRetries {
			_ = closeClient(client)
			return nil, fmt.Errorf("max tries reached %d, with err: %v", tryNum, err)
		}

		// Alternating tries between getting a new connection from the pool
		// And attempting to reconnect
		if tryNum%2 == 1 {
			client, err = c.pool.Get()
		} else {
			// Hit this logic on the 2nd (or all even) passes through the loop
			// If the method pulls a client from the Pool that it can't connect to,
			// The next loop will attempt to reconnect to that client
			// Before pulling another from the pool
			err = client.Reconnect(c.urlStr)
		}
		if err != nil {
			_ = closeClient(client)
			return nil, err
		}
		tryNum++

		c.mutex.Lock()
		rawResponse, err = client.ExecQuery(query)
		c.mutex.Unlock()
		if err == nil { // success, break out of the retry loop
			break
		}

		_, isNetErr := err.(*net.OpError) // check if err is a network error
		if err != nil && !isNetErr {      // if it's not network error, so something else went wrong, no point in retrying
			_ = closeClient(client)
			return nil, err
		}
		// if it is a network error, repeat the loop

	}

	// should we worry about an error here?
	_ = c.pool.Put(client)

	return rawResponse, nil
}

func (c *GremlinClient) PingDatabase(ctx context.Context) error {
	err := c.pool.MaintainPool(c.urlStr)
	return err
}

func closeClient(client GoGremlin) error {
	if client == nil {
		return nil
	}
	return client.Close()
}
