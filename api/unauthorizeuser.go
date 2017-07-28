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
	"io/ioutil"

	"github.com/labstack/echo"
)

type unauthorizationPayload struct {
	UserID string   `json:"userId"`
	Rooms  []string `json:"rooms"`
}

// UnauthorizeUsersHandler is the handler responsible for unauthorizing users in rooms
func UnauthorizeUsersHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.WithField(
			"handler", "UnauthorizeUsersHandler",
		)
		lg.Debug("Retrieving redis connection...")
		redisConn := app.RedisClient.Pool.Get()
		defer redisConn.Close()

		var err error
		var jsonPayload authorizationPayload
		err = WithSegment("payload", c, func() error {
			body := c.Request().Body
			b, err := ioutil.ReadAll(body)
			if err != nil {
				return err
			}
			return json.Unmarshal(b, &jsonPayload)
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}
		if jsonPayload.UserID == "" || len(jsonPayload.Rooms) == 0 {
			return FailWith(400, "Missing user or rooms", c)
		}
		for _, topic := range jsonPayload.Rooms {
			lg.Debug("unauthorizing user")
			unauthorizationString := fmt.Sprintf("%s-%s", jsonPayload.UserID, topic)
			err = WithSegment("redis", c, func() error {
				_, err = redisConn.Do("del", unauthorizationString)
				return err
			})
			if err != nil {
				lg.WithError(err).Error("Failed to unauthorize user in redis.")
				return FailWith(500, err.Error(), c)
			}
			lg.Info("unauthorized user into rooms")
		}
		return SucceedWith(map[string]interface{}{}, c)
	}
}
