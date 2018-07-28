package wsservice

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/samoslab/nebula/client/common"
	"github.com/samoslab/nebula/client/daemon"
	"github.com/sirupsen/logrus"
)

const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Poll file for changes with this period.
	filePeriod = 10 * time.Second
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type WSController struct {
	log  logrus.FieldLogger
	cm   **daemon.ClientManager
	quit chan struct{}
	done chan struct{}
}

func NewWSController(log logrus.FieldLogger, m **daemon.ClientManager) *WSController {
	c := &WSController{
		log:  log,
		quit: make(chan struct{}),
		done: make(chan struct{}),
		cm:   new(*daemon.ClientManager),
	}
	*c.cm = *m
	return c
}

// SetClientManager use http service client manager
func (c *WSController) SetClientManager(m **daemon.ClientManager) {
	*c.cm = *m
}

func (c *WSController) Shutdown() {
	close(c.quit)
	<-c.done
}

func (c *WSController) answerWriter(ws *websocket.Conn, msgType string) {
	log := c.log
	pingTicker := time.NewTicker(pingPeriod)
	defer func() {
		pingTicker.Stop()
		ws.Close()
	}()
	fmt.Printf(" client connect --------------\n")
	for {
		select {
		case <-c.quit:
			log.Info("shutdown answer writter")
			return

		case <-pingTicker.C:
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		case msg := <-(*c.cm).GetMsgChan():
			(*c.cm).DecreaseMsgCount()
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				return
			}
		}
	}
}

func (c *WSController) ServeWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	msgType := r.FormValue("type")

	go c.answerWriter(ws, msgType)
}

func (c *WSController) Consume() {
	log := c.log
	fileTicker := time.NewTicker(filePeriod)
	defer func() {
		fileTicker.Stop()
		close(c.done)
	}()
	for {
		select {
		case <-c.quit:
			log.Info("Shutdown message consumer")
			return
		case <-fileTicker.C:
			cnt := (*c.cm).GetMsgCount()
			if cnt > uint32(common.MsgQueueLen-common.MsgQueueLen+0) {
				for i := 0; i < int(cnt); i++ {
					select {
					case msg := <-(*c.cm).GetMsgChan():
						log.Infof("active consume msg %+v", msg)
						(*c.cm).DecreaseMsgCount()
					}
				}
			}
		}
	}
}

func (c *WSController) Run(addr string) error {
	http.HandleFunc("/message", c.ServeWs)
	var wg sync.WaitGroup
	errC := make(chan error)
	go c.Consume()
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := http.ListenAndServe(addr, nil); err != nil {
			errC <- err
			return
		}
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case err := <-errC:
		return err
	case <-c.quit:
		return nil
	case <-done:
		return nil
	}
}
