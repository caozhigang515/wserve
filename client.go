package wserve

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
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
				if c.o.DeBug {
					fmt.Println(err.Error())
				}
			}
			break
		}
		var rb Body
		if err = json.Unmarshal(message, &rb); err != nil {
			_ = c.conn.WriteJSON(Body{
				Message: err.Error(),
			})
			continue
		}
		rb.IFrom = c.u
		rb.From = c.u.Major()
		msg := &Message{rb: &rb, cli: c, Request: c.req}
		switch rb.Type {
		case "system":
			if c.o.Permissions == nil || c.o.Permissions(c.req, rb.Operate) {
				c.hub.operate <- msg
			} else {
				_ = c.conn.WriteJSON(Body{
					Message: "do not have permission",
				})
			}
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
