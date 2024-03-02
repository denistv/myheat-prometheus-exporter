package services

import (
	"context"
	"fmt"
	"time"

	"github.com/denistv/evan-prometheus-exporter/clients/evan"
	"github.com/denistv/wdlogger"
)

func NewExporter(evanClient *evan.Client, l wdlogger.Logger, metricsService *Metrics) *Exporter {
	return &Exporter{
		logger:         l,
		evanClient:     evanClient,
		metricsService: metricsService,
	}
}

type Exporter struct {
	logger         wdlogger.Logger
	evanClient     *evan.Client
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
	getDevicesResp, err := e.evanClient.GetDevices(ctx)
	if err != nil {
		return fmt.Errorf("getting devices: %w", err)
	}

	if len(getDevicesResp.Data["devices"]) == 0 {
		return nil
	}

	for _, device := range getDevicesResp.Data["devices"] {
		deviceInfo, err := e.evanClient.GetDeviceInfo(ctx, device.ID)
		if err != nil {
			e.logger.Error(
				"get device info error",
				wdlogger.NewErrorField("error", err), wdlogger.NewInt64Field("id", device.ID),
			)
			continue
		}

		if len(deviceInfo.Data.Envs) == 0 {
			e.logger.Info("empty device info data", wdlogger.NewInt64Field("id", device.ID))
			continue
		}

		for _, env := range deviceInfo.Data.Envs {
			e.metricsService.SetEnvironmentTemp(env.ID, env.Name, env.Value)
		}
	}

	return nil
}
