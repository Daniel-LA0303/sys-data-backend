package chat_ws

import (
	"context"
	"log"
	"net/http"

	chat_dto "core/project/core/internal/chat-organization/dtos"
	chat_services "core/project/core/internal/chat-organization/services"
	auth "core/project/core/internal/users/auth"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// En producción valida el origen correctamente
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHandler struct {
	hub     *Hub
	service *chat_services.Service
}

func NewWSHandler(hub *Hub, service *chat_services.Service) *WSHandler {
	return &WSHandler{hub: hub, service: service}
}

// ServeWS es el endpoint HTTP que hace el upgrade a WebSocket.
// URL: /ws/chat/{roomId}?orgId=xxx
//
// El JWT viene en:
//   - Query param:  ?token=xxx   (más fácil desde el browser)
//   - O header:     Authorization: Bearer xxx
func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	roomId := vars["roomId"]
	orgId := r.URL.Query().Get("orgId")
	log.Printf("[WS] incoming connection roomId=%s orgId=%s", roomId, orgId)

	if roomId == "" || orgId == "" {
		http.Error(w, "roomId and orgId are required", http.StatusBadRequest)
		return
	}

	// --- Validar JWT ---
	// Intentamos leerlo del query param primero (WebSocket no soporta headers custom fácilmente)
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token is required", http.StatusUnauthorized)
		return
	}

	claims, err := auth.ValidateJWT(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	userId := claims.UserID
	username := claims.Username

	// --- Upgrade a WebSocket ---
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WSHandler] upgrade error: %v", err)
		return
	}

	client := NewClient(h.hub, conn, userId, username, orgId, roomId)
	h.hub.register <- client

	// saveFn persiste el mensaje y retorna el objeto enriquecido para broadcast
	saveFn := func(msg *WSMessage) (*WSMessage, error) {
		req := chat_dto.CreateMessageRequest{
			RoomId:      msg.RoomId,
			UserId:      msg.UserId,
			Content:     msg.Content,
			MessageType: msg.MessageType,
		}

		saved, err := h.service.CreateMessage(context.Background(), req)
		if err != nil {
			return nil, err
		}

		return &WSMessage{
			Type:        "chat",
			MessagesId:  saved.MessagesId,
			RoomId:      saved.RoomId,
			OrgId:       orgId,
			UserId:      saved.UserId,
			Username:    username,
			Content:     saved.Content,
			MessageType: saved.MessageType,
			CreatedAt:   saved.CreatedAt,
		}, nil
	}

	// ReadPump y WritePump corren en goroutines separadas (patrón gorilla estándar)
	go client.WritePump()
	go client.ReadPump(saveFn)
}

// RegisterWSRoutes registra la ruta WebSocket en el router
func RegisterWSRoutes(r *mux.Router, handler *WSHandler) {
	r.HandleFunc("/ws/chat/{roomId}", handler.ServeWS)
}
