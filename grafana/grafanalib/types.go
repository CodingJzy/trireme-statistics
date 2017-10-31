package grafana

import (
	"github.com/aporeto-inc/grafanaclient"
)

// Grafanauis is the structure which holds required grafana fields
// Implements Grafanaui interface
type Grafanauis struct {
	session   *grafanaclient.Session
	dashboard *grafanaclient.Dashboard
	row       grafanaclient.Row
}
