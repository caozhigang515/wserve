package wserve

import (
	"net/http"
)

type Message struct {
	rb      *Body
	cli     *Client
	Request *http.Request
}

func (m *Message) GetBody() *Body {
	return m.rb
}

func (m *Message) SendMessage(v interface{}) {
	m.rb.Message = v
	m.rb.To = m.rb.From
	//m.rb.From = "system"
	m.cli.hub.alone <- m
}

func (m *Message) SendMessageTo(u interface{}, v interface{}) {
	m.rb.Message = v
	m.rb.From = "system"
	m.rb.To = u
	m.cli.hub.alone <- m
}

func (m *Message) RadioMessage(v interface{}) {
	m.rb.Message = v
	m.rb.From = "system"
	m.cli.hub.broadcast <- m
}

func (m *Message) OffClient(v interface{}) {
	if v == nil {
		m.cli.closed()
		return
	}
	for client, _ := range m.cli.hub.clients {
		if client.u.compare(v) {
			client.closed()
		}
	}
}

func (m *Message) OnlineClients() interface{} {
	var list []IUser
	for client := range m.cli.hub.clients {
		list = append(list, client.u)
	}
	return list
}
