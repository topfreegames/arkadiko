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

var _ = Describe("Healthcheck Handler", func() {
	Describe("Specs", func() {
		It("Should respond with default WORKING string", func() {
			a := GetDefaultTestApp()
			status, body := Get(a, "/healthcheck")

			Expect(status).To(Equal(http.StatusOK))
			Expect(body).To(Equal("WORKING"))
		})

		It("Should respond with customized WORKING string", func() {
			a := GetDefaultTestApp()
			a.Config.SetDefault("healthcheck.workingText", "OTHERWORKING")
			status, body := Get(a, "/healthcheck")

			Expect(status).To(Equal(http.StatusOK))
			Expect(body).To(Equal("OTHERWORKING"))
		})
	})

	Describe("Perf", func() {
		HTTPMeasure("healthcheck", func(data map[string]interface{}) {
			testJSON := map[string]interface{}{
				"message": "hello",
			}
			payload, err := json.Marshal(testJSON)
			Expect(err).NotTo(HaveOccurred())
			data["payload"] = payload
		}, func(ts *httptest.Server, data map[string]interface{}) {
			app := data["app"].(*api.App)
			status, body := Get(app, "/healthcheck")
			Expect(status).To(Equal(http.StatusOK), string(body))
		}, 0.05)
	})
})
