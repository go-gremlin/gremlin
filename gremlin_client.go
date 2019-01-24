package gremlin

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"time"

	"github.com/cbinsights/gremlin/lock"
)

type GremlinClient struct {
	pool       *gremlinPool
	urlStr     string
	argRegexp  *regexp.Regexp
	LockClient lock.LockClient_i
	maxRetries int
	quit       chan struct{}
	done       chan struct{}
}

func NewGremlinClient(urlStr string, maxCap int, maxRetries int, verboseLogging bool, lockClient lock.LockClient_i, options ...OptAuth) (*GremlinClient, error) {
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
		LockClient: lockClient,
		maxRetries: maxRetries,
	}

	return c, nil
}

func (c *GremlinClient) ExecQueryF(ctx context.Context, gremlinQuery GremlinQuery) (string, error) {
	args := EscapeArgs(gremlinQuery.Args, EscapeGremlin)
	for _, arg := range args {
		// if the argument is not a string (i.e. an int) or matches the regex string, then we're good
		if InterfaceToString(arg) != "" && !c.argRegexp.MatchString(InterfaceToString(arg)) {
			return "", fmt.Errorf("Invalid character in your query argument: %s", InterfaceToString(arg))
		}
	}
	query := fmt.Sprintf(gremlinQuery.Query, args...)
	rawResponse, err := c.execWithRetry(ctx, query, gremlinQuery.LockKey)
	if err != nil {
		return "", err
	}

	return string(rawResponse), nil
}

func getLock(c *GremlinClient, key string) (lock.Lock_i, bool, error) {
	hasKey := false
	var lock lock.Lock_i
	if key == "" {
		return lock, hasKey, nil
	}
	lock, err := c.LockClient.LockKey(key)
	if err != nil {
		return nil, false, err
	}
	return lock, true, nil
}

func (c *GremlinClient) execWithRetry(ctx context.Context, query string, queryId string) (rawResponse []byte, err error) {
	var (
		client GoGremlin
		tryNum = 1
	)
	hasKey := false
	lock, hasKey, err := getLock(c, queryId)
	if err != nil {
		return nil, err
	}
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

		if hasKey {
			err = lock.Lock()
			if err != nil {
				return nil, err
			}
		}
		rawResponse, err = client.ExecQuery(query)
		if hasKey {
			err = lock.Unlock()
			if err != nil {
				return nil, err
			}
			lock.Destroy()
		}
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

func (c *GremlinClient) Close(ctx context.Context) error {
	c.quit <- struct{}{}
	<-c.done
	err := c.pool.Close()
	return err
}

func (c *GremlinClient) StartMonitor(ctx context.Context, interval time.Duration) error {
	ticker := time.NewTicker(interval * time.Second)
	quit := make(chan struct{})
	done := make(chan struct{})
	go runMonitorTilQuit(ctx, quit, done, ticker, c)
	c.quit = quit
	c.done = done
	return nil
}

func runMonitorTilQuit(ctx context.Context, quit, done chan struct{}, ticker *time.Ticker, c *GremlinClient) {
	defer func() {
		done <- struct{}{}
	}()
	for {
		select {
		case <-quit:
			return
		case <-ticker.C:
			_ = c.pool.MaintainPool(c.urlStr)
		}
	}
}

func closeClient(client GoGremlin) error {
	if client == nil {
		return nil
	}
	return client.Close()
}
