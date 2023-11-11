package golokia

import (
	"context"
	"fmt"
	"testing"
)

// Note that these test currently expect a jolokia java process to be
// running on 7025. Currently tested with a cassandra process2

var (
	host       = "localhost"
	port       = "8080"
	jolokia    = "jolokia-war-1.2.3"
	targetHost = "localhost"
	targetPort = "9999"
)

func makeClient() *Client {
	return &Client{
		BaseURL:  "http://127.0.0.1:37150/jolokia",
		Username: "hengwei",
		Password: "iug7wr6lksd4fg",
	}
}

func makeTarget() *Target {
	// "service:jmx:rmi:///jndi/rmi://" + targetHost + ":" + targetPort + "/jmxrmi"
	return nil
}

func TestClientListDomains(t *testing.T) {
	client := makeClient()
	target := makeTarget()

	domains, err := client.ListDomains(context.Background(), target)
	if err != nil {
		t.Errorf("err(%s) returned", err)
	}
	fmt.Println("Domains: ", domains)
}

func TestClientListBeans(t *testing.T) {
	client := makeClient()
	target := makeTarget()

	beans, err := client.ListBeans(context.Background(), target, "java.lang")
	if err != nil {
		t.Errorf("err(%s) returned", err)
	}
	fmt.Println("Beans: ", beans)
}

func TestClientExecuteOperation(t *testing.T) {
	client := makeClient()
	target := makeTarget()

	result, err := client.Exec(context.Background(), target, "java.lang:type=Threading", "getThreadInfo([J,boolean,boolean)",
		[]interface{}{[]int{153, 263}, true, true})
	if err != nil {
		t.Errorf("err(%s) returned", err)
	}
	t.Logf("Operation result: %#v", result)
}

func TestClientListProperties(t *testing.T) {
	client := makeClient()
	target := makeTarget()
	props, response, err := client.ListProperties(context.Background(), target, "java.lang:type=Threading", "")
	if err != nil {
		t.Errorf("err(%s), returned", err)
	}
	t.Logf("ListProperties result: %#v", props)
	t.Logf("ListProperties result: %#v", response)
}

// func TestClientGetAttr(t *testing.T) {
// 	client := makeClient()
// 	target := makeTarget()
// 	val, err := client.GetAttr("java.lang", []string{"type=Threading"}, "PeakThreadCount")
// 	if err != nil {
// 		t.Errorf("err(%s), returned", err)
// 		return
// 	}
// 	fmt.Println("Value:", val)
// }
