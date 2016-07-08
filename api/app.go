// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/kataras/iris"
	"github.com/rcrowley/go-metrics"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"

	"github.com/topfreegames/arkadiko/mqttclient"
)

// JSON type
type JSON map[string]interface{}

// App is a struct that represents a arkadiko API Application
type App struct {
	Debug      bool
	Port       int
	Host       string
	ConfigPath string
	Errors     metrics.EWMA
	App        *iris.Framework
	Config     *viper.Viper
	Logger     zap.Logger
	MqttClient *mqttclient.MqttClient
}

// GetApp returns a new arkadiko API Application
func GetApp(host string, port int, configPath string, debug bool) *App {
	app := &App{
		Host:       host,
		Port:       port,
		ConfigPath: configPath,
		Config:     viper.New(),
		Debug:      debug,
		MqttClient: nil,
	}
	app.Configure()
	return app
}

// Configure instantiates the required dependencies for arkadiko Api Application
func (app *App) Configure() {
	app.Logger = zap.NewJSON(zap.WarnLevel)

	app.setConfigurationDefaults()
	app.loadConfiguration()
	app.configureApplication()
}

func (app *App) setConfigurationDefaults() {
	app.Config.SetDefault("healthcheck.workingText", "WORKING")
}

func (app *App) loadConfiguration() {
	app.Config.SetConfigFile(app.ConfigPath)
	app.Config.SetEnvPrefix("arkadiko")
	app.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	app.Config.AutomaticEnv()

	if err := app.Config.ReadInConfig(); err == nil {
		app.Logger.Info("Loaded config file.", zap.String("configFile", app.Config.ConfigFileUsed()))
	} else {
		panic(fmt.Sprintf("Could not load configuration file from: %s", app.ConfigPath))
	}
}

func (app *App) configureApplication() {
	app.App = iris.New()
	a := app.App

	a.Config.DisablePathEscape = false
	a.Get("/healthcheck", HealthCheckHandler(app))

	// MQTT Route
	a.Post("/sendmqtt/:topic", SendMqttHandler(app))

	app.Errors = metrics.NewEWMA15()

	app.MqttClient = mqttclient.GetMqttClient(app.ConfigPath, nil)

	go func() {
		app.Errors.Tick()
		time.Sleep(5 * time.Second)
	}()
}

func (app *App) addError() {
	app.Errors.Update(1)
}

// Start starts listening for web requests at specified host and port
func (app *App) Start() {
	app.App.Listen(fmt.Sprintf("%s:%d", app.Host, app.Port))
}
