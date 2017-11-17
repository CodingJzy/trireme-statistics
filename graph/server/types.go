package server

import (
	"time"

	"github.com/aporeto-inc/trireme-statistics/influxdb"
)

// GraphData is the struct that holds the json format required for graph to generate nodes and link
type GraphData struct {
	Nodes []Node `json:"nodes"`
	Links []Link `json:"links"`
}

// Node which holds pu information
type Node struct {
	Time      time.Time `json:"time"`
	ContextID string    `json:"id"`
	PodName   string    `json:"name"`
	IPAddress string    `json:"ipaddress"`
	Namespace string    `json:"namespace"`
}

// Link which holds the links between pu's
type Link struct {
	Time      time.Time `json:"time"`
	Source    string    `json:"source"`
	Target    string    `json:"target"`
	Action    string    `json:"action"`
	Namespace string    `json:"namespace"`
}

// Graph which holds the fields for graph creation
type Graph struct {
	jsonData   *GraphData
	httpClient influxdb.DataAdder
	dbname     string
	nodes      []Node
	links      []Link
	nodesChan  chan []Node
	linksChan  chan []Link
	nodeMap    map[string]*Node
	linkMap    map[string]*Link
	tagValue   string
}

// ContainerEvents struct to hold container event attributes
type ContainerEvents struct {
	contextID string
	ipAddress string
	timestamp string
	tags      string
	event     string
}

// FlowEvents struct to hold flow event attributes
type FlowEvents struct {
	timestamp string
	srcID     string
	srcIP     string
	dstID     string
	dstIP     string
	action    string
	tags      string
}
