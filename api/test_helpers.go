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
	"net/http/httptest"
	"strings"
	"testing"

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
	app, err := GetApp("0.0.0.0", 8890, "../config/test.yml", false, logger)
	if err != nil {
		logger.Fatal("Could not get test application.", zap.Error(err))
	}
	err = app.Configure()
	if err != nil {
		logger.Fatal("Could not get test application.", zap.Error(err))
	}
	return app
}

func request(method, path, body string, app *App) (int, string) {
	var req *http.Request
	if body != "" {
		reader := strings.NewReader(body) //Convert string to reader
		req, _ = http.NewRequest(method, path, reader)
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}
	rec := httptest.NewRecorder()
	app.App.ServeHTTP(rec, req)
	return rec.Code, rec.Body.String()
}

// Get returns a test request against specified URL
func Get(app *App, url string, t *testing.T) (int, string) {
	return request("GET", url, "", app)
}

// PostBody returns a test request against specified URL
func PostBody(app *App, url string, payload string) (int, string) {
	return sendBody(app, "POST", url, payload)
}

func sendBody(app *App, method, url, payload string) (int, string) {
	return request(method, url, payload, app)
}

// PostJSON returns a test request against specified URL
func PostJSON(app *App, url string, payload JSON) (int, string) {
	return sendJSON(app, "POST", url, payload)
}

func sendJSON(app *App, method, url string, payloadJSON JSON) (int, string) {
	payload, _ := json.Marshal(payloadJSON)
	return request(method, url, string(payload), app)
}
