// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/arkadiko/api"
	. "github.com/topfreegames/arkadiko/testing"
)

var _ = Describe("Send to MQTT Handler", func() {
	Describe("Specs", func() {
		It("Should respond with 200 for a valid message", func() {
			a := GetDefaultTestApp()
			testJSON := map[string]interface{}{
				"message": "hello",
			}
			response := `{"topic": "test", "payload": {"message":"hello"}}`
			status, body := PostJSON(a, "/sendmqtt/test", testJSON)

			Expect(status).To(Equal(http.StatusOK))
			Expect(body).To(Equal(response))
		})

		It("Should respond with 200 for a valid message with hierarchical topic", func() {
			a := GetDefaultTestApp()
			testJSON := map[string]interface{}{
				"message": "hello",
			}
			response := `{"topic": "test/topic", "payload": {"message":"hello"}}`
			url := "/sendmqtt/test/topic"
			status, body := PostJSON(a, url, testJSON)

			Expect(status).To(Equal(http.StatusOK))
			Expect(body).To(Equal(response))
		})

		It("Should respond with 400 if malformed JSON", func() {
			a := GetDefaultTestApp()
			testJSON := `{"message": "hello"}}`
			status, _ := PostBody(a, "/sendmqtt/test/topic", testJSON)

			Expect(status).To(Equal(400))
		})
	})
	Describe("Perf", func() {
		HTTPMeasure("Should send fast to MQTT", func(data map[string]interface{}) {
			testJSON := map[string]interface{}{
				"message": "hello",
			}
			payload, err := json.Marshal(testJSON)
			Expect(err).NotTo(HaveOccurred())
			data["payload"] = payload
		}, func(ts *httptest.Server, data map[string]interface{}) {
			app := data["app"].(*api.App)
			payload := string(data["payload"].([]byte))
			status, body := PostBody(app, "/sendmqtt/test/topic", payload)
			Expect(status).To(Equal(http.StatusOK), string(body))
		}, 0.005)
	})
})
