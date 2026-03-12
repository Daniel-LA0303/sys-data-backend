package chat_handlers

import (
	"core/project/core/common/response"
	chat_dto "core/project/core/internal/chat-organization/dtos"
	chat_services "core/project/core/internal/chat-organization/services"
	"core/project/core/internal/users/auth"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type ChatHandler struct {
	service *chat_services.Service
}

func NewUserHandler(service *chat_services.Service) *ChatHandler {
	return &ChatHandler{service: service}
}

func (h *ChatHandler) GetChatsByUserHandler(w http.ResponseWriter, r *http.Request) {

	// 1️Obtener query params
	query := r.URL.Query()
	userId := query.Get("userId")
	orgId := query.Get("orgId")

	// 2️Llamar al service
	rooms, err := h.service.GetChatsByUser(
		r.Context(),
		userId,
		orgId,
	)

	if err != nil {
		response.Error(w, err)
		return
	}

	// 3️Si es nil, enviar array vacío
	if rooms == nil {
		rooms = []chat_dto.ChatsRoomsByUser{}
	}

	// 4️Respuesta exitosa
	response.Success(w, rooms, "Chats fetched successfully")
}

func (h *ChatHandler) GetMessagesByRoomHandler(w http.ResponseWriter, r *http.Request) {

	// 1️Obtener query params
	query := r.URL.Query()
	roomId := query.Get("roomId")

	// 2️Llamar al service
	rooms, err := h.service.GetMessagesByRoom(
		r.Context(),
		roomId,
	)

	if err != nil {
		response.Error(w, err)
		return
	}

	// 3️Si es nil, enviar array vacío
	if rooms == nil {
		rooms = []chat_dto.ChatMessage{}
	}

	// 4️Respuesta exitosa
	response.Success(w, rooms, "Chats fetched successfully")
}

func (h *ChatHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req chat_dto.CreateMessageRequest

	// Decodificar JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validaciones básicas
	if req.RoomId == "" || req.UserId == "" || strings.TrimSpace(req.Content) == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	// Llamar al service
	res, err := h.service.CreateMessage(ctx, req)
	if err != nil {
		log.Printf("CREATE MESSAGE HANDLER ERROR: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Respuesta JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Printf("ENCODE RESPONSE ERROR: %v\n", err)
	}
}

func (h *ChatHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	var req chat_dto.CreateChatRoomRequest

	// 1️⃣ Decode request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, err)
		return
	}

	// 2️⃣ Call service
	res, err := h.service.CreateRoom(ctx, req)
	if err != nil {
		log.Printf("CREATE ROOM ERROR: %v\n", err)
		response.Error(w, err)
		return
	}

	// 3️⃣ Success response
	response.Success(w, res, "Chat room created successfully")
}
func RegisterChatRoutes(r *mux.Router, handler *ChatHandler) {

	r.HandleFunc("/chat/chats-by-user",
		auth.WithJWTAuth(handler.GetChatsByUserHandler),
	).Methods("GET")
	r.HandleFunc("/chat/messages-by-room",
		auth.WithJWTAuth(handler.GetMessagesByRoomHandler),
	).Methods("GET")

	r.HandleFunc("/chat/new-message",
		auth.WithJWTAuth(handler.CreateMessage),
	).Methods("POST")

	r.HandleFunc("/chat/create-room",
		auth.WithJWTAuth(handler.CreateRoom),
	).Methods("POST")
}
