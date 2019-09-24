// Copyright 2018 CSOIO.COM, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 基于ETcd实现的指令通讯

package notify

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"log"
	"strconv"
)

type Notify struct {
	root    string
	client  *clientv3.Client
	context context.Context
}

// 发送
// uuid 唯一id,command 指令, data 指令数据
func (n *Notify) SendTo(uuid string, command string, data string) error {
	path := fmt.Sprintf("/%s/notify/command/%s/%s", n.root, command, uuid)
	_, err := n.client.Put(n.context, path, data)
	if err != nil {
		return err
	}
	return err
}

// 发送
// uuid 唯一id,command 指令, data 指令数据, ttl 存活时间, keepAlive 持续保持
func (n *Notify) SendToWithTTL(uuid string, command string, data string, ttl int64, keepAlive bool) error {
	path := fmt.Sprintf("/%s/notify/command/%s/%s", n.root, command, uuid)
	resp, err := n.client.Grant(n.context, ttl)
	if err != nil {
		return err
	}
	_, err = n.client.Put(n.context, path, data, clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}

	if keepAlive {
		ch, err := n.client.KeepAlive(n.context, resp.ID)
		if err != nil {
			return err
		}

		for range ch {
		}
	}

	return err
}

// 接收
// uuid 唯一id, command 指令名
func (n *Notify) RecvFrom(uuid string, command string) ([]string, error) {
	path := fmt.Sprintf("/%s/notify/command/%s/%s", n.root, command, uuid)
	var response []string
	resp, err := n.client.Get(n.context, path, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	for _, ev := range resp.Kvs {
		response = append(response, string(ev.Value))
	}
	// TODO 此处待优化, 不够优雅
	_, _ = n.client.Delete(n.context, path, clientv3.WithPrefix())
	return response, err
}

// 持续接收
// uuid 唯一id, command 指令名
func (n *Notify) RecvFromLoop(uuid string, command string, EventFunc func(value string)) {
	path := fmt.Sprintf("/%s/notify/command/%s/%s", n.root, command, uuid)
	log.Println("监听:", path)
	rch := n.client.Watch(n.context, path, clientv3.WithPrefix(), clientv3.WithRev(n.getRevision()))
	for wResp := range rch {
		for _, ev := range wResp.Events {
			err := n.setRevision(ev.Kv.ModRevision + 1)
			if err != nil {
				log.Println(err)
			}
			if ev.Type == clientv3.EventTypePut {
				// TODO 此处待优化, 不够优雅
				_, _ = n.client.Delete(n.context, path, clientv3.WithPrefix())
				EventFunc(string(ev.Kv.Value))
			}
		}
	}
}

// 接收广播
func (n *Notify) RecvBroadcast(command string, EventFunc func(event *clientv3.Event)) {
	path := fmt.Sprintf("/%s/notify/command/%s", n.root, command)
	rch := n.client.Watch(n.context, path, clientv3.WithPrefix(), clientv3.WithRev(n.getRevision()))
	for wResp := range rch {
		for _, event := range wResp.Events {
			err := n.setRevision(event.Kv.ModRevision + 1)
			if err != nil {
				log.Println(err)
			}
			EventFunc(event)
		}
	}
}

// 发送广播
func (n *Notify) SendBroadcast(command string, data string, ttl int64) error {
	path := fmt.Sprintf("/%s/notify/command/%s", n.root, command)
	resp, err := n.client.Grant(n.context, ttl)
	if err != nil {
		return err
	}
	_, err = n.client.Put(n.context, path, data, clientv3.WithLease(resp.ID))
	return err
}

// 保存当前命令版本
func (n *Notify) setRevision(revision int64) error {
	path := fmt.Sprintf("/%s/notify/revision", n.root)
	_, err := n.client.Put(n.context, path, strconv.FormatInt(revision, 10))
	return err
}

// 获取命令版本
func (n *Notify) getRevision() int64 {
	var curRevision int64
	path := fmt.Sprintf("/%s/notify/revision", n.root)
	resp, _ := n.client.Get(context.Background(), path)
	// 当保存的revision为空情下,用最新revision
	if len(resp.Kvs) == 0 {
		return resp.Header.Revision + 1
	}
	for _, ev := range resp.Kvs {
		curRevision = BytesToInt64(ev.Value)
	}
	return curRevision + 1
}

func New(root string, client *clientv3.Client) *Notify {
	return &Notify{
		root:    root,
		context: context.Background(),
		client:  client,
	}
}
