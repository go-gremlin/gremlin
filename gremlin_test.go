package gremlin

import (
	"context"
	"testing"
)

func TestImplementsGremlinI(t *testing.T) {
	t.Parallel()
	var g Gremlin_i
	g = &GremlinClient{}
	g = GremlinLogger{}
	_ = g
}

// TODO: Test that bad conns don't get added back when ExecQuery returns an error

func TestPoolMaintainsConnections(t *testing.T) {
	t.Parallel()
	max := 3
	pool, err := newGremlinPool(max, makeMockGoGremlinClient(""))
	assert(t, err == nil, "Expected nil error")
	defer pool.Close()

	client := &GremlinClient{
		pool: pool,
	}

	for i := 1; i <= 10; i++ {
		client.ExecQueryF(context.Background(), `g.V()`)
		gotLen := pool.Len()
		assert(t, gotLen == max, "expected", max, "got", gotLen)
	}
}

func TestPoolLenAfterConstruction(t *testing.T) {
	t.Parallel()
	max := 10
	p, err := newGremlinPool(max, makeMockGoGremlinClient(""))
	assert(t, err == nil, "Expected nil error")

	gotLen := p.Len()
	assert(t, gotLen == max, "Expected", max, "got", gotLen)
}

func TestPoolLenAfterGet(t *testing.T) {
	t.Parallel()
	max := 10
	expectedConn := mockGoGremlinClient{"esurient"}
	p, err := newGremlinPool(max, makeMockGoGremlinClient("esurient"))
	assert(t, err == nil, "Expected nil error")
	defer p.Close()

	// get them all
	for i := 0; i < max; i++ {
		conn, err := p.Get()
		assert(t, err == nil, "Expected nil error")
		assert(t, conn == expectedConn, "Expecting", expectedConn, "got", conn)

		expectedLen := max - (i + 1)
		gotLen := p.Len()
		assert(t, gotLen == expectedLen, "Expecting", expectedLen, "got", gotLen)
	}

	assert(t, p.Len() == 0, "Expecting", 0, "got", p.Len())

	conn, err := p.Get()
	assert(t, err == nil, "Expected nil error")
	assert(t, conn == expectedConn, "Expecting", expectedConn, "got", conn)
}

func TestPool_Close(t *testing.T) {
	t.Parallel()
	max := 4
	p, err := newGremlinPool(max, makeMockGoGremlinClient("godoggo"))
	assert(t, err == nil, "Expected nil error")

	err = p.Close()
	assert(t, err == nil, "Expected nil error")
	assert(t, p.Len() == 0, "Expected connections to be 0")
}

func TestPool_Reconnect(t *testing.T) {
	t.Parallel()
	max := 4

	p, err := newGremlinPool(max, makeMockGoGremlinClient("godoggo"))
	assert(t, err == nil, "Expected nil error")

	conn, err := p.Get()
	assert(t, err == nil, "Expected nil error")

	err = conn.Reconnect("abc")
	assert(t, err == nil, "Expected nil error")
	assert(t, p.Len() == max-1, "Expected connections to be 3")

}

func TestPoolGetWhenUnitialized(t *testing.T) {
	t.Parallel()
	p := &gremlinPool{}

	_, err := p.Get()

	assert(t, err == ErrClosed, "Expecting", ErrClosed, "got", err)
}

func TestPoolGetWhenClosed(t *testing.T) {
	t.Parallel()
	p := &gremlinPool{
		conns: make(chan GoGremlin),
	}

	close(p.conns)

	_, err := p.Get()

	assert(t, err == ErrClosed, "Expecting", ErrClosed, "got", err)
}

func TestPoolGetReusedConn(t *testing.T) {
	t.Parallel()
	p := &gremlinPool{
		conns: make(chan GoGremlin, 1),
	}

	client := mockGoGremlinClient{"gogogo"}

	p.conns <- client

	conn, err := p.Get()
	assert(t, err == nil, "Expected nil error")

	assert(t, conn == client, "Expecting", client, "got", conn)
}

func TestPoolGetNewConn(t *testing.T) {
	t.Parallel()
	p := &gremlinPool{
		conns:     make(chan GoGremlin),
		newClient: makeMockGoGremlinClient("yayaya"),
	}

	conn, err := p.Get()
	assert(t, err == nil, "Expected nil error")

	expected := mockGoGremlinClient{"yayaya"}
	assert(t, conn == expected, "Expecting", expected, "got", conn)
}

//
// HELPERS / MOCKS
//
func assert(t *testing.T, isTrue bool, args ...interface{}) {
	if !isTrue {
		t.Fatal(args...)
	}
}

func makeMockGoGremlinClient(secret string) newClientFn {
	return func() (GoGremlin, error) {
		return mockGoGremlinClient{secret}, nil
	}
}

type mockGoGremlinClient struct {
	secret string
}

func (g mockGoGremlinClient) ExecQuery(query string) ([]byte, error) {
	return []byte("dummy response"), nil
}

func (g mockGoGremlinClient) Close() error {
	return nil
}

func (g mockGoGremlinClient) Reconnect(urlStr string) error {
	return nil
}

func (g mockGoGremlinClient) MaintainConnection(urlStr string) error {
	return nil
}
