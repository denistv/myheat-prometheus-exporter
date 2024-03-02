package services

import (
	"strconv"

	"github.com/denistv/wdlogger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricNameEnvTempCurrent = "myheat_env_temp_current"
	metricNameEnvTempTarget  = "myheat_env_temp_target"
	metricNameEnvHeatDemand  = "myheat_env_heat_demand"

	metricNameDeviceWeatherTemp = "myheat_dev_weather_temp"
)

func NewMetrics(logger wdlogger.Logger) *Metrics {
	// Environment current temperature
	envTempCurrOpts := prometheus.GaugeOpts{
		Name: metricNameEnvTempCurrent,
		Help: "Температура помещения в данный момент",
	}
	envTempCurrLabels := []string{"id", "name"}
	envTempCurrMetric := promauto.NewGaugeVec(envTempCurrOpts, envTempCurrLabels)

	// Environment target temperature
	envTempTargetOpts := prometheus.GaugeOpts{
		Name: metricNameEnvTempTarget,
		Help: "Целевая температура помещения",
	}
	envTempTargetLabels := []string{"id", "name"}
	envTempTargetMetric := promauto.NewGaugeVec(envTempTargetOpts, envTempTargetLabels)

	// Env heat demand
	envHeatDemandOpts := prometheus.GaugeOpts{
		Name: metricNameEnvHeatDemand,
		Help: "Запрошен нагрев для достижения целевой температуры",
	}
	envHeatDemandLabels := []string{"id", "name"}
	envHeatDemandMetric := promauto.NewGaugeVec(envHeatDemandOpts, envHeatDemandLabels)

	// Device weather temperature
	deviceWeatherTempOpts := prometheus.GaugeOpts{
		Name: metricNameDeviceWeatherTemp,
		Help: "Температура на улице",
	}
	deviceWeatherTempLabels := []string{"id", "name", "city"}
	deviceWeatherTempMetric := promauto.NewGaugeVec(deviceWeatherTempOpts, deviceWeatherTempLabels)

	return &Metrics{
		logger: logger,

		envTempCurrMetric:       envTempCurrMetric,
		envTempTargetMetric:     envTempTargetMetric,
		envHeatDemandMetric:     envHeatDemandMetric,
		deviceWeatherTempMetric: deviceWeatherTempMetric,
	}
}

type Metrics struct {
	logger wdlogger.Logger

	envTempCurrMetric       *prometheus.GaugeVec
	envTempTargetMetric     *prometheus.GaugeVec
	envHeatDemandMetric     *prometheus.GaugeVec
	deviceWeatherTempMetric *prometheus.GaugeVec
}

func envLabels(id int64, name string) map[string]string {
	return map[string]string{
		"id":   strconv.FormatInt(id, 10),
		"name": name,
	}
}

func (m *Metrics) SetEnvironmentTempCurrent(id int64, name string, value float64) {
	m.logger.Info(
		"set value",
		wdlogger.NewStringField("metric_name", metricNameEnvTempCurrent),
		wdlogger.NewInt64Field("id", id),
		wdlogger.NewStringField("name", name),
		wdlogger.NewFloat64Field("value", value),
	)

	labels := envLabels(id, name)
	m.envTempCurrMetric.With(labels).Set(value)
}

func (m *Metrics) SetEnvironmentTempTarget(id int64, name string, value float64) {
	m.logger.Info(
		"set value",
		wdlogger.NewStringField("metric_name", metricNameEnvTempTarget),
		wdlogger.NewInt64Field("id", id),
		wdlogger.NewStringField("name", name),
		wdlogger.NewFloat64Field("value", value),
	)

	labels := envLabels(id, name)
	m.envTempTargetMetric.With(labels).Set(value)
}

func bTf(v bool) float64 {
	if v {
		return 1
	}
	return 0
}

func (m *Metrics) SetEnvironmentHeatDemand(id int64, name string, value bool) {
	m.logger.Info(
		"set value",
		wdlogger.NewStringField("metric_name", metricNameEnvHeatDemand),
		wdlogger.NewInt64Field("id", id),
		wdlogger.NewStringField("name", name),
		wdlogger.NewBoolField("value", value),
	)

	labels := envLabels(id, name)
	m.envHeatDemandMetric.With(labels).Set(bTf(value))
}

func (m *Metrics) SetDeviceWeatherTemp(id int64, name string, city string, value float64) {
	m.logger.Info(
		"set value",
		wdlogger.NewStringField("metric_name", metricNameDeviceWeatherTemp),
		wdlogger.NewInt64Field("id", id),
		wdlogger.NewStringField("name", name),
		wdlogger.NewStringField("city", city),
		wdlogger.NewFloat64Field("value", value),
	)

	labels := map[string]string{
		"id":   strconv.FormatInt(id, 10),
		"name": name,
		"city": city,
	}
	m.deviceWeatherTempMetric.With(labels).Set(value)
}
