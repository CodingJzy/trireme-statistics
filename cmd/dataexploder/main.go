package main

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/aporeto-inc/trireme-statistics/influxdb"
	"github.com/aporeto-inc/trireme-statistics/models"
	"go.aporeto.io/trireme-lib/collector"
	"go.aporeto.io/trireme-lib/policy"
)

var wg sync.WaitGroup

func explode() {
	defer wg.Done()
	var flowModel models.FlowModel
	var contModel models.ContainerModel
	var source collector.EndPoint
	var destination collector.EndPoint
	samplesize := 500
	counter := 0
	httpCli, _ := influxdb.NewDBConnection("aporeto", "aporeto", "http://influxdb:8086", "flowDB", false)

	httpCli.Start()
	for i := 0; i < samplesize; i++ {
		fmt.Printf("Sending data %d \n", i)

		flowModel.FlowRecord.ContextID = "1ascasd7t"
		flowModel.FlowRecord.Count = counter
		flowModel.Counter = counter

		source.ID = "srcID"
		source.IP = "192.168.0.1"
		source.Port = 1234 + uint16(i)
		source.Type = collector.Address

		flowModel.FlowRecord.Source = &source

		destination.ID = "dstID"
		destination.IP = "192.1688.2.2"
		destination.Port = 880
		destination.Type = collector.Address

		flowModel.FlowRecord.Destination = &destination

		var tags policy.TagStore
		tags.Tags = []string{"&{[k8s-app=kube-dns pod-template-hash=3468831164 @namespace=kube-system AporetoContextID=02f4ebf65b05]}"}
		flowModel.FlowRecord.Tags = &tags
		var actype policy.ActionType
		actype.Accepted()
		actype.ActionString()
		flowModel.FlowRecord.Action = actype
		flowModel.FlowRecord.DropReason = "None"
		flowModel.FlowRecord.PolicyID = "sampleID"

		httpCli.CollectFlowEvent(&flowModel.FlowRecord)
		var policy policy.TagStore
		policy.Tags = []string{"&{[@sys:image=gcr.io/google_containers/pause-amd64:3.0 @sys:name=/k8s_POD_aporeto-collector-13rvx_kube-system_b86c7f27-ba0d-11e7-8725-42010a8001d7_0 @usr:io.kubernetes.pod.namespace=kube-system @usr:io.kubernetes.pod.uid=b86c7f27-ba0d-11e7-8725-42010a8001d7 @usr:annotation.kubernetes.io/config.source=api @usr:io.kubernetes.container.name=POD @usr:io.kubernetes.docker.type=podsandbox @usr:io.kubernetes.pod.name=aporeto-collector-13rvx @usr:annotation.kubernetes.io/config.seen=2017-10-26T05:22:54.960841543Z @usr:annotation.kubernetes.io/created-by={\"kind\":\"SerializedReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"ReplicaSet\",\"namespace\":\"kube-system\",\"name\":\"aporeto-collector\",\"uid\":\"b86841e6-ba0d-11e7-8725-42010a8001d7\",\"apiVersion\":\"extensions\",\"resourceVersion\":\"1837901\"}} @usr:app=aporeto-collector]}"}
		contModel.ContainerRecord.ContextID = "1ascasd7t"
		contModel.ContainerRecord.Event = "start"
		contModel.ContainerRecord.Tags = &policy
		httpCli.CollectContainerEvent(&contModel.ContainerRecord)
		counter++
	}
	wg.Wait()
	httpCli.Stop()
}

func main() {
	setLogs("human", "debug")
	wg.Add(1)
	time.Sleep(time.Second * 1)
	go explode()
	wg.Wait()
	fmt.Println("Done main")
}

// setLogs setups Zap to
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
