// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package lib

import (
	"encoding/json"
	"fmt"
)

// RequestError contains the code and body of a failed request
type RequestError struct {
	statusCode int
	body       string
}

// NewRequestError returns a new RequestError
func NewRequestError(statusCode int, body string) *RequestError {
	return &RequestError{
		statusCode: statusCode,
		body:       body,
	}
}

func (r *RequestError) Error() string {
	return fmt.Sprintf("request error: %d %s", r.statusCode, r.body)
}

// Status returns the status code of the error
func (r *RequestError) Status() int {
	return r.statusCode
}

// SendMQTTResponse is the result of the SendMQTT request
type SendMQTTResponse struct {
	Payload  json.RawMessage
	Retained bool
	Topic    string
}
