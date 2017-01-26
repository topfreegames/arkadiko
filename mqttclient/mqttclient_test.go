// arkadiko
// https://github.com/topfreegames/arkadiko
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package mqttclient_test

import (
	"time"

	"github.com/eclipse/paho.mqtt.golang"
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
				for !connected {
					time.Sleep(100 * time.Millisecond)
				}
				err := mc.SendMessage("test", `{"message": "hello"}`)
				Expect(err).To(BeNil())
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
