// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"github.com/topfreegames/arkadiko/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// SendMqttHandler is the handler responsible for sending messages to mqtt
func SendMqttHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		ctx, rootSpan := app.Otel.StartSpan(c.Request().Context(), "SendMqttHandler")
		defer rootSpan.End()

		lg := app.Logger.WithFields(log.Fields{
			"handler": "SendMqttHandler",
		})

		retainedValue := c.QueryParam("retained")
		retained := true
		if retainedValue != "true" {
			retained = false
		}

		source := c.QueryParam("source")

		var body []byte
		var err error

		err = app.Otel.WithSpan(ctx, "read_request_body", func(ctx context.Context) error {
			b, readErr := io.ReadAll(c.Request().Body)
			if readErr != nil {
				return readErr
			}
			body = b
			return nil
		})
		if err != nil {
			otel.SetSpanStatus(ctx, codes.Error, err.Error())
			return FailWith(http.StatusPreconditionFailed, err.Error(), c)
		}

		var msgPayload map[string]any
		err = app.Otel.WithSpan(ctx, "unmarshal_payload", func(ctx context.Context) error {
			unmarshalErr := json.Unmarshal(body, &msgPayload)
			return unmarshalErr
		})
		if err != nil {
			otel.SetSpanStatus(ctx, codes.Error, err.Error())
			return FailWith(http.StatusBadRequest, err.Error(), c)
		}

		// Default should_moderate to false so messages sent from the server side are not moderated
		if _, exists := msgPayload["should_moderate"]; !exists {
			msgPayload["should_moderate"] = false
		}

		topic := c.ParamValues()[0]
		rootSpan.SetAttributes(
			attribute.String("mqtt.topic", topic),
			attribute.Bool("mqtt.retained", retained),
			attribute.String("request.source", source),
		)

		body, err = json.Marshal(msgPayload)
		if err != nil {
			otel.SetSpanStatus(ctx, codes.Error, err.Error())
			return FailWith(http.StatusBadRequest, err.Error(), c)
		}
		workingString := fmt.Sprintf(`{"topic": "%s", "retained": %t, "payload": %v}`, topic, retained, string(body))

		lg = lg.WithFields(log.Fields{
			"topic":    topic,
			"retained": retained,
			"payload":  string(body),
			"source":   source,
		})

		var mqttLatency time.Duration
		var beforeMqttTime time.Time
		attrs := []attribute.KeyValue{
			attribute.String("mqtt.topic", topic),
			attribute.Bool("mqtt.retained", retained),
			attribute.String("mqtt.message_size", fmt.Sprintf("%d", len(body))),
		}
		if source != "" {
			attrs = append(attrs, attribute.String("request.source", source))
		}

		err = app.Otel.WithSpanAttributes(ctx, "mqtt_publish", attrs, func(ctx context.Context) error {
			beforeMqttTime = time.Now()
			sendMqttErr := app.MqttClient.PublishMessage(ctx, topic, string(body), retained)
			mqttLatency = time.Since(beforeMqttTime)

			otel.AddSpanAttributes(ctx, attribute.Int64("mqtt.latency_ns", mqttLatency.Nanoseconds()))

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
		app.Metrics.MQTTLatency.WithLabelValues(fmt.Sprintf("%t", err != nil), fmt.Sprintf("%t", retained)).Observe(mqttLatency.Seconds())
		lg = lg.WithField("mqttLatency", mqttLatency.Nanoseconds())
		lg.Debug("sent mqtt message")
		c.Set("mqttLatency", mqttLatency)
		c.Set("requestor", source)
		c.Set("retained", retained)

		if err != nil {
			otel.SetSpanStatus(ctx, codes.Error, err.Error())
			lg.WithError(err).Error("failed to send mqtt message")
			return FailWith(http.StatusInternalServerError, err.Error(), c)
		}

		otel.SetSpanStatus(ctx, codes.Ok, "")
		return c.String(http.StatusOK, workingString)
	}
}
