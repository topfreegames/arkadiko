package httpclient

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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

type mqttPost struct {
	topic     string
	payload   string
	qos       int
	retain    bool
	client_id string
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

func (mc *HttpClient) SendMessage(topic string, payload string, retainBool bool) error {
	retain := "0"
	if retainBool {
		retain = "1"
	}

	form := url.Values{
		"topic":   {topic},
		"message": {payload},
		"retain":  {retain},
		"qos":     {"2"},
	}.Encode()

	req, _ := http.NewRequest(
		"POST",
		mc.HttpServerUrl+"/mqtt/publish",
		strings.NewReader(form),
	)

	req.SetBasicAuth(mc.user, mc.password)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form)))
	_, err := mc.httpClient.Do(req)

	return err
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
	mc.Config.SetDefault("httpserver.pass", "admin")
}

func (mc *HttpClient) configureClient() {
	mc.httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}

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
