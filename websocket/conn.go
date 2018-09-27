package websocket

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

	maxMessageSize = 5120
)
const (
	TextMessage = 1

	BinaryMessage = 2

	CloseMessage = 8

	PingMessage = 9

	PongMessage = 10
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Connection struct {
	ws *websocket.Conn

	send chan message

	recv chan int

	addr *http.Request
}

func (c *Connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// 写消息给客户端
func (c *Connection) writePump(h *Hub, fd int64) {
	sendCloseConn := true
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()

		go func() {
			if sendCloseConn {
				h.requestChan <- CloseConn{fd: fd}
			}
		}()
	}()

WRITE_LOOP:
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.write(CloseMessage, []byte{})
				c.ws.Close()
				sendCloseConn = false
				break WRITE_LOOP
			}

			if err := c.write(message.Type, message.Msg); err != nil {
				log.Println("-------error!!!!!!!", err.Error())
				break WRITE_LOOP
			}

		case <-ticker.C:
			if err := c.write(PingMessage, []byte{}); err != nil {
				break WRITE_LOOP
			}
		}
	}
}

//接受客户端发来的消息
func (c *Connection) readPump(h *Hub, fd int64) {

	sendCloseConn := true
	defer func() {
		if sendCloseConn {
			go func() {

				h.requestChan <- CloseConn{fd: fd}

			}()
		}

	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
READ_LOOP:
	for {
		select {
		case <-c.recv:
			sendCloseConn = false
			break READ_LOOP
		default:
		}

		messageType, messageContent, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v-----unexpectedCloseError", err)
			}
			log.Printf("error: %v--------readMessageError", err)
			break
		}
		// 区分文本消息和二进制消息
		if messageType == TextMessage {
			go func() {
				h.requestChan <- ReceiveMsg{fd: fd, payLoad: messageContent, payLoadType: TextMessage}
			}()
		} else if messageType == BinaryMessage {
			go func() {
				h.requestChan <- ReceiveMsg{fd: fd, payLoad: messageContent, payLoadType: BinaryMessage}
			}()
		} else {
			log.Printf("errorType-----------")
			break
		}

	}

}
