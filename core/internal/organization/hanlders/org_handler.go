package organization_hanlders

import (
	"core/project/core/common/errors"
	"core/project/core/common/response"
	organization_dto "core/project/core/internal/organization/dtos"
	organization_service "core/project/core/internal/organization/services"
	"core/project/core/internal/users/auth"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type OrgHandler struct {
	service *organization_service.Service
}

func NewOrgHandler(service *organization_service.Service) *OrgHandler {
	return &OrgHandler{service: service}
}

func (h *OrgHandler) GetOrgHandler(w http.ResponseWriter, r *http.Request) {

	orgId := r.URL.Query().Get("orgId")

	org, err := h.service.GetOrganizationSmallInfo(r.Context(), orgId)
	if err != nil {
		response.Error(w, err) // ← Manejo centralizado
		return
	}

	response.Success(w, org, "Info retrieved successfully")

}

/*func (h *OrgHandler) NewOrgHandler(w http.ResponseWriter, r *http.Request) {

	var input organization_dto.OrganizationInfoSmallDTO

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, errors.NewValidationError("Invalid request body"))
		return
	}

	if err := h.service.NewOrganization(r.Context(), input); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, nil, "Organization created successfully")
}*/

func (h *OrgHandler) UpdateOrganizationInfo(w http.ResponseWriter, r *http.Request) {

	var input organization_dto.OrganizationFullInfoDTO

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, errors.NewValidationError("Invalid request body"))
		return
	}

	if err := h.service.UpdateOrganizationInfo(r.Context(), input); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, nil, "Organization information updated successfully")
}

func (h *OrgHandler) GetDepartmentInfo(w http.ResponseWriter, r *http.Request) {

	departmentId := r.URL.Query().Get("departmentId")

	dept, err := h.service.GetDepartmentInfo(r.Context(), departmentId)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, dept, "Department retrieved successfully")
}

func (h *OrgHandler) NewDepartment(w http.ResponseWriter, r *http.Request) {

	var input organization_dto.DeparmentInfoDTO

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, errors.NewValidationError("Invalid request body"))
		return
	}

	if err := h.service.NewDepartment(r.Context(), input); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, nil, "Department created successfully")
}

func RegisterOrgRoutes(r *mux.Router, handler *OrgHandler) {
	r.HandleFunc("/org/get-org-by-id", auth.WithJWTAuth(handler.GetOrgHandler)).Methods("GET")
	//r.HandleFunc("/org/new-org", auth.WithJWTAuth(handler.NewOrgHandler)).Methods("POST")
	r.HandleFunc("/org/update-org", handler.UpdateOrganizationInfo).Methods("PUT")
	r.HandleFunc("/org/get-department-by-id", handler.GetDepartmentInfo).Methods("GET")
	r.HandleFunc("/org/new-department", handler.NewDepartment).Methods("POST")

}
