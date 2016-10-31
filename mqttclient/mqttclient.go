// arkadiko
// https://github.com/topfreegames/arkadiko
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package mqttclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"
)

// MqttClient contains the data needed to connect the client
type MqttClient struct {
	MqttServerHost string
	MqttServerPort int
	ConfigPath     string
	Config         *viper.Viper
	Logger         zap.Logger
	MqttClient     mqtt.Client
}

var client *MqttClient
var once sync.Once

// GetMqttClient creates the mqttclient and returns it
func GetMqttClient(configPath string, onConnectHandler mqtt.OnConnectHandler, l zap.Logger) *MqttClient {
	once.Do(func() {
		client = &MqttClient{
			ConfigPath: configPath,
			Config:     viper.New(),
		}
		client.configure(l)
		client.start(onConnectHandler)
	})
	return client
}

// SendMessage sends the message with the given payload to topic
func (mc *MqttClient) SendMessage(topic string, message string) error {
	fmt.Println("CAMILA CAMILA CAMILA a")
	if token := mc.MqttClient.Publish(topic, 2, false, message); token.Wait() && token.Error() != nil {
		fmt.Println("CAMILA CAMILA CAMILA b")
		mc.Logger.Error(fmt.Sprintf("%v", token.Error()))
		return token.Error()
	}
	fmt.Println("CAMILA CAMILA CAMILA c")
	return nil
}

func (mc *MqttClient) configure(l zap.Logger) {
	mc.Logger = l

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
}

func (mc *MqttClient) loadConfiguration() {
	mc.Config.SetConfigFile(mc.ConfigPath)
	mc.Config.SetEnvPrefix("arkadiko")
	mc.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	mc.Config.AutomaticEnv()

	if err := mc.Config.ReadInConfig(); err == nil {
		mc.Logger.Info("Loaded config file.", zap.String("configFile", mc.Config.ConfigFileUsed()))
	} else {
		panic(fmt.Sprintf("Could not load configuration file from: %s", mc.ConfigPath))
	}
}

func (mc *MqttClient) configureClient() {
	mc.MqttServerHost = mc.Config.GetString("mqttserver.host")
	mc.MqttServerPort = mc.Config.GetInt("mqttserver.port")
}

func (mc *MqttClient) start(onConnectHandler mqtt.OnConnectHandler) {
	mc.Logger.Info("Initializing mqtt client", zap.String("host", mc.MqttServerHost),
		zap.Int("port", mc.MqttServerPort), zap.String("ca_cert_file", mc.Config.GetString("mqttserver.ca_cert_file")))

	useTls := mc.Config.GetBool("mqttserver.usetls")

	protocol := "tcp"
	if useTls {
		protocol = "ssl"
	}
	opts := mqtt.NewClientOptions().AddBroker(fmt.Sprintf("%s://%s:%d", protocol, mc.MqttServerHost, mc.MqttServerPort)).SetClientID("arkadiko")

	if useTls {
		mc.Logger.Info("using tls", zap.Bool("insecure_skip_verify", mc.Config.GetBool("mqttserver.insecure_tls")))
		certpool := x509.NewCertPool()
		if mc.Config.GetString("mqttserver.ca_cert_file") != "" {
			pemCerts, err := ioutil.ReadFile(mc.Config.GetString("mqttserver.ca_cert_file"))
			if err == nil {
				certpool.AppendCertsFromPEM(pemCerts)
			} else {
				mc.Logger.Error(err.Error())
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
	mc.MqttClient = mqtt.NewClient(opts)

	c := mc.MqttClient

	if token := c.Connect(); token.Wait() && token.Error() != nil {
		mc.Logger.Fatal("Error connecting to mqttserver", zap.Error(token.Error()))
	}

	mc.Logger.Info(fmt.Sprintf("Successfully connected to mqtt server at %s:%d!",
		mc.MqttServerHost, mc.MqttServerPort))
}
