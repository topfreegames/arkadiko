// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2018 Top Free Games <backend@tfgco.com>

package echo

import (
	"fmt"

	"github.com/labstack/echo"
	"github.com/opentracing/opentracing-go"
	"github.com/topfreegames/extensions/jaeger"
)

// Instrument adds Jaeger instrumentation on an Echo app
func Instrument(app *echo.Echo) {
	app.Use(middleware)
}

func middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		tracer := opentracing.GlobalTracer()

		request := c.Request()
		route := c.Path()

		header := opentracing.HTTPHeadersCarrier(request.Header)
		parent, _ := tracer.Extract(opentracing.HTTPHeaders, header)

		operationName := fmt.Sprintf("HTTP %s %s", request.Method, route)
		reference := opentracing.ChildOf(parent)
		tags := opentracing.Tags{
			"http.method":   request.Method,
			"http.host":     request.Host,
			"http.pathname": request.URL.Path,
			"http.query":    c.QueryString(),

			"span.kind": "server",
		}

		span := opentracing.StartSpan(operationName, reference, tags)
		defer span.Finish()
		defer jaeger.LogPanic(span)

		ctx := request.Context()
		ctx = opentracing.ContextWithSpan(ctx, span)
		request = request.WithContext(ctx)

		c.SetRequest(request)

		err := next(c)
		if err != nil {
			message := err.Error()
			jaeger.LogError(span, message)
		}

		response := c.Response()
		statusCode := response.Status

		span.SetTag("http.status_code", statusCode)

		return err
	}
}
