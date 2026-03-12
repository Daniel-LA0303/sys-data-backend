package main

import (
	"core/project/core/api"
	dbpost "core/project/core/db"
	chat_handlers "core/project/core/internal/chat-organization/handlers"
	chat_repository "core/project/core/internal/chat-organization/repositories"
	chat_services "core/project/core/internal/chat-organization/services"
	chat_ws "core/project/core/internal/chat-ws/handlers"
	organization_hanlders "core/project/core/internal/organization/hanlders"
	organization_repository "core/project/core/internal/organization/repositories"
	organization_service "core/project/core/internal/organization/services"
	module_users "core/project/core/internal/users/module"
	global_ws "core/project/core/internal/ws-global"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("Service starting...")

	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Construir DSN desde variables
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
		os.Getenv("DB_TIMEZONE"),
	)

	// 1. Conexión a DB (usando tu paquete db)
	conn := dbpost.ConnectPostgres(dsn)
	if conn == nil {
		log.Fatal("There was an error connecting to the database")
	}

	// 2. Inicialización de la cadena de dependencias

	orgRepo := organization_repository.NewRepository(conn)
	//userRepo := user_repository.NewRepository(conn)
	chatRepo := chat_repository.NewRepository(conn)

	//userServ := user_service.NewService(userRepo, orgRepo, conn)
	orgServ := organization_service.NewService(orgRepo)
	chatServ := chat_services.NewService(chatRepo, conn)

	orgHandler := organization_hanlders.NewOrgHandler(orgServ)
	//userHandler := user_handler.NewUserHandler(userServ)
	chatHandler := chat_handlers.NewUserHandler(chatServ)

	userModule := module_users.NewModule(
		conn,
		orgRepo, // <- can be orgModule.Repo if we create a module for organizations
	)

	// chat ws
	wsHub := chat_ws.NewHub()
	go wsHub.Run()

	// global ws
	globalHub := global_ws.NewHub()
	go globalHub.Run()

	// websocket to send chat messages
	wsHandler := chat_ws.NewWSHandler(wsHub, chatServ)

	// websocket to global state by org
	globalWsHandler := global_ws.NewWSHandler(globalHub)

	server := api.NewApiServer(
		":8081",
		//userHandler,
		userModule.Handler,
		orgHandler,
		chatHandler,
		wsHandler,
		globalWsHandler,
	)

	if err := server.Run(); err != nil {
		log.Fatal("Error starting the server:", err)
	}
}
