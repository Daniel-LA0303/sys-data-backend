package module_users

import (
	user_handler "core/project/core/internal/users/handlers"
	user_repository "core/project/core/internal/users/repositories"
	user_service "core/project/core/internal/users/services"

	organization_repository "core/project/core/internal/organization/repositories"

	"github.com/jmoiron/sqlx"
)

type Module struct {
	Handler *user_handler.UserHandler
}

func NewModule(db *sqlx.DB, orgRepo *organization_repository.Repository) *Module {

	userRepo := user_repository.NewRepository(db)

	service := user_service.NewService(
		userRepo,
		orgRepo,
		db,
	)

	handler := user_handler.NewUserHandler(service)

	return &Module{
		Handler: handler,
	}
}
