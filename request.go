package gremlin

import (
	"encoding/json"
	_ "fmt"
	"github.com/google/uuid"
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
	Sasl              string            `json:"sasl,omitempty"`
	BatchSize         int               `json:"batchSize,omitempty"`
	ManageTransaction bool              `json:"manageTransaction,omitempty"`
	Aliases           map[string]string `json:"aliases,omitempty"`
}

// Formats the requests in the appropriate way
type FormattedReq struct {
	Op        string       `json:"op"`
	RequestId interface{}  `json:"requestId"`
	Args      *RequestArgs `json:"args"`
	Processor string       `json:"processor"`
}

func GraphSONSerializer(req *Request) ([]byte, error) {
	form := NewFormattedReq(req)
	msg, err := json.Marshal(form)
	if err != nil {
		return nil, err
	}
	mimeType := []byte("application/vnd.gremlin-v2.0+json")
	var mimeLen = []byte{0x21}
	res := append(mimeLen, mimeType...)
	res = append(res, msg...)
	return res, nil

}

func NewFormattedReq(req *Request) FormattedReq {
	rId := map[string]string{"@type": "g:UUID", "@value": req.RequestId}
	sr := FormattedReq{RequestId: rId, Processor: req.Processor, Op: req.Op, Args: req.Args}

	return sr
}

type Bind map[string]interface{}

func Query(query string) *Request {
	args := &RequestArgs{
		Gremlin:  query,
		Language: "gremlin-groovy",
	}
	u := uuid.New()
	uuidString := u.String()
	req := &Request{
		RequestId: uuidString,
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
