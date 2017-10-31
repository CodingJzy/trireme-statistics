package grafana

import (
	"github.com/aporeto-inc/grafanaclient"
)

// Grafanaui is the interface which has all methods to interact with the grafana ui
type Grafanaui interface {
	CreateDataSource(name string, dbname string, dbuname string, dbpass string, dburl string, access string) error
	CreateDashboard(dbr string)
	AddCharts(panel PanelType, title string, fields string) grafanaclient.Panel
	AddRows(panel PanelType, rowname string, paneltitle string, events string)
	UploadToDashboard()
}
