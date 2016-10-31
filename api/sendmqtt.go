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
	"net/http"

	"github.com/labstack/echo"
	"github.com/topfreegames/arkadiko/log"
	"github.com/uber-go/zap"
)

// SendMqttHandler is the handler responsible for sending messages to mqtt
func SendMqttHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.With(
			zap.String("handler", "SendMqttHandler"),
		)

		body := c.Request().Body()
		b, err := ioutil.ReadAll(body)
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		if string(b) == "null" {

			return FailWith(400, "Invalid JSON", c)
		}

		var msgPayload map[string]interface{}
		err = WithSegment("payload", c, func() error {
			return json.Unmarshal(b, &msgPayload)
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		topic := c.ParamValues()[0]
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		b, err = json.Marshal(msgPayload)
		if err != nil {
			return FailWith(400, err.Error(), c)
		}
		workingString := fmt.Sprintf(`{"topic": "%s", "payload": %v}`, topic, string(b))
		log.I(lg, "sending message on topic", func(cm log.CM) {
			cm.Write(zap.String("topic", topic), zap.String("payload", string(b)))
		})
		err = app.MqttClient.SendMessage(topic, string(b))
		if err != nil {
			return FailWith(400, err.Error(), c)
		}
		return c.String(http.StatusOK, workingString)
	}
}
