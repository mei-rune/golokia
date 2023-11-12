package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/mei-rune/golokia"
	"github.com/mei-rune/httpdump"
)

func main() {
	var client golokia.Client
	var target golokia.Target

	flag.StringVar(&client.BaseURL, "url", "127.0.0.1", "jolokia 服务器的地址")
	flag.StringVar(&client.Username, "username", "", "jolokia 服务器的用户名")
	flag.StringVar(&client.Password, "password", "", "jolokia 服务器的密码")

	var targetHost string
	flag.StringVar(&targetHost, "target_host", "", "采集对象 jmx 的IP和端口")
	flag.StringVar(&target.Username, "target_username", "", "采集对象 jmx 的用户名")
	flag.StringVar(&target.Password, "target_password", "", "采集对象 jmx 的密码")

	var dir string
	flag.StringVar(&dir, "dir", ".", "采集数据保存的目录")

	flag.Parse()

	target.URL = "service:jmx:rmi:///jndi/rmi://" + targetHost + "/jmxrmi"

	httpdump.SetDebugProvider(httpdump.Dir(dir))
	ctx := context.Background()

	domains, err := client.ListDomains(ctx, &target)
	if err != nil {
		fmt.Println("ListDomains", err)
		return
	}

	for _, domain := range domains {
		beans, err := client.ListBeans(ctx, &target, domain)
		if err != nil {
			fmt.Println("ListBeans", err)
			return
		}
		for _, bean := range beans {
			response, _, err := client.ListProperties(ctx, &target, domain+":"+bean, "")
			if err != nil {
				fmt.Println("ListProperties", err)
				return
			}
			fmt.Printf("%#v\r\n", response)
		}
	}
}
