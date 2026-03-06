package chat_ws

import (
	"encoding/json"
	"log"
	"sync"
)

// Hub mantiene todas las rooms y gestiona broadcast multi-tenant.
// La clave del mapa es "orgId:roomId" para aislar tenants.
type Hub struct {
	mu         sync.RWMutex
	rooms      map[string]map[*Client]bool // roomKey -> set de clientes
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
}

type BroadcastMessage struct {
	RoomKey string // "orgId:roomId"
	Sender  *Client
	Payload []byte
}

// WSMessage es el formato que viaja por el WebSocket
type WSMessage struct {
	Type        string `json:"type"` // "chat" | "ping"
	MessagesId  string `json:"messagesId,omitempty"`
	RoomId      string `json:"roomId,omitempty"`
	OrgId       string `json:"orgId,omitempty"`
	UserId      string `json:"userId,omitempty"`
	Username    string `json:"username,omitempty"`
	Content     string `json:"content,omitempty"`
	MessageType string `json:"messageType,omitempty"` // "text" | "image" ...
	CreatedAt   string `json:"createdAt,omitempty"`
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan *BroadcastMessage, 512),
	}
}

// Run corre en su propia goroutine
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.rooms[client.roomKey] == nil {
				h.rooms[client.roomKey] = make(map[*Client]bool)
			}
			h.rooms[client.roomKey][client] = true
			h.mu.Unlock()
			log.Printf("[Hub] client joined  room=%s  user=%s", client.roomKey, client.userId)

		case client := <-h.unregister:
			h.mu.Lock()
			if members, ok := h.rooms[client.roomKey]; ok {
				delete(members, client)
				if len(members) == 0 {
					delete(h.rooms, client.roomKey)
				}
			}
			h.mu.Unlock()
			close(client.send)
			log.Printf("[Hub] client left    room=%s  user=%s", client.roomKey, client.userId)

		case msg := <-h.broadcast:
			h.mu.RLock()
			members := h.rooms[msg.RoomKey]
			h.mu.RUnlock()

			for client := range members {
				// No reenviar al propio emisor
				if client == msg.Sender {
					continue
				}
				select {
				case client.send <- msg.Payload:
				default:
					// Canal lleno → desconectar cliente lento
					h.unregister <- client
				}
			}
		}
	}
}

// Broadcast envía un mensaje a todos los miembros de una room (excepto el emisor)
func (h *Hub) Broadcast(roomKey string, sender *Client, msg *WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[Hub] marshal error: %v", err)
		return
	}
	h.broadcast <- &BroadcastMessage{
		RoomKey: roomKey,
		Sender:  sender,
		Payload: data,
	}
}
