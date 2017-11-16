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
