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
func NewGraph(httpClient *influxdb.Influxdb, dbname string) *Graph {

	return &Graph{
		httpClient: httpClient,
		dbname:     dbname,
		nodesChan:  make(chan []Nodes),
		linksChan:  make(chan []Links),
	}
}

// GetData is called by the client which generates json with a logic that defines the nodes and links for graph
func (g *Graph) GetData(w http.ResponseWriter, r *http.Request) {
	var graphData *GraphData

	starttime, err := time.Parse(time.RFC3339, r.URL.Query().Get("starttime")+"Z")
	if err != nil {
		zap.L().Warn("Error: Parsing Time ", zap.Error(err))
	}

	endtime, err := time.Parse(time.RFC3339, r.URL.Query().Get("endtime")+"Z")
	if err != nil {
		zap.L().Warn("Error: Parsing Time ", zap.Error(err))
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
	var nodes []Nodes

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
	var links []Links

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
				zap.L().Error("Error: Retrieving container events from DB", zap.Error(err))
			}
			g.jsonData, err = g.transform(res)
			if err != nil {
				zap.L().Error("Error: Transforming to nodes and links", zap.Error(err))
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
		return nil, fmt.Errorf("Error: Resource Unavailabe %s", err)
	}

	return res, nil
}

func (g *Graph) getFlowEvents(httpClient *influxdb.Influxdb, dbname string) (*client.Response, error) {
	zap.L().Info("Retrieving FlowEvents from DB")
	res, err := g.executeQuery(FlowEventsQuery)
	if err != nil {
		return nil, fmt.Errorf("Error: Resource Unavailabe %s", err)
	}

	return res, nil
}

func (g *Graph) executeQuery(query string) (*client.Response, error) {

	res, err := g.httpClient.ExecuteQuery(query, g.dbname)
	if err != nil {
		return nil, fmt.Errorf("Error: Resource Unavailabe %s", err)
	}

	return res, nil
}

func (g *Graph) deleteContainerEvents(contextIDs []string) []Nodes {

	for _, node := range g.nodes {
		for _, contextID := range contextIDs {
			if node.ContextID == contextID {
				node.Delete = true
			}
		}
	}

	for k := len(g.nodes) - 1; k >= 0; k-- {
		if g.nodes[k].Delete {
			g.nodes = append(g.nodes[:k], g.nodes[k+1:]...)
		}
	}

	return g.nodes
}

// transform will convert the JSON response from influxdb to nodes and links to generate graph
// nodes struct will have nodeid, nodeipaddress and nodename
// links struct will have source, target and action
// the nodes are extracted from the influx data and stored in the array of structure
// then later this array is sent to the link generator which process the links between the nodes
// the link generator basically generates the link by comparing the nodeip with the flows src and dst ip's
func (g *Graph) transform(res *client.Response) (*GraphData, error) {
	zap.L().Info("Transforming to Nodes and Links")
	var node Nodes
	var contextID []string
	g.links = nil
	var startEvents = []string{ContainerUpdate}

	if len(res.Results[0].Series) > 0 {
		if res.Results[0].Series[0].Name == ContainerEvent {
			for _, containerEvent := range res.Results[0].Series[0].Values {
				if containerEvent[2] == ContainerUpdate {
					for k := range startEvents {
						if containerEvent[2].(string) == startEvents[k] {
							time, err := time.Parse(time.RFC3339, containerEvent[0].(string))
							if err != nil {
								return nil, fmt.Errorf("Error: Parsing Time %s", err)
							}
							node.Time = time
							node.ContextID = containerEvent[1].(string)
							node.IPAddress = containerEvent[5].(string)
							node.IPIDHash = getHash(containerEvent[1].(string), containerEvent[5].(string))
							if containerEvent[6].(string) != "" {
								node.Namespace = g.parseTag(containerEvent[6].(string), PODNamespaceFromContainerTags)
								node.PodName = g.parseTag(containerEvent[6].(string), PODNameFromContainerTags)
							}
							g.removeDuplicateNodes(node)
						}
					}
				} else if containerEvent[2].(string) == ContainerDelete {
					contextID = append(contextID, containerEvent[1].(string))
				}
			}
			if len(contextID) > 0 {
				g.deleteContainerEvents(contextID)
			}
		}
	}

	err := g.generateLinks()
	if err != nil {
		return nil, fmt.Errorf("Error: Generating Links %s", err)
	}

	jsonData := GraphData{Nodes: g.nodes, Links: g.links}

	return &jsonData, nil
}

func (g *Graph) generateLinks() error {

	res, err := g.getFlowEvents(g.httpClient, g.dbname)
	if err != nil {
		return fmt.Errorf("Error: Retrieving Flow Events %s", err)
	}

	var link Links
	var isSrcPod, isDstPod bool

	if len(res.Results[0].Series) > 0 {
		if res.Results[0].Series[0].Name == FlowEvent {
			for _, flowEvent := range res.Results[0].Series[0].Values {
				for _, node := range g.nodes {
					if node.IPIDHash == getHash(flowEvent[2].(string), flowEvent[5].(string)) {
						link.Target = node.ContextID
						isDstPod = true
					} else if node.IPIDHash == getHash(flowEvent[12].(string), flowEvent[13].(string)) {
						link.Source = node.ContextID
						isSrcPod = true
					}
				}
				if isSrcPod && isDstPod && link.Source != "" && link.Target != "" {
					link.Action = flowEvent[1].(string)
					link.Namespace = g.parseTag(flowEvent[16].(string), PODNamespaceFromFlowTags)
					time, err := time.Parse(time.RFC3339, flowEvent[0].(string))
					if err != nil {
						return fmt.Errorf("Error: Parsing Time %s", err)
					}
					link.Time = time
					g.removeDuplicateLinks(link)
					g.checkIfReject(link)
					isSrcPod = false
					isDstPod = false
				}
				link.Source = ""
				link.Target = ""
			}
		}
	}

	if len(g.links) == 0 {
		g.links = append(g.links, DefaultLink())
	}

	return nil
}

func (g *Graph) removeDuplicateNodes(node Nodes) []Nodes {
	var isNodePresent bool

	for l := range g.nodes {
		if g.nodes[l].IPIDHash == node.IPIDHash {
			isNodePresent = true
		}
	}
	if !isNodePresent {
		g.nodes = append(g.nodes, node)
	}

	return g.nodes
}

func (g *Graph) removeDuplicateLinks(targetLink Links) []Links {
	var isLinkPresent bool

	for _, link := range g.links {
		if link.Source == targetLink.Source && link.Target == targetLink.Target && link.Action == targetLink.Action {
			isLinkPresent = true
		}
	}

	if !isLinkPresent {
		g.links = append(g.links, targetLink)
	}

	return g.links
}

func (g *Graph) checkIfReject(targetLink Links) []Links {
	var rejectedLink Links

	for _, link := range g.links {
		if link.Source == targetLink.Source && link.Target == targetLink.Target && link.Action != targetLink.Action {
			if link.Action == FlowAccept && targetLink.Action == FlowReject {
				rejectedLink.Namespace = link.Namespace
				rejectedLink.Time = link.Time
				rejectedLink.Source = link.Source
				rejectedLink.Target = link.Target
				rejectedLink.Action = FlowNowRejected
			}
		}
	}

	if rejectedLink.Action != "" {
		g.links = g.removeDuplicateLinks(rejectedLink)
	}

	return g.links
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
