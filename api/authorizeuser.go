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
	"github.com/topfreegames/arkadiko/log"
	"github.com/uber-go/zap"
)

type authorizationPayload struct {
	UserId string   `json:"userId"`
	Rooms  []string `json:"rooms"`
}

// AuthorizeUsersHandler is the handler responsible for authorizing users in rooms
func AuthorizeUsersHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.With(
			zap.String("handler", "AuthorizeUsersHandler"),
		)
		log.D(lg, "Retrieving redis connection...")
		var redisConn redis.Conn
		WithSegment("redis", c, func() error {
			redisConn = app.RedisClient.Pool.Get()
			return nil
		})
		defer redisConn.Close()

		var err error
		var jsonPayload authorizationPayload
		err = WithSegment("payload", c, func() error {
			body := c.Request().Body()
			b, err := ioutil.ReadAll(body)
			if err != nil {
				return err
			}
			return json.Unmarshal(b, &jsonPayload)
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}
		if jsonPayload.UserId == "" || len(jsonPayload.Rooms) == 0 {
			return FailWith(400, "Missing user or rooms", c)
		}
		for _, topic := range jsonPayload.Rooms {
			log.D(lg, "authorizing user", func(cm log.CM) {
				cm.Write(zap.String("user", jsonPayload.UserId), zap.String("room", topic))
			})
			authorizationString := fmt.Sprintf("%s-%s", jsonPayload.UserId, topic)
			err = WithSegment("redis", c, func() error {
				_, err = redisConn.Do("set", authorizationString, 2)
				return err
			})
			if err != nil {
				log.E(lg, "Failed to authorize user in redis.", func(cm log.CM) {
					cm.Write(zap.Error(err))
				})
				return FailWith(500, err.Error(), c)
			}
			log.I(lg, "authorized user into rooms", func(cm log.CM) {
				cm.Write(zap.String("user", jsonPayload.UserId), zap.String("room", topic))
			})
		}
		return SucceedWith(map[string]interface{}{}, c)
	}
}
