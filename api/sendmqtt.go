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
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

// SendMqttHandler is the handler responsible for sending messages to mqtt
func SendMqttHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.WithFields(log.Fields{
			"handler": "SendMqttHandler",
		})

		var gameID string
		defer func(route, method, status string) {
			if gameID == "" {
				gameID = "unknown" // if request fails early gameID will be unknown at metrics
			}
			app.Metrics.SendMqttRequests.WithLabelValues(route, method, status, gameID).Inc()
		}(c.Path(), c.Request().Method, fmt.Sprintf("%d", c.Response().Status))

		retainedValue := c.QueryParam("retained")
		retained := true
		if retainedValue != "true" {
			retained = false
		}

		source := c.QueryParam("source")

		body := c.Request().Body
		b, err := io.ReadAll(body)
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

		// Default should_moderate to false so messages sent from the server side are not moderated
		if _, exists := msgPayload["should_moderate"]; !exists {
			msgPayload["should_moderate"] = false
		}

		topic := c.ParamValues()[0]
		gameID = getGameID(topic, msgPayload)

		b, err = json.Marshal(msgPayload)
		if err != nil {
			return FailWith(400, err.Error(), c)
		}
		workingString := fmt.Sprintf(`{"topic": "%s", "retained": %t, "payload": %v}`, topic, retained, string(b))

		lg = lg.WithFields(log.Fields{
			"topic":    topic,
			"retained": retained,
			"payload":  string(b),
			"source":   source,
		})

		var mqttLatency time.Duration
		var beforeMqttTime time.Time

		err = WithSegment("mqtt", c, func() error {
			beforeMqttTime = time.Now()
			sendMqttErr := app.MqttClient.PublishMessage(c.Request().Context(), topic, string(b), retained)
			mqttLatency = time.Now().Sub(beforeMqttTime)

			return sendMqttErr
		})

		tags := []string{
			fmt.Sprintf("error:%t", err != nil),
			fmt.Sprintf("retained:%t", retained),
			fmt.Sprintf("game_id:%s", gameID),
		}
		if source != "" {
			tags = append(tags, fmt.Sprintf("requestor:%s", source))
		}

		app.DDStatsD.Timing("mqtt_latency", mqttLatency, tags...)
		app.Metrics.MQTTLatency.WithLabelValues(fmt.Sprintf("%t", err != nil), fmt.Sprintf("%t", retained), gameID).Observe(mqttLatency.Seconds())
		lg = lg.WithField("mqttLatency", mqttLatency.Nanoseconds())
		lg.Debug("sent mqtt message")
		c.Set("mqttLatency", mqttLatency)
		c.Set("requestor", source)
		c.Set("topic", topic)
		c.Set("game_id", gameID)
		c.Set("retained", retained)

		if err != nil {
			lg.WithError(err).Error("failed to send mqtt message")
			return FailWith(500, err.Error(), c)
		}
		return c.String(http.StatusOK, workingString)
	}
}

func getGameID(topic string, msgPayload map[string]interface{}) string {
	for _, k := range []string{"game_id", "gameID"} {
		if v, ok := msgPayload[k]; ok && v != nil {
			return fmt.Sprint(v)
		}
	}
	gameID := ""
	if strings.Contains(topic, "/") {
		gameID = strings.Split(topic, "/")[1]
	} else {
		gameID = topic
	}

	return gameID
}
