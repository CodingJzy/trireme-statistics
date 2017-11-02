package main

import (
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/aporeto-inc/trireme-statistics/configuration"
	"github.com/aporeto-inc/trireme-statistics/grafana/grafanalib"
	"github.com/aporeto-inc/trireme-statistics/version"
)

func banner(version, revision string) {
	fmt.Printf(`


	  _____     _
	 |_   _| __(_)_ __ ___ _ __ ___   ___
	   | || '__| | '__/ _ \ '_'' _ \ / _ \
	   | || |  | | | |  __/ | | | | |  __/
	   |_||_|  |_|_|  \___|_| |_| |_|\___|
		GRAFANA-INITIALIZER

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

	// Creating grafana dashboards
	err = setupGrafana(cfg.GrafanaUsername, cfg.GrafanaPassword, cfg.GrafanaURL, cfg.GrafanaDBAccess, cfg.InfluxUsername, cfg.InfluxPassword, cfg.InfluxURL, cfg.InfluxDBName)
	if err != nil {
		zap.L().Fatal("Error: Connecting to GrafanaServer", zap.Error(err))
	}

	zap.L().Info("Grafana configuration finished successfully. Exiting in 5 seconds")

	<-time.After(time.Second * 5)
}

// setupGrafana sets up Grafana to create the Flow and Container dashboard
func setupGrafana(uiUser, uiPassword, uiAddress, uiAccess, influxUser, influxPassword, influxAddress, influxDB string) error {

	grafanaClient, err := grafana.NewUISession(uiUser, uiPassword, uiAddress)
	if err != nil {
		return fmt.Errorf("Error: Initiating Connection to Grafana Server %s", err)
	}

	err = grafanaClient.CreateDataSource("Events", influxDB, influxUser, influxPassword, influxAddress, uiAccess)
	if err != nil {
		return fmt.Errorf("Error: Creating Datasource %s", err)
	}

	grafanaClient.CreateDashboard("StatisticBoard")
	initSingleStatPanels(grafanaClient)
	initTablePanels(grafanaClient)
	grafanaClient.CreateDashboard("Graphs")
	initGraphPanels(grafanaClient)

	return nil
}

func initSingleStatPanels(grafanaClient grafana.GrafanaManipulator) {
	grafanaClient.CreateRow("SingleStat")
	grafanaClient.AddPanel(grafana.SingleStat, grafana.ContainerEventsCount, grafana.ContainerEvent, []string{"IPAddress"})
	grafanaClient.AddPanel(grafana.SingleStat, grafana.FlowEventsCount, grafana.FlowEvent, []string{"Action"})
	grafanaClient.UploadToDashboard()
}

func initTablePanels(grafanaClient grafana.GrafanaManipulator) {
	grafanaClient.CreateRow("Table")
	FourTupleFields := []string{"Action", "SourceIP", "SourcePort", "DestinationIP", "DestinationPort", "Tags"}
	grafanaClient.AddPanel(grafana.Table, grafana.FourTupleWithAction, grafana.FlowEvent, FourTupleFields)
	grafanaClient.AddPanel(grafana.Table, grafana.ContainerEventFields, grafana.ContainerEvent, []string{grafana.AllFields})
	grafanaClient.AddPanel(grafana.Table, grafana.FlowEventFields, grafana.FlowEvent, []string{grafana.AllFields})
	grafanaClient.UploadToDashboard()
}

func initGraphPanels(grafanaClient grafana.GrafanaManipulator) {
	grafanaClient.CreateRow("Graph")
	grafanaClient.AddPanel(grafana.Graph, grafana.ContainerEventsGraph, grafana.ContainerEvent, []string{"ContextID", "IPAddress", "Tags"})
	grafanaClient.AddPanel(grafana.Graph, grafana.FlowEventsGraph, grafana.FlowEvent, []string{"ContextID", "Tags"})
	grafanaClient.UploadToDashboard()
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
