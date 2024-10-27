package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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

	// Configure tariff selector
	tariffs := []services.Tariff{}

	envTariff2FromRaw := os.Getenv("MYHEAT_TARIFF2_FROM")
	envTariff2ToRaw := os.Getenv("MYHEAT_TARIFF2_TO")

	if envTariff2FromRaw != "" && envTariff2ToRaw != "" {
		tariff2From, err := strconv.ParseInt(envTariff2FromRaw, 10, 32)
		if err != nil {
			logger.Fatal(err.Error())
		}

		tariff2To, err := strconv.ParseInt(envTariff2ToRaw, 10, 32)
		if err != nil {
			logger.Fatal(err.Error())
		}

		nt := services.NewNightTariff(int(tariff2From), int(tariff2To))
		tariffs = append(tariffs, nt)

		logger.Info(
			"night tariff applied",
			wdlogger.NewInt64Field("from", tariff2From),
			wdlogger.NewInt64Field("to", tariff2To),
		)
	}

	tariffSelector := services.NewTariffSelector(time.Now, tariffs)
	tariffSelector = tariffSelector

	metricsService := services.NewMetrics(logger, tariffSelector)
	go metricsService.Run(ctx)

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
