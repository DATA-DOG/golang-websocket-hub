package wshub

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
)

var (
	// Time allowed to write a message to the peer.
	WriteWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	PongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	PingPeriod = (PongWait * 9) / 10
	// Maximum message size allowed from peer.
	MaxMessageSize int64 = 64 * 1024
)

type subscription struct {
	Username   string
	Token      string
	connection *connection
}

type subscriber struct {
	connections map[*connection]bool
	Username    string
}

type connection struct {
	ws     *websocket.Conn
	send   chan []byte
	hub    *Hub
	closed bool
}

func (c *connection) close() {
	if !c.closed {
		if err := c.ws.Close(); err != nil {
			c.hub.log.Println("[DEBUG] websocket was already closed:", err)
		}
		close(c.send)
		c.closed = true
	}
}

func (c *connection) listenRead() {
	defer func() {
		c.hub.unregister <- c
		c.close()
	}()
	c.ws.SetReadLimit(MaxMessageSize)
	if err := c.ws.SetReadDeadline(time.Now().Add(PongWait)); err != nil {
		c.hub.log.Println("[ERROR] failed to set socket read deadline:", err)
	}
	c.ws.SetPongHandler(func(string) error {
		return c.ws.SetReadDeadline(time.Now().Add(PongWait))
	})
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			c.hub.log.Println("[DEBUG] read message error:", err)
			break
		}

		s := &subscription{connection: c}
		if err := json.Unmarshal(message, s); err != nil {
			c.hub.log.Println("[ERROR] invalid data sent for subscription:", string(message))
			continue
		}
		c.hub.subscribe <- s
	}
}

func (c *connection) listenWrite() {
	write := func(mt int, payload []byte) error {
		if err := c.ws.SetWriteDeadline(time.Now().Add(WriteWait)); err != nil {
			return err
		}
		return c.ws.WriteMessage(mt, payload)
	}
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				if err := write(websocket.CloseMessage, []byte{}); err != nil {
					c.hub.log.Println("[DEBUG] socket already closed:", err)
				}
				return
			}
			if err := write(websocket.TextMessage, message); err != nil {
				c.hub.log.Println("[DEBUG] failed to write socket message:", err)
				return
			}
		case <-ticker.C:
			if err := write(websocket.PingMessage, []byte{}); err != nil {
				c.hub.log.Println("[DEBUG] failed to ping socket:", err)
				return
			}
		}
	}
}
