package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denistv/myheat-prometheus-exporter/internal/clients/myheat"
	"github.com/denistv/myheat-prometheus-exporter/internal/services"
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

	clientCfg := myheat.NewDefaultConfig()
	clientCfg.Login = os.Getenv("MYHEAT_LOGIN")
	clientCfg.Key = os.Getenv("MYHEAT_KEY")

	if err := clientCfg.Validate(); err != nil {
		logger.Fatal("validating MyHeat client config", wdlogger.NewErrorField("error", err))
	}

	myheatClient := myheat.NewClient(clientCfg, logger)
	metricsService := services.NewMetrics(logger)

	exporterPullInterval, err := time.ParseDuration(os.Getenv("MYHEAT_EXPORTER_PULL_INTERVAL"))
	if err != nil {
		logger.Fatal("validating exporter config", wdlogger.NewErrorField("error", err))
	}

	expCfg := services.NewExporterConfig(exporterPullInterval)
	exp := services.NewExporter(expCfg, myheatClient, logger, metricsService)

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
