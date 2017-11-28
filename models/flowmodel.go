package models

import "github.com/aporeto-inc/trireme-lib/collector"

// FlowModel ...
type FlowModel struct {
	Counter    int
	FlowRecord collector.FlowRecord
}

// ContainerModel ...
type ContainerModel struct {
	Counter         int
	ContainerRecord collector.ContainerRecord
}
