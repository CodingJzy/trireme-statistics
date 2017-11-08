package server

import (
	"fmt"
	"testing"
	"time"

	"github.com/aporeto-inc/trireme-statistics/influxdb/mock"
	gomock "github.com/golang/mock/gomock"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/influxdata/influxdb/models"
	. "github.com/smartystreets/goconvey/convey"
)

func getSampleInlfuxDBResponse(eventType string) client.Response {
	var testResponse client.Response
	var testResultArray []client.Result
	var testResult client.Result
	var testRowArray []models.Row
	var testRow models.Row

	if eventType == ContainerEvent {
		testRow.Name = ContainerEvent
		testValues := make([][]interface{}, 2)
		testValues[0] = make([]interface{}, 7)
		testValues[1] = make([]interface{}, 7)
		testValues[0][0] = "2017-11-08T06:14:44.843219756Z"
		testValues[0][1] = "6f4b63dde673"
		testValues[0][2] = "update"
		testValues[0][5] = "10.20.0.1"
		testValues[0][6] = `&{[@sys:image=gcr.io/google_containers/pause-amd64:3.0 @sys:name=/k8s_POD_aporeto-collector-sp9v9_kube-system_1b326fb4-c44c-11e7-bcd7-42010a8001e2_0 @usr:annotation.kubernetes.io/created-by={"kind":"SerializedReference","apiVersion":"v1","reference":{"kind":"ReplicaSet","namespace":"kube-system","name":"aporeto-collector","uid":"1b3135f2-c44c-11e7-bcd7-42010a8001e2","apiVersion":"extensions","resourceVersion":"1220761"}}
 @usr:io.kubernetes.docker.type=podsandbox @usr:io.kubernetes.pod.name=aporeto-collector-sp9v9 @usr:io.kubernetes.pod.uid=1b326fb4-c44c-11e7-bcd7-42010a8001e2 @usr:annotation.kubernetes.io/config.seen=2017-11-08T06:14:41.101573506Z @usr:annotation.kubernetes.io/config.source=api @usr:app=aporeto-collector @usr:io.kubernetes.container.name=POD @usr:io.kubernetes.pod.namespace=kube-system]}]`
		testValues[1][0] = "2017-11-08T06:14:44.843219756Z"
		testValues[1][1] = "14138259f129"
		testValues[1][2] = "update"
		testValues[1][5] = "10.20.2.59"
		testValues[1][6] = `&{[@sys:image=gcr.io/google_containers/pause-amd64:3.0 @sys:name=/k8s_POD_aporeto-influxdb-j5hm6_kube-system_123de570-c44c-11e7-bcd7-42010a8001e2_0 @usr:annotation.kubernetes.io/config.seen=2017-11-08T06:14:26.074574104Z @usr:annotation.kubernetes.io/config.source=api @usr:io.kubernetes.container.name=POD @usr:io.kubernetes.pod.name=aporeto-influxdb-j5hm6 @usr:io.kubernetes.pod.uid=123de570-c44c-11e7-bcd7-42010a8001e2 @usr:annotation.kubernetes.io/created-by={"kind":"SerializedReference","apiVersion":"v1","reference":{"kind":"ReplicaSet","namespace":"kube-system","name":"aporeto-influxdb","uid":"123d2106-c44c-11e7-bcd7-42010a8001e2","apiVersion":"extensions","resourceVersion":"1220731"}}
 @usr:app=aporeto-influxdb @usr:io.kubernetes.docker.type=podsandbox @usr:io.kubernetes.pod.namespace=kube-system]}]]`
		testRow.Values = testValues
	} else if eventType == FlowEvent {
		testRow.Name = FlowEvent
		testValues := make([][]interface{}, 1)
		testValues[0] = make([]interface{}, 17)
		testValues[0][0] = "2017-11-08T06:14:46.314517734Z"
		testValues[0][12] = "6f4b63dde673"
		testValues[0][13] = "10.20.0.1"
		testValues[0][2] = "14138259f129"
		testValues[0][5] = "10.20.2.59"
		testValues[0][1] = "accept"
		testValues[0][16] = `&{[app=aporeto-influxdb @namespace=kube-system AporetoContextID=14138259f129]}]`
		testRow.Values = testValues
	}

	testRowArray = append(testRowArray, testRow)
	testResult.Series = testRowArray
	testResultArray = append(testResultArray, testResult)
	testResponse.Results = testResultArray

	return testResponse
}

func sampleGraphData(reverse bool) (*GraphData, []Node, []Link) {
	var testGraphData GraphData
	var srcNode Node
	var dstNode Node
	var link Link
	nodes := make([]Node, 2)
	links := make([]Link, 1)

	srcNode.ContextID = "6f4b63dde673"
	srcNode.IPAddress = "10.20.0.1"
	srcNode.Namespace = "kube-system"
	srcNode.PodName = "aporeto-collector-sp9v9"
	parsedTime, _ := time.Parse(time.RFC3339, "2017-11-08T06:14:44.843219756Z")
	srcNode.Time = parsedTime

	dstNode.ContextID = "14138259f129"
	dstNode.IPAddress = "10.20.2.59"
	dstNode.Namespace = "kube-system"
	dstNode.PodName = "aporeto-influxdb-j5hm6"
	parsedTime, _ = time.Parse(time.RFC3339, "2017-11-08T06:14:44.843219756Z")
	dstNode.Time = parsedTime

	if !reverse {
		nodes[0] = srcNode
		nodes[1] = dstNode
	} else {
		nodes[1] = srcNode
		nodes[0] = dstNode
	}

	link.Source = "6f4b63dde673"
	link.Target = "14138259f129"
	link.Action = "accept"
	link.Namespace = "kube-system"
	parsedTime, _ = time.Parse(time.RFC3339, "2017-11-08T06:14:46.314517734Z")
	link.Time = parsedTime

	links[0] = link

	testGraphData = GraphData{Nodes: nodes, Links: links}

	return &testGraphData, nodes, links
}

func TestNewGraph(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDataAdder := mockinfluxdb.NewMockDataAdder(ctrl)

	Convey("Given I create a new graph instance", t, func() {
		newTestGraph := NewGraph(mockDataAdder, "testDB")

		Convey("Then I should get a new graph instance", func() {
			So(newTestGraph, ShouldNotBeNil)
		})
	})
}

func TestGetContainerEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDataAdder := mockinfluxdb.NewMockDataAdder(ctrl)

	Convey("Given I create a new graph instance", t, func() {
		newTestGraph := NewGraph(mockDataAdder, "testDB")

		Convey("Given I try to get container events", func() {
			testContainerResponse := getSampleInlfuxDBResponse(ContainerEvent)
			mockDataAdder.EXPECT().ExecuteQuery(ContainerEventsQuery, "testDB").Return(&testContainerResponse, nil).Times(1)
			res, err := newTestGraph.getContainerEvents()
			Convey("I should get no error", func() {
				So(res, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})
		})

		Convey("Given I try to get container events with errors", func() {
			mockDataAdder.EXPECT().ExecuteQuery(ContainerEventsQuery, "testDB").Return(nil, fmt.Errorf("Error")).Times(1)
			res, err := newTestGraph.getContainerEvents()
			Convey("I should get an error", func() {
				So(res, ShouldBeNil)
				So(err, ShouldResemble, fmt.Errorf("Error: Executing Query Error: Resource Unavailabe Error"))
			})
		})
	})
}

func TestGetFlowEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDataAdder := mockinfluxdb.NewMockDataAdder(ctrl)

	Convey("Given I create a new graph instance", t, func() {
		newTestGraph := NewGraph(mockDataAdder, "testDB")

		Convey("Given I try to get flow events", func() {
			testFlowResponse := getSampleInlfuxDBResponse(FlowEvent)
			mockDataAdder.EXPECT().ExecuteQuery(FlowEventsQuery, "testDB").Return(&testFlowResponse, nil).Times(1)
			res, err := newTestGraph.getFlowEvents(mockDataAdder, "testDB")
			Convey("I should get no error", func() {
				So(res, ShouldNotBeNil)
				So(err, ShouldBeNil)
			})
		})

		Convey("Given I try to get flow events with errors", func() {
			mockDataAdder.EXPECT().ExecuteQuery(FlowEventsQuery, "testDB").Return(nil, fmt.Errorf("Error")).Times(1)
			res, err := newTestGraph.getFlowEvents(mockDataAdder, "testDB")
			Convey("I should get an error", func() {
				So(res, ShouldBeNil)
				So(err, ShouldResemble, fmt.Errorf("Error: Executing Query Error: Resource Unavailabe Error"))
			})
		})
	})
}

func TestDeleteContainerEvents(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDataAdder := mockinfluxdb.NewMockDataAdder(ctrl)

	Convey("Given I create a new graph instance", t, func() {
		newTestGraph := NewGraph(mockDataAdder, "testDB")

		Convey("Given I try to delete nodes", func() {
			var testNode Node
			testNode.ContextID = "testContextID"
			testNode.IPAddress = "10.1.1.0"
			testKey := getHash(testNode.ContextID, testNode.IPAddress)
			newTestGraph.nodeMap[testKey] = &testNode
			newTestGraph.deleteContainerEvents("testContextID")
			Convey("I should get see empty map", func() {
				So(len(newTestGraph.nodeMap), ShouldBeZeroValue)
			})
		})

		Convey("Given I try to delete nodes with invalid ID", func() {
			var testNode Node
			testNode.ContextID = "testContextID"
			testNode.IPAddress = "InvalidIP"
			testKey := getHash(testNode.ContextID, testNode.IPAddress)
			newTestGraph.nodeMap[testKey] = &testNode
			newTestGraph.deleteContainerEvents("InvalidContextID")
			Convey("I should have an entry in map", func() {
				So(len(newTestGraph.nodeMap), ShouldEqual, 1)
			})
		})
	})
}

func TestTransform(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDataAdder := mockinfluxdb.NewMockDataAdder(ctrl)

	Convey("Given I create a new graph instance", t, func() {
		newTestGraph := NewGraph(mockDataAdder, "testDB")

		Convey("Given I try to transform response form influxdb to nodes and links", func() {
			testContainerResponse := getSampleInlfuxDBResponse(ContainerEvent)
			testFlowResponse := getSampleInlfuxDBResponse(FlowEvent)
			mockDataAdder.EXPECT().ExecuteQuery(FlowEventsQuery, "testDB").Return(&testFlowResponse, nil).Times(1)
			res, err := newTestGraph.transform(&testContainerResponse)
			Convey("I should get no error", func() {
				testSampleGraphData, _, _ := sampleGraphData(false)
				if res.Nodes[0].ContextID == testSampleGraphData.Nodes[0].ContextID {
					So(res, ShouldResemble, testSampleGraphData)
					So(err, ShouldBeNil)
				} else {
					testSampleReverseGraphData, _, _ := sampleGraphData(true)
					So(res, ShouldResemble, testSampleReverseGraphData)
					So(err, ShouldBeNil)
				}
			})
		})

		Convey("Given I transform response form influxdb to nodes and links with errors", func() {
			testContainerResponse := getSampleInlfuxDBResponse(ContainerEvent)
			mockDataAdder.EXPECT().ExecuteQuery(FlowEventsQuery, "testDB").Return(nil, fmt.Errorf("Error")).Times(1)
			res, err := newTestGraph.transform(&testContainerResponse)
			Convey("I should get error", func() {
				So(res, ShouldBeNil)
				So(err, ShouldResemble, fmt.Errorf("Error: Generating Link Error: Retrieving Flow Events Error: Executing Query Error: Resource Unavailabe Error"))
			})
		})
	})
}

func TestClearDataStores(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDataAdder := mockinfluxdb.NewMockDataAdder(ctrl)

	Convey("Given I create a new graph instance", t, func() {
		newTestGraph := NewGraph(mockDataAdder, "testDB")

		Convey("Given I try to empty the data stores", func() {
			_, nodes, links := sampleGraphData(false)
			newTestGraph.nodes = nodes
			newTestGraph.links = links
			So(newTestGraph.nodes, ShouldNotBeNil)
			So(newTestGraph.links, ShouldNotBeNil)

			newTestGraph.clearDataStores()
			Convey("I should see empty data stores", func() {
				So(newTestGraph.nodes, ShouldBeNil)
				So(newTestGraph.links, ShouldBeNil)
				So(len(newTestGraph.linkMap), ShouldBeZeroValue)
			})
		})
	})
}

func TestParseTag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDataAdder := mockinfluxdb.NewMockDataAdder(ctrl)

	Convey("Given I create a new graph instance", t, func() {
		newTestGraph := NewGraph(mockDataAdder, "testDB")

		Convey("Given I try to parse tags", func() {
			testPodTag := `&{[@sys:image=gcr.io/google_containers/pause-amd64:3.0 @sys:name=/k8s_POD_aporeto-collector-sp9v9_kube-system_1b326fb4-c44c-11e7-bcd7-42010a8001e2_0 @usr:annotation.kubernetes.io/created-by={"kind":"SerializedReference","apiVersion":"v1","reference":{"kind":"ReplicaSet","namespace":"kube-system","name":"aporeto-collector","uid":"1b3135f2-c44c-11e7-bcd7-42010a8001e2","apiVersion":"extensions","resourceVersion":"1220761"}}
     @usr:io.kubernetes.docker.type=podsandbox @usr:io.kubernetes.pod.name=aporeto-collector-sp9v9 @usr:io.kubernetes.pod.uid=1b326fb4-c44c-11e7-bcd7-42010a8001e2 @usr:annotation.kubernetes.io/config.seen=2017-11-08T06:14:41.101573506Z @usr:annotation.kubernetes.io/config.source=api @usr:app=aporeto-collector @usr:io.kubernetes.container.name=POD @usr:io.kubernetes.pod.namespace=kube-system]}]`

			Convey("Given I try to parse pod name", func() {
				PodName := newTestGraph.parseTag(testPodTag, PODNameFromContainerTags)
				Convey("I should see pod name", func() {
					So(PodName, ShouldEqual, "aporeto-collector-sp9v9")
				})
			})

			Convey("Given I try to parse pod namespace", func() {
				PodNamespace := newTestGraph.parseTag(testPodTag, PODNamespaceFromContainerTags)
				Convey("I should see pod namespace", func() {
					So(PodNamespace, ShouldEqual, "kube-system")
				})
			})

			Convey("Given I try to parse flow namespace", func() {
				testFlowTag := `&{[app=aporeto-influxdb @namespace=kube-system AporetoContextID=14138259f129]}]`
				flowNamespace := newTestGraph.parseTag(testFlowTag, PODNamespaceFromFlowTags)
				Convey("I should see flow namespace", func() {
					So(flowNamespace, ShouldEqual, "kube-system")
				})
			})

			Convey("Given I try to parse invalid tag", func() {
				testFlowTag := "invalidTag"
				flowNamespace := newTestGraph.parseTag(testFlowTag, "InvalidTagType")
				Convey("I should see empty namespace", func() {
					So(flowNamespace, ShouldEqual, "")
				})
			})
		})
	})
}
