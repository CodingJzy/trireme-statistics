package server

const defaultGraphDataAddress = "/get"

const (
	// ContainerEventsQuery is the query used to retrieve ContainerEvents from database
	ContainerEventsQuery = "SELECT * FROM ContainerEvents"
	// FlowEventsQuery is the query used to retrieve FlowEvents from database
	FlowEventsQuery = "SELECT * FROM FlowEvents"
)

const (
	// FlowReject indicates that a flow was rejected
	FlowReject = "reject"
	// FlowAccept logs that a flow is accepted
	FlowAccept = "accept"
	// FlowNowRejected logs that a flow is accepted and later rejected
	FlowNowRejected = "nowrejected"
	// ContainerStart indicates a container start event
	ContainerStart = "start"
	// ContainerStop indicates a container stop event
	ContainerStop = "stop"
	// ContainerCreate indicates a container create event
	ContainerCreate = "create"
	// ContainerDelete indicates a container delete event
	ContainerDelete = "delete"
	// ContainerUpdate indicates a container policy update event
	ContainerUpdate = "update"
	// ContainerFailed indicates an event that a container was stopped because of policy issues
	ContainerFailed = "forcestop"
	// ContainerIgnored indicates that the container will be ignored by Trireme
	ContainerIgnored = "ignore"
	// UnknownContainerDelete indicates that policy for an unknown  container was deleted
	UnknownContainerDelete = "unknowncontainer"
)

const (
	// ContainerEvent is the Container events measurement name
	ContainerEvent = "ContainerEvents"
	// FlowEvent is the Flow events measurement name
	FlowEvent = "FlowEvents"
)

const (
	// PODNameFromContainerTags is tha tag used to retrieve pod name from tags in ContainerEvents
	PODNameFromContainerTags = "@usr:io.kubernetes.pod.name"
	// PODNamespaceFromContainerTags is the tag used to retrieve pod namespace from tags in ContainerEvents
	PODNamespaceFromContainerTags = "@usr:io.kubernetes.pod.namespace"
	// PODNamespaceFromFlowTags is the tag used to retrieve flow associated to a particular pod namespace from tags in FlowEvents
	PODNamespaceFromFlowTags = "@namespace"
)
