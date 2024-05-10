package services

import (
	"context"
	"fmt"
	"time"

	"github.com/denistv/myheat-prometheus-exporter/internal/clients/myheat"
	"github.com/denistv/wdlogger"
)

func NewExporterConfig(pullInterval time.Duration) ExporterConfig {
	return ExporterConfig{
		PullInterval: pullInterval,
	}
}

type ExporterConfig struct {
	PullInterval time.Duration
}

func (e ExporterConfig) Validate() error {
	if e.PullInterval.Seconds() <= 0 {
		return fmt.Errorf("exporter pull interval must be positive number")
	}

	return nil
}

func NewExporter(cfg ExporterConfig, evanClient *myheat.Client, l wdlogger.Logger, metricsService *Metrics) *Exporter {
	return &Exporter{
		cfg:            cfg,
		logger:         l,
		myheat:         evanClient,
		metricsService: metricsService,
	}
}

type Exporter struct {
	cfg            ExporterConfig
	logger         wdlogger.Logger
	myheat         *myheat.Client
	metricsService *Metrics
}

func (e *Exporter) Run(ctx context.Context) {
	e.logger.Info("exporter started")

	ticker := time.NewTicker(time.Second * 30)

	err := e.pull(ctx)
	if err != nil {
		e.logger.Error("error while pulling data", wdlogger.NewErrorField("error", err))
	}

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("received shutdown signal, exiting")
			return

		case <-ticker.C:
			err := e.pull(ctx)
			if err != nil {
				e.logger.Error("error while pulling data", wdlogger.NewErrorField("error", err))
			}
		}
	}
}

func (e *Exporter) pull(ctx context.Context) error {
	e.logger.Info("pull data from myheat")
	defer func() {
		e.logger.Info("pull data from myheat complete")
	}()

	getDevicesResp, err := e.myheat.GetDevices(ctx)
	if err != nil {
		return fmt.Errorf("getting devices: %w", err)
	}

	if len(getDevicesResp.Data["devices"]) == 0 {
		return nil
	}

	for _, device := range getDevicesResp.Data["devices"] {
		deviceInfo, err := e.myheat.GetDeviceInfo(ctx, device.ID)
		if err != nil {
			e.logger.Error(
				"get device info error",
				wdlogger.NewErrorField("error", err), wdlogger.NewInt64Field("id", device.ID),
			)
			continue
		}

		if len(deviceInfo.Data.Envs) == 0 {
			e.logger.Warn("empty device info data", wdlogger.NewInt64Field("id", device.ID))
			continue
		}

		e.metricsService.SetDeviceWeatherTemp(device.ID, device.Name, device.City, deviceInfo.Data.WeatherTemp)
		e.metricsService.SetDeviceSeverity(device.ID, device.Name, device.Severity, device.SeverityDesc)

		for _, env := range deviceInfo.Data.Envs {
			if env.Type != myheat.EnvTypeRoomTemperature {
				continue
			}

			e.metricsService.SetEnvironmentTempCurrent(env.ID, env.Name, env.Value)
			e.metricsService.SetEnvironmentTempTarget(env.ID, env.Name, env.Target)
			e.metricsService.SetEnvironmentHeatDemand(env.ID, env.Name, env.Demand)
			e.metricsService.CountEnvHeatDemandSeconds(env.ID, env.Name, env.Demand)
		}
	}

	return nil
}
