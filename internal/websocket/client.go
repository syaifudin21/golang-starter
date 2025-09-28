package websocket

import (
	"encoding/json"
	"exam/internal/dtos"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	maxMessageSize = 512
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	Room   *Room
	Conn   *websocket.Conn
	Send   chan []byte
	UserID uint
}

// ReadPump pumps messages from the websocket connection to the room.
func (c *Client) ReadPump() {
	defer func() {
		c.Room.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(maxMessageSize)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg dtos.WebsocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}

		inboundMsg := &InboundMessage{
			Client:  c,
			Type:    msg.Type,
			Payload: msg.Payload,
		}

		c.Room.Inbound <- inboundMsg
	}
}

// WritePump pumps messages from the Send channel to the websocket connection.
func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The room closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}