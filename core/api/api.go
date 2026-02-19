package api

import (
	organization_hanlders "core/project/core/internal/organization/hanlders"
	users_handler "core/project/core/internal/users/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type APIServer struct {
	addr        string
	userHandler *users_handler.UserHandler
	orgHandler  *organization_hanlders.OrgHandler
}

func NewApiServer(
	addr string,
	userHandler *users_handler.UserHandler,
	orgHandler *organization_hanlders.OrgHandler,
) *APIServer {
	return &APIServer{
		addr:        addr,
		userHandler: userHandler,
		orgHandler:  orgHandler,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()

	users_handler.RegisterUserRoutes(api, s.userHandler)
	organization_hanlders.RegisterOrgRoutes(api, s.orgHandler)

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
