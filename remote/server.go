// arkadiko
// https://github.com/topfreegames/arkadiko
//
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>

package remote

import (
	"fmt"
	"net"
	"strings"
	"sync"

	"google.golang.org/grpc"

	raven "github.com/getsentry/raven-go"
	newrelic "github.com/newrelic/go-agent"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/topfreegames/arkadiko/mqttclient"
	context "golang.org/x/net/context"
)

//Server represents the server that replies to RPC messages
type Server struct {
	Debug      bool
	Port       int
	Host       string
	ConfigPath string
	Config     *viper.Viper
	Logger     log.FieldLogger
	MqttClient *mqttclient.MqttClient
	NewRelic   newrelic.Application
	grpcServer *grpc.Server
}

//NewServer returns a new RPC Server
func NewServer(host string, port int, configPath string, debug bool, logger log.FieldLogger) (*Server, error) {
	server := &Server{
		Host:       host,
		Port:       port,
		ConfigPath: configPath,
		Config:     viper.New(),
		Debug:      debug,
		MqttClient: nil,
		Logger:     logger,
	}
	err := server.configure()
	if err != nil {
		return nil, err
	}
	return server, nil
}

func (s *Server) configure() error {
	s.setConfigurationDefaults()
	err := s.loadConfiguration()
	if err != nil {
		return err
	}
	s.configureSentry()
	err = s.configureNewRelic()
	if err != nil {
		return err
	}

	err = s.configureRPC()
	if err != nil {
		return err
	}

	defaultLogger := s.Logger.WithFields(log.Fields{})

	defaultLogger.Debug("Connecting to mqtt...")
	s.MqttClient = mqttclient.GetMqttClient(s.ConfigPath, nil, defaultLogger)
	defaultLogger.Info("Connected to mqtt successfully.")

	return nil
}

func (s *Server) configureSentry() {
	l := s.Logger.WithFields(log.Fields{
		"source":    "rpc",
		"operation": "configureSentry",
	})
	sentryURL := s.Config.GetString("sentry.url")
	raven.SetDSN(sentryURL)
	l.Info("Configured sentry successfully.")
}

func (s *Server) configureNewRelic() error {
	newRelicKey := s.Config.GetString("newrelic.key")

	l := s.Logger.WithFields(log.Fields{
		"source":    "rpc",
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

	s.NewRelic = nr
	l.Info("Initialized New Relic successfully.")

	return nil
}

func (s *Server) setConfigurationDefaults() {
	s.Config.SetDefault("healthcheck.workingText", "WORKING")
}

func (s *Server) loadConfiguration() error {
	s.Config.SetConfigFile(s.ConfigPath)
	s.Config.SetEnvPrefix("arkadiko")
	s.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	s.Config.AutomaticEnv()

	if err := s.Config.ReadInConfig(); err == nil {
		s.Logger.WithField(
			"configFile", s.Config.ConfigFileUsed(),
		).Info("Loaded config file.")
	} else {
		return fmt.Errorf("Could not load configuration file from: %s", s.ConfigPath)
	}

	return nil
}

//OnErrorHandler handles panics
func (s *Server) OnErrorHandler(err error, stack []byte) {
	s.Logger.WithError(err).Error("Panic occurred.")

	tags := map[string]string{
		"source": "s",
		"type":   "panic",
	}
	raven.CaptureError(err, tags)
}

func (s *Server) configureRPC() error {
	l := s.Logger.WithField("operation", "configureRPC")

	opts := []grpc.ServerOption{}

	//TODO: instrument with Jaeger
	s.grpcServer = grpc.NewServer(opts...)
	RegisterMQTTServer(s.grpcServer, s)
	l.Debug("MQTT Server configured properly")

	return nil
}

// Start starts listening for web requests at specified host and port
func (s *Server) Start() error {
	var wg sync.WaitGroup

	l := s.Logger.WithFields(log.Fields{
		"source":    "rpc",
		"operation": "Start",
	})

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port))
	if err != nil {
		l.WithError(err).Error("Failed to start RPC Server")
		return err
	}

	l.WithFields(log.Fields{
		"host": s.Host,
		"port": s.Port,
	}).Info("RPC Server started.")

	wg.Add(1)
	go func(s *Server, lis net.Listener) {
		wg.Done()
		if err := s.grpcServer.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}(s, lis)
	wg.Wait()

	return nil
}

//SendMessage to MQTT Server
func (s *Server) SendMessage(ctx context.Context, message *Message) (*SendMessageResult, error) {
	l := s.Logger.WithFields(log.Fields{
		"source":    "rpc",
		"operation": "Start",
		"Topic":     message.Topic,
	})
	var err error
	if message.Retained {
		l.Debug("Sending retained message.")
		err = s.MqttClient.SendRetainedMessage(ctx, message.Topic, message.Payload)
	} else {
		l.Debug("Sending message.")
		err = s.MqttClient.SendMessage(ctx, message.Topic, message.Payload)
	}
	if err != nil {
		l.WithError(err).Error("Failed to send message to MQTT.")
		return nil, err
	}

	return &SendMessageResult{
		Topic:    message.Topic,
		Retained: message.Retained,
	}, nil
}
