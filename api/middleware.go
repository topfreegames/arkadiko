// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	raven "github.com/getsentry/raven-go"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

const responseTimeMillisecondsMetricName = "response_time_milliseconds"

// NewVersionMiddleware with API version
func NewVersionMiddleware() *VersionMiddleware {
	return &VersionMiddleware{
		Version: VERSION,
	}
}

// VersionMiddleware inserts the current version in all requests
type VersionMiddleware struct {
	Version string
}

// Serve serves the middleware
func (v *VersionMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderServer, fmt.Sprintf("Arkadiko/v%s", v.Version))
		c.Response().Header().Set("Arkadiko-Server", fmt.Sprintf("Arkadiko/v%s", v.Version))
		return next(c)
	}
}

// NewSentryMiddleware returns a new sentry middleware
func NewSentryMiddleware(app *App) *SentryMiddleware {
	return &SentryMiddleware{
		App: app,
	}
}

// SentryMiddleware is responsible for sending all exceptions to sentry
type SentryMiddleware struct {
	App *App
}

// Serve serves the middleware
func (s *SentryMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err != nil {
			if httpErr, ok := err.(*echo.HTTPError); ok {
				if httpErr.Code < 500 {
					return err
				}
			}
			tags := map[string]string{
				"source": "app",
				"type":   "Internal server error",
				"url":    c.Request().URL.String(),
				"status": fmt.Sprintf("%d", c.Response().Status),
			}
			raven.SetHttpContext(newHTTPFromCtx(c))
			raven.CaptureError(err, tags)
		}
		return err
	}
}

// ResponseTimeMetricsMiddleware struct encapsulating DDStatsD
type ResponseTimeMetricsMiddleware struct {
	DDStatsD      MetricsReporter
	latencyMetric *prometheus.HistogramVec
}

// ResponseTimeMetricsMiddleware returns a new ResponseTimeMetricsMiddleware
func NewResponseTimeMetricsMiddleware(
	ddStatsD MetricsReporter,
	latencyMetric *prometheus.HistogramVec,
) *ResponseTimeMetricsMiddleware {

	return &ResponseTimeMetricsMiddleware{
		DDStatsD:      ddStatsD,
		latencyMetric: latencyMetric,
	}
}

// ResponseTimeMetricsMiddleware is a middleware to measure the response time
// of a route and send it do StatsD
func (responseTimeMiddleware ResponseTimeMetricsMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		startTime := time.Now()
		result := next(c)
		status := c.Response().Status
		route := c.Path()
		method := c.Request().Method

		timeUsed := time.Since(startTime)

		tags := []string{
			fmt.Sprintf("route:%s", route),
			fmt.Sprintf("method:%s", method),
			fmt.Sprintf("status:%d", status),
		}

		// Keeping both for retro compatibility in the short term
		responseTimeMiddleware.DDStatsD.Timing(responseTimeMillisecondsMetricName, timeUsed, tags...)
		responseTimeMiddleware.latencyMetric.WithLabelValues(route, method, fmt.Sprintf("%d", status)).Observe(timeUsed.Seconds())

		return result
	}
}

func getHTTPParams(ctx echo.Context) (string, map[string]string, string) {
	qs := ""
	if len(ctx.QueryParams()) > 0 {
		qsBytes, _ := json.Marshal(ctx.QueryParams())
		qs = string(qsBytes)
	}

	//TODO: Fix this
	//headers := map[string]string{}
	//for _, headerKey := range ctx.Response().Header(). {
	//headers[string(headerKey)] = string(ctx.Response().Header().Get(headerKey))
	//}

	cookies := string(ctx.Response().Header().Get("Cookie"))
	return qs, map[string]string{}, cookies
}

func newHTTPFromCtx(ctx echo.Context) *raven.Http {
	qs, headers, cookies := getHTTPParams(ctx)

	h := &raven.Http{
		Method:  string(ctx.Request().Method),
		Cookies: cookies,
		Query:   qs,
		URL:     ctx.Request().URL.String(),
		Headers: headers,
	}
	return h
}

// NewRecoveryMiddleware returns a configured middleware
func NewRecoveryMiddleware(onError func(error, []byte)) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		OnError: onError,
	}
}

// RecoveryMiddleware recovers from errors
type RecoveryMiddleware struct {
	OnError func(error, []byte)
}

// Serve executes on error handler when errors happen
func (r *RecoveryMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		defer func() {
			if err := recover(); err != nil {
				eError, ok := err.(error)
				if !ok {
					eError = fmt.Errorf(fmt.Sprintf("%v", err))
				}
				if r.OnError != nil {
					r.OnError(eError, debug.Stack())
				}
				c.Error(eError)
			}
		}()
		return next(c)
	}
}

// NewLoggerMiddleware returns the logger middleware
func NewLoggerMiddleware(theLogger log.FieldLogger) *LoggerMiddleware {
	l := &LoggerMiddleware{Logger: theLogger}
	return l
}

// LoggerMiddleware is responsible for logging to Zap all requests
type LoggerMiddleware struct {
	Logger log.FieldLogger
}

// Serve serves the middleware
func (l *LoggerMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		l := l.Logger.WithFields(log.Fields{
			"source": "request",
		})

		//all except latency to string
		var ip, method, path string
		var status int
		var latency time.Duration
		var startTime, endTime time.Time

		path = c.Path()
		method = c.Request().Method

		startTime = time.Now()

		err := next(c)

		//no time.Since in order to format it well after
		endTime = time.Now()
		latency = endTime.Sub(startTime)

		status = c.Response().Status
		ip = c.Request().RemoteAddr

		route := c.Path()
		reqLog := l.WithFields(log.Fields{
			"route":      route,
			"endTime":    endTime,
			"statusCode": status,
			"latency":    latency,
			"ip":         ip,
			"method":     method,
			"path":       path,
		})

		mqttLatencyInterface := c.Get("mqttLatency")
		if mqttLatencyInterface != nil {
			mqttLatency := mqttLatencyInterface.(time.Duration)
			reqLog = reqLog.WithField("mqttLatency", mqttLatency)
		}

		requestorInterface := c.Get("requestor")
		if requestorInterface != nil {
			requestor := requestorInterface.(string)
			reqLog = reqLog.WithField("requestor", requestor)
		}

		retainedInterface := c.Get("retained")
		if retainedInterface != nil {
			retained := retainedInterface.(bool)
			reqLog = reqLog.WithField("retained", retained)
		}

		//request failed
		if status > 399 && status < 500 {
			reqLog.WithError(err).Warn("Request failed.")
			return err
		}

		//request is ok, but server failed
		if status > 499 {
			reqLog.WithError(err).Error("Response failed.")
			return err
		}

		//Everything went ok
		reqLog.Debug("Request successful.")
		return err
	}
}

// NewNewRelicMiddleware returns the logger middleware
func NewNewRelicMiddleware(app *App, theLogger log.FieldLogger) *NewRelicMiddleware {
	l := &NewRelicMiddleware{App: app, Logger: theLogger}
	return l
}

// NewRelicMiddleware is responsible for logging to Zap all requests
type NewRelicMiddleware struct {
	App    *App
	Logger log.FieldLogger
}

// Serve serves the middleware
func (nr *NewRelicMiddleware) Serve(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		route := c.Path()
		txn := nr.App.NewRelic.StartTransaction(route, nil, nil)
		c.Set("txn", txn)
		defer func() {
			c.Set("txn", nil)
			txn.End()
		}()

		err := next(c)
		if err != nil && c.Response().Status != 404 {
			txn.NoticeError(err)
			return err
		}

		return nil
	}
}
