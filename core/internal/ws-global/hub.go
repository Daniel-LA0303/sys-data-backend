package global_ws

import (
	"encoding/json"
	"log"
	"sync"
)

// Hub global — clave de canal: orgId (un canal por tenant)
type Hub struct {
	mu         sync.RWMutex
	orgs       map[string]map[*Client]bool // orgId -> set de clientes
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	notify     chan *NotifyMessage
}

type BroadcastMessage struct {
	OrgId   string
	Payload []byte
}

// NotifyMessage permite enviar un evento a un userId específico desde cualquier parte del backend
type NotifyMessage struct {
	OrgId   string
	UserId  string
	Payload []byte
}

// WSEvent es el formato que viaja por este WebSocket
type WSEvent struct {
	Type    string          `json:"type"`              // "presence" | "notification" | "ping"
	Payload json.RawMessage `json:"payload,omitempty"` // flexible según el type
}

// PresencePayload se usa para type="presence"
type PresencePayload struct {
	Event string     `json:"event"`           // "joined" | "left" | "snapshot"
	Users []UserInfo `json:"users,omitempty"` // snapshot al conectarse
	User  *UserInfo  `json:"user,omitempty"`  // joined/left
	Count int        `json:"count"`
}

type UserInfo struct {
	UserId   string `json:"userId"`
	Username string `json:"username"`
}

func NewHub() *Hub {
	return &Hub{
		orgs:       make(map[string]map[*Client]bool),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan *BroadcastMessage, 512),
		notify:     make(chan *NotifyMessage, 512),
	}
}

func (h *Hub) Run() {
	for {
		select {

		case client := <-h.register:
			h.mu.Lock()
			if h.orgs[client.orgId] == nil {
				h.orgs[client.orgId] = make(map[*Client]bool)
			}
			h.orgs[client.orgId][client] = true
			h.mu.Unlock()

			log.Printf("[GlobalHub] joined  org=%s user=%s", client.orgId, client.userId)

			// 1. send a snapshot message to he know that the client is connected
			h.sendSnapshot(client)

			// 2. Broadcast "joined" with only users from a org
			h.broadcastPresence(client.orgId, client, "joined", &UserInfo{
				UserId:   client.userId,
				Username: client.username,
			})

		case client := <-h.unregister:
			h.mu.Lock()
			if members, ok := h.orgs[client.orgId]; ok {
				delete(members, client)
				if len(members) == 0 {
					delete(h.orgs, client.orgId)
				}
			}
			h.mu.Unlock()
			close(client.send)

			log.Printf("[GlobalHub] left    org=%s user=%s", client.orgId, client.userId)

			// Broadcast "left" al resto de la org
			h.broadcastPresence(client.orgId, nil, "left", &UserInfo{
				UserId:   client.userId,
				Username: client.username,
			})

		case msg := <-h.broadcast:
			h.mu.RLock()
			members := h.orgs[msg.OrgId]
			h.mu.RUnlock()

			for client := range members {
				select {
				case client.send <- msg.Payload:
				default:
					h.unregister <- client
				}
			}

		case msg := <-h.notify:
			// we deliver only to the userId and org
			h.mu.RLock()
			members := h.orgs[msg.OrgId]
			h.mu.RUnlock()

			for client := range members {
				if client.userId == msg.UserId {
					select {
					case client.send <- msg.Payload:
					default:
						h.unregister <- client
					}
				}
			}
		}
	}
}

/*
websocket to send notifications
ws://localhost:8081/ws/global?orgId=xxx&token=xxx
*/
func (h *Hub) SendNotification(orgId, userId string, payload any) error {
	data, err := json.Marshal(WSEvent{
		Type:    "notification",
		Payload: mustMarshalRaw(payload),
	})
	if err != nil {
		return err
	}
	h.notify <- &NotifyMessage{OrgId: orgId, UserId: userId, Payload: data}
	return nil
}

// --- helpers internos ---

func (h *Hub) sendSnapshot(to *Client) {
	h.mu.RLock()
	members := h.orgs[to.orgId]
	h.mu.RUnlock()

	users := make([]UserInfo, 0, len(members))
	for c := range members {
		if c.userId == to.userId {
			continue // no incluirse a uno mismo
		}
		users = append(users, UserInfo{UserId: c.userId, Username: c.username})
	}

	payload := PresencePayload{
		Event: "snapshot",
		Users: users,
		Count: len(users),
	}
	event := WSEvent{Type: "presence", Payload: mustMarshalRaw(payload)}
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	select {
	case to.send <- data:
	default:
	}
}

func (h *Hub) broadcastPresence(orgId string, exclude *Client, event string, user *UserInfo) {
	h.mu.RLock()
	members := h.orgs[orgId]
	count := len(members)
	h.mu.RUnlock()

	payload := PresencePayload{Event: event, User: user, Count: count}
	wsEvent := WSEvent{Type: "presence", Payload: mustMarshalRaw(payload)}
	data, err := json.Marshal(wsEvent)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range members {
		if c == exclude {
			continue
		}
		select {
		case c.send <- data:
		default:
			h.unregister <- c
		}
	}
}

func mustMarshalRaw(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
