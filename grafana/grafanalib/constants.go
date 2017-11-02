package grafana

// PanelType is the type of panels used to append to rows
type PanelType string

const (
	// SingleStat - Type of panel used in grafana
	SingleStat PanelType = "singlestat"
	// Graph - Type of panel used in grafana
	Graph PanelType = "graph"
	// Diagram - Type of panel used in grafana
	Diagram PanelType = "jdbranham-diagram-panel"
	// Table - Type of panel used in grafana
	Table PanelType = "table"
)

const (
	// FlowEventsCount - Title for Singlestat flowevents panel
	FlowEventsCount = "FlowEventsCount"
	// ContainerEventsCount - Title for Singlestat containerevents panel
	ContainerEventsCount = "ContainerEventsCount"
	// FourTupleWithAction - Title for Table flowtuple panel
	FourTupleWithAction = "FourTupleWithAction"
	// ContainerEventFields - Title for Table containerevents panel
	ContainerEventFields = "ContainerEventFields"
	// FlowEventFields -  Title for Table flowevents panel
	FlowEventFields = "FlowEventFields"
	// ContainerEventsGraph - Title for Graph containerevents  panel
	ContainerEventsGraph = "ContainerEventsGraph"
	// FlowEventsGraph - Title for Graph flowevents panel
	FlowEventsGraph = "FlowEventsGraph"
	// AllFields - To retrieve all the fields from DB
	AllFields = "*"
)

const (
	// InfluxDB - datasource DB
	InfluxDB = "influxdb"
)

const (
	// ContainerEvent is the Container events measurement name
	ContainerEvent = "ContainerEvents"
	// FlowEvent is the Flow events measurement name
	FlowEvent = "FlowEvents"
)

const (
	// Count - aggregate function
	Count = "count"
)
