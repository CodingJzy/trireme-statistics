package grafana

import (
	"github.com/aporeto-inc/grafanaclient"
)

// Grafana is the structure which holds required grafana fields
// Implements GrafanaManipulator interface
type Grafana struct {
	session         *grafanaclient.Session
	dashboard       *grafanaclient.Dashboard
	panelInstance   grafanaclient.Panel
	panelCollection []grafanaclient.Panel
	Row             grafanaclient.Row
}
