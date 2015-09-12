package gremlin

import (
	"github.com/satori/go.uuid"
)

var MimeType = "application/json"

type Request struct {
	RequestId string `json:"requestId"`
	Op string `json:"op"`
	Processor string `json:"processor"`
	Args *RequestArgs `json:"args"`
}

type RequestArgs struct {
	Gremlin string `json:"gremlin,omitempty"`
	Session string `json:"session,omitempty"`
	Bindings Bindings `json:"bindings,omitempty"`
	Language string `json:"language,omitempty"`
	Rebindings Rebindings `json:"rebindings,omitempty"`
	Sasl []byte `json:"sasl,omitempty"`
	BatchSize int `json:"batchSize,omitempty"`
}

type Bindings map[string]interface{}

type Rebindings map[string]string

func NewRequest(query string) *Request {
	return NewRequestWithBindings(query, Bindings{})
}

func NewRequestWithBindings(query string, bindings Bindings) *Request {
	args := &RequestArgs{
		Gremlin: query,
		Bindings: bindings,
		Language: "gremlin-groovy",
	}
	req := &Request{
		RequestId: uuid.NewV4().String(),
		Op: "eval",
		Processor: "",
		Args: args,
	}
	return req
}

func (req *Request) Exec() *Response {
	return &Response{Request: req}
}
