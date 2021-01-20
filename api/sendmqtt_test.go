// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	uuid "github.com/satori/go.uuid"
	"github.com/topfreegames/arkadiko/api"
	. "github.com/topfreegames/arkadiko/testing"
)

var _ = Describe("Send to MQTT Handler", func() {
	Describe("Specs", func() {
		Describe("Regular Message", func() {
			It("Should respond with 200 for a valid message", func() {
				a := GetDefaultTestApp()
				testJSON := map[string]interface{}{
					"message": "hello",
				}
				response := `{"topic": "test", "retained": false, "payload": {"message":"hello"}}`
				status, body := PostJSON(a, "/sendmqtt/test", testJSON)

				Expect(status).To(Equal(http.StatusOK))
				Expect(body).To(Equal(response))
			})

			It("Should respond with 200 for a valid message with hierarchical topic", func() {
				a := GetDefaultTestApp()
				testJSON := map[string]interface{}{
					"message": "hello",
				}
				response := `{"topic": "test/topic", "retained": false, "payload": {"message":"hello"}}`
				url := "/sendmqtt/test/topic"
				status, body := PostJSON(a, url, testJSON)

				Expect(status).To(Equal(http.StatusOK))
				Expect(body).To(Equal(response))
			})

			It("Should respond with 400 if malformed JSON", func() {
				a := GetDefaultTestApp()
				testJSON := `{"message": "hello"}}`
				status, _ := PostBody(a, "/sendmqtt/test/topic", testJSON)

				Expect(status).To(Equal(400))
			})
		})
		Describe("Retained Message", func() {
			It("Should respond with 200 for a valid message", func() {
				a := GetDefaultTestApp()
				client := a.MqttClient
				testJSON := map[string]interface{}{
					"message": "hello",
				}
				topic := uuid.NewV4().String()
				expectedMsg := `{"message":"hello"}`
				response := fmt.Sprintf(
					`{"topic": "%s", "retained": true, "payload": %s}`,
					topic,
					expectedMsg,
				)
				url := fmt.Sprintf("/sendmqtt/%s?retained=true", topic)
				status, body := PostJSON(a, url, testJSON)

				Expect(status).To(Equal(http.StatusOK))
				Expect(body).To(Equal(response))

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

			It("Should respond with 200 for a valid message with hierarchical topic", func() {
				a := GetDefaultTestApp()
				testJSON := map[string]interface{}{
					"message": "hello",
				}
				response := `{"topic": "test/topic", "retained": true, "payload": {"message":"hello"}}`
				url := "/sendmqtt/test/topic?retained=true"
				status, body := PostJSON(a, url, testJSON)

				Expect(status).To(Equal(http.StatusOK))
				Expect(body).To(Equal(response))
			})

			It("Should respond with 400 if malformed JSON", func() {
				a := GetDefaultTestApp()
				testJSON := `{"message": "hello"}}`
				status, _ := PostBody(a, "/sendmqtt/test/topic?retained=true", testJSON)

				Expect(status).To(Equal(400))
			})
		})
	})

	Describe("Perf", func() {
		HTTPMeasure("send to MQTT", func(data map[string]interface{}) {
			testJSON := map[string]interface{}{
				"message": "hello",
			}
			payload, err := json.Marshal(testJSON)
			Expect(err).NotTo(HaveOccurred())
			data["payload"] = payload
		}, func(ts *httptest.Server, data map[string]interface{}) {
			app := data["app"].(*api.App)
			payload := string(data["payload"].([]byte))
			status, body := PostBody(app, "/sendmqtt/test/topic", payload)
			Expect(status).To(Equal(http.StatusOK), string(body))
		}, 0.015)
	})
})
