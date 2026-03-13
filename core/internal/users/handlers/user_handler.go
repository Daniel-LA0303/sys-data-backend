package users_handler

import (
	"core/project/core/common/errors"
	"core/project/core/common/response"
	"core/project/core/internal/users/auth"
	user_dto "core/project/core/internal/users/dtos"
	users_service "core/project/core/internal/users/services"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type UserHandler struct {
	service *users_service.Service
}

func NewUserHandler(service *users_service.Service) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var input user_dto.CreateUserRequestDTO

	// 1. get body
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, errors.NewValidationError("Invalid request body"))
		return
	}

	// 2. call service
	authData, err := h.service.RegisterUser(r.Context(), input)
	if err != nil {
		response.Error(w, err)
		return
	}

	// 3. build response
	response.Success(w, authData, "User registered successfully")
}

func (h *UserHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var input user_dto.LoginRequestDTO

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, errors.NewValidationError("Invalid request body"))
		return
	}

	authResponse, err := h.service.Login(r.Context(), input)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, authResponse, "Login successful")
}

func (h *UserHandler) IniviteUserHandler(w http.ResponseWriter, r *http.Request) {
	var input user_dto.InviteUserRequestDTO

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, errors.NewValidationError("Invalid request body"))
		return
	}

	log.Printf("hanlder invite")
	user, err := h.service.InviteUser(r.Context(), input)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.Success(w, user, "User invited successfully")
}

func (h *UserHandler) GetUserByEmailHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	user, err := h.service.GetUserByEmail(r.Context(), email)
	if err != nil {
		response.Error(w, err) // ← Manejo centralizado
		return
	}

	response.Success(w, user, "")
}

func (h *UserHandler) GetPaginatedUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 1. Parsear query params: /api/auth/users?page=1&limit=10
	query := r.URL.Query()

	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	// 2. Llamar al service
	users, err := h.service.GetUsers(r.Context(), page, limit)
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	// 3. Responder (puedes mandarlo vacío si no hay nada)
	if users == nil {
		users = []user_dto.UserResponseDTO{}
	}

	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) GetPaginatedUsersByOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	orgId := query.Get("orgId")
	page, _ := strconv.Atoi(query.Get("page"))
	limit, _ := strconv.Atoi(query.Get("limit"))

	users, total, err := h.service.GetUsersByOrganizationPaginated(
		r.Context(),
		page,
		limit,
		orgId,
	)

	if err != nil {
		http.Error(w, "Error fetching users paginated by org", http.StatusInternalServerError)
		return
	}

	if users == nil {
		users = []user_dto.UsersByOrganizationResponseDTO{}
	}

	response.SuccessPaginated(w, users, page, limit, total)
}

func (h *UserHandler) GetPaginatedUsersByOrganizationChatHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query()

	orgId := query.Get("orgId")

	users, err := h.service.GetUsersByOrganizationChatPaginated(
		r.Context(),
		orgId,
	)

	if err != nil {
		http.Error(w, "Error fetching users paginated by org", http.StatusInternalServerError)
		return
	}

	if users == nil {
		users = []user_dto.UsersByOrganizationResponseDTO{}
	}

	response.Success(w, users, "Users fetched successfully")
}

func (h *UserHandler) CreateOrUpdateInfoUserCustomSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var input user_dto.UpdateUserCustomSettingsDTO

	// 1. check if its a valid request
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, errors.NewValidationError("Invalid request body"))
		return
	}

	err := h.service.CreateOrUpdateUserCustomSettings(r.Context(), input)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, nil, "User settings updated successfully")
}

func (h *UserHandler) UpdateLanguagePreference(w http.ResponseWriter, r *http.Request) {
	var input user_dto.UpdateLanguagePreferenceDTO

	// 1. check if its a valid request
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, errors.NewValidationError("Invalid request body"))
		return
	}

	language, err := h.service.UpdateLanguagePreference(r.Context(), input.Language)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, language, "Update language preference successfully")
}

func (h *UserHandler) UpdateThemePreference(w http.ResponseWriter, r *http.Request) {
	var input user_dto.UpdateThemePreferenceDTO

	// 1. check if its a valid request
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, errors.NewValidationError("Invalid request body"))
		return
	}

	theme, err := h.service.UpdateThemePreference(r.Context(), input.Theme)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, theme, "Update theme preference successfully")
}

func (h *UserHandler) GetProfileHandler(w http.ResponseWriter, r *http.Request) {

	userID := auth.GetUserIDFromContext(r.Context())

	// SIN validación extra
	response.Success(w, map[string]string{
		"user_id": userID,
	}, "Access granted")
}

func (h *UserHandler) ProtectedTestHandler(w http.ResponseWriter, r *http.Request) {
	response.Success(w, map[string]string{
		"message": "You have access to this protected route",
	}, "")
}

func RegisterUserRoutes(r *mux.Router, handler *UserHandler) {
	r.HandleFunc("/auth/register", handler.RegisterUserHandler).Methods("POST")
	r.HandleFunc("/auth/login", handler.LoginHandler).Methods("POST")
	r.HandleFunc("/auth/user", handler.GetUserByEmailHandler).Methods("GET")
	r.HandleFunc("/auth/users", handler.GetPaginatedUsersHandler).Methods("GET")

	r.HandleFunc("/auth/me",
		auth.WithJWTAuth(handler.GetProfileHandler),
	).Methods("GET")

	r.HandleFunc("/auth/users/users-by-org",
		auth.WithJWTAuth(handler.GetPaginatedUsersByOrganizationHandler),
	).Methods("GET")

	r.HandleFunc("/auth/users/users-by-org-chat",
		auth.WithJWTAuth(handler.GetPaginatedUsersByOrganizationChatHandler),
	).Methods("GET")

	r.HandleFunc(
		"/auth/protected",
		auth.WithJWTAuth(handler.ProtectedTestHandler),
	).Methods("GET")

	r.HandleFunc(
		"/auth/user/invite-user",
		auth.WithJWTAuth(handler.IniviteUserHandler),
	).Methods("POST")

	// update user settings or create
	r.HandleFunc(
		"/auth/user/create-or-update-user-custom-settings",
		auth.WithJWTAuth(handler.CreateOrUpdateInfoUserCustomSettingsHandler),
	).Methods("PUT")
	r.HandleFunc(
		"/user/update-language-preference",
		auth.WithJWTAuth(handler.UpdateLanguagePreference),
	).Methods("POST")

	r.HandleFunc(
		"/user/update-theme-preference",
		auth.WithJWTAuth(handler.UpdateThemePreference),
	).Methods("POST")
}
