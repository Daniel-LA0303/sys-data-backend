package global_ws

import (
	auth "core/project/core/internal/users/auth"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WSHandler struct {
	hub *Hub
}

func NewWSHandler(hub *Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// ServeWS — endpoint: /ws/global?orgId=xxx&token=xxx
func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	orgId := r.URL.Query().Get("orgId")
	token := r.URL.Query().Get("token")

	if orgId == "" || token == "" {
		http.Error(w, "orgId and token are required", http.StatusBadRequest)
		return
	}

	// validate token
	claims, err := auth.ValidateJWT(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[GlobalWSHandler] upgrade error: %v", err)
		return
	}

	client := NewClient(h.hub, conn, claims.UserID, claims.Username, orgId)
	h.hub.register <- client

	go client.WritePump()
	go client.ReadPump()
}

func RegisterWSRoutes(r *mux.Router, handler *WSHandler) {
	r.HandleFunc("/ws/global", handler.ServeWS)
}
