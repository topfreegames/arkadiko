package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	uuid "github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	ehttp "github.com/topfreegames/extensions/http"
)

type HttpClient struct {
	HttpServerUrl string
	user          string
	password      string
	ConfigPath    string
	Config        *viper.Viper
	Logger        log.FieldLogger
	httpClient    *http.Client
}

type MqttPost struct {
	Topic     string `json:"topic"`
	Payload   string `json:"payload"`
	Qos       int    `json:"qos"`
	Retain    bool   `json:"retain"`
	Client_id string `json:"client_id"`
}

var client *HttpClient
var once sync.Once

// GetMqttClient creates the mqttclient and returns it
func GetHttpClient(configPath string, l log.FieldLogger) *HttpClient {
	once.Do(func() {
		client = &HttpClient{
			ConfigPath: configPath,
			Config:     viper.New(),
		}
		client.configure(l)
	})
	return client
}

func (mc *HttpClient) SendMessage(ctx context.Context, topic string, payload string, retainBool bool) error {
	form := &MqttPost{
		Topic:     topic,
		Payload:   payload,
		Retain:    retainBool,
		Qos:       2,
		Client_id: fmt.Sprintf("arkadiko-%s", uuid.NewV4().String()),
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(form)

	req, _ := http.NewRequest(
		"POST",
		mc.HttpServerUrl+"/api/v2/mqtt/publish",
		b,
	)

	if ctx == nil {
		ctx = context.Background()
	}
	req = req.WithContext(ctx)

	req.SetBasicAuth(mc.user, mc.password)
	req.Header.Add("Content-Type", "application/json")
	res, err := mc.httpClient.Do(req)

	if err != nil {
		return err
	}

	io.Copy(ioutil.Discard, res.Body)
	res.Body.Close()

	return nil
}

func (mc *HttpClient) configure(l log.FieldLogger) {
	mc.Logger = l

	mc.setConfigurationDefaults()
	mc.loadConfiguration()
	mc.configureClient()
}

func (mc *HttpClient) setConfigurationDefaults() {
	mc.Config.SetDefault("httpserver.url", "http://localhost:8080")
	mc.Config.SetDefault("httpserver.user", "admin")
	mc.Config.SetDefault("httpserver.pass", "public")
}

func (mc *HttpClient) configureClient() {

	mc.httpClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 1024,
		},
	}
	ehttp.Instrument(mc.httpClient)

	mc.HttpServerUrl = mc.Config.GetString("httpserver.url")
	mc.user = mc.Config.GetString("httpserver.user")
	mc.password = mc.Config.GetString("httpserver.pass")
}

func (mc *HttpClient) loadConfiguration() {
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
