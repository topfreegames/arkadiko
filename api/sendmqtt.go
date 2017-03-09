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
	"time"

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

		retainedValue := c.QueryParam("retained")
		retained := true
		if retainedValue != "true" {
			retained = false
		}

		source := c.QueryParam("source")
		lg = lg.With(zap.String("sender", source))

		body := c.Request().Body
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
		lg = lg.With(zap.String("topic", topic), zap.String("payload", string(b)))
		workingString := fmt.Sprintf(`{"topic": "%s", "retained": %t, "payload": %v}`, topic, retained, string(b))

		var mqttLatency time.Duration
		var beforeMqttTime, afterMqttTime time.Time

		err = WithSegment("mqtt", c, func() error {
			beforeMqttTime = time.Now()
			if retained {
				err = app.MqttClient.SendRetainedMessage(topic, string(b))
			} else {
				err = app.MqttClient.SendMessage(topic, string(b))
			}
			afterMqttTime = time.Now()
			return err
		})

		mqttLatency = afterMqttTime.Sub(beforeMqttTime)
		lg = lg.With(zap.Int64("mqttLatency", mqttLatency.Nanoseconds()))
		log.I(lg, "sent mqtt message")
		c.Set("mqttLatency", mqttLatency)

		if err != nil {
			return FailWith(400, err.Error(), c)
		}
		return c.String(http.StatusOK, workingString)
	}
}
