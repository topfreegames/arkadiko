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
	"net/url"

	"github.com/kataras/iris"
)

func failWith(status int, message string, c *iris.Context) {
	c.SetStatusCode(status)
	c.Write(message)
}

// SendMqttHandler is the handler responsible for sending messages to mqtt
func SendMqttHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		if string(c.RequestCtx.Request.Body()[:]) == "null" {
			failWith(400, "Invalid JSON", c)
			return
		}
		var msgPayload map[string]interface{}
		err := json.Unmarshal(c.RequestCtx.Request.Body(), &msgPayload)
		if err != nil {
			failWith(400, err.Error(), c)
			return
		}

		topic, err := url.QueryUnescape(c.Param("topic"))
		if err != nil {
			failWith(400, err.Error(), c)
			return
		}

		b, err := json.Marshal(msgPayload)
		if err != nil {
			failWith(400, err.Error(), c)
			return
		}
		workingString := fmt.Sprintf(`{"topic": "%s", "payload": %v}`, topic, string(b[:]))

		err = app.MqttClient.SendMessage(topic, string(b[:]))
		if err != nil {
			failWith(400, err.Error(), c)
			return
		}

		c.SetStatusCode(iris.StatusOK)
		c.Write(workingString)
	}
}
