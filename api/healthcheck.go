// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

// HealthCheckHandler is the handler responsible for validating that the app is still up
func HealthCheckHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "Healthcheck")
		workingString := app.Config.GetString("healthcheck.workingText")
		workingString = strings.TrimSpace(workingString)
		return c.String(http.StatusOK, workingString)
	}
}
