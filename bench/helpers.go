// podium
// https://github.com/topfreegames/podium
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package bench

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/topfreegames/arkadiko/remote"
	"github.com/topfreegames/arkadiko/testing"
	"github.com/valyala/fasthttp"
)

func getSmallMessage() []byte {
	msg := map[string]interface{}{}
	for i := 1; i <= 10; i++ {
		msg[fmt.Sprintf("item-%d", i)] = fmt.Sprintf("some small text %d", i)
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		panic(fmt.Sprintf("Could not serialize big message: %s.", err.Error()))
	}
	return payload
}

func getLargeMessage() []byte {
	msg := map[string]interface{}{}
	for i := 1; i <= 100; i++ {
		msg[fmt.Sprintf("item-%d", i)] = fmt.Sprintf("a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself like a very large text repeating itself %d", i)
	}

	payload, err := json.Marshal(msg)
	if err != nil {
		panic(fmt.Sprintf("Could not serialize big message: %s.", err.Error()))
	}
	return payload
}

func getRoute(url string) string {
	return fmt.Sprintf("http://localhost:52344%s", url)
}

func fastGet(url string) (int, []byte, error) {
	return fastSendTo("GET", url, nil)
}

func fastDelete(url string) (int, []byte, error) {
	return fastSendTo("DELETE", url, nil)
}

func fastPostTo(url string, payload []byte) (int, []byte, error) {
	return fastSendTo("POST", url, payload)
}

func fastPutTo(url string, payload []byte) (int, []byte, error) {
	return fastSendTo("PUT", url, payload)
}

func fastPatchTo(url string, payload []byte) (int, []byte, error) {
	return fastSendTo("PATCH", url, payload)
}

var fastClient *fasthttp.Client

func fastGetClient() *fasthttp.Client {
	if fastClient == nil {
		fastClient = &fasthttp.Client{}
	}
	return fastClient
}

func fastSendTo(method, url string, payload []byte) (int, []byte, error) {
	c := fastGetClient()
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetMethod(method)
	if payload != nil {
		req.AppendBody(payload)
	}
	resp := fasthttp.AcquireResponse()
	err := c.Do(req, resp)
	return resp.StatusCode(), resp.Body(), err
}

func validateResp(statusCode int, body string, err error) {
	if err != nil {
		panic(err)
	}
	if statusCode != 200 {
		fmt.Printf("Request failed with status code %d\n", statusCode)
		panic(body)
	}
}

var client remote.MQTTClient

func rpcSendMessage(message *remote.Message) (*remote.SendMessageResult, error) {
	var err error
	if client == nil {
		client, err = testing.GetRPCTestClient(52345)
		if err != nil {
			return nil, err
		}
	}

	response, err := client.SendMessage(context.Background(), message)
	if err != nil {
		return nil, err
	}

	return response, nil
}
