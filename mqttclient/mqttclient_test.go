package mqttclient

import (
	"testing"

	"github.com/eclipse/paho.mqtt.golang"

	. "github.com/franela/goblin"
	. "github.com/onsi/gomega"
)

func TestMqttClient(t *testing.T) {
	g := Goblin(t)

	// special hook for gomega
	RegisterFailHandler(func(m string, _ ...int) { g.Fail(m) })

	g.Describe("Mqtt Client creation", func() {
		g.It("It should create a mqtt client", func(done Done) {
			var onConnectHandler = func(client mqtt.Client) {
				done()
			}
			mc := GetMqttClient("../config/test.yml", onConnectHandler)

			g.Assert(mc.ConfigPath).Equal("../config/test.yml")
		})
	})
}
