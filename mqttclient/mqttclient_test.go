// arkadiko
// https://github.com/topfreegames/arkadiko
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package mqttclient

import (
	"testing"
	"time"

	"github.com/eclipse/paho.mqtt.golang"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestMqttClient(t *testing.T) {
	g := Goblin(t)

	// special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Mqtt Client", func() {
		g.It("It should send message and receive nil", func() {
			connected := false
			var onConnectHandler = func(client mqtt.Client) {
				connected = true
			}
			mc := GetMqttClient("../config/test.yml", onConnectHandler)

			g.Assert(mc.ConfigPath).Equal("../config/test.yml")
			for !connected {
				time.Sleep(100 * time.Millisecond)
			}
			err := mc.SendMessage("test", `{"message": "hello"}`)
			Expect(err).To(BeNil())
		})
	})
}
