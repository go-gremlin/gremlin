package gremlin

import (
	"github.com/satori/go.uuid"
)

type Request struct {
	RequestId string       `json:"requestId"`
	Op        string       `json:"op"`
	Processor string       `json:"processor"`
	Args      *RequestArgs `json:"args"`
}

type RequestArgs struct {
	Gremlin           string            `json:"gremlin,omitempty"`
	Session           string            `json:"session,omitempty"`
	Bindings          Bind              `json:"bindings,omitempty"`
	Language          string            `json:"language,omitempty"`
	Rebindings        Bind              `json:"rebindings,omitempty"`
	Sasl              []byte            `json:"sasl,omitempty"`
	BatchSize         int               `json:"batchSize,omitempty"`
	ManageTransaction bool              `json:"manageTransaction,omitempty"`
	Aliases           map[string]string `json:"aliases,omitempty"`
}

type Bind map[string]interface{}

func Query(query string) *Request {
	args := &RequestArgs{
		Gremlin:  query,
		Language: "gremlin-groovy",
	}
	req := &Request{
		RequestId: uuid.NewV4().String(),
		Op:        "eval",
		Processor: "",
		Args:      args,
	}
	return req
}

func (req *Request) Bindings(bindings Bind) *Request {
	req.Args.Bindings = bindings
	return req
}

func (req *Request) ManageTransaction(flag bool) *Request {
	req.Args.ManageTransaction = flag
	return req
}

func (req *Request) Aliases(aliases map[string]string) *Request {
	req.Args.Aliases = aliases
	return req
}

func (req *Request) Session(session string) *Request {
	req.Args.Session = session
	return req
}

func (req *Request) SetProcessor(processor string) *Request {
	req.Processor = processor
	return req
}
