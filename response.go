package gremlin

import (
	"encoding/json"
	"golang.org/x/net/websocket"
	"errors"
)

type Response struct {
	RequestId string `json:"requestId"`
	Status *ResponseStatus `json:"status"`
	Result *ResponseResult `json:"result"`
}

type ResponseStatus struct {
	Code int `json:"code"`
	Attributes map[string]interface{} `json:"attributes"`
	Message string `json:"message"`
}

type ResponseResult struct {
	Data json.RawMessage `json:"data"`
	Meta map[string]interface{} `json:"meta"`
}

func (req *Request) Exec() (data []byte, err error) {
	// Check if we are connected
	if conn == nil {
		err = errors.New("You are currently not connected to any database. Please open a connection first.")
		return
	}
	// Submit request
	if err = websocket.JSON.Send(conn, req); err != nil {
		return
	}
	// Data buffer
	var dataItems []json.RawMessage
	inBatchMode := false
	// Receive data
	for {
		var res *Response
		if err = websocket.JSON.Receive(conn, &res); err != nil {
			return
		}
		var items []json.RawMessage
		switch res.Status.Code {
		case StatusNoContent:
			return

		case StatusPartialContent:
			inBatchMode = true
			if err = json.Unmarshal(res.Result.Data, &items); err != nil {
				return
			}
			dataItems = append(dataItems, items...)

		case StatusSuccess:
			if inBatchMode {
				if err = json.Unmarshal(res.Result.Data, &items); err != nil {
					return
				}
				dataItems = append(dataItems, items...)
				data, err = json.Marshal(dataItems)
				} else {
					data = res.Result.Data
				}
				return

			default:
				if msg, exists := ErrorMsg[res.Status.Code]; exists {
					err = errors.New(msg)
				} else {
					err = errors.New("An unknown error occured")
				}
				return
			}
		}
		return
	}
