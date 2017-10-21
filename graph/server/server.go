package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/aporeto-inc/trireme-statistics/influxdb"
	"github.com/influxdata/influxdb/client/v2"
)

// GraphData is the struct that holds the json format required for graph to generate nodes and link
type GraphData struct {
	Nodes []Nodes `json:"nodes"`
	Links []Links `json:"links"`
}

// Nodes which holds pu information
type Nodes struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IPAddress string `json:"ipaddress"`
}

// Links which holds the links between pu's
type Links struct {
	Source int    `json:"source"`
	Target int    `json:"target"`
	Action string `json:"action"`
}

// GetData is called by the client which generates json with a logic that defines the nodes and links
func GetData(httpClient *influxdb.Influxdb, dbname string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		res, err := getContainerEvents(httpClient, dbname)
		if err != nil {
			http.Error(w, err.Error(), 0)
		}

		jsonData, err := transform(res, httpClient, dbname)
		if err != nil {
			http.Error(w, err.Error(), 2)
		}

		err = json.NewEncoder(w).Encode(jsonData)
		if err != nil {
			http.Error(w, err.Error(), 3)
		}
	}
}

// GetGraph is used to parse html with custom address to request for json
func GetGraph(w http.ResponseWriter, r *http.Request) {

	htmlData, err := template.New("graph").Parse(js)
	if err != nil {
		http.Error(w, err.Error(), 0)
	}

	data := struct {
		Address string
	}{
		Address: r.URL.Query().Get("address"),
	}

	err = htmlData.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), 1)
	}

	w.Header().Set("Content-Type", "text/html")
}

func getContainerEvents(httpClient *influxdb.Influxdb, dbname string) (*client.Response, error) {

	res, err := executeQuery("SELECT * FROM ContainerEvents", httpClient, dbname)
	if err != nil {
		return nil, fmt.Errorf("Error: Resource Unavailabe %s", err)
	}

	return res, nil
}

func getFlowEvents(httpClient *influxdb.Influxdb, dbname string) (*client.Response, error) {

	res, err := executeQuery("SELECT * FROM FlowEvents", httpClient, dbname)
	if err != nil {
		return nil, fmt.Errorf("Error: Resource Unavailabe %s", err)
	}

	return res, nil
}

func executeQuery(query string, httpClient *influxdb.Influxdb, dbname string) (*client.Response, error) {

	res, err := httpClient.ExecuteQuery(query, dbname)
	if err != nil {
		return nil, fmt.Errorf("Error: Resource Unavailabe %s", err)
	}

	return res, nil
}

func deleteContainerEvents(id []string, httpClient *influxdb.Influxdb, dbname string) ([]Nodes, error) {
	var node Nodes
	var nodes []Nodes

	for i := 0; i < len(id); i++ {
		_, err := executeQuery("DELETE FROM \"ContainerEvents\" WHERE \"EventID\" = '"+id[i]+"'", httpClient, dbname)
		if err != nil {
			return nil, fmt.Errorf("Error: Executing Query %s", err)
		}
	}

	res, _ := getContainerEvents(httpClient, dbname)

	for j := 0; j < len(res.Results[0].Series[0].Values); j++ {
		node.ID = res.Results[0].Series[0].Values[j][1].(string)
		if res.Results[0].Series[0].Values[j][6].(string) != "" {
			name := getName(res.Results[0].Series[0].Values[j][6].(string))
			node.Name = name
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

// transform will convert the JSON response from influxdb to nodes and links to generate graph
// nodes struct will have nodeid, nodeipaddress and nodename
// links struct will have source, target and action
// the nodes are extracted from the influx data and stored in the array of structure
// then later this array is sent to the link generator which process the links between the nodes
// the link generator basically generates the link by comparing the nodeip with the flows src and dst ip's
func transform(res *client.Response, httpClient *influxdb.Influxdb, dbname string) (*GraphData, error) {
	var nodes []Nodes
	var links []Links
	var node Nodes
	var err error
	var id []string
	var startEvents = []string{"start", "update", "create"}

	if len(res.Results[0].Series) > 0 {
		if res.Results[0].Series[0].Name == "ContainerEvents" {
			for j := 0; j < len(res.Results[0].Series[0].Values); j++ {
				if res.Results[0].Series[0].Values[j][2].(string) != "delete" {
					for k := 0; k < len(startEvents); k++ {
						if res.Results[0].Series[0].Values[j][2].(string) == startEvents[k] {
							node.ID = res.Results[0].Series[0].Values[j][1].(string)
							node.IPAddress = res.Results[0].Series[0].Values[j][5].(string)
							if res.Results[0].Series[0].Values[j][6].(string) != "" {
								name := getName(res.Results[0].Series[0].Values[j][6].(string))
								node.Name = name
							}
							nodes = append(nodes, node)
						}
					}
				} else {
					id = append(id, res.Results[0].Series[0].Values[j][1].(string))
					nodes, err = deleteContainerEvents(id, httpClient, dbname)
					if err != nil {
						return nil, fmt.Errorf("Error: Reading from Resoponse %s", err)
					}
				}
			}
		}
	}

	links, err = generateLinks(nodes, httpClient, dbname)
	if err != nil {
		return nil, fmt.Errorf("Error: Generating Links %s", err)
	}

	jsonData := GraphData{Nodes: nodes, Links: links}

	return &jsonData, nil
}

func generateLinks(nodea []Nodes, httpClient *influxdb.Influxdb, dbname string) ([]Links, error) {

	res, err := getFlowEvents(httpClient, dbname)
	if err != nil {
		return nil, fmt.Errorf("Error: Retriving Flow Events %s", err)
	}

	var links []Links
	var link Links
	var isSrc, isDst bool

	if len(res.Results[0].Series) > 0 {
		if res.Results[0].Series[0].Name == "FlowEvents" {
			for j := 0; j < len(res.Results[0].Series[0].Values); j++ {
				for i := 0; i < len(nodea); i++ {
					if nodea[i].IPAddress == res.Results[0].Series[0].Values[j][5] {
						link.Target = i
						isSrc = true
					} else if nodea[i].IPAddress == res.Results[0].Series[0].Values[j][13] {
						link.Source = i
						isDst = true
					}
				}
				if isSrc && isDst {
					link.Action = res.Results[0].Series[0].Values[j][1].(string)
					links = append(links, link)
					isSrc = false
					isDst = false
				}
			}
		}
	}

	if len(links) == 0 {
		link.Source = 0
		link.Target = 0
		links = append(links, link)
	}

	return links, nil
}

func getName(tag string) string {
	var name string

	if strings.Contains(tag, "@usr:io.kubernetes.pod.name") {
		eachTag := strings.Split(tag, " ")
		for i := 0; i < len(eachTag); i++ {
			podName := strings.SplitAfter(eachTag[i], "=")
			for j := 0; j < len(podName); j++ {
				if podName[j] == "@usr:io.kubernetes.pod.name=" {
					name = podName[1]
				}
			}
		}
	} else {
		eachTag := strings.Split(tag, " ")
		containerName := strings.SplitAfter(eachTag[0], "=")
		if containerName != nil {
			name = containerName[1]
		} else {
			name = "unknown"
		}
	}

	return name
}
