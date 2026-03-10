package response

// common/response/api_response.go

import (
	"encoding/json"
	"net/http"

	"core/project/core/common/errors"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
	Type    string      `json:"type"`
	Message string      `json:"message"`
	Fields  interface{} `json:"fields,omitempty"`
}

// Success response
func Success(w http.ResponseWriter, data interface{}, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

// Error response
func Error(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	// Si es AppError, usamos su info
	if appErr, ok := err.(*errors.AppError); ok {
		w.WriteHeader(appErr.StatusCode)

		response := APIResponse{
			Success: false,
			Error: &ErrorInfo{
				Type:    string(appErr.Type),
				Message: appErr.Message,
				Fields:  appErr.Fields,
			},
		}

		json.NewEncoder(w).Encode(response)
		return
	}

	// Error genérico
	w.WriteHeader(http.StatusInternalServerError)
	response := APIResponse{
		Success: false,
		Error: &ErrorInfo{
			Type:    string(errors.ErrorTypeInternal),
			Message: "An unexpected error occurred",
		},
	}

	json.NewEncoder(w).Encode(response)
}

// Paginación
func SuccessPaginated(w http.ResponseWriter, data interface{}, page, limit, total int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := struct {
		APIResponse
		Pagination PaginationInfo `json:"pagination"`
	}{
		APIResponse: APIResponse{
			Success: true,
			Data:    data,
		},
		Pagination: PaginationInfo{
			Page:       page,
			Limit:      limit,
			TotalCount: total,
			TotalPages: (total + limit - 1) / limit,
		},
	}

	json.NewEncoder(w).Encode(response)
}

type PaginationInfo struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalCount int `json:"total_count"`
	TotalPages int `json:"total_pages"`
}
