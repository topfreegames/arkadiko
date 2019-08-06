// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package lib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/viper"

	ehttp "github.com/topfreegames/extensions/http"
)

// Arkadiko represents an arkadiko API application
type Arkadiko struct {
	client *http.Client

	baseURL string
	pass    string
	user    string
}

// NewArkadiko returns a new arkadiko API application
func NewArkadiko(config *viper.Viper) *Arkadiko {
	config.SetDefault("arkadiko.maxIdleConns", 100)
	config.SetDefault("arkadiko.maxIdleConnsPerHost", http.DefaultMaxIdleConnsPerHost)
	config.SetDefault("arkadiko.timeout", 500*time.Millisecond)

	return &Arkadiko{
		baseURL: config.GetString("arkadiko.url"),
		pass:    config.GetString("arkadiko.pass"),
		user:    config.GetString("arkadiko.user"),
		client: getHTTPClient(
			config.GetDuration("arkadiko.timeout"),
			config.GetInt("arkadiko.maxIdleConns"),
			config.GetInt("arkadiko.maxIdleConnsPerHost"),
		),
	}
}

func getHTTPClient(timeout time.Duration, maxIdleConns, maxIdleConnsPerHost int) *http.Client {
	client := &http.Client{
		Timeout:   timeout,
		Transport: getHTTPTransport(maxIdleConns, maxIdleConnsPerHost),
	}

	ehttp.Instrument(client)
	return client
}

func getHTTPTransport(maxIdleConns, maxIdleConnsPerHost int) http.RoundTripper {
	// Tests use a mock transport.
	if _, ok := http.DefaultTransport.(*http.Transport); !ok {
		return http.DefaultTransport
	}

	dialer := &net.Dialer{
		DualStack: true,
		KeepAlive: 30 * time.Second,
		Timeout:   30 * time.Second,
	}

	// We can't get http.DefaultTransport here and update its
	// fields since it's an exported variable, so other libs could
	// also change it and overwrite. This hardcoded values are copied
	// from http.DefaultTransport but could be configurable too.
	return &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		ExpectContinueTimeout: 1 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		IdleConnTimeout:       90 * time.Second,
		MaxIdleConns:          maxIdleConns,
		MaxIdleConnsPerHost:   maxIdleConnsPerHost,
	}
}

// SendMQTT publishes an MQTT message on the given topic
func (a *Arkadiko) SendMQTT(ctx context.Context, topic string, payload interface{}, retained bool,
) (*SendMQTTResponse, error) {
	path := fmt.Sprintf("/sendmqtt/%s?retained=%t", url.QueryEscape(topic), retained)

	response, err := a.sendRequest(ctx, "POST", path, payload)
	if err != nil {
		return nil, err
	}

	var result *SendMQTTResponse
	err = json.Unmarshal(response, &result)

	return result, err
}

func (a *Arkadiko) sendRequest(ctx context.Context, method, path string, payload interface{}) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, a.baseURL+path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(a.user, a.pass)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 399 {
		return nil, NewRequestError(resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
