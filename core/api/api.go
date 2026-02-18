package api

import (
	users_handler "core/project/core/internal/users/handlers"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type APIServer struct {
	addr        string
	userHandler *users_handler.UserHandler
}

func NewApiServer(addr string, userHandler *users_handler.UserHandler) *APIServer {
	return &APIServer{
		addr:        addr,
		userHandler: userHandler, // ← Ahora coincide
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()

	users_handler.RegisterUserRoutes(api, s.userHandler)

	log.Println("Server running on", s.addr)
	return http.ListenAndServe(s.addr, router)
}
