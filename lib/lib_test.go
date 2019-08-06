package lib_test

import (
	"context"

	"github.com/jarcoal/httpmock"
	"github.com/spf13/viper"
	"github.com/topfreegames/arkadiko/lib"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lib", func() {
	var (
		arkadiko *lib.Arkadiko
		config   *viper.Viper
	)

	BeforeSuite(func() {
		config = viper.New()
		httpmock.Activate()
	})

	BeforeEach(func() {
		config.Set("arkadiko.url", "http://arkadiko")
		config.Set("arkadiko.user", "user")
		config.Set("arkadiko.pass", "pass")

		arkadiko = lib.NewArkadiko(config)
		httpmock.Reset()
	})

	Describe("NewArkadiko", func() {
		It("Should start a new instance of Arkadiko Lib", func() {
			arkadiko = lib.NewArkadiko(config)
			Expect(arkadiko).ToNot(BeNil())
		})
	})

	Describe("SendMQTT", func() {
		It("Should call arkadiko API to send MQTT", func() {
			httpmock.RegisterResponder(
				"POST", "http://arkadiko/sendmqtt/topic?retained=false",
				httpmock.NewStringResponder(200, `{
            "payload": {"message": "message"},
            "retained": false,
            "topic": "topic"
          }`,
				),
			)

			payload := map[string]string{
				"message": "message",
			}

			response, err := arkadiko.SendMQTT(context.Background(), "topic", payload, false)

			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
			Expect(response.Payload).To(BeEquivalentTo(`{"message": "message"}`))
			Expect(response.Retained).To(BeFalse())
			Expect(response.Topic).To(Equal("topic"))
		})

		It("Should URL escape topic", func() {
			httpmock.RegisterResponder(
				"POST", "http://arkadiko/sendmqtt/some%2Ftopic?retained=true",
				httpmock.NewStringResponder(200, `{
            "payload": {"message": "message"},
            "retained": true,
            "topic": "some/topic"
          }`,
				),
			)

			payload := map[string]string{
				"message": "message",
			}

			response, err := arkadiko.SendMQTT(context.Background(), "some/topic", payload, true)

			Expect(err).To(BeNil())
			Expect(response).ToNot(BeNil())
			Expect(response.Payload).To(BeEquivalentTo(`{"message": "message"}`))
			Expect(response.Retained).To(BeTrue())
			Expect(response.Topic).To(Equal("some/topic"))
		})

		It("Should return meaningful error", func() {
			httpmock.RegisterResponder(
				"POST", "http://arkadiko/sendmqtt/topic?retained=false",
				httpmock.NewStringResponder(404, "Not Found"),
			)

			payload := map[string]string{
				"message": "message",
			}

			response, err := arkadiko.SendMQTT(context.Background(), "topic", payload, false)

			Expect(err).To(Equal(lib.NewRequestError(404, "Not Found")))
			Expect(response).To(BeNil())
		})
	})

	AfterSuite(func() {
		httpmock.DeactivateAndReset()
	})
})
