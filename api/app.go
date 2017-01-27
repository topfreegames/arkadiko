// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"fmt"
	"os"
	"strings"
	"time"

	raven "github.com/getsentry/raven-go"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	newrelic "github.com/newrelic/go-agent"
	"github.com/rcrowley/go-metrics"
	"github.com/spf13/viper"
	"github.com/uber-go/zap"

	"github.com/topfreegames/arkadiko/log"
	"github.com/topfreegames/arkadiko/mqttclient"
	"github.com/topfreegames/arkadiko/redisclient"
)

// JSON type
type JSON map[string]interface{}

// App is a struct that represents a arkadiko API Application
type App struct {
	Debug       bool
	Port        int
	Host        string
	ConfigPath  string
	Errors      metrics.EWMA
	App         *echo.Echo
	Config      *viper.Viper
	Logger      zap.Logger
	MqttClient  *mqttclient.MqttClient
	RedisClient *redisclient.RedisClient
	NewRelic    newrelic.Application
}

// GetApp returns a new arkadiko API Application
func GetApp(host string, port int, configPath string, debug bool, logger zap.Logger) (*App, error) {
	app := &App{
		Host:       host,
		Port:       port,
		ConfigPath: configPath,
		Config:     viper.New(),
		Debug:      debug,
		MqttClient: nil,
		Logger:     logger,
	}
	err := app.Configure()
	if err != nil {
		return nil, err
	}
	return app, nil
}

// Configure instantiates the required dependencies for arkadiko Api Application
func (app *App) Configure() error {
	app.setConfigurationDefaults()
	err := app.loadConfiguration()
	if err != nil {
		return err
	}
	app.configureSentry()
	err = app.configureNewRelic()
	if err != nil {
		return err
	}

	err = app.configureApplication()
	if err != nil {
		return err
	}

	return nil
}

func (app *App) configureSentry() {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "configureSentry"),
	)
	sentryURL := app.Config.GetString("sentry.url")
	raven.SetDSN(sentryURL)
	l.Info("Configured sentry successfully.")
}

func (app *App) configureNewRelic() error {
	newRelicKey := app.Config.GetString("newrelic.key")

	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "configureNewRelic"),
	)

	config := newrelic.NewConfig("arkadiko", newRelicKey)
	if newRelicKey == "" {
		l.Info("New Relic is not enabled..")
		config.Enabled = false
	}
	nr, err := newrelic.NewApplication(config)
	if err != nil {
		l.Error("Failed to initialize New Relic.", zap.Error(err))
		return err
	}

	app.NewRelic = nr
	l.Info("Initialized New Relic successfully.")

	return nil
}

func (app *App) setConfigurationDefaults() {
	app.Config.SetDefault("healthcheck.workingText", "WORKING")
	app.Config.SetDefault("redis.password", "")
}

func (app *App) loadConfiguration() error {
	app.Config.SetConfigFile(app.ConfigPath)
	app.Config.SetEnvPrefix("arkadiko")
	app.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	app.Config.AutomaticEnv()

	if err := app.Config.ReadInConfig(); err == nil {
		app.Logger.Info("Loaded config file.", zap.String("configFile", app.Config.ConfigFileUsed()))
	} else {
		return fmt.Errorf("Could not load configuration file from: %s", app.ConfigPath)
	}

	return nil
}

//OnErrorHandler handles panics
func (app *App) OnErrorHandler(err error, stack []byte) {
	app.Logger.Error(
		"Panic occurred.",
		zap.Object("panicText", err),
		zap.String("stack", string(stack)),
	)

	tags := map[string]string{
		"source": "app",
		"type":   "panic",
	}
	raven.CaptureError(err, tags)
}

func (app *App) configureApplication() error {
	l := app.Logger.With(
		zap.String("operation", "configureApplication"),
	)

	app.App = echo.New()
	a := app.App

	_, w, _ := os.Pipe()
	a.Logger.SetOutput(w)

	basicAuthUser := app.Config.GetString("basicauth.username")
	if basicAuthUser != "" {
		basicAuthPass := app.Config.GetString("basicauth.password")

		a.Use(middleware.BasicAuth(func(username, password string) bool {
			return username == basicAuthUser && password == basicAuthPass
		}))
	}

	a.Pre(middleware.RemoveTrailingSlash())
	a.Use(NewLoggerMiddleware(app.Logger).Serve)
	a.Use(NewRecoveryMiddleware(app.OnErrorHandler).Serve)
	a.Use(NewVersionMiddleware().Serve)
	a.Use(NewSentryMiddleware(app).Serve)
	a.Use(NewNewRelicMiddleware(app, app.Logger).Serve)

	// Routes
	// Healthcheck
	a.GET("/healthcheck", HealthCheckHandler(app))

	// MQTT Route
	a.POST("/sendmqtt/*", SendMqttHandler(app))
	a.POST("/authorize_user", AuthorizeUsersHandler(app))
	a.POST("/unauthorize_user", UnauthorizeUsersHandler(app))

	app.Errors = metrics.NewEWMA15()

	redisHost := app.Config.GetString("redis.host")
	redisPort := app.Config.GetInt("redis.port")
	redisPass := app.Config.GetString("redis.password")
	rl := l.With(
		zap.String("host", redisHost),
		zap.Int("port", redisPort),
	)

	log.D(rl, "Connecting to redis...")
	app.RedisClient = redisclient.GetRedisClient(redisHost, redisPort, redisPass, l)
	log.I(rl, "Connected to redis successfully.")

	log.D(l, "Connecting to mqtt...")
	app.MqttClient = mqttclient.GetMqttClient(app.ConfigPath, nil, l)
	log.I(l, "Connected to mqtt successfully.")

	go func() {
		app.Errors.Tick()
		time.Sleep(5 * time.Second)
	}()
	return nil
}

func (app *App) addError() {
	app.Errors.Update(1)
}

// Start starts listening for web requests at specified host and port
func (app *App) Start() error {
	l := app.Logger.With(
		zap.String("source", "app"),
		zap.String("operation", "Start"),
	)

	log.I(l, "App started.", func(cm log.CM) {
		cm.Write(zap.String("host", app.Host), zap.Int("port", app.Port))
	})
	err := app.App.Start(fmt.Sprintf("%s:%d", app.Host, app.Port))
	if err != nil {
		log.E(l, "App failed to start.", func(cm log.CM) {
			cm.Write(
				zap.String("host", app.Host),
				zap.Int("port", app.Port),
				zap.Error(err),
			)
		})
		return err
	}
	return nil
}
