package grafana

import "github.com/aporeto-inc/grafanaclient"

// DefaultRow is a default row generator
func DefaultRow() grafanaclient.Row {
	return grafanaclient.Row{Height: ""}
}

// DefaultSelectAttribute is a default select attribute generator
func DefaultSelectAttribute() grafanaclient.Select {
	return grafanaclient.Select{}
}
