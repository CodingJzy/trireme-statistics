package server

// DefaultLink is the default links struct for graph
func DefaultLink() Links {

	return Links{Source: "", Target: ""}
}

// DefaultNode is the default nodes struct for graph
func DefaultNode() Nodes {
	return Nodes{}
}

func getHash(contextID string, ipAddress string) string {

	return contextID + ":" + ipAddress
}
