package main

import (
	"fmt"
	"github.com/dean2021/notify"
	"go.etcd.io/etcd/clientv3"
	"time"
)

func main() {

	// 连接ETcd
	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{
			"127.0.0.1:2379",
		},
		DialTimeout: time.Second * 30,
	})
	if err != nil {
		print(err)
	}

	n := notify.New("hids", client)

	// 唯一id
	uuid := "a21527cb7ea88402a7ec796447d9faa9"

	// 发送一个必须消费的指令
	err = n.SendTo(uuid, "upgrade", "xxx")
	if err != nil {
		print(err)
	}

	// 发送一个4秒后过期的指令
	err = n.SendToWithTTL(uuid, "upgrade", "xxx", 4)
	if err != nil {
		print(err)
	}

	// 接收并读取指令内容
	resp, err := n.RecvFrom(uuid, "upgrade")
	if err != nil {
		print(err)
	}
	fmt.Println(resp)

	// 循环接收uuid的指令
	n.RecvFromLoop(uuid, "upgrade", func(value string) {
		fmt.Println(value)
	})

	// 接收广播
	n.RecvBroadcast("register", func(event *clientv3.Event) {
		fmt.Println(event)
	})
}
