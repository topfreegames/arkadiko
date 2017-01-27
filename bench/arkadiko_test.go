// arkadiko
// https://github.com/topfreegames/arkadiko
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package bench

import (
	"fmt"
	"testing"

	"github.com/satori/go.uuid"
	"github.com/topfreegames/arkadiko/remote"
)

var keeper interface{}

func BenchmarkHTTPSendSmallMessage(b *testing.B) {
	msg := getSmallMessage()
	topic := uuid.NewV4().String()

	route := getRoute(fmt.Sprintf("/sendmqtt/%s", topic))
	status, body, err := fastPostTo(route, msg)
	validateResp(status, string(body), err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		status, body, err = fastPostTo(route, msg)
		validateResp(status, string(body), err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkHTTPSendLargeMessage(b *testing.B) {
	msg := getLargeMessage()
	topic := uuid.NewV4().String()

	route := getRoute(fmt.Sprintf("/sendmqtt/%s", topic))
	status, body, err := fastPostTo(route, msg)
	validateResp(status, string(body), err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		status, body, err = fastPostTo(route, msg)
		validateResp(status, string(body), err)
		b.SetBytes(int64(len([]byte(body))))

		keeper = body
	}
}

func BenchmarkGRPCSendSmallMessage(b *testing.B) {
	msg := getSmallMessage()
	topic := uuid.NewV4().String()
	message := &remote.Message{
		Topic:    topic,
		Retained: false,
		Payload:  string(msg),
	}

	_, err := rpcSendMessage(message)
	if err != nil {
		panic(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		response, err := rpcSendMessage(message)
		if err != nil {
			panic(err)
		}
		b.SetBytes(int64(len(msg)))

		keeper = response
	}
}

func BenchmarkGRPCSendLargeMessage(b *testing.B) {
	msg := getLargeMessage()
	topic := uuid.NewV4().String()
	message := &remote.Message{
		Topic:    topic,
		Retained: false,
		Payload:  string(msg),
	}

	_, err := rpcSendMessage(message)
	if err != nil {
		panic(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		response, err := rpcSendMessage(message)
		if err != nil {
			panic(err)
		}
		b.SetBytes(int64(len(msg)))

		keeper = response
	}
}
