package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivories/market_base/websocket"
)

var clients sync.Map //并发map
func ReadMq(h *websocket.Hub) {

	for {
		job := h.Pop()
		if job == nil {
			time.Sleep(1 * time.Millisecond)
			continue
		} else {
			switch job.MsgType {
			case websocket.CONNECT_ESTABLISH:
				clients.Store(job.Fd, true)

			case websocket.CONNECT_CLOSE:

				clients.Delete(job.Fd)

			case websocket.RECEIVE_MSG:
				h.SendMsgByMultichat(job.Fd, job.PayLoad, websocket.TextMessage)

			default:

			}

		}

	}
}
func main() {
	router := gin.Default()

	h := websocket.New()

	router.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "index.html")
	})
	router.GET("/channel/:name", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "chan.html")
	})
	router.GET("/channel/:name/ws", func(c *gin.Context) {
		c.Header("access-control-allow-credentials", "true")
		c.Header("Access-Control-Allow-Headers", "content-type, authorization, x-websocket-extensions, x-websocket-version, x-websocket-protocol")
		c.Header("access-control-allow-origin", "*")
		websocket.ServeWs(c.Writer, c.Request, h)
	})
	go ReadMq(h)

	router.Run(":5000")
}
