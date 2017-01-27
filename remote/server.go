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

	"google.golang.org/grpc"

	raven "github.com/getsentry/raven-go"
	newrelic "github.com/newrelic/go-agent"
	"github.com/spf13/viper"
	"github.com/topfreegames/arkadiko/log"
	"github.com/topfreegames/arkadiko/mqttclient"
	"github.com/topfreegames/arkadiko/redisclient"
	"github.com/uber-go/zap"
	context "golang.org/x/net/context"
)

//Server represents the server that replies to RPC messages
type Server struct {
	Debug       bool
	Port        int
	Host        string
	ConfigPath  string
	Config      *viper.Viper
	Logger      zap.Logger
	MqttClient  *mqttclient.MqttClient
	RedisClient *redisclient.RedisClient
	NewRelic    newrelic.Application
	grpcServer  *grpc.Server
}

//NewServer returns a new RPC Server
func NewServer(host string, port int, configPath string, debug bool, logger zap.Logger) (*Server, error) {
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

	redisHost := s.Config.GetString("redis.host")
	redisPort := s.Config.GetInt("redis.port")
	redisPass := s.Config.GetString("redis.password")
	rl := s.Logger.With(
		zap.String("host", redisHost),
		zap.Int("port", redisPort),
	)

	log.D(rl, "Connecting to redis...")
	s.RedisClient = redisclient.GetRedisClient(redisHost, redisPort, redisPass, s.Logger)
	log.I(rl, "Connected to redis successfully.")

	log.D(s.Logger, "Connecting to mqtt...")
	s.MqttClient = mqttclient.GetMqttClient(s.ConfigPath, nil, s.Logger)
	log.I(s.Logger, "Connected to mqtt successfully.")

	return nil
}

func (s *Server) configureSentry() {
	l := s.Logger.With(
		zap.String("source", "rpc"),
		zap.String("operation", "configureSentry"),
	)
	sentryURL := s.Config.GetString("sentry.url")
	raven.SetDSN(sentryURL)
	l.Info("Configured sentry successfully.")
}

func (s *Server) configureNewRelic() error {
	newRelicKey := s.Config.GetString("newrelic.key")

	l := s.Logger.With(
		zap.String("source", "rpc"),
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

	s.NewRelic = nr
	l.Info("Initialized New Relic successfully.")

	return nil
}

func (s *Server) setConfigurationDefaults() {
	s.Config.SetDefault("healthcheck.workingText", "WORKING")
	s.Config.SetDefault("redis.password", "")
}

func (s *Server) loadConfiguration() error {
	s.Config.SetConfigFile(s.ConfigPath)
	s.Config.SetEnvPrefix("arkadiko")
	s.Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	s.Config.AutomaticEnv()

	if err := s.Config.ReadInConfig(); err == nil {
		s.Logger.Info("Loaded config file.", zap.String("configFile", s.Config.ConfigFileUsed()))
	} else {
		return fmt.Errorf("Could not load configuration file from: %s", s.ConfigPath)
	}

	return nil
}

//OnErrorHandler handles panics
func (s *Server) OnErrorHandler(err error, stack []byte) {
	s.Logger.Error(
		"Panic occurred.",
		zap.Object("panicText", err),
		zap.String("stack", string(stack)),
	)

	tags := map[string]string{
		"source": "s",
		"type":   "panic",
	}
	raven.CaptureError(err, tags)
}

func (s *Server) configureRPC() error {
	l := s.Logger.With(
		zap.String("operation", "configureRPC"),
	)

	opts := []grpc.ServerOption{}

	s.grpcServer = grpc.NewServer(opts...)
	RegisterMQTTServer(s.grpcServer, s)
	l.Debug("MQTT Server configured properly")

	return nil
}

// Start starts listening for web requests at specified host and port
func (s *Server) Start() error {
	l := s.Logger.With(
		zap.String("source", "rpc"),
		zap.String("operation", "Start"),
	)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port))
	if err != nil {
		l.Error("Failed to start RPC Server", zap.Error(err))
		return err
	}

	log.I(l, "RPC Server started.", func(cm log.CM) {
		cm.Write(zap.String("host", s.Host), zap.Int("port", s.Port))
	})
	s.grpcServer.Serve(lis)
	return nil
}

//SendMessage to MQTT Server
func (s *Server) SendMessage(ctx context.Context, message *Message) (*SendMessageResult, error) {
	l := s.Logger.With(
		zap.String("source", "rpc"),
		zap.String("operation", "Start"),
	)

	var err error
	if message.Retained {
		l.Debug("Sending retained message.", zap.String("Topic", message.Topic))
		err = s.MqttClient.SendRetainedMessage(message.Topic, message.Payload)
	} else {
		l.Debug("Sending message.", zap.String("Topic", message.Topic))
		err = s.MqttClient.SendMessage(message.Topic, message.Payload)
	}
	if err != nil {
		l.Error("Failed to send message to MQTT.", zap.Error(err))
		return nil, err
	}

	return &SendMessageResult{
		Topic:    message.Topic,
		Retained: message.Retained,
	}, nil
}
