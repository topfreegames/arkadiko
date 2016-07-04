// mqttbridge
// https://github.com/topfreegames/mqttbridge
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"strings"

	"github.com/kataras/iris"
)

// HealthCheckHandler is the handler responsible for validating that the app is still up
func HealthCheckHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		workingString := app.Config.GetString("healthcheck.workingText")

		c.SetStatusCode(iris.StatusOK)
		workingString = strings.TrimSpace(workingString)
		c.Write(workingString)
	}
}
