package services

import (
	"strconv"

	"github.com/denistv/wdlogger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func NewMetrics(logger wdlogger.Logger) *Metrics {
	envTempOpts := prometheus.GaugeOpts{
		Name: "environment_temperature",
		Help: "temperature of environment",
	}
	envTempLabels := []string{"id", "name"}
	envTempMetric := promauto.NewGaugeVec(envTempOpts, envTempLabels)

	return &Metrics{
		logger:        logger,
		envTempMetric: envTempMetric,
	}
}

type Metrics struct {
	logger wdlogger.Logger

	envTempMetric *prometheus.GaugeVec
}

// SetEnvironmentTemp Задать текущую температуру
func (m *Metrics) SetEnvironmentTemp(id int64, name string, value float64, target float64) {
	m.logger.Info(
		"setting environment temp",
		wdlogger.NewInt64Field("id", id),
		wdlogger.NewStringField("name", name),
		wdlogger.NewFloat64Field("value", value),
	)

	labels := map[string]string{
		"id":   strconv.FormatInt(id, 10),
		"name": name,
	}
	m.envTempMetric.With(labels).Set(value)
}
