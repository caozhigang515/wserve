package wserve

import (
	"fmt"
	"testing"
	"time"
)

func TestWServe(t *testing.T) {
	serve, hub := New(Debug())

	hub.UseOperate("test", func(message *Message) {
		fmt.Println("test 操作,", message.GetBody())
		message.SendMessage("qqwwee")
	})

	hub.UseOperate("close", func(message *Message) {
		message.OffClient(nil)
	})

	hub.UseOperate("list", func(message *Message) {
		message.SendMessage(message.OnlineClients())
	})

	hub.UseOperate("tik", func(message *Message) {
		time.Sleep(5 * time.Second)
		ticker := time.NewTicker(time.Second * 2)
		for {
			select {
			case <-ticker.C:
				message.SendMessageTo(234, "测试消息")
			}
		}
	})

	_ = serve.Run(":7897")
}
