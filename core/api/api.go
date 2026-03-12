package api

import (
	chat_handlers "core/project/core/internal/chat-organization/handlers"
	chat_ws "core/project/core/internal/chat-ws/handlers"
	organization_hanlders "core/project/core/internal/organization/hanlders"
	users_handler "core/project/core/internal/users/handlers"
	global_ws "core/project/core/internal/ws-global"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type APIServer struct {
	addr            string
	userHandler     *users_handler.UserHandler
	orgHandler      *organization_hanlders.OrgHandler
	chatHandler     *chat_handlers.ChatHandler
	wsHandler       *chat_ws.WSHandler
	globalWsHandler *global_ws.WSHandler
}

func NewApiServer(
	addr string,
	userHandler *users_handler.UserHandler,
	orgHandler *organization_hanlders.OrgHandler,
	chatHandler *chat_handlers.ChatHandler,
	wsHandler *chat_ws.WSHandler,
	globalWsHandler *global_ws.WSHandler,
) *APIServer {
	return &APIServer{
		addr:            addr,
		userHandler:     userHandler,
		orgHandler:      orgHandler,
		chatHandler:     chatHandler,
		wsHandler:       wsHandler,
		globalWsHandler: globalWsHandler,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()

	users_handler.RegisterUserRoutes(api, s.userHandler)
	organization_hanlders.RegisterOrgRoutes(api, s.orgHandler)
	chat_handlers.RegisterChatRoutes(api, s.chatHandler)

	chat_ws.RegisterWSRoutes(router, s.wsHandler)

	global_ws.RegisterWSRoutes(router, s.globalWsHandler)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"}, // frontend
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	handler := c.Handler(router)

	log.Println("Server running on", s.addr)
	return http.ListenAndServe(s.addr, handler)
}
