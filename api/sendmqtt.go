// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"encoding/json"

	"github.com/kataras/iris"
)

func failWith(status int, message string, c *iris.Context) {
	c.SetStatusCode(status)
	c.Write(message)
}

type sendMqttPayload struct {
	Topic   string `json:"topic"`
	Payload JSON   `json:"payload"`
}

// SendMqttHandler is the handler responsible for sending messages to mqtt
func SendMqttHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		var jsonPayload sendMqttPayload
		err := json.Unmarshal(c.RequestCtx.Request.Body(), &jsonPayload)
		if err != nil {
			failWith(400, err.Error(), c)
			return
		}
		if jsonPayload.Topic == "" || len(jsonPayload.Payload) == 0 {
			failWith(400, "Missing topic or payload", c)
			return
		}

		b, err := json.Marshal(jsonPayload.Payload)
		if err != nil {
			failWith(400, err.Error(), c)
			return
		}
		workingString := string(b[:])

		err = app.MqttClient.SendMessage(jsonPayload.Topic, string(b[:]))
		if err != nil {
			failWith(400, err.Error(), c)
			return
		}

		c.SetStatusCode(iris.StatusOK)
		c.Write(workingString)
	}
}
