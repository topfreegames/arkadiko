// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package remote_test

import (
	"context"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/arkadiko/remote"
	. "github.com/topfreegames/arkadiko/testing"
)

var _ = Describe("RPC Server", func() {
	Describe("Specs", func() {
		Describe("Create Server", func() {
			It("Should create valid server", func() {
				s, err := GetDefaultTestServer()
				Expect(err).NotTo(HaveOccurred())
				Expect(s).NotTo(BeNil())
			})
		})

		Describe("sending messages", func() {
			It("Should send message", func() {
				s, err := GetDefaultTestServer()
				s.Start()

				topic := uuid.NewV4().String()

				cli, err := GetRPCTestClient()
				Expect(cli).NotTo(BeNil())
				Expect(err).NotTo(HaveOccurred())

				expectedMsg := `{ "qwe": 123 }`
				result, err := cli.SendMessage(context.Background(), &remote.Message{
					Topic:    topic,
					Payload:  expectedMsg,
					Retained: true,
				})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())

				client := s.MqttClient
				var msg mqtt.Message
				var onMessageHandler = func(client mqtt.Client, message mqtt.Message) {
					msg = message
				}
				client.MqttClient.Subscribe(topic, 2, onMessageHandler)

				//Have to wait so the goroutine can call our handler
				time.Sleep(50 * time.Millisecond)

				Expect(msg).NotTo(BeNil())
				Expect(msg.Retained()).To(BeTrue())
				Expect(string(msg.Payload())).To(Equal(expectedMsg))
			})
		})
	})
})
