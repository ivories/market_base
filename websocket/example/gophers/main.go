package main

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ivories/market_base/websocket"
)

type GopherInfo struct {
	ID, X, Y string
}

var clients sync.Map //并发map
func ReadMq(h *websocket.Hub) {
	gophers := make(map[int64]*GopherInfo)
	counter := 1
	for {
		job := h.Pop()
		if job == nil {
			time.Sleep(1 * time.Millisecond)
			continue
		} else {
			switch job.MsgType {
			case websocket.CONNECT_ESTABLISH:
				clients.Store(job.Fd, true)

				for _, info := range gophers {
					h.SendMsg(job.Fd, []byte("set "+info.ID+" "+info.X+" "+info.Y), websocket.TextMessage)

				}

				gophers[job.Fd] = &GopherInfo{strconv.Itoa(counter), "0", "0"}
				h.SendMsg(job.Fd, []byte("iam "+gophers[job.Fd].ID), websocket.TextMessage)
				counter++

			case websocket.CONNECT_CLOSE:

				clients.Range(func(key, value interface{}) bool {
					fd := key.(int64)
					if fd != job.Fd {
						h.SendMsg(fd, []byte("dis "+gophers[job.Fd].ID), websocket.TextMessage)
					}
					return true
				})
				clients.Delete(job.Fd)
				delete(gophers, job.Fd)

			case websocket.RECEIVE_MSG:
				_, ok := clients.Load(job.Fd)
				if ok {
					p := strings.Split(string(job.PayLoad), " ")

					info := gophers[job.Fd]
					if len(p) == 2 {
						info.X = p[0]
						info.Y = p[1]
						clients.Range(func(key, value interface{}) bool {
							fd := key.(int64)
							if fd != job.Fd {
								h.SendMsg(fd, []byte("set "+info.ID+" "+info.X+" "+info.Y), websocket.TextMessage)
							}
							return true
						})
					}

				}

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

	router.GET("/ws", func(c *gin.Context) {
		c.Header("access-control-allow-credentials", "true")
		c.Header("Access-Control-Allow-Headers", "content-type, authorization, x-websocket-extensions, x-websocket-version, x-websocket-protocol")
		c.Header("access-control-allow-origin", "*")
		websocket.ServeWs(c.Writer, c.Request, h)
	})

	go ReadMq(h)

	router.Run(":5000")
}
