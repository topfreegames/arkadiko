// arkadiko
// https://github.com/topfreegames/arkadiko
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package mqttclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	emqtt "github.com/topfreegames/extensions/mqtt"
	"github.com/topfreegames/extensions/mqtt/interfaces"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
)

// MqttClient contains the data needed to connect the client
type MqttClient struct {
	MqttServerHost string
	MqttServerPort int
	Timeout        time.Duration
	ConfigPath     string
	Config         *viper.Viper
	Logger         log.FieldLogger
	MqttClient     interfaces.Client
	maxRetries     int
}

var client *MqttClient
var once sync.Once

// GetMqttClient creates the mqttclient and returns it
func GetMqttClient(
	configPath string,
	onConnectHandler mqtt.OnConnectHandler,
	onConnectionLost mqtt.ConnectionLostHandler,
	onReconnecting mqtt.ReconnectHandler,
	l log.FieldLogger,
) *MqttClient {
	defaultOnConnectHandler := func(client mqtt.Client) {
		l.Info("Connected to MQTT server")
	}

	defaultOnConnectionLostHandler := func(client mqtt.Client, err error) {
		l.WithError(err).Error("Connection to MQTT server lost")
	}

	defaultOnReconnectingHandler := func(client mqtt.Client, options *mqtt.ClientOptions) {
		l.Info("Reconnecting to MQTT server")
	}

	if onConnectHandler == nil {
		onConnectHandler = defaultOnConnectHandler
	}

	if onConnectionLost == nil {
		onConnectionLost = defaultOnConnectionLostHandler
	}

	if onReconnecting == nil {
		onReconnecting = defaultOnReconnectingHandler
	}

	once.Do(func() {
		client = &MqttClient{
			ConfigPath: configPath,
			Config:     viper.New(),
		}
		client.configure(l)
		client.start(onConnectHandler, onConnectionLost, onReconnecting)
	})
	return client
}

// SendMessage sends the message with the given payload to topic
func (mc *MqttClient) SendMessage(ctx context.Context, topic string, message string) error {
	return mc.PublishMessage(ctx, topic, message, false)
}

// SendRetainedMessage sends the message with the given payload to topic
func (mc *MqttClient) SendRetainedMessage(ctx context.Context, topic string, message string) error {
	return mc.PublishMessage(ctx, topic, message, true)
}

func (mc *MqttClient) PublishMessage(ctx context.Context, topic string, message string, retained bool) error {
	l := mc.Logger.WithFields(
		log.Fields{
			"method":   "PublishMessage",
			"topic":    topic,
			"message":  message,
			"retained": retained,
		},
	)

	l.Debug("Publishing message to mqtt")

	for i := 0; i < 3; i++ { // Retry up to 3 times
		token := mc.MqttClient.WithContext(ctx).Publish(topic, 2, retained, message)
		if token.WaitTimeout(mc.Timeout) {
			l.Debug("message published to mqtt")
			break
		}

		if token.Error() != nil {
			l.WithError(token.Error()).Error("Error publishing message to mqtt")
		} else {
			l.Debug("message timed out")
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// WaitForConnection to mqtt server
func (mc *MqttClient) WaitForConnection(timeout int) error {
	start := time.Now()
	timedOut := func() bool {
		return time.Now().Sub(start) > time.Duration(timeout)*time.Millisecond
	}
	for !mc.MqttClient.IsConnected() || timedOut() {
		time.Sleep(1 * time.Millisecond)
	}

	if timedOut() {
		return fmt.Errorf("Connection to MQTT timed out")
	}
	return nil
}

func (mc *MqttClient) configure(l log.FieldLogger) {
	mc.Logger = l.WithField("source", "MqttClient")

	mc.setConfigurationDefaults()
	mc.loadConfiguration()
	mc.configureClient()
}

func (mc *MqttClient) setConfigurationDefaults() {
	mc.Config.SetDefault("mqttserver.host", "localhost")
	mc.Config.SetDefault("mqttserver.port", 1883)
	mc.Config.SetDefault("mqttserver.user", "admin")
	mc.Config.SetDefault("mqttserver.pass", "admin")
	mc.Config.SetDefault("mqttserver.ca_cert_file", "")
	mc.Config.SetDefault("mqttserver.timeout", 500*time.Millisecond)
}

func (mc *MqttClient) loadConfiguration() {
	mc.Config.SetConfigFile(mc.ConfigPath)
	mc.Config.SetEnvPrefix("arkadiko")
	mc.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	mc.Config.AutomaticEnv()

	if err := mc.Config.ReadInConfig(); err == nil {
		mc.Logger.WithFields(log.Fields{
			"configFile": mc.Config.ConfigFileUsed(),
		}).Info("Loaded config file.")
	} else {
		panic(fmt.Sprintf("Could not load configuration file from: %s", mc.ConfigPath))
	}
}

func (mc *MqttClient) configureClient() {
	mc.MqttServerHost = mc.Config.GetString("mqttserver.host")
	mc.MqttServerPort = mc.Config.GetInt("mqttserver.port")
	mc.Timeout = mc.Config.GetDuration("mqttserver.timeout")
}

func (mc *MqttClient) start(onConnectHandler mqtt.OnConnectHandler, onConnectionLost mqtt.ConnectionLostHandler, onReconnecting mqtt.ReconnectHandler) {
	mc.Logger.WithFields(log.Fields{
		"host":         mc.MqttServerHost,
		"port":         mc.MqttServerPort,
		"ca_cert_file": mc.Config.GetString("mqttserver.ca_cert_file"),
	}).Info("Initializing mqtt client")

	useTLS := mc.Config.GetBool("mqttserver.usetls")

	protocol := "tcp"
	if useTLS {
		protocol = "ssl"
	}

	clientID := fmt.Sprintf("arkadiko-%s", uuid.NewV4().String())
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("%s://%s:%d", protocol, mc.MqttServerHost, mc.MqttServerPort)).SetClientID(clientID)

	if useTLS {
		mc.Logger.WithFields(log.Fields{
			"insecure_skip_verify": mc.Config.GetBool("mqttserver.insecure_tls"),
		}).Info("using tls")
		certpool := x509.NewCertPool()
		if mc.Config.GetString("mqttserver.ca_cert_file") != "" {
			pemCerts, err := ioutil.ReadFile(mc.Config.GetString("mqttserver.ca_cert_file"))
			if err == nil {
				certpool.AppendCertsFromPEM(pemCerts)
			} else {
				mc.Logger.WithError(err).Error()
			}
		}
		tlsConfig := &tls.Config{InsecureSkipVerify: mc.Config.GetBool("mqttserver.insecure_tls"), ClientAuth: tls.NoClientCert, RootCAs: certpool}
		opts.SetTLSConfig(tlsConfig)
	}
	opts.SetUsername(mc.Config.GetString("mqttserver.user"))
	opts.SetPassword(mc.Config.GetString("mqttserver.pass"))
	opts.SetKeepAlive(3 * time.Second)
	opts.SetPingTimeout(5 * time.Second)
	opts.SetMaxReconnectInterval(30 * time.Second)
	opts.SetOnConnectHandler(onConnectHandler)
	opts.SetAutoReconnect(true)
	opts.SetConnectionLostHandler(onConnectionLost)
	opts.SetReconnectingHandler(onReconnecting)

	c := emqtt.NewClient(opts)
	mc.MqttClient = c

	if token := c.Connect(); token.Wait() && token.Error() != nil {
		mc.Logger.WithError(token.Error()).Info("Error connecting to mqttserver")
	}

	mc.Logger.Info(fmt.Sprintf("Successfully connected to mqtt server at %s:%d!",
		mc.MqttServerHost, mc.MqttServerPort))
}
