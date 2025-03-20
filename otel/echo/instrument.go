// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2018 Top Free Games <backend@tfgco.com>

package echo

import (
	"fmt"
	"strings"

	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

// // Instrument adds Jaeger instrumentation on an Echo app
// func Instrument(app *echo.Echo) {
// 	app.Use(otelecho.Middleware("my-server"))
// }
func Instrument(app *echo.Echo) {
	app.Use(otelecho.Middleware("my-server", otelecho.WithSkipper(func(c echo.Context) bool {
		return strings.Contains(c.Path(), "/healthcheck") 
	})))
}
