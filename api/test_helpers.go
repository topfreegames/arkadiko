// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
	echostandard "github.com/labstack/echo/engine/standard"
	"github.com/uber-go/zap"
)

// GetDefaultTestApp returns a new arkadiko API Application bound to 0.0.0.0:8888 for test
func GetDefaultTestApp() *App {
	logger := zap.New(
		zap.NewJSONEncoder(),
		zap.FatalLevel,
	).With(
		zap.String("source", "app"),
	)
	app, err := GetApp("0.0.0.0", 8890, "../config/test.yml", true, false, logger)
	if err != nil {
		logger.Fatal("Could not get test application.", zap.Error(err))
	}
	err = app.Configure()
	if err != nil {
		logger.Fatal("Could not get test application.", zap.Error(err))
	}
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
	app.Engine.SetHandler(app.App)
	handler := http.Handler(app.Engine.(*echostandard.Server))

	e := httpexpect.WithConfig(httpexpect.Config{
		Reporter: httpexpect.NewAssertReporter(t),
		Client: &http.Client{
			Transport: httpexpect.NewBinder(handler),
		},
	})

	return e.Request(method, url)
}
