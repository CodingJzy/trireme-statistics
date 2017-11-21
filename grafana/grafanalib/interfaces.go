package grafana

import (
	"github.com/aporeto-inc/grafanaclient"
)

// Grafanamanipulator is the interface which has all methods to interact with the grafana ui
type Grafanamanipulator interface {
	CreateDataSource(name string, dbname string, dbuname string, dbpass string, dburl string, access string) error
	CreateDashboard(dbr string)
	AddPanel(panel PanelType, title string, measurement string, fields []string) grafanaclient.Panel
	CreateRow(rowname string)
	CreateTarget(measurement string, fields []string, aggregateFunction string)
	ConstructSelectQueriesFromFields(fields []string, aggregareFuntionSelects grafanaclient.Select) []grafanaclient.Selects
	UploadToDashboard()
}
