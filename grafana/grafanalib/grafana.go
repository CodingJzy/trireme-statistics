package grafana

import (
	"time"

	"github.com/aporeto-inc/grafanaclient"
)

// NewUISession is used to create a new session and return grafana handle
func NewUISession(user string, pass string, addr string) (Grafanamanipulator, error) {

	session, err := createSession(user, pass, addr)
	if err != nil {
		return nil, err
	}

	return &Grafana{
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
func (g *Grafana) CreateDataSource(name string, dbname string, dbuname string, dbpass string, dburl string, access string) error {

	datasourceName, err := g.session.GetDataSource(name)
	if err != nil {
		return err
	}

	if datasourceName.Name != name {
		ds := grafanaclient.DataSource{Name: name,
			Type:     InfluxDB,
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
func (g *Grafana) CreateDashboard(dbtitle string) {

	dashboard := grafanaclient.Dashboard{Editable: true}
	dashboard.Title = dbtitle
	g.dashboard = &dashboard
}

// UploadToDashboard is used to push all created panels ans rows into the dashboard
func (g *Grafana) UploadToDashboard() {

	for _, panel := range g.panelCollection {
		g.Row.AddPanel(panel)
	}

	g.dashboard.AddRow(g.Row)
	g.dashboard.SetTimeFrame(time.Now().Add(-5*time.Minute), time.Now().Add(10*time.Minute))
	g.session.UploadDashboard(*g.dashboard, true)
	g.panelCollection = nil
}

// CreateRow is used to create a new row in the dashboard
func (g *Grafana) CreateRow(rowname string) {

	newRow := grafanaclient.NewRow()
	newRow.Title = rowname
	newRow.Height = "250px"
	g.Row = newRow
}

// AddPanel is used to add different panels into rows
func (g *Grafana) AddPanel(paneltype PanelType, paneltitle string, measurement string, fields []string) grafanaclient.Panel {

	g.panelInstance = grafanaclient.NewPanel()
	g.panelInstance.DataSource = "Events"

	switch paneltype {
	case SingleStat:
		g.generateSingleStatPanel(measurement, paneltitle, fields)
	case Graph:
		g.generateGraphPanel(measurement, paneltitle, fields)
	case Table:
		g.generateTablePanel(measurement, paneltitle, fields)
	}

	g.panelCollection = append(g.panelCollection, g.panelInstance)

	return g.panelInstance
}

// CreateTarget is used to create targets to query panels
func (g *Grafana) CreateTarget(measurement string, fields []string, aggregateFunction string) {

	var selectAttributeCount grafanaclient.Select
	if aggregateFunction != "" {
		selectAttributeCount.Type = aggregateFunction
	}

	for _, field := range fields {
		target := grafanaclient.NewTarget()
		target.Measurement = measurement
		selectCollection := g.ConstructSelectQueriesFromFields([]string{field}, selectAttributeCount)
		target.Select = selectCollection
		target.Alias = field

		g.panelInstance.AddTarget(target)
	}
}

// ConstructSelectQueriesFromFields is used to create select queries for panels to visualize
func (g *Grafana) ConstructSelectQueriesFromFields(fields []string, aggregareFuntionSelects grafanaclient.Select) []grafanaclient.Selects {

	var selectAttribute grafanaclient.Select
	var selectAttributeCollection grafanaclient.Selects
	var selectCollection []grafanaclient.Selects

	for _, field := range fields {
		selectAttribute.Type = "field"
		selectAttribute.Params = []string{field}
		selectAttributeCollection = append(selectAttributeCollection, selectAttribute)
		if aggregareFuntionSelects.Type != "" {
			selectAttributeCollection = append(selectAttributeCollection, aggregareFuntionSelects)
		}
		selectCollection = append(selectCollection, selectAttributeCollection)
		selectAttributeCollection = nil
	}

	return selectCollection
}

func (g *Grafana) generateGraphPanel(measurement string, paneltitle string, fields []string) {

	g.panelInstance.Title = paneltitle
	g.panelInstance.Type = "graph"
	g.panelInstance.ValueName = "total"
	g.panelInstance.Span = 12
	g.panelInstance.Stack = true
	g.panelInstance.Fill = 1

	switch measurement {
	case FlowEvent:
		g.CreateTarget(FlowEvent, fields, Count)
	case ContainerEvent:
		g.CreateTarget(ContainerEvent, fields, Count)
	}
}

func (g *Grafana) generateSingleStatPanel(measurement string, paneltitle string, fields []string) {

	g.panelInstance.Title = paneltitle
	g.panelInstance.Type = "singlestat"
	g.panelInstance.ValueName = "total"
	g.panelInstance.Span = 6
	g.panelInstance.Stack = true
	g.panelInstance.Fill = 1

	target := grafanaclient.NewTarget()

	switch measurement {
	case FlowEvent:
		target.Measurement = FlowEvent
	case ContainerEvent:
		target.Measurement = ContainerEvent
	}

	var selectAttributeCount grafanaclient.Select
	selectAttributeCount.Type = Count
	selectCollection := g.ConstructSelectQueriesFromFields(fields, selectAttributeCount)
	target.Select = selectCollection

	g.panelInstance.AddTarget(target)
}

func (g *Grafana) generateTablePanel(measurement string, paneltitle string, fields []string) {

	g.panelInstance.Title = paneltitle
	g.panelInstance.Type = "table"
	g.panelInstance.Span = 12
	g.panelInstance.Stack = true
	g.panelInstance.Fill = 1

	target := grafanaclient.NewTarget()

	switch measurement {
	case FlowEvent:
		target.Measurement = FlowEvent
	case ContainerEvent:
		target.Measurement = ContainerEvent
	}

	selectCollection := g.ConstructSelectQueriesFromFields(fields, DefaultSelectAttribute())
	target.Select = selectCollection

	groupBy := grafanaclient.NewGroupBy()
	groupBy[0].Type = "tag"
	groupBy[0].Params = []string{"EventName"}
	target.GroupBy = groupBy
	target.Format = "table"

	g.panelInstance.AddTarget(target)
}
