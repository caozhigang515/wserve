package wserve

import (
	"github.com/gorilla/websocket"
	"net/http"
)

var (
	upgrader websocket.Upgrader
)

type WServe struct {
	hub *Hub
	o   *options
}

func (ws *WServe) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		panic(err)
	}

	var u IUser
	if ws.o.Certification != nil {
		u, err = ws.o.Certification(req)
		if err != nil {
			_ = conn.CloseHandler()(1008, err.Error())
			_ = conn.Close()
			return
		}
	}

	cli := &Client{
		o:      ws.o,
		u:      u,
		hub:    ws.hub,
		conn:   conn,
		req:    req,
		writer: w,
		send:   make(chan []byte, 256),
	}

	cli.hub.register <- cli
	go cli.readHeartbeat()
	go cli.writHeartbeat()

}

func (ws *WServe) Run(addr string) error {
	return http.ListenAndServe(addr, ws)
}

func New(opts ...Option) (*WServe, *Hub) {
	ws := &WServe{}
	ws.o = DefaultOptions()
	for _, opt := range opts {
		opt(ws.o)
	}
	ws.hub = newHub(ws.o)
	upgrader = websocket.Upgrader{
		ReadBufferSize:  ws.o.ReadBufferSize,
		WriteBufferSize: ws.o.WriteBufferSize,
		CheckOrigin:     ws.o.CheckOrigin,
	}
	go ws.hub.listen()
	return ws, ws.hub
}
