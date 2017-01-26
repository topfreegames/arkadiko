// arkadiko
// https://github.com/topfreegames/arkadiko
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package mqttclient_test

import (
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/arkadiko/mqttclient"
	"github.com/uber-go/zap"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MQTT Client", func() {
	Describe("Specs", func() {

		logger := zap.New(
			zap.NewJSONEncoder(),
			zap.FatalLevel,
		).With(
			zap.String("source", "app"),
		)

		Describe("Specs", func() {
			It("It should send message and receive nil", func() {
				connected := false
				var onConnectHandler = func(client mqtt.Client) {
					connected = true
				}
				mc := mqttclient.GetMqttClient("../config/test.yml", onConnectHandler, logger)

				Expect(mc.ConfigPath).To(Equal("../config/test.yml"))

				err := mc.WaitForConnection(100)
				Expect(err).NotTo(HaveOccurred())

				err = mc.SendMessage("test", `{"message": "hello"}`)
				Expect(err).To(BeNil())
			})

			It("It should send retained message", func() {
				mc := mqttclient.GetMqttClient("../config/test.yml", nil, logger)

				Expect(mc.ConfigPath).To(Equal("../config/test.yml"))

				err := mc.WaitForConnection(100)
				Expect(err).NotTo(HaveOccurred())

				topic := uuid.NewV4().String()
				expectedMsg := `{"message": "hello"}`

				err = mc.SendRetainedMessage(topic, expectedMsg)
				Expect(err).NotTo(HaveOccurred())
				//TODO: REALLY need to wait 50ms?
				time.Sleep(50 * time.Millisecond)

				var msg mqtt.Message
				var onMessageHandler = func(client mqtt.Client, message mqtt.Message) {
					msg = message
				}
				mc.MqttClient.Subscribe(topic, 2, onMessageHandler)

				//Have to wait so the goroutine can call our handler
				time.Sleep(1 * time.Millisecond)

				Expect(msg).NotTo(BeNil())
				Expect(msg.Retained()).To(BeTrue())
				Expect(string(msg.Payload())).To(Equal(expectedMsg))
			})
		})

		Describe("Perf", func() {
			Measure("it should send message", func(b Benchmarker) {
				var onConnectHandler = func(client mqtt.Client) {}
				mc := mqttclient.GetMqttClient("../config/test.yml", onConnectHandler, logger)

				runtime := b.Time("runtime", func() {
					err := mc.SendMessage("test", `{"message": "hello"}`)
					Expect(err).NotTo(HaveOccurred())
				})

				Expect(runtime.Seconds()).Should(BeNumerically("<", 0.01), "Sending message shouldn't take too long.")
			}, 200)
		})
	})
})
