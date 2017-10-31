package grafana

import (
	"time"

	"github.com/aporeto-inc/grafanaclient"
)

// NewUISession is used to create a new session and return grafana handle
func NewUISession(user string, pass string, addr string) (Grafanaui, error) {

	session, err := createSession(user, pass, addr)
	if err != nil {
		return nil, err
	}

	return &Grafanauis{
		session: session,
	}, nil
}

func createSession(user string, pass string, addr string) (*grafanaclient.Session, error) {

	session := grafanaclient.NewSession(user, pass, addr)
	err := session.DoLogon()
	if err != nil {
		return nil, err
	}

	return session, nil
}

// CreateDataSource is used to create a new datasource based on users arguements
func (g *Grafanauis) CreateDataSource(name string, dbname string, dbuname string, dbpass string, dburl string, access string) error {

	datasourceName, err := g.session.GetDataSource(name)
	if err != nil {
		return err
	}

	if datasourceName.Name != name {
		ds := grafanaclient.DataSource{Name: name,
			Type:     "influxdb",
			Access:   access,
			URL:      dburl,
			User:     dbuname,
			Password: dbpass,
			Database: dbname,
		}

		err := g.session.CreateDataSource(ds)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateDashboard is used to create a new dashboard
func (g *Grafanauis) CreateDashboard(dbtitle string) {

	dashboard := grafanaclient.Dashboard{Editable: true}
	dashboard.Title = dbtitle
	g.dashboard = &dashboard
}

// AddRows is used to add a new row in the dashboard
func (g *Grafanauis) AddRows(panel PanelType, rowname string, fields string, events string) {

	graphRow := grafanaclient.NewRow()
	graphRow.Title = rowname
	g.row = graphRow

	newpanel := g.AddCharts(panel, events, fields)

	g.row.AddPanel(newpanel)

	g.dashboard.AddRow(g.row)

	g.UploadToDashboard()
}

// UploadToDashboard is used to push all created rows into the dashboard
func (g *Grafanauis) UploadToDashboard() {

	g.dashboard.SetTimeFrame(time.Now().Add(-5*time.Minute), time.Now().Add(10*time.Minute))

	g.session.UploadDashboard(*g.dashboard, true)
}

// AddCharts is used to add different charts into rows
func (g *Grafanauis) AddCharts(paneltype PanelType, paneltitle string, fields string) grafanaclient.Panel {

	graphPanel := grafanaclient.NewPanel()

	graphPanel.Title = paneltitle
	if paneltype == SingleStat {
		graphPanel.Type = "singlestat"
		graphPanel.DataSource = "Events"
	}
	graphPanel.ValueName = "total"
	graphPanel.Span = 12
	graphPanel.Stack = true
	graphPanel.Fill = 1

	target := grafanaclient.NewTarget()
	if paneltitle == "FlowEvents" {
		target.Measurement = "FlowEvents"
	} else if paneltitle == "ContainerEvents" {
		target.Measurement = "ContainerEvents"
	}

	var selectAttributeType grafanaclient.Select
	selectAttributeType.Type = "field"
	selectAttributeType.Params = []string{fields}

	var selectAttributeCount grafanaclient.Select
	selectAttributeCount.Type = "count"

	var selectAttributeCollection grafanaclient.Selects
	selectAttributeCollection = append(selectAttributeCollection, selectAttributeType)
	selectAttributeCollection = append(selectAttributeCollection, selectAttributeCount)

	var selectCollection []grafanaclient.Selects
	selectCollection = append(selectCollection, selectAttributeCollection)

	target.Select = selectCollection
	target.Alias = fields

	graphPanel.AddTarget(target)

	return graphPanel
}
