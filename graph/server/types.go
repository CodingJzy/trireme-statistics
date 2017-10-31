package server

import (
	"time"

	"github.com/aporeto-inc/trireme-statistics/influxdb"
)

// GraphData is the struct that holds the json format required for graph to generate nodes and link
type GraphData struct {
	Nodes []Nodes `json:"nodes"`
	Links []Links `json:"links"`
}

// Nodes which holds pu information
type Nodes struct {
	Time      time.Time `json:"time"`
	ContextID string    `json:"id"`
	PodName   string    `json:"name"`
	IPAddress string    `json:"ipaddress"`
	IPIDHash  string    `json:"ipidhash"`
	Namespace string    `json:"namespace"`
	Delete    bool      `json:"delete"`
}

// Links which holds the links between pu's
type Links struct {
	Time      time.Time `json:"time"`
	Source    string    `json:"source"`
	Target    string    `json:"target"`
	Action    string    `json:"action"`
	Namespace string    `json:"namespace"`
}

// Graph which holds the fields for graph creation
type Graph struct {
	jsonData   *GraphData
	httpClient *influxdb.Influxdb
	dbname     string
	nodes      []Nodes
	links      []Links
	nodesChan  chan []Nodes
	linksChan  chan []Links
	tagValue   string
}
