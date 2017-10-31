package grafana

// PanelType is the type of panels used to append to rows
type PanelType string

const (
	//SingleStat - Type of panel used in grafana
	SingleStat PanelType = "singlestat"
	//Graph - Type of panel used in grafana
	Graph PanelType = "graph"
	//Diagram - Type of panel used in grafana
	Diagram PanelType = "jdbranham-diagram-panel"
)
