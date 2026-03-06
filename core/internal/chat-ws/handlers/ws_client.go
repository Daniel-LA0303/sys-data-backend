package chat_ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

// Client representa una conexión WebSocket activa
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte

	// Identidad del cliente
	userId   string
	username string
	orgId    string
	roomId   string
	roomKey  string // "orgId:roomId"
}

func NewClient(hub *Hub, conn *websocket.Conn, userId, username, orgId, roomId string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		userId:   userId,
		username: username,
		orgId:    orgId,
		roomId:   roomId,
		roomKey:  orgId + ":" + roomId,
	}
}

// ReadPump lee mensajes entrantes del WebSocket y los manda al hub para broadcast.
// Debe correr en una goroutine propia.
func (c *Client) ReadPump(saveFn func(msg *WSMessage) (*WSMessage, error)) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[Client] read error user=%s: %v", c.userId, err)
			}
			break
		}

		var incoming WSMessage
		if err := json.Unmarshal(raw, &incoming); err != nil {
			log.Printf("[Client] unmarshal error: %v", err)
			continue
		}

		// Ignorar pings de keepalive del cliente
		if incoming.Type == "ping" {
			continue
		}

		// Forzar los datos de identidad desde la conexión (no confiar en el payload)
		incoming.Type = "chat"
		incoming.UserId = c.userId
		incoming.Username = c.username
		incoming.RoomId = c.roomId
		incoming.OrgId = c.orgId

		// Persistir en DB y obtener el mensaje con ID y timestamps
		saved, err := saveFn(&incoming)
		if err != nil {
			log.Printf("[Client] save error: %v", err)
			continue
		}

		// 1. Enviar el mensaje guardado de vuelta al propio emisor
		if data, err := json.Marshal(saved); err == nil {
			select {
			case c.send <- data:
			default:
			}
		}

		// 2. Broadcast al resto de la room
		c.hub.Broadcast(c.roomKey, c, saved)
	}
}

// WritePump escribe mensajes desde el canal send hacia la conexión WebSocket.
// Debe correr en una goroutine propia.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("[Client] write error user=%s: %v", c.userId, err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
