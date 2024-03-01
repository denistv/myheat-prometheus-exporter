package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/denistv/evan-prometheus-exporter/clients/evan"
	"github.com/denistv/evan-prometheus-exporter/internal/services"
	"github.com/denistv/wdlogger"
	"github.com/denistv/wdlogger/wrappers/stdwrap"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	ctx, _ := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	logger := stdwrap.NewSTDWrapper()

	clientCfg := evan.NewDefaultConfig()
	clientCfg.Login = os.Getenv("EVAN_EXPORTER_LOGIN")
	clientCfg.Key = os.Getenv("EVAN_EXPORTER_KEY")

	if err := clientCfg.Validate(); err != nil {
		logger.Fatal("validating config", wdlogger.NewErrorField("error", err))
	}

	evanClient := evan.NewClient(clientCfg, logger)
	metricsService := services.NewMetrics(logger)
	exp := services.NewExporter(evanClient, logger, metricsService)

	go exp.Run(ctx)

	httpServer := http.Server{
		Addr: ":3000",
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := httpServer.ListenAndServe()
		if err != nil {
			logger.Panic("unexpected error", wdlogger.NewErrorField("error", err))
		}

		<-ctx.Done()

		err = httpServer.Shutdown(ctx)
		if err != nil {
			logger.Fatal("error while shutting down http server", wdlogger.NewErrorField("error", err))
			return
		}
	}()

	<-ctx.Done()
}
