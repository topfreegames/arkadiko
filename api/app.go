// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	raven "github.com/getsentry/raven-go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	newrelic "github.com/newrelic/go-agent"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"github.com/topfreegames/arkadiko/httpclient"
	"github.com/topfreegames/arkadiko/mqttclient"
	"github.com/topfreegames/arkadiko/otel"
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
	App        *echo.Echo
	Config     *viper.Viper
	Logger     log.FieldLogger
	MqttClient *mqttclient.MqttClient
	HttpClient *httpclient.HttpClient
	NewRelic   newrelic.Application
	DDStatsD   *DogStatsD
	Metrics    *Metrics
	Otel       *otel.OtelImpl
	ctx        context.Context
}

// GetApp returns a new arkadiko API Application
func GetApp(host string, port int, configPath string, debug bool, logger log.FieldLogger) (*App, error) {
	app := &App{
		Host:       host,
		Port:       port,
		ConfigPath: configPath,
		Config:     viper.New(),
		Debug:      debug,
		MqttClient: nil,
		HttpClient: nil,
		Logger:     logger,
		ctx:        context.Background(),
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

	err = app.configureStatsD()
	if err != nil {
		return err
	}

	err = app.configureApplication()
	if err != nil {
		return err
	}

	app.configureOtel()

	return nil
}

func (app *App) configureSentry() {
	l := app.Logger.WithFields(log.Fields{
		"source":    "app",
		"operation": "configureSentry",
	})
	sentryURL := app.Config.GetString("sentry.url")
	raven.SetDSN(sentryURL)
	l.Info("Configured sentry successfully.")
}

func (app *App) configureStatsD() error {
	l := app.Logger.WithFields(log.Fields{
		"source":    "app",
		"operation": "configureStatsD",
	})

	ddstatsd, err := NewDogStatsD(app.Config)
	if err != nil {
		l.WithError(err).Error("Failed to initialize DogStatsD.")
		return err
	}
	app.DDStatsD = ddstatsd
	l.Info("Initialized DogStatsD successfully.")

	return nil
}

func (app *App) configureNewRelic() error {
	newRelicKey := app.Config.GetString("newrelic.key")

	l := app.Logger.WithFields(log.Fields{
		"source":    "app",
		"operation": "configureNewRelic",
	})

	config := newrelic.NewConfig("arkadiko", newRelicKey)
	if newRelicKey == "" {
		l.Info("New Relic is not enabled..")
		config.Enabled = false
	}
	nr, err := newrelic.NewApplication(config)
	if err != nil {
		l.WithError(err).Error("Failed to initialize New Relic.")
		return err
	}

	app.NewRelic = nr
	l.Info("Initialized New Relic successfully.")

	return nil
}

func (app *App) configureOtel() {
	l := app.Logger.WithFields(log.Fields{
		"source":    "app",
		"operation": "configureOtel",
	})

	if app.Config.GetBool("otel.disabled") {
		l.Info("OpenTelemetry tracing is disabled.")
		return
	}

	serviceName := app.Config.GetString("otel.service.name")
	if serviceName == "" {
		serviceName = "arkadiko"
	}
	host := app.Config.GetString("otel.collector.host")
	port := app.Config.GetString("otel.collector.port")
	samplingProbability := app.Config.GetFloat64("otel.samplingProbability")
	otelImpl, err := otel.NewOtelImpl(app.ctx, serviceName, host, port, samplingProbability)
	if err != nil {
		l.WithError(err).Error("Failed to initialize Open Telemetry.")
		return
	}
	app.Otel = otelImpl

	app.App.Use(otelecho.Middleware(serviceName, otelecho.WithSkipper(func(c echo.Context) bool {
		return strings.Contains(c.Path(), "/healthcheck")
	})))

	l.WithFields(log.Fields{
		"serviceName":         serviceName,
		"samplingProbability": samplingProbability,
	}).Info("OpenTelemetry tracing initialized successfully.")
}

func (app *App) setConfigurationDefaults() {
	app.Config.SetDefault("healthcheck.workingText", "WORKING")
	app.Config.SetDefault("httpserver.metricsServer", 9090)
}

func (app *App) loadConfiguration() error {
	app.Config.SetConfigFile(app.ConfigPath)
	app.Config.SetEnvPrefix("arkadiko")
	app.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	app.Config.AutomaticEnv()

	if err := app.Config.ReadInConfig(); err == nil {
		app.Logger.WithFields(log.Fields{
			"configFile": app.Config.ConfigFileUsed(),
		}).Info("Loaded config file.")
	} else {
		return fmt.Errorf("could not load configuration file from: %s", app.ConfigPath)
	}

	return nil
}

// OnErrorHandler handles panics
func (app *App) OnErrorHandler(err error, stack []byte) {
	app.Logger.WithError(err).Error("Panic occurred.")

	tags := map[string]string{
		"source": "app",
		"type":   "panic",
	}
	raven.CaptureError(err, tags)
}

func (app *App) configureApplication() error {
	l := app.Logger.WithFields(log.Fields{
		"operation": "configureApplication",
	})

	app.App = echo.New()

	a := app.App

	_, w, _ := os.Pipe()
	a.Logger.SetOutput(w)

	basicAuthUser := app.Config.GetString("basicauth.username")
	if basicAuthUser != "" {
		basicAuthPass := app.Config.GetString("basicauth.password")
		a.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
			if username == basicAuthUser && password == basicAuthPass {
				return true, nil
			}
			return false, nil
		}))

	}

	app.Metrics = NewMetrics()

	a.Pre(middleware.RemoveTrailingSlash())
	a.Use(NewLoggerMiddleware(app.Logger).Serve)
	a.Use(NewRecoveryMiddleware(app.OnErrorHandler).Serve)
	a.Use(NewResponseTimeMetricsMiddleware(app.DDStatsD, app.Metrics.APILatency).Serve)
	a.Use(NewVersionMiddleware().Serve)
	a.Use(NewSentryMiddleware(app).Serve)
	a.Use(NewNewRelicMiddleware(app, app.Logger).Serve)

	// Routes
	// Healthcheck
	a.GET("/healthcheck", HealthCheckHandler(app))

	// MQTT Route
	a.POST("/sendmqtt/*", SendMqttHandler(app))

	app.Errors = metrics.NewEWMA15()

	l.Debug("Connecting to mqtt...")
	onConnectionLost := func(client mqtt.Client, err error) {
		l.WithError(err).Error("Connection to MQTT server lost")
		app.Metrics.DisconnectionCounter.WithLabelValues(err.Error()).Inc()
	}
	app.MqttClient = mqttclient.GetMqttClient(app.ConfigPath, nil, onConnectionLost, nil, l)
	l.Info("Connected to mqtt successfully.")

	app.HttpClient = httpclient.GetHttpClient(app.ConfigPath, l)

	go func() {
		app.Errors.Tick()
		time.Sleep(5 * time.Second)
	}()
	return nil
}

// Start starts listening for web requests at specified host and port
func (app *App) Start() error {
	l := app.Logger.WithFields(log.Fields{
		"source":    "app",
		"operation": "Start",
	})

	// start metrics server
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{}))
	metricsServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.Config.GetInt("httpserver.metricsServer")),
		Handler: mux,
	}

	go func() {
		err := metricsServer.ListenAndServe()
		if err != nil {
			l.WithError(err).Error("Failed to start metrics server.")
		}
	}()

	l.WithFields(log.Fields{
		"host": app.Host,
		"port": app.Port,
	}).Infof("App starting on %s:%d", app.Host, app.Port)

	app.App.Server.BaseContext = func(net.Listener) context.Context {
		return app.ctx
	}

	shutdown, cancel := signal.NotifyContext(app.ctx, os.Interrupt)
	defer cancel()

	go func() {
		err := app.App.Start(fmt.Sprintf("%s:%d", app.Host, app.Port))
		if err != nil {
			l.WithError(err).Error("App failed to start.")
		}
	}()

	<-shutdown.Done()

	err := app.App.Shutdown(app.ctx)
	if err != nil {
		l.WithError(err).Error("App failed to stop.")
	}

	if app.Otel != nil {
		return app.Otel.CloserFunc(app.ctx)
	}
	return nil
}
