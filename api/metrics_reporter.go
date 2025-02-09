package api

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/dogstatsd"
)

// MetricTypes constants
var MetricTypes = struct {
	APIRequestPath string
}{
	APIRequestPath: "api_request_path",
}

// MetricsReporter interface
type MetricsReporter interface {
	Timing(metric string, value time.Duration, tags ...string) error
	Gauge(metrics string, value float64, tags ...string) error
	Increment(metric string, tags ...string) error
}

// NewMetricsReporter ctor
func NewMetricsReporter(config *viper.Viper) (MetricsReporter, error) {
	return NewDogStatsD(config)
}

// DogStatsD metrics reporter struct
type DogStatsD struct {
	client     dogstatsd.Client
	rate       float64
	tagsPrefix string
}

func loadDefaultConfigsDogStatsD(config *viper.Viper) {
	config.SetDefault("dogstatsd.host", "localhost:8125")
	config.SetDefault("dogstatsd.prefix", "arkadiko.")
	config.SetDefault("dogstatsd.tags_prefix", "arkadiko.")
	config.SetDefault("dogstatsd.rate", "1")
}

// NewDogStatsD ctor
func NewDogStatsD(config *viper.Viper) (*DogStatsD, error) {
	loadDefaultConfigsDogStatsD(config)
	host := config.GetString("dogstatsd.host")
	prefix := config.GetString("dogstatsd.prefix")
	tagsPrefix := config.GetString("dogstatsd.tags_prefix")
	rate, err := strconv.ParseFloat(config.GetString("dogstatsd.rate"), 64)
	if err != nil {
		return nil, err
	}
	c, err := dogstatsd.New(host, prefix)
	if err != nil {
		return nil, err
	}
	return &DogStatsD{
		client:     c,
		rate:       rate,
		tagsPrefix: tagsPrefix,
	}, nil
}

func prefixTags(prefix string, tags ...string) {
	for i, t := range tags {
		tags[i] = fmt.Sprintf("%s%s", prefix, t)
	}
}

// Timing reports time interval taken for something
func (d *DogStatsD) Timing(
	metric string, value time.Duration, tags ...string,
) error {
	prefixTags(d.tagsPrefix, tags...)
	return d.client.Timing(metric, value, tags, d.rate)
}

// Gauge reports a numeric value that can go up or down
func (d *DogStatsD) Gauge(
	metric string, value float64, tags ...string,
) error {
	prefixTags(d.tagsPrefix, tags...)
	return d.client.Gauge(metric, value, tags, d.rate)
}

// Increment reports an increment to some metric
func (d *DogStatsD) Increment(metric string, tags ...string) error {
	prefixTags(d.tagsPrefix, tags...)
	return d.client.Incr(metric, tags, d.rate)
}

type Metrics struct {
	APILatency           *prometheus.HistogramVec
	MQTTLatency          *prometheus.HistogramVec
	DisconnectionCounter *prometheus.CounterVec
}

var metricsOnce sync.Once
var metricsSingleton *Metrics

func NewMetrics() *Metrics {
	metricsOnce.Do(func() {
		metricsSingleton = &Metrics{
			APILatency: promauto.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "arkadiko",
				Name:      "response_time",
				Help:      "API response time",
			}, []string{"route", "method", "status"}),
			MQTTLatency: promauto.NewHistogramVec(prometheus.HistogramOpts{
				Namespace: "arkadiko",
				Name:      "mqtt_latency",
				Help:      "MQTT latency",
			}, []string{"error", "retained"}),
			DisconnectionCounter: promauto.NewCounterVec(prometheus.CounterOpts{
				Namespace: "arkadiko",
				Name:      "mqtt_disconnections",
				Help:      "MQTT disconnections",
			}, []string{"error"}),
		}
	})

	return metricsSingleton
}
