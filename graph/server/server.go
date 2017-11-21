package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/aporeto-inc/trireme-statistics/influxdb"
	"github.com/influxdata/influxdb/client/v2"
)

// NewGraph is the handler for graph generators
func NewGraph(httpClient influxdb.DataAdder, dbname string) *Graph {

	return &Graph{
		httpClient: httpClient,
		dbname:     dbname,
		nodesChan:  make(chan []Node),
		linksChan:  make(chan []Link),
		nodeMap:    make(map[string]*Node),
		linkMap:    make(map[string]*Link),
	}
}

// GetData is called by the client which generates json with a logic that defines the nodes and links for graph
func (g *Graph) GetData(w http.ResponseWriter, r *http.Request) {
	var graphData *GraphData

	starttime, err := time.Parse(time.RFC3339, r.URL.Query().Get("starttime")+"Z")
	if err != nil {
		zap.L().Warn("Parsing Time ", zap.Error(err))
	}

	endtime, err := time.Parse(time.RFC3339, r.URL.Query().Get("endtime")+"Z")
	if err != nil {
		zap.L().Warn("Parsing Time ", zap.Error(err))
	}

	namespace := r.URL.Query().Get("namespace")

	if r.URL.Query().Get("starttime") != "" || r.URL.Query().Get("endtime") != "" || namespace != "" {
		// Launching parallely to aggregate nodes and links for given input
		go g.FindLinksBetweenGivenTimeAndOrNamespace(starttime, endtime, namespace)
		go g.FindNodesBetweenGivenTimeAndOrNamespace(starttime, endtime, namespace)
		graphData = &GraphData{Nodes: <-g.nodesChan, Links: <-g.linksChan}
	} else {
		graphData = g.jsonData
	}

	err = json.NewEncoder(w).Encode(graphData)
	if err != nil {
		http.Error(w, err.Error(), 3)
	}
}

// FindNodesBetweenGivenTimeAndOrNamespace will aggregate nodes within the specified time, namespaces or both
func (g *Graph) FindNodesBetweenGivenTimeAndOrNamespace(starttime time.Time, endtime time.Time, namespace string) {
	var nodes []Node

	for _, node := range g.jsonData.Nodes {
		switch {
		case node.Time.After(starttime) && node.Time.Before(endtime) && node.Namespace == namespace:
			nodes = append(nodes, node)
		case node.Time.After(starttime) && node.Time.Before(endtime) && namespace == "":
			nodes = append(nodes, node)
		case node.Namespace == namespace:
			nodes = append(nodes, node)
		}
	}
	g.nodesChan <- nodes
	return
}

// FindLinksBetweenGivenTimeAndOrNamespace will aggregate links within the specified time, namespaces or both
func (g *Graph) FindLinksBetweenGivenTimeAndOrNamespace(starttime time.Time, endtime time.Time, namespace string) {
	var links []Link

	for _, link := range g.jsonData.Links {
		switch {
		case link.Time.After(starttime) && link.Time.Before(endtime) && link.Namespace == namespace:
			links = append(links, link)
		case link.Time.After(starttime) && link.Time.Before(endtime) && namespace == "":
			links = append(links, link)
		case link.Namespace == namespace:
			links = append(links, link)
		default:
			links = append(links, DefaultLink())
		}
	}
	g.linksChan <- links
	return
}

// Start is used to start generating jsonData for every 15 seconds
func (g *Graph) Start(interval int) {
	zap.L().Info("Starting to Generate JSON every", zap.Any("Interval", interval))
	go func() {
		for range time.Tick(time.Second * time.Duration(interval)) {
			res, err := g.getContainerEvents()
			if err != nil {
				zap.L().Error("Retrieving container events from DB", zap.Error(err))
			}
			g.jsonData, err = g.transform(res)
			if err != nil {
				zap.L().Error("Transforming to nodes and links", zap.Error(err))
			}
		}
	}()
}

// GetGraph is used to parse html with custom address to request for json
func (g *Graph) GetGraph(w http.ResponseWriter, r *http.Request) {

	htmlData, err := template.New("graph").Parse(js)
	if err != nil {
		http.Error(w, err.Error(), 0)
	}

	graphDataAddress := r.URL.Query().Get("address")
	if graphDataAddress == "" {
		graphDataAddress = defaultGraphDataAddress
	}

	r.ParseForm()

	data := struct {
		Address string
	}{
		Address: graphDataAddress,
	}

	switch {
	case len(r.Form["starttime"]) > 0 && len(r.Form["endtime"]) > 0 && len(r.Form["namespace"]) > 0:
		data.Address = data.Address + "?starttime=" + r.Form["starttime"][0] + "&endtime=" + r.Form["endtime"][0] + "&namespace=" + r.Form["namespace"][0]
	case len(r.Form["starttime"]) > 0 && len(r.Form["endtime"]) > 0:
		data.Address = data.Address + "?starttime=" + r.Form["starttime"][0] + "&endtime=" + r.Form["endtime"][0]
	case len(r.Form["namespace"]) > 0:
		data.Address = data.Address + "?namespace=" + r.Form["namespace"][0]
	}

	err = htmlData.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), 1)
	}

	w.Header().Set("Content-Type", "text/html")
}

func (g *Graph) getContainerEvents() (*client.Response, error) {
	zap.L().Info("Retrieving ContainerEvents from DB")
	res, err := g.executeQuery(ContainerEventsQuery)
	if err != nil {
		return nil, fmt.Errorf("Executing Query %s", err)
	}

	return res, nil
}

func (g *Graph) getFlowEvents(httpClient influxdb.DataAdder, dbname string) (*client.Response, error) {
	zap.L().Info("Retrieving FlowEvents from DB")
	res, err := g.executeQuery(FlowEventsQuery)
	if err != nil {
		return nil, fmt.Errorf("Executing Query %s", err)
	}

	return res, nil
}

func (g *Graph) executeQuery(query string) (*client.Response, error) {

	res, err := g.httpClient.ExecuteQuery(query, g.dbname)
	if err != nil {
		return nil, fmt.Errorf("Resource Unavailabe %s", err)
	}

	return res, nil
}

// transform will convert the JSON response from influxdb to nodes and links to generate graph
// the nodes are retrieved from influxdb and stored in map of nodes
// then later this map is used to generate links and links are stored in map of links
// the link generator basically generates the link by comparing the ipidhash with the flows hash
func (g *Graph) transform(res *client.Response) (*GraphData, error) {
	zap.L().Info("Transforming to Node and Link")

	if res == nil {
		return nil, fmt.Errorf("No Response from InfluxDB")
	}

	var startEvents = []string{ContainerUpdate}

	if len(res.Results[0].Series) > 0 {
		if res.Results[0].Series[0].Name == ContainerEvent {
			for _, containerEvent := range res.Results[0].Series[0].Values {
				var node Node
				containerAttr := extractContainerEventAttributes(containerEvent)
				if containerAttr == nil {
					return nil, fmt.Errorf("Empty Container Attributes ")
				}
				if containerAttr.event == ContainerUpdate {
					for _, containerEvent := range startEvents {
						if containerAttr.event == containerEvent {
							ipIDHash := getHash(containerAttr.contextID, containerAttr.ipAddress)
							if _, ok := g.nodeMap[ipIDHash]; !ok {
								node.ContextID = containerAttr.contextID
								parsedTime, err := time.Parse(time.RFC3339, containerAttr.timestamp)
								if err != nil {
									return nil, fmt.Errorf("Parsing Time %s", err)
								}
								node.Time = parsedTime
								node.IPAddress = containerAttr.ipAddress
								node.Namespace = g.parseTag(containerAttr.tags, PODNamespaceFromContainerTags)
								node.PodName = g.parseTag(containerAttr.tags, PODNameFromContainerTags)
								g.nodeMap[ipIDHash] = &node
							}
						}
					}
				} else if containerAttr.event == ContainerDelete {
					go g.deleteContainerEvents(containerAttr.contextID)
				}
			}
		}
	}

	err := g.generateLinks()
	if err != nil {
		return nil, fmt.Errorf("Generating Link %s", err)
	}

	g.populateNodesAndLinks()

	jsonData := GraphData{Nodes: g.nodes, Links: g.links}

	// Clears the structures and maps
	go g.clearDataStores()

	return &jsonData, nil
}

func (g *Graph) generateLinks() error {

	res, err := g.getFlowEvents(g.httpClient, g.dbname)
	if err != nil {
		return fmt.Errorf("Retrieving Flow Events %s", err)
	}

	if len(res.Results[0].Series) > 0 {
		if res.Results[0].Series[0].Name == FlowEvent {
			for _, flowEvent := range res.Results[0].Series[0].Values {
				var link Link
				flowAttr := extractFlowEventAttributes(flowEvent)
				if flowAttr == nil {
					return fmt.Errorf("Empty Flow Attributes ")
				}
				srcHash := getHash(flowAttr.srcID, flowAttr.srcIP)
				dstHash := getHash(flowAttr.dstID, flowAttr.dstIP)
				key := srcHash + dstHash
				if _, ok := g.linkMap[key]; !ok {
					if srcNode, ok := g.nodeMap[srcHash]; ok {
						link.Source = srcNode.ContextID
					}
					if dstNode, ok := g.nodeMap[dstHash]; ok {
						link.Target = dstNode.ContextID
					}

					if link.Source != "" && link.Target != "" {
						link.Action = flowAttr.action
						link.Namespace = g.parseTag(flowAttr.tags, PODNamespaceFromFlowTags)
						parsedTime, err := time.Parse(time.RFC3339, flowAttr.timestamp)
						if err != nil {
							return fmt.Errorf("Parsing Time %s", err)
						}
						link.Time = parsedTime
						g.linkMap[key] = &link
					}
				} else {
					if g.linkMap[key].Action != flowAttr.action {
						g.linkMap[key].Action = FlowNowRejected
					}
				}
			}
		}
	}

	return nil
}

func (g *Graph) deleteContainerEvents(contextID string) {

	for _, node := range g.nodeMap {
		if node.ContextID == contextID {
			ipIDHash := getHash(contextID, node.IPAddress)
			delete(g.nodeMap, ipIDHash)
		}
	}
	return
}

func (g *Graph) populateNodesAndLinks() {

	for _, node := range g.nodeMap {
		g.nodes = append(g.nodes, *node)
	}

	for _, link := range g.linkMap {
		g.links = append(g.links, *link)
	}
}

func (g *Graph) clearDataStores() {

	g.links = nil
	g.nodes = nil
	for k := range g.linkMap {
		delete(g.linkMap, k)
	}
	return
}

func (g *Graph) parseTag(tag string, parseType string) string {
	var result string

	switch parseType {
	case PODNameFromContainerTags:
		result = g.getNameOrNamespaceFromTag(tag, PODNameFromContainerTags)
	case PODNamespaceFromContainerTags:
		result = g.getNameOrNamespaceFromTag(tag, PODNamespaceFromContainerTags)
	case PODNamespaceFromFlowTags:
		result = g.getNameOrNamespaceFromTag(tag, PODNamespaceFromFlowTags)
	default:
		return ""
	}

	return result
}

func (g *Graph) getNameOrNamespaceFromTag(tags string, tagExtractor string) string {

	if strings.Contains(tags, tagExtractor) {
		for _, tag := range strings.Split(tags, " ") {
			tagCollection := strings.SplitAfter(tag, "=")
			for _, tagcollection := range tagCollection {
				g.extractTagValue(tagCollection, tagcollection, tagExtractor)
			}
		}
	}

	return g.tagValue
}

func (g *Graph) extractTagValue(tagCollection []string, tagcollection string, tagExtractor string) {

	if tagcollection == tagExtractor+"=" || tagcollection == "&{["+tagExtractor+"=" {
		if index := strings.IndexByte(tagCollection[1], ']'); index >= 0 {
			g.tagValue = tagCollection[1][:index]
		} else {
			g.tagValue = tagCollection[1]
		}
	}
}
