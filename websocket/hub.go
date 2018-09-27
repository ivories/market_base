package websocket

import (
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	MsgType     int64
	Fd          int64
	PayLoad     []byte
	PayLoadType int
	Created     time.Time
}

// type EstablishConn struct {
// 	conn *Connection
// }
type CloseConn struct {
	fd int64
}

type ReceiveMsg struct {
	fd          int64
	payLoad     []byte
	payLoadType int
}

type SendMsg struct {
	fd          int64
	payLoad     []byte
	payLoadType int
}

type Hub struct {
	clients     sync.Map //并发map
	requestChan chan interface{}
	mq          Queue // 消息队列
	autoFd      int64
	rwmutex     *sync.RWMutex
}

func (h *Hub) run() {
	timer := time.Tick(time.Second * 30)
	for {
		select {
		case c := <-h.requestChan:
			switch object := c.(type) {
			// case EstablishConn: //建立长连接

			// 	h.createConn(object)

			case CloseConn: //断开长连接

				h.closeConn(object.fd)

			case ReceiveMsg: //接收客户端消息
				h.receiveMsg(object)

			case SendMsg:
				h.sendMsg(object)

			}
		case <-timer:
			numGoroutine := runtime.NumGoroutine()
			log.Printf("Goroutine %v", numGoroutine)

		}
	}
}

// //创建链接
// func (h *Hub) createConn(establishConn EstablishConn) {
// 	fd := h.createFd()
// 	h.clients.Store(fd, establishConn.conn)
// 	msg := h.newMessage(fd, MSG_ESTABLISH, []byte{}, 0)
// 	h.put(&msg)
// 	go establishConn.conn.writePump(h, fd)
// 	go establishConn.conn.readPump(h, fd)
// }

//接收客户端消息
func (h *Hub) receiveMsg(m ReceiveMsg) {
	messag := h.newMessage(m.fd, RECEIVE_MSG, m.payLoad, m.payLoadType)
	h.put(&messag)

}

//发送消息

func (h *Hub) sendMsg(m SendMsg) {
	value, ok := h.clients.Load(m.fd)
	if !ok {
		return
	}
	conn := value.(*Connection)
	// if isClosed(conn.send) {
	// 	log.Println("channle closed!")
	// 	go func() { h.requestChan <- CloseConn{fd: m.fd} }()
	// 	return
	// }
	msg := message{
		Type: websocket.TextMessage,
		Msg:  m.payLoad,
	}

	select {
	case conn.send <- msg:
	default:
		log.Printf("Message error \t\t%v", msg)
		go func() { h.requestChan <- CloseConn{fd: m.fd} }()

	}
}

func (h *Hub) newMessage(fd, msgType int64, payLoad []byte, payLoadType int) Message {

	return Message{
		MsgType:     msgType,
		Fd:          fd,
		PayLoad:     payLoad,
		PayLoadType: payLoadType,
		Created:     time.Now(),
	}
}

//入队
func (h *Hub) put(message *Message) {
	h.mq.Push(message)
}

//生成客户端id
func (h *Hub) createFd() int64 {
	h.autoFd++
	autoFd := h.autoFd
	return autoFd
}

//Close 关闭客户端
func (h *Hub) closeConn(fd int64) {

	value, ok := h.clients.Load(fd)
	if !ok {
		return
	}
	conn := value.(*Connection)
	err := conn.ws.Close()
	if err != nil {
		log.Println(err.Error())
	}
	h.clients.Delete(fd)
	msg := h.newMessage(fd, CONNECT_CLOSE, []byte{}, 0)
	h.put(&msg)

}

//判断channl 是否已经关闭
func isClosed(ch <-chan message) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}

//发送消息给指定客户端
func (h *Hub) SendMsg(fd int64, msgContent []byte, msgType int) {
	h.requestChan <- SendMsg{fd: fd, payLoad: msgContent, payLoadType: msgType}
}

//广播
func (h *Hub) SendBroadCastMsg(msgContent []byte, msgType int) {
	h.clients.Range(func(key, value interface{}) bool {

		h.SendMsg(key.(int64), msgContent, msgType)
		return true
	})
}

//主动断开客户端链接
func (h *Hub) CloseClinetConn(fd int64) {
	_, ok := h.clients.Load(fd)
	if !ok {
		return
	}
	go func() { h.requestChan <- CloseConn{fd: fd} }()

}

//获取链接数
func (h *Hub) GetOnlineCount() int {
	var slice []int64
	h.clients.Range(func(key, value interface{}) bool {
		slice = append(slice, key.(int64))
		return true
	})
	return len(slice)
}

//出队 需开放给外部
func (h *Hub) Pop() *Message {
	msg := h.mq.Pop()
	if msg != nil {
		return msg.(*Message)
	}
	return nil
}

//Demo
func (h *Hub) SendMsgByMultichat(fd int64, msgContent []byte, msgType int) {
	h.clients.Range(func(key, value interface{}) bool {
		my, ok := h.clients.Load(fd)
		if !ok {
			return true
		}
		myconn := my.(*Connection)
		conn := value.(*Connection)
		if myconn.addr.URL.Path == conn.addr.URL.Path {
			h.SendMsg(key.(int64), msgContent, msgType)
		}
		return true
	})
}
