package main

import (
	"context"
	"net/http"

	"github.com/denistv/evan-prometheus-exporter/clients/evan"
	"github.com/denistv/evan-prometheus-exporter/internal/services"
	"github.com/denistv/wdlogger"
	"github.com/denistv/wdlogger/wrappers/stdwrap"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	ctx := context.Background()

	logger := stdwrap.NewSTDWrapper()
	clientCfg := evan.NewDefaultConfig()
	evanClient := evan.NewClient(clientCfg, logger)
	exp := services.NewExporter(evanClient, logger)

	go exp.Run(ctx)

	http.Handle("/metrics", promhttp.Handler())
	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		logger.Panic("unexpected error", wdlogger.NewErrorField("error", err))
	}

	<-ctx.Done()
}
