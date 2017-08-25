// arkadiko
// https://github.com/topfreegames/arkadiko
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package httpclient_test

import (
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/topfreegames/arkadiko/httpclient"
	"github.com/topfreegames/arkadiko/mqttclient"
)

var _ = Describe("HTTP Client", func() {
	Describe("Specs", func() {

		l, _ := test.NewNullLogger()

		logger := l.WithFields(log.Fields{})

		Describe("Specs", func() {
			It("It should send message and receive nil", func() {
				mc := httpclient.GetHttpClient("../config/test.yml", logger)

				Expect(mc.ConfigPath).To(Equal("../config/test.yml"))

				err := mc.SendMessage("test", `{"message": "hello"}`, false)
				Expect(err).To(BeNil())
			})

			It("It should send retained message", func() {
				hc := httpclient.GetHttpClient("../config/test.yml", logger)

				Expect(hc.ConfigPath).To(Equal("../config/test.yml"))

				topic := uuid.NewV4().String()
				expectedMsg := `{"message": "hello"}`

				err := hc.SendMessage(topic, expectedMsg, true)
				Expect(err).NotTo(HaveOccurred())

				mc := mqttclient.GetMqttClient("../config/test.yml", nil, logger)
				var msg mqtt.Message
				var onMessageHandler = func(client mqtt.Client, message mqtt.Message) {
					msg = message
				}
				mc.MqttClient.Subscribe(topic, 2, onMessageHandler)

				//Have to wait so the goroutine can call our handler
				time.Sleep(10 * time.Millisecond)

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
			}, 20)
		})
	})
})
