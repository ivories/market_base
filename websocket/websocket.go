package websocket

import (
	"log"
	"net/http"
	_ "strings"
	"sync"
)

func ServeWs(w http.ResponseWriter, r *http.Request, h *Hub) {

	log.Printf("remote: %v ------------%v\n%v", r.RemoteAddr, r.Host, *r.URL)
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	c := &Connection{send: make(chan message, 256), recv: make(chan int), ws: ws, addr: r}
	// h.requestChan <- EstablishConn{conn: c}
	fd := h.createFd()
	h.clients.Store(fd, c)
	msg := h.newMessage(fd, CONNECT_ESTABLISH, []byte{}, 0)
	h.put(&msg)
	go c.writePump(h, fd)
	c.readPump(h, fd)

}
func newHub() *Hub {
	return &Hub{
		clients:     sync.Map{},
		requestChan: make(chan interface{}),
		mq:          Queue{}, // 消息队列
		autoFd:      START_FD,
		rwmutex:     &sync.RWMutex{},
	}
}

func New() *Hub {
	hub := newHub()
	go hub.run()
	return hub
}
