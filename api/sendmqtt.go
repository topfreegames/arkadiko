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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/labstack/echo"
	log "github.com/sirupsen/logrus"
)

var sendMqttLatencyMetric *prometheus.HistogramVec

func init() {
	sendMqttLatencyMetric = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "arkadiko",
		Name:      "send_mqtt_latency",
		Help:      "The latency of sending a message to mqtt",
	},
		[]string{"error", "retained"},
	)
}

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

		// Default should_moderate to false so messages sent from the server side are not moderated
		if _, exists := msgPayload["should_moderate"]; !exists {
			msgPayload["should_moderate"] = false
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
		}
		if source != "" {
			tags = append(tags, fmt.Sprintf("requestor:%s", source))
		}

		app.DDStatsD.Timing("mqtt_latency", mqttLatency, tags...)
		sendMqttLatencyMetric.WithLabelValues(fmt.Sprintf("%t", err != nil), fmt.Sprintf("%t", retained)).Observe(mqttLatency.Seconds())
		lg = lg.WithField("mqttLatency", mqttLatency.Nanoseconds())
		lg.Debug("sent mqtt message")
		c.Set("mqttLatency", mqttLatency)
		c.Set("requestor", source)
		c.Set("retained", retained)

		if err != nil {
			lg.WithError(err).Error("failed to send mqtt message")
			return FailWith(500, err.Error(), c)
		}
		return c.String(http.StatusOK, workingString)
	}
}
