// mqttbridge
// https://github.com/topfreegames/mqttbridge
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
)

// GetDefaultTestApp returns a new mqttbridge API Application bound to 0.0.0.0:8888 for test
func GetDefaultTestApp() *App {
	app := GetApp("0.0.0.0", 8890, "../config/test.yml", true)
	app.Configure()
	return app
}

// Get returns a test request against specified URL
func Get(app *App, url string, t *testing.T) *httpexpect.Response {
	req := sendRequest(app, "GET", url, t)
	return req.Expect()
}

// PostBody returns a test request against specified URL
func PostBody(app *App, url string, t *testing.T, payload string) *httpexpect.Response {
	return sendBody(app, "POST", url, t, payload)
}

func sendBody(app *App, method string, url string, t *testing.T, payload string) *httpexpect.Response {
	req := sendRequest(app, method, url, t)
	return req.WithBytes([]byte(payload)).Expect()
}

// PostJSON returns a test request against specified URL
func PostJSON(app *App, url string, t *testing.T, payload JSON) *httpexpect.Response {
	return sendJSON(app, "POST", url, t, payload)
}

func sendJSON(app *App, method, url string, t *testing.T, payload JSON) *httpexpect.Response {
	req := sendRequest(app, method, url, t)
	return req.WithJSON(payload).Expect()
}

func sendRequest(app *App, method, url string, t *testing.T) *httpexpect.Request {
	handler := app.App.NoListen().Handler

	e := httpexpect.WithConfig(httpexpect.Config{
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewFastBinder(handler),
		},
	})

	return e.Request(method, url)
}
