package wserve

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

type Client struct {
	o      *options
	u      IUser
	hub    *Hub
	conn   *websocket.Conn
	req    *http.Request
	writer http.ResponseWriter
	send   chan []byte
}

func (c *Client) closed() {
	c.hub.unregister <- c
}

func (c *Client) readHeartbeat() {
	defer func() {
		c.closed()
	}()
	c.conn.SetReadLimit(c.o.MaxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.o.ReadDeadline) * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(time.Duration(c.o.ReadDeadline) * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		var rb Body
		if err = json.Unmarshal(message, &rb); err != nil {
			fmt.Println("无效数据")
			continue
		}
		//rb.From = c.parse(rb.ITo)
		rb.IFrom = c.u
		rb.From = c.u.Major()
		msg := &Message{rb: &rb, cli: c, Request: c.req}
		switch rb.Type {
		case "system":
			c.hub.operate <- msg
		case "user":
			c.hub.alone <- msg
		case "radio":
			c.hub.broadcast <- msg
		}
	}
}

func (c *Client) writHeartbeat() {
	ticker := time.NewTicker((time.Duration(c.o.ReadDeadline) * time.Second * 9) / 10)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(time.Duration(c.o.WriteDeadline) * time.Second))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.send)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(time.Duration(c.o.WriteDeadline) * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
