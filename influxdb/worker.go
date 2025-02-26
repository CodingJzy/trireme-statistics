package influxdb

import (
	"fmt"

	"git.cloud.top/DSec/trireme-lib/collector"
	"go.uber.org/zap"
)

const (
	// EventName is the constant used to store the name of the event type
	EventName = "EventName"

	// EventTypeFlow is the constant used to store event of type flows
	EventTypeFlow = "FlowEvents"

	// EventTypeContainer is the constant used to store event of type container
	EventTypeContainer = "ContainerEvents"

	// EventTypeContainerStart is the constant used to store event of type container start
	EventTypeContainerStart = "ContainerStartEvents"

	// EventTypeContainerStop is the constant used to store event of type container stop
	EventTypeContainerStop = "ContainerStopEvents"
)

// A worker manages the workload for the InfluxDB collector
type worker struct {
	events chan *workerEvent
	stop   chan struct{}
	db     DataAdder
}

type eventType int

const (
	containerEvent eventType = iota
	flowEvent      eventType = iota
)

// a workerEvent is an event that the worker need to process
type workerEvent struct {
	event           eventType
	containerRecord *collector.ContainerRecord
	flowRecord      *collector.FlowRecord
}

func newWorker(stop chan struct{}, db DataAdder) *worker {
	return &worker{
		events: make(chan *workerEvent, 500),
		stop:   stop,
		db:     db,
	}
}

func (w *worker) addEvent(wevent *workerEvent) {
	select {
	case w.events <- wevent: // Put event in channel unless it is full
		zap.L().Debug("Adding event to InfluxDBProcessingQueue.")
	default:
		zap.L().Warn("Event queue full for InfluxDB. Dropping event.")
	}
}

// startWorker start processing the event for this worker.
// Blocking... Use go.
func (w *worker) startWorker() {
	zap.L().Info("Starting InfluxDBworker")
	for {
		select {
		case event := <-w.events:
			w.processEvent(event)
		case <-w.stop:
			return
		}
	}
}

func (w *worker) processEvent(wevent *workerEvent) {
	zap.L().Debug("Processing event for InfluxDB")

	switch wevent.event {
	case containerEvent:
		if err := w.doCollectContainerEvent(wevent.containerRecord); err != nil {
			zap.L().Error("Couldn't process influxDB Request ContainerRequest", zap.Error(err))
		}

	case flowEvent:
		if err := w.doCollectFlowEvent(wevent.flowRecord); err != nil {
			zap.L().Error("Couldn't process influxDB Request FlowRequest", zap.Error(err))
		}
	}
}

// CollectContainerEvent implements trireme collector interface
func (w *worker) doCollectContainerEvent(record *collector.ContainerRecord) error {
	var eventName string

	switch record.Event {
	case collector.ContainerStart, collector.ContainerUpdate, collector.ContainerCreate:
		eventName = EventTypeContainerStart
	case collector.ContainerDelete, collector.ContainerStop:
		eventName = EventTypeContainerStop
	case collector.ContainerIgnored:
		// Used for non relevant container events.
		return nil
	case collector.ContainerFailed:
		// TODO: handle ContainerFailed event type
		return nil
	default:
		return fmt.Errorf("Unrecognized container event name %s ", record.Event)
	}
	var IPAddress string
	for _, v := range record.IPAddress {
		IPAddress = v
	}

	return w.db.AddData(map[string]string{
		"EventName": eventName,
		"EventID":   record.ContextID,
	}, map[string]interface{}{
		"ContextID": record.ContextID,
		"IPAddress": IPAddress,
		"Tags":      record.Tags,
		"Event":     record.Event,
	})
}

// CollectFlowEvent implements trireme collector interface
func (w *worker) doCollectFlowEvent(record *collector.FlowRecord) error {
	return w.db.AddData(map[string]string{
		"EventName": EventTypeFlow,
		"EventID":   record.ContextID,
	}, map[string]interface{}{
		"ContextID":       record.ContextID,
		"Counter":         record.Count,
		"SourceID":        record.Source.ID,
		"SourceIP":        record.Source.IP,
		"SourcePort":      record.Source.Port,
		"SourceType":      record.Source.Type,
		"DestinationID":   record.Destination.ID,
		"DestinationIP":   record.Destination.IP,
		"DestinationPort": record.Destination.Port,
		"DestinationType": record.Destination.Type,
		"Tags":            record.Tags,
		"Action":          record.Action,
		"DropReason":      record.DropReason,
		"PolicyID":        record.PolicyID,
	})
}
