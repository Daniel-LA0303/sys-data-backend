package main

import (
	"core/project/core/api"
	dbpost "core/project/core/db"
	chat_handlers "core/project/core/internal/chat-organization/handlers"
	chat_repository "core/project/core/internal/chat-organization/repositories"
	chat_services "core/project/core/internal/chat-organization/services"
	organization_hanlders "core/project/core/internal/organization/hanlders"
	organization_repository "core/project/core/internal/organization/repositories"
	organization_service "core/project/core/internal/organization/services"
	user_handler "core/project/core/internal/users/handlers"
	user_repository "core/project/core/internal/users/repositories"
	user_service "core/project/core/internal/users/services"
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
	userRepo := user_repository.NewRepository(conn)
	chatRepo := chat_repository.NewRepository(conn)

	userServ := user_service.NewService(userRepo, orgRepo, conn)
	orgServ := organization_service.NewService(orgRepo)
	chatServ := chat_services.NewService(chatRepo)

	orgHandler := organization_hanlders.NewOrgHandler(orgServ)
	userHandler := user_handler.NewUserHandler(userServ)
	chatHandler := chat_handlers.NewUserHandler(chatServ)

	server := api.NewApiServer(
		":8081",
		userHandler,
		orgHandler,
		chatHandler,
	)

	if err := server.Run(); err != nil {
		log.Fatal("Error starting the server:", err)
	}
}
