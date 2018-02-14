package gremlin

import (
	"encoding/json"
	"log"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

type testServer struct {
	Done      chan struct{}
	Responses [][]Response
	srv       *http.Server
}

var upgrader = websocket.Upgrader{} // use default options

func newTestServer() *testServer {
	return &testServer{
		Done: make(chan struct{}),
		srv:  &http.Server{Addr: ":8080"},
	}
}

func (t *testServer) Start() {
	http.HandleFunc("/gremlin", t.gremlin)
	err := t.srv.ListenAndServe()
	if err != nil {
		log.Println("ListenAndServe: ", err)
	}
}

func (t *testServer) Stop() {
	if err := t.srv.Shutdown(nil); err != nil {
		log.Fatal("Shutdown: ", err)
	}
}

func (t *testServer) gremlin(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		resps := t.Responses[0]
		for _, r := range resps {
			//fmt.Printf("write %v %v %v\n", r.RequestId, r.Status.Code, string(r.Result.Data))
			msg, err := json.Marshal(r)
			if err != nil {
				log.Println("write:", err)
				return
			}
			c.WriteMessage(websocket.BinaryMessage, msg)
		}
		t.Responses = t.Responses[1:]
	}
}

var (
	s *testServer
	c *Client
)

func init() {
	s = newTestServer()
	go s.Start()
	c = NewClient("ws://localhost:8080/gremlin")
	c.Connect()
}

func TestSingleResponse(t *testing.T) {
	reqID := uuid.NewV4().String()

	s.Responses = [][]Response{
		[]Response{
			Response{
				RequestId: reqID,
				Status: &ResponseStatus{
					Code: 200,
				},
				Result: &ResponseResult{
					Data: json.RawMessage("[1, 2, 3]"),
				},
			},
		},
	}

	q := Query("")
	q.RequestId = reqID
	data, _ := c.Send(q)

	assert.Equal(t, "[1,2,3]", string(data), "Response should contain data from a single response")
}

func TestEmptyResponse(t *testing.T) {
	reqID := uuid.NewV4().String()

	s.Responses = [][]Response{
		[]Response{
			Response{
				RequestId: reqID,
				Status: &ResponseStatus{
					Code: 204,
				},
				Result: &ResponseResult{
					Data: json.RawMessage("[1, 2, 3]"),
				},
			},
		},
	}

	q := Query("")
	q.RequestId = reqID
	data, _ := c.Send(q)

	assert.Equal(t, "", string(data), "Response should not contain any data")
}

func TestPartialResponse(t *testing.T) {
	reqID := uuid.NewV4().String()

	s.Responses = [][]Response{
		[]Response{
			Response{
				RequestId: reqID,
				Status: &ResponseStatus{
					Code: 206,
				},
				Result: &ResponseResult{
					Data: json.RawMessage("[1, 2]"),
				},
			},
			Response{
				RequestId: reqID,
				Status: &ResponseStatus{
					Code: 206,
				},
				Result: &ResponseResult{
					Data: json.RawMessage("[3, 4]"),
				},
			},
			Response{
				RequestId: reqID,
				Status: &ResponseStatus{
					Code: 200,
				},
				Result: &ResponseResult{
					Data: json.RawMessage("[5, 6]"),
				},
			},
		},
	}

	q := Query("")
	q.RequestId = reqID
	data, _ := c.Send(q)

	assert.Equal(t, "[1,2,3,4,5,6]", string(data), "Response should contain aggregated data from responses")
}

func TestConcurrentRequest(t *testing.T) {
	reqID1 := uuid.NewV4().String()
	reqID2 := uuid.NewV4().String()

	s.Responses = [][]Response{
		[]Response{
			Response{
				RequestId: reqID1,
				Status: &ResponseStatus{
					Code: 206,
				},
				Result: &ResponseResult{
					Data: json.RawMessage("[1, 2]"),
				},
			},
			Response{
				RequestId: reqID1,
				Status: &ResponseStatus{
					Code: 200,
				},
				Result: &ResponseResult{
					Data: json.RawMessage("[5, 6]"),
				},
			},
		},
		[]Response{
			Response{
				RequestId: reqID2,
				Status: &ResponseStatus{
					Code: 200,
				},
				Result: &ResponseResult{
					Data: json.RawMessage("[3, 4]"),
				},
			},
		},
	}

	q1 := Query("")
	q1.RequestId = reqID1
	q2 := Query("")
	q2.RequestId = reqID2

	rs := make(chan string, 2)

	go func() {
		data1, _ := c.Send(q1)
		rs <- string(data1)
	}()
	go func() {
		data2, _ := c.Send(q2)
		rs <- string(data2)
	}()

	r1 := <-rs
	if r1 == "[1,2,5,6]" {
		r2 := <-rs
		assert.Equal(t, "[3,4]", string(r2), "Response should contain reqID2 results")
	} else {
		r2 := <-rs
		assert.Equal(t, "[1,2,5,6]", string(r2), "Response should contain reqID1 results")
	}

}

func TestError(t *testing.T) {
	reqID := uuid.NewV4().String()

	s.Responses = [][]Response{
		[]Response{
			Response{
				RequestId: reqID,
				Status: &ResponseStatus{
					Code: StatusServerError,
				},
			},
		},
	}

	q := Query("")
	q.RequestId = reqID
	_, err := c.Send(q)

	assert.Equal(t, ErrorMsg[StatusServerError], err.Error(), "Should get server error")

}
