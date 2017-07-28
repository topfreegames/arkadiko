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

	"github.com/garyburd/redigo/redis"
	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

type authorizationPayload struct {
	UserID string   `json:"userId"`
	Rooms  []string `json:"rooms"`
}

// AuthorizeUsersHandler is the handler responsible for authorizing users in rooms
func AuthorizeUsersHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.WithFields(log.Fields{
			"handler": "AuthorizeUsersHandler",
		})
		lg.Debug("Retrieving redis connection...")
		var redisConn redis.Conn
		WithSegment("redis", c, func() error {
			redisConn = app.RedisClient.Pool.Get()
			return nil
		})
		defer redisConn.Close()

		var err error
		var jsonPayload authorizationPayload
		err = WithSegment("payload", c, func() error {
			b, err := ioutil.ReadAll(c.Request().Body)
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
			lg.Debug("authorizing user")
			authorizationString := fmt.Sprintf("%s-%s", jsonPayload.UserID, topic)
			err = WithSegment("redis", c, func() error {
				_, err = redisConn.Do("set", authorizationString, 2)
				return err
			})
			if err != nil {
				lg.WithError(err).Error("Failed to authorize user in redis.")
				return FailWith(500, err.Error(), c)
			}
			lg.Info("authorized user into rooms")
		}
		return SucceedWith(map[string]interface{}{}, c)
	}
}
