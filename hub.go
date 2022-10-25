package wserve

import "fmt"

type Hub struct {
	o          *options
	register   chan *Client
	unregister chan *Client
	// radio
	broadcast chan *Message
	// p2p
	alone chan *Message
	// operate
	operate chan *Message
	// all clients
	clients map[*Client]struct{}
	// operates collection
	operates map[string]Operate
}

func newHub(o *options) *Hub {
	return &Hub{
		o:          o,
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message),
		alone:      make(chan *Message),
		operate:    make(chan *Message),
		clients:    make(map[*Client]struct{}),
		operates:   make(map[string]Operate),
	}
}

func (h *Hub) UseOperate(operate string, handle Operate) {
	if _, ok := h.operates[operate]; ok {
		panic("repeatable operation.")
	}
	h.operates[operate] = handle
}

func (h *Hub) listen() {
	for {
		select {
		case client := <-h.register:
			if h.o.DeBug {
				fmt.Println("client register.", client.req.RemoteAddr)
			}
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			if h.o.DeBug {
				fmt.Println("client unregister.", client.req.RemoteAddr)
			}
			_ = client.conn.CloseHandler()(1000, "")
			_ = client.conn.Close()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case msg := <-h.broadcast:
			if h.o.DeBug {
				fmt.Println("broadcast message: ", string(msg.GetBody().Bytes()))
			}
			for client := range h.clients {
				select {
				case client.send <- msg.rb.Bytes():
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		case msg := <-h.alone:
			if h.o.DeBug {
				fmt.Println("p2p message: ", msg.rb.From, msg.rb.To, string(msg.GetBody().Bytes()))
			}
			for client := range h.clients {
				if client.u.compare(msg.rb.To) {
					select {
					case client.send <- msg.rb.Bytes():
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		case msg := <-h.operate:
			if h.o.DeBug {
				fmt.Println("system operate: ", msg.rb.Operate, msg.rb.Message)
			}
			handle, ok := h.operates[msg.rb.Operate]
			if !ok {
				go msg.SendMessage("no such operation")
				break
			}
			go handle(msg)
		}
	}
}
