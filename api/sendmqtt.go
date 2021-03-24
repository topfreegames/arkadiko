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
	log "github.com/sirupsen/logrus"
)

// SendMqttHandler is the handler responsible for sending messages to mqtt
func SendMqttHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.WithFields(log.Fields{
			"handler": "SendMqttHandler",
		})

		retainedValue := c.QueryParam("retained")
		retained := true
		if retainedValue != "true" {
			retained = false
		}

		source := c.QueryParam("source")

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
		workingString := fmt.Sprintf(`{"topic": "%s", "retained": %t, "payload": %v}`, topic, retained, string(b))

		var mqttLatency time.Duration
		var beforeMqttTime, afterMqttTime time.Time

		err = WithSegment("mqtt", c, func() error {
			beforeMqttTime = time.Now()
			// Bringing back the MQTT client due to issues with HTTP layer in the broker
			if retained {
				err = app.MqttClient.SendRetainedMessage(c.Request().Context(), topic, string(b))
			} else {
				err = app.MqttClient.SendMessage(c.Request().Context(), topic, string(b))
			}
			afterMqttTime = time.Now()
			return err
		})

		tags := []string{
			fmt.Sprintf("error:%t", err != nil),
			fmt.Sprintf("retained:%t", retained),
		}
		if source != "" {
			tags = append(tags, fmt.Sprintf("requestor:%s", source))
		}
		mqttLatency = afterMqttTime.Sub(beforeMqttTime)
		app.DDStatsD.Timing("mqtt_latency", mqttLatency, tags...)
		lg = lg.WithField("mqttLatency", mqttLatency.Nanoseconds())
		lg.Debug("sent mqtt message")
		c.Set("mqttLatency", mqttLatency)
		c.Set("requestor", source)
		c.Set("retained", retained)

		if err != nil {
			return FailWith(500, err.Error(), c)
		}
		return c.String(http.StatusOK, workingString)
	}
}
