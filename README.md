# trireme-statistics

trireme-statistics holds all the libraries and executables related to the metrics exported from the Trireme library.

More specifically:
* Trireme-Graph: A simple graphic implementation of the traffic flowing in your cluster.
* InfluxDB-Collector: A library implementing the collector interface from Trireme in order to send the data to InfluxDB.
* Grafana-Initializer: A library that connects to Grafana and initializes a series of relevant metrics that show Trireme's activity.

All of those can be launched as part of the Trireme-Kubernetes

# Trireme-Graph

Trireme-Graph is a simple implementation of a graphic display of all your network connections in a specific namespace for a specific timerange.

Trireme-Graph can be launched on Kubernetes and will by default try to connect to an InfluxDB available at `influxdb:8086`

```
 kubectl create -f https://github.com/aporeto-inc/trireme-kubernetes/blob/master/deployment/statistics/collector.yaml
```

# InfluxDB-Collector

InfluxDB-Collector is an [Implementation](https://github.com/aporeto-inc/trireme-lib/blob/master/collector/interfaces.go) of the Trireme Collector interface:

```
// EventCollector is the interface for collecting events.
type EventCollector interface {

	// CollectFlowEvent collect a  flow event.
	CollectFlowEvent(record *FlowRecord)

	// CollectContainerEvent collects a container events
	CollectContainerEvent(record *ContainerRecord)
}
```

which sends the events directly to InfluxDB.

# Grafana-Initializer

A library that connects to Grafana and initialize a dashboard with a couple predefined graphs that display information about the Data collected from Trireme into InfluxDB.