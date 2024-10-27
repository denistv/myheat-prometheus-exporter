package services

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/denistv/wdlogger"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	metricNameEnvTempCurrent       = "myheat_env_temp_current"
	metricNameEnvTempTarget        = "myheat_env_temp_target"
	metricNameEnvHeatDemand        = "myheat_env_heat_demand"
	metricNameEnvHeatDemandSeconds = "myheat_env_heat_demand_seconds_total"

	metricNameDeviceWeatherTemp = "myheat_dev_weather_temp"
	metricNameDeviceSeverity    = "myheat_dev_severity"
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

	// Env heat demand seconds
	envHeatDemandSecondsOpts := prometheus.CounterOpts{
		Name: metricNameEnvHeatDemandSeconds,
		Help: "Подсчитывает время, в течение которого запрошен нагрев",
	}
	envHeatDemandSecondsLabels := []string{"id", "name"}
	envHeatDemandSecondsMetric := promauto.NewCounterVec(envHeatDemandSecondsOpts, envHeatDemandSecondsLabels)

	// Device weather temperature
	deviceWeatherTempOpts := prometheus.GaugeOpts{
		Name: metricNameDeviceWeatherTemp,
		Help: "Температура на улице",
	}
	deviceWeatherTempLabels := []string{"id", "name", "city"}
	deviceWeatherTempMetric := promauto.NewGaugeVec(deviceWeatherTempOpts, deviceWeatherTempLabels)

	// Severity
	deviceSeverityOpts := prometheus.GaugeOpts{
		Name: metricNameDeviceSeverity,
		Help: "Состояние устройства",
	}
	deviceSeverityLabels := []string{"id", "name"}
	deviceSeverityMetric := promauto.NewGaugeVec(deviceSeverityOpts, deviceSeverityLabels)

	return &Metrics{
		logger: logger,

		envTempCurrMetric:          envTempCurrMetric,
		envTempTargetMetric:        envTempTargetMetric,
		envHeatDemandMetric:        envHeatDemandMetric,
		envHeatDemandSecondsMetric: envHeatDemandSecondsMetric,
		envHeatDemandSecondsState:  make(map[int64]envHeatDemandState),
		deviceWeatherTempMetric:    deviceWeatherTempMetric,
		deviceSeverityMetric:       deviceSeverityMetric,
	}
}

type Metrics struct {
	logger wdlogger.Logger

	envTempCurrMetric          *prometheus.GaugeVec
	envTempTargetMetric        *prometheus.GaugeVec
	envHeatDemandMetric        *prometheus.GaugeVec
	envHeatDemandSecondsMetric *prometheus.CounterVec

	deviceWeatherTempMetric *prometheus.GaugeVec
	deviceSeverityMetric    *prometheus.GaugeVec

	envHeatDemandSecondsStateMu sync.RWMutex
	envHeatDemandSecondsState   map[int64]envHeatDemandState
}

func (m *Metrics) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			m.envHeatDemandSecondsStateMu.RLock()

			for _, state := range m.envHeatDemandSecondsState {
				// Пропускаем env's, нагрев которых выключен
				if !state.value {
					continue
				}

				m.envHeatDemandSecondsMetric.With(state.labels).Inc()
			}

			m.envHeatDemandSecondsStateMu.RUnlock()
		case <-ctx.Done():
			return
		}
	}
}

// Дефолтные лейблы для большинства метрик
func defaultLabels(id int64, name string) map[string]string {
	return map[string]string{
		"id":   strconv.FormatInt(id, 10),
		"name": name,
	}
}

func (m *Metrics) SetEnvironmentTempCurrent(id int64, name string, value float64) {
	m.logger.Info(
		"set",
		wdlogger.NewStringField("metric_name", metricNameEnvTempCurrent),
		wdlogger.NewInt64Field("id", id),
		wdlogger.NewStringField("name", name),
		wdlogger.NewFloat64Field("value", value),
	)

	labels := defaultLabels(id, name)
	m.envTempCurrMetric.With(labels).Set(value)
}

func (m *Metrics) SetEnvironmentTempTarget(id int64, name string, value float64) {
	m.logger.Info(
		"set",
		wdlogger.NewStringField("metric_name", metricNameEnvTempTarget),
		wdlogger.NewInt64Field("id", id),
		wdlogger.NewStringField("name", name),
		wdlogger.NewFloat64Field("value", value),
	)

	labels := defaultLabels(id, name)
	m.envTempTargetMetric.With(labels).Set(value)
}

func boolToFloat64(v bool) float64 {
	if v {
		return 1
	}
	return 0
}

func (m *Metrics) SetEnvironmentHeatDemand(id int64, name string, value bool) {
	m.logger.Info(
		"set",
		wdlogger.NewStringField("metric_name", metricNameEnvHeatDemand),
		wdlogger.NewInt64Field("id", id),
		wdlogger.NewStringField("name", name),
		wdlogger.NewBoolField("value", value),
	)

	labels := defaultLabels(id, name)
	m.envHeatDemandMetric.With(labels).Set(boolToFloat64(value))
}

func (m *Metrics) SetDeviceWeatherTemp(id int64, name string, city string, value float64) {
	m.logger.Info(
		"set",
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

func (m *Metrics) SetDeviceSeverity(id int64, name string, value int64, desc string) {
	m.logger.Info(
		"set",
		wdlogger.NewStringField("metric_name", metricNameDeviceSeverity),
		wdlogger.NewInt64Field("id", id),
		wdlogger.NewStringField("name", name),
		wdlogger.NewInt64Field("value", value),
		wdlogger.NewStringField("desc", desc),
	)

	labels := defaultLabels(id, name)

	m.deviceSeverityMetric.Reset()
	m.deviceSeverityMetric.With(labels).Set(float64(value))
}

type envHeatDemandState struct {
	labels map[string]string
	value  bool
}

func (m *Metrics) CountEnvHeatDemandSeconds(id int64, name string, value bool) {
	m.logger.Info(
		"set",
		wdlogger.NewStringField("metric_name", metricNameEnvHeatDemandSeconds),
		wdlogger.NewInt64Field("id", id),
		wdlogger.NewStringField("name", name),
		wdlogger.NewBoolField("value", value),
	)

	m.envHeatDemandSecondsStateMu.Lock()
	defer m.envHeatDemandSecondsStateMu.Unlock()

	state, ok := m.envHeatDemandSecondsState[id]
	if !ok {
		state = envHeatDemandState{
			labels: defaultLabels(id, name),
		}
	}

	// update state
	state.value = value
	m.envHeatDemandSecondsState[id] = state
}
