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

	"github.com/kataras/iris"
	"github.com/uber-go/zap"
)

type authorizationPayload struct {
	UserId string   `json:"userId"`
	Rooms  []string `json:"rooms"`
}

// AuthorizeUsersHandler is the handler responsible for authorizing users in rooms
func AuthorizeUsersHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		redisConn := app.RedisClient.Pool.Get()
		defer redisConn.Close()
		var jsonPayload authorizationPayload
		err := json.Unmarshal(c.RequestCtx.Request.Body(), &jsonPayload)
		if err != nil {
			failWith(400, err.Error(), c)
			return
		}
		if jsonPayload.UserId == "" || len(jsonPayload.Rooms) == 0 {
			failWith(400, "Missing user or rooms", c)
			return
		}
		for _, topic := range jsonPayload.Rooms {
			app.Logger.Debug("authorizing user", zap.String("user", jsonPayload.UserId), zap.String("room", topic))
			authorizationString := fmt.Sprintf("%s-%s", jsonPayload.UserId, topic)
			_, err := redisConn.Do("set", authorizationString, 2)
			if err != nil {
				app.Logger.Error(err.Error())
				c.SetStatusCode(iris.StatusInternalServerError)
				return
			}
			app.Logger.Info("authorized user into rooms", zap.String("user", jsonPayload.UserId), zap.String("room", topic))
		}
		c.SetStatusCode(iris.StatusOK)
	}
}
