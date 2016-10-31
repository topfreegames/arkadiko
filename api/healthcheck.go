// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/garyburd/redigo/redis"
	"github.com/labstack/echo"
)

// HealthCheckHandler is the handler responsible for validating that the app is still up
func HealthCheckHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "Healthcheck")
		workingString := app.Config.GetString("healthcheck.workingText")

		err := WithSegment("redis", c, func() error {
			var redisConn redis.Conn
			WithSegment("redis", c, func() error {
				redisConn = app.RedisClient.Pool.Get()
				return nil
			})
			defer redisConn.Close()
			res, err := redisConn.Do("ping")
			if err != nil || res != "PONG" {
				return fmt.Errorf("Error connecting to redis: %s", err)
			}

			return nil
		})
		if err != nil {
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		workingString = strings.TrimSpace(workingString)
		return c.String(http.StatusOK, workingString)
	}
}
