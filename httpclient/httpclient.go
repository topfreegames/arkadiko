package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	ehttp "github.com/topfreegames/extensions/http"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// HTTPError is the error in a http call
type HTTPError struct {
	StatusCode int
}

// Error returns an http error string
func (e *HTTPError) Error() string {
	return fmt.Sprintf("received status %d", e.StatusCode)
}

// NewHTTPError is the HTTPError ctor
func NewHTTPError(statusCode int) *HTTPError {
	return &HTTPError{statusCode}
}

type HttpClient struct {
	HttpServerUrl string
	user          string
	password      string
	ConfigPath    string
	Config        *viper.Viper
	Logger        log.FieldLogger
	httpClient    *http.Client
}

type MqttPost struct {
	Topic    string `json:"topic"`
	Payload  string `json:"payload"`
	Qos      int    `json:"qos"`
	Retain   bool   `json:"retain"`
	ClientId string `json:"clientid"`
}

var client *HttpClient
var once sync.Once

func getHTTPTransport(
	maxIdleConns, maxIdleConnsPerHost int,
) http.RoundTripper {
	if _, ok := http.DefaultTransport.(*http.Transport); !ok {
		return http.DefaultTransport // tests use a mock transport
	}

	// We can't get http.DefaultTransport here and update its
	// fields since it's an exported variable, so other libs could
	// also change it and overwrite. This hardcoded values are copied
	// from http.DefaultTransport but could be configurable too.
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          maxIdleConns,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConnsPerHost:   maxIdleConnsPerHost,
	}
}

// GetHqttClient creates the hqttclient and returns it
func GetHttpClient(configPath string, l log.FieldLogger) *HttpClient {
	once.Do(func() {
		client = &HttpClient{
			ConfigPath: configPath,
			Config:     viper.New(),
		}
		client.configure(l)
	})
	return client
}

// SendMessage sends a message to mqqt using a HTTP POST request
func (mc *HttpClient) SendMessage(ctx context.Context, topic string, payload string, retainBool bool) error {
	// Create a child span for the HTTP SendMessage operation if a parent span exists in the context
	if ctx == nil {
		ctx = context.Background()
	}
	var span trace.Span
	ctx, span = trace.SpanFromContext(ctx).TracerProvider().Tracer("arkadiko").Start(ctx, "HttpClient.SendMessage")
	defer span.End()

	// Add relevant attributes to the span
	span.SetAttributes(
		attribute.String("mqtt.topic", topic),
		attribute.Bool("mqtt.retained", retainBool),
		attribute.Int("mqtt.message_length", len(payload)),
		attribute.String("http.url", mc.HttpServerUrl+"/api/v4/mqtt/publish"),
		attribute.String("http.method", "POST"),
	)

	lg := mc.Logger.WithFields(log.Fields{
		"topic":   topic,
		"retain":  retainBool,
		"payload": payload,
	})
	form := &MqttPost{
		Topic:    topic,
		Payload:  payload,
		Retain:   retainBool,
		Qos:      2,
		ClientId: fmt.Sprintf("arkadiko-%s", uuid.NewV4().String()),
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(form)

	startTime := time.Now()

	span.AddEvent("http_request_start")

	req, err := http.NewRequest(
		"POST",
		mc.HttpServerUrl+"/api/v4/mqtt/publish",
		b,
	)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to build HTTP request")
		lg.WithError(err).Error("failed to build http request")
		return err
	}
	req = req.WithContext(ctx)

	req.SetBasicAuth(mc.user, mc.password)
	req.Header.Add("Content-Type", "application/json")

	res, err := mc.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to make HTTP request")
		lg.WithError(err).Error("failed to make request")
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to read response body")
		lg.WithError(err).Error("failed to read body")
		return err
	}

	// Calculate and record request latency
	latency := time.Since(startTime)
	span.SetAttributes(attribute.Int64("http.latency_ns", latency.Nanoseconds()))

	// Add response information
	span.SetAttributes(
		attribute.Int("http.status_code", res.StatusCode),
		attribute.Int("http.response_size", len(body)),
	)

	if res.StatusCode > 399 {
		err := NewHTTPError(res.StatusCode)
		span.RecordError(err)
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP error: %d", res.StatusCode))
		lg.WithError(err).WithField("body", body).Error("failed request")
		return err
	}

	span.SetStatus(codes.Ok, "")
	span.AddEvent("http_request_success")
	return nil
}

func (mc *HttpClient) configure(l log.FieldLogger) {
	mc.Logger = l

	mc.setConfigurationDefaults()
	mc.loadConfiguration()
	mc.configureClient()
}

func (mc *HttpClient) setConfigurationDefaults() {
	mc.Config.SetDefault("httpserver.url", "http://localhost:8081")
	mc.Config.SetDefault("httpserver.user", "admin")
	mc.Config.SetDefault("httpserver.pass", "public")
	mc.Config.SetDefault("httpserver.timeout", 500)
	mc.Config.SetDefault("httpserver.maxIdleConnsPerHost", http.DefaultMaxIdleConnsPerHost)
	mc.Config.SetDefault("httpserver.maxIdleConns", 100)
}

func (mc *HttpClient) configureClient() {
	timeout := time.Duration(mc.Config.GetInt("httpserver.timeout")) * time.Millisecond
	maxIdleConns := mc.Config.GetInt("httpserver.maxIdleConns")
	maxIdleConnsPerHost := mc.Config.GetInt("httpserver.maxIdleConnsPerHost")

	mc.httpClient = &http.Client{
		Transport: getHTTPTransport(maxIdleConns, maxIdleConnsPerHost),
		Timeout:   timeout,
	}
	ehttp.Instrument(mc.httpClient)

	mc.HttpServerUrl = mc.Config.GetString("httpserver.url")
	mc.user = mc.Config.GetString("httpserver.user")
	mc.password = mc.Config.GetString("httpserver.pass")
}

func (mc *HttpClient) loadConfiguration() {
	mc.Config.SetConfigFile(mc.ConfigPath)
	mc.Config.SetEnvPrefix("arkadiko")
	mc.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	mc.Config.AutomaticEnv()

	if err := mc.Config.ReadInConfig(); err == nil {
		mc.Logger.WithFields(log.Fields{
			"configFile": mc.Config.ConfigFileUsed(),
		}).Info("Loaded config file.")
	} else {
		panic(fmt.Sprintf("Could not load configuration file from: %s", mc.ConfigPath))
	}
}
