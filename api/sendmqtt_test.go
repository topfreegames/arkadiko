// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"net/http"
	"testing"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestSendMqtt(t *testing.T) {
	g := Goblin(t)

	// special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Send Mqtt", func() {
		g.It("Should respond with 200 for a valid message", func() {
			a := GetDefaultTestApp()
			var jsonPayload JSON
			testJSON := `{"topic": "test topic", "payload": {"message": "hello"}}`
			payload := `{"message":"hello"}`
			json.Unmarshal([]byte(testJSON), &jsonPayload)
			res := PostJSON(a, "/sendmqtt", t, jsonPayload)

			g.Assert(res.Raw().StatusCode).Equal(http.StatusOK)
			res.Body().Equal(payload)
		})

		g.It("Should respond with 400 if missing topic", func() {
			a := GetDefaultTestApp()
			var jsonPayload JSON
			testJSON := `{"topi": "test topic", "payload": {"message": "hello"}}`
			json.Unmarshal([]byte(testJSON), &jsonPayload)
			res := PostJSON(a, "/sendmqtt", t, jsonPayload)

			g.Assert(res.Raw().StatusCode).Equal(400)
		})

		g.It("Should respond with 400 if missing payload", func() {
			a := GetDefaultTestApp()
			var jsonPayload JSON
			testJSON := `{"topic": "test topic", "payloa": {"message": "hello"}}`
			json.Unmarshal([]byte(testJSON), &jsonPayload)
			res := PostJSON(a, "/sendmqtt", t, jsonPayload)

			g.Assert(res.Raw().StatusCode).Equal(400)
		})
	})
}
