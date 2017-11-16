package server

// DefaultLink is the default links struct for graph
func DefaultLink() Link {

	return Link{Source: "", Target: ""}
}

// DefaultNode is the default nodes struct for graph
func DefaultNode() Node {
	return Node{}
}

func getHash(contextID string, ipAddress string) string {

	return contextID + ":" + ipAddress
}

func extractContainerEventAttributes(containerEvent []interface{}) *ContainerEvents {
	var containerAttr ContainerEvents

	if value := containerEvent[ContainerTimestampIndex]; value != nil {
		containerAttr.timestamp = value.(string)
	}
	if value := containerEvent[ContainerContextIDIndex]; value != nil {
		containerAttr.contextID = value.(string)
	}
	if value := containerEvent[ContainerEventIndex]; value != nil {
		containerAttr.event = value.(string)
	}
	if value := containerEvent[ContainerIPAddressIndex]; value != nil {
		containerAttr.ipAddress = value.(string)
	}
	if value := containerEvent[ContainerTagsIndex]; value != nil {
		containerAttr.tags = value.(string)
	}

	return &containerAttr
}

func extractFlowEventAttributes(flowEvent []interface{}) *FlowEvents {
	var flowAttr FlowEvents

	if value := flowEvent[FlowTimestampIndex]; value != nil {
		flowAttr.timestamp = value.(string)
	}
	if value := flowEvent[FlowSourceIDIndex]; value != nil {
		flowAttr.srcID = value.(string)
	}
	if value := flowEvent[FlowSourceIPIndex]; value != nil {
		flowAttr.srcIP = value.(string)
	}
	if value := flowEvent[FlowDestinationIDIndex]; value != nil {
		flowAttr.dstID = value.(string)
	}
	if value := flowEvent[FlowDestinationIPIndex]; value != nil {
		flowAttr.dstIP = value.(string)
	}
	if value := flowEvent[FlowActionIndex]; value != nil {
		flowAttr.action = value.(string)
	}
	if value := flowEvent[FlowTagsIndex]; value != nil {
		flowAttr.tags = value.(string)
	}

	return &flowAttr
}
