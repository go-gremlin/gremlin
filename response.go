package gremlin

import (
	"encoding/json"
	"fmt"
)

type Response struct {
	RequestId string          `json:"requestId"`
	Status    *ResponseStatus `json:"status"`
	Result    *ResponseResult `json:"result"`
	Err       error
}

type ResponseStatus struct {
	Code       int                    `json:"code"`
	Attributes map[string]interface{} `json:"attributes"`
	Message    string                 `json:"message"`
}

type ResponseResult struct {
	Data json.RawMessage        `json:"data"`
	Meta map[string]interface{} `json:"meta"`
}

// Implementation of the stringer interface. Useful for exploration
func (r Response) String() string {
	return fmt.Sprintf("Response \nRequestId: %v, \nStatus: {%#v}, \nResult: {%#v}\n", r.RequestId, r.Status, r.Result)
}
