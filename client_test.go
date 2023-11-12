package golokia

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func makeClient() *Client {
	return &Client{
		BaseURL:  os.Getenv("golokia_url"),
		Username: os.Getenv("golokia_username"),
		Password: os.Getenv("golokia_password"),
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
