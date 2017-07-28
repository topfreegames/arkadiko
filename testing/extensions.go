// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package testing

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"google.golang.org/grpc"

	"github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/types"
	"github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"github.com/topfreegames/arkadiko/api"
	"github.com/topfreegames/arkadiko/remote"
)

//InitializeTestServer method
func InitializeTestServer(app *api.App) *httptest.Server {
	return httptest.NewServer(app.App)
}

//BeforeOnce runs the before each block only once
func BeforeOnce(beforeBlock func()) {
	hasRun := false

	ginkgo.BeforeEach(func() {
		if !hasRun {
			beforeBlock()
			hasRun = true
		}
	})
}

// GetDefaultTestApp returns a new arkadiko API Application bound to 0.0.0.0:8888 for test
func GetDefaultTestApp() *api.App {
	logger := log.WithField("source", "app")
	app, err := api.GetApp("0.0.0.0", 8890, "../config/test.yml", false, logger)
	if err != nil {
		logger.WithError(err).Fatal("Could not get test application.")
	}
	err = app.Configure()
	if err != nil {
		logger.WithError(err).Fatal("Could not get test application.")
	}
	return app
}

// GetDefaultTestServer returns a new arkadiko RPC Server bound to 0.0.0.0:8891 for test
func GetDefaultTestServer() (*remote.Server, error) {
	logger := log.WithField("source", "rpc")
	server, err := remote.NewServer("0.0.0.0", 8891, "../config/test.yml", false, logger)
	if err != nil {
		return nil, err
	}
	return server, nil
}

//GetRPCTestClient returns a connected test client
func GetRPCTestClient(portOrNil ...int) (remote.MQTTClient, error) {
	port := 8891
	if len(portOrNil) == 1 {
		port = portOrNil[0]
	}
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}

	conn, err := grpc.Dial(fmt.Sprintf("0.0.0.0:%d", port), opts...)
	if err != nil {
		return nil, err
	}
	return remote.NewMQTTClient(conn), err
}

//HTTPMeasure runs the specified specs in an http test
func HTTPMeasure(description string, setup func(map[string]interface{}), f func(*httptest.Server, map[string]interface{}), timeout float64) bool {
	return measure(description, setup, f, timeout, types.FlagTypeNone)
}

//FHTTPMeasure runs the specified specs in an http test
func FHTTPMeasure(description string, setup func(map[string]interface{}), f func(*httptest.Server, map[string]interface{}), timeout float64) bool {
	return measure(description, setup, f, timeout, types.FlagTypeFocused)
}

//XHTTPMeasure runs the specified specs in an http test
func XHTTPMeasure(description string, setup func(map[string]interface{}), f func(*httptest.Server, map[string]interface{}), timeout float64) bool {
	return measure(description, setup, f, timeout, types.FlagTypePending)
}

func measure(description string, setup func(map[string]interface{}), f func(*httptest.Server, map[string]interface{}), timeout float64, flagType types.FlagType) bool {
	app := GetDefaultTestApp()
	d := func(t string, f func()) { ginkgo.Describe(t, f) }
	if flagType == types.FlagTypeFocused {
		d = func(t string, f func()) { ginkgo.FDescribe(t, f) }
	}
	if flagType == types.FlagTypePending {
		d = func(t string, f func()) { ginkgo.XDescribe(t, f) }
	}

	d("Measure", func() {
		var ts *httptest.Server
		var loops int
		var ctx map[string]interface{}

		BeforeOnce(func() {
			ts = InitializeTestServer(app)
			ctx = map[string]interface{}{"app": app}
			setup(ctx)
		})

		ginkgo.AfterEach(func() {
			loops++
			if loops == 200 {
				ts.Close()
			}
		})

		ginkgo.Measure(description, func(b ginkgo.Benchmarker) {
			runtime := b.Time("runtime", func() {
				f(ts, ctx)
			})
			gomega.Expect(runtime.Seconds()).Should(
				gomega.BeNumerically("<", timeout),
				fmt.Sprintf("%s shouldn't take too long.", description),
			)
		}, 20)
	})

	return true
}

func request(method, path, body string, app *api.App) (int, string) {
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
func Get(app *api.App, url string) (int, string) {
	return request("GET", url, "", app)
}

// PostBody returns a test request against specified URL
func PostBody(app *api.App, url string, payload string) (int, string) {
	return sendBody(app, "POST", url, payload)
}

func sendBody(app *api.App, method, url, payload string) (int, string) {
	return request(method, url, payload, app)
}

// PostJSON returns a test request against specified URL
func PostJSON(app *api.App, url string, payload interface{}) (int, string) {
	return sendJSON(app, "POST", url, payload)
}

func sendJSON(app *api.App, method, url string, payloadJSON interface{}) (int, string) {
	payload, _ := json.Marshal(payloadJSON)
	return request(method, url, string(payload), app)
}
