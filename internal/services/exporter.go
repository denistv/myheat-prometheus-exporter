package services

import (
	"context"
	"github.com/denistv/evan-prometheus-exporter/clients/evan"
	logger "github.com/denistv/wdlogger"
)

func NewExporter(evanClient *evan.Client, l logger.Logger) *Exporter {
	return &Exporter{
		logger:     l,
		evanClient: evanClient,
	}
}

type Exporter struct {
	logger     logger.Logger
	evanClient *evan.Client
}

func (e *Exporter) Run(ctx context.Context) {
	e.logger.Info("exporter started")

	select {
	case <-ctx.Done():
		e.logger.Info("received shutdown signal, exiting")
		return
	}
}
