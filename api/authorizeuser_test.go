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

var _ = Describe("Authorize Handler", func() {
	Describe("Specs", func() {
		Describe("Should authorize users", func() {
			It("Should respond with 200 for a valid message", func() {
				a := GetDefaultTestApp()
				var jsonPayload map[string]interface{}
				testJSON := `{"userId": "felipe", "rooms": ["room1", "room2"]}`
				json.Unmarshal([]byte(testJSON), &jsonPayload)

				status, _ := PostJSON(a, "/authorize_user", jsonPayload)
				Expect(status).To(Equal(http.StatusOK))
			})

			It("Should respond with 400 if malformed map[string]interface{}", func() {
				a := GetDefaultTestApp()
				var jsonPayload map[string]interface{}
				testJSON := `{"message": "hello"}}`
				json.Unmarshal([]byte(testJSON), &jsonPayload)

				status, _ := PostJSON(a, "/authorize_user", jsonPayload)
				Expect(status).To(Equal(400))
			})

			It("Should respond with 400 if no userId is provided", func() {
				a := GetDefaultTestApp()
				var jsonPayload map[string]interface{}
				testJSON := `{"message": "hello", "rooms": ["room1"]}`
				json.Unmarshal([]byte(testJSON), &jsonPayload)

				status, _ := PostJSON(a, "/authorize_user", jsonPayload)
				Expect(status).To(Equal(400))
			})

			It("Should respond with 400 if no rooms is provided", func() {
				a := GetDefaultTestApp()
				var jsonPayload map[string]interface{}
				testJSON := `{"userId": "hello", "roo": ["room1"]}`
				json.Unmarshal([]byte(testJSON), &jsonPayload)

				status, _ := PostJSON(a, "/authorize_user", jsonPayload)
				Expect(status).To(Equal(400))
			})
		})
	})

	Describe("Perf", func() {
		HTTPMeasure("authorize user", func(data map[string]interface{}) {
			testJSON := map[string]interface{}{
				"userId": "felipe",
				"rooms":  []string{"room1", "room2"},
			}
			payload, err := json.Marshal(testJSON)
			Expect(err).NotTo(HaveOccurred())
			data["payload"] = payload
		}, func(ts *httptest.Server, data map[string]interface{}) {
			payload := string(data["payload"].([]byte))
			app := data["app"].(*api.App)
			status, body := PostBody(app, "/authorize_user", payload)
			Expect(status).To(Equal(http.StatusOK), string(body))
		}, 0.05)
	})
})
