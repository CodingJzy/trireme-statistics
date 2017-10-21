package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/rs/cors"

	"github.com/aporeto-inc/trireme-csr/version"
	"github.com/aporeto-inc/trireme-statistics/configuration"
	"github.com/aporeto-inc/trireme-statistics/grafana"
	"github.com/aporeto-inc/trireme-statistics/graph/server"
	"github.com/aporeto-inc/trireme-statistics/influxdb"
)

func banner(version, revision string) {
	fmt.Printf(`


	  _____     _
	 |_   _| __(_)_ __ ___ _ __ ___   ___
	   | || '__| | '__/ _ \ '_'' _ \ / _ \
	   | || |  | | | |  __/ | | | | |  __/
	   |_||_|  |_|_|  \___|_| |_| |_|\___|
		STATISTICS

_______________________________________________________________
             %s - %s
                                                 🚀  by Aporeto

`, version, revision)
}

func main() {
	banner(version.VERSION, version.REVISION)

	cfg, err := configuration.LoadConfiguration()
	if err != nil {
		log.Fatal("Error parsing configuration", err)
	}

	err = setLogs(cfg.LogFormat, cfg.LogLevel)
	if err != nil {
		log.Fatalf("Error setting up logs: %s", err)
	}

	zap.L().Debug("Config used", zap.Any("Config", cfg))

	influxClient, err := influxdb.NewDBConnection(cfg.DBUserName, cfg.DBPassword, cfg.DBAddress, cfg.DBName, cfg.DBSkipTLS)
	if err != nil {
		zap.L().Fatal("Error: Initiating Connection to DB", zap.Error(err))
	}

	// Creating grafana dashboards
	err = setupGrafana(cfg.UIUserName, cfg.UIPassword, cfg.UIAddress, cfg.UIDBAccess, cfg.DBUserName, cfg.DBPassword, cfg.DBAddress, cfg.DBName)
	if err != nil {
		zap.L().Fatal("Error: Connecting to GrafanaServer", zap.Error(err))
	}

	// serveGraph is blocking
	go func() {
		err = serveGraph(influxClient, cfg.ListenAddress, cfg.DBName)
		if err != nil {
			zap.L().Fatal("Error: Connecting to GraphServer", zap.Error(err))
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	zap.L().Info("Everything started. Waiting for Stop signal")
	// Waiting for a Sig
	<-c

}

// setLogs setups Zap to log at the specified log level and format
func setLogs(logFormat, logLevel string) error {
	var zapConfig zap.Config

	switch logFormat {
	case "json":
		zapConfig = zap.NewProductionConfig()
		zapConfig.DisableStacktrace = true
	default:
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.DisableStacktrace = true
		zapConfig.DisableCaller = true
		zapConfig.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {}
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set the logger
	switch logLevel {
	case "trace":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "debug":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "fatal":
		zapConfig.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(logger)
	return nil
}

// setupGrafana sets up Grafana to create the Flow and Container dashboard
func setupGrafana(uiUser, uiPassword, uiAddress, uiAccess, influxUser, influxPassword, influxAddress, influxDB string) error {
	grafanaClient, err := grafana.NewUISession(uiUser, uiPassword, uiAddress)
	if err != nil {
		return fmt.Errorf("Error: Initiating Connection to Grafana Server %s", err)
	}

	err = grafanaClient.CreateDataSource("Events", influxDB, influxUser, influxPassword, influxDB, uiAccess)
	if err != nil {
		return fmt.Errorf("Error: Creating Datasource %s", err)
	}

	grafanaClient.CreateDashboard("StatisticBoard")
	grafanaClient.AddRows(grafana.SingleStat, "events", "Action", "FlowEvents")
	grafanaClient.AddRows(grafana.SingleStat, "events", "IPAddress", "ContainerEvents")

	return nil
}

func serveGraph(influxClient *influxdb.Influxdb, listenAddress string, dbname string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/get", server.GetData(influxClient, dbname))
	mux.HandleFunc("/graph", server.GetGraph)

	handler := cors.Default().Handler(mux)

	err := http.ListenAndServe(listenAddress, handler)
	if err != nil {
		return fmt.Errorf("ListenAndServe: %s", err)
	}

	fmt.Println("Server Listening at ", listenAddress)
	return nil
}
