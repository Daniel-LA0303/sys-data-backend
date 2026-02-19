package organization_service

import (
	"context"
	"core/project/core/common/errors"
	organization_dto "core/project/core/internal/organization/dtos"
	organization_repository "core/project/core/internal/organization/repositories"
	"core/project/core/internal/users/auth"
	"database/sql"
	stdErrors "errors"
	"log"
)

type Service struct {
	repo *organization_repository.Repository
}

func NewService(repo *organization_repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetOrganizationSmallInfo(
	ctx context.Context,
	organizationId string,
) (*organization_dto.OrganizationInfoSmallDTO, error) {

	userID := auth.GetUserIDFromContext(ctx)
	log.Printf("userid from token conetxt is: %v\n", userID)

	// 1. Validate user
	/*if userID == "" {
		return nil, errors.NewUnauthorizedError("Unauthorized")
	}*/

	// 2. Validate input
	if organizationId == "" {
		return nil, errors.NewValidationError("OrganizationId is required")
	}

	// 3. Get organization
	org, err := s.repo.GetOrganizationById(ctx, organizationId)
	if err != nil {

		if stdErrors.Is(err, sql.ErrNoRows) {
			return nil, errors.NewNotFoundError("Organization")
		}

		return nil, errors.NewDatabaseError(err)
	}

	return org, nil
}

func (s *Service) NewOrganization(
	ctx context.Context,
	o organization_dto.OrganizationInfoSmallDTO,
) error {

	userID := auth.GetUserIDFromContext(ctx)

	// 1. validation user
	if userID == "" {
		return errors.NewUnauthorizedError("Unauthorized")
	}

	// 1. validations
	if o.OrgName == "" {
		return errors.NewValidationError("OrgName is required")
	}
	if o.OwnerUserId == "" {
		return errors.NewValidationError("OwnerUserId is required")
	}

	// 2. create organization
	if err := s.repo.CreateOrganization(ctx, o); err != nil {
		return errors.NewDatabaseError(err)
	}

	return nil
}

func (s *Service) UpdateOrganizationInfo(
	ctx context.Context,
	o organization_dto.OrganizationFullInfoDTO,
) error {

	userID := auth.GetUserIDFromContext(ctx)

	// 1. validation user
	if userID == "" {
		return errors.NewUnauthorizedError("Unauthorized")
	}

	// 2. validations
	if o.OrgId == "" {
		return errors.NewValidationError("OrgId is required")
	}

	if o.LegalName == "" {
		return errors.NewValidationError("Legal name is required")
	}

	if o.Industry == "" {
		return errors.NewValidationError("Industry is required")
	}

	existing, err := s.repo.GetOrganizationFullInfoById(ctx, o.OrgId)
	// 2. create organization
	if err != nil {
		if stdErrors.Is(err, sql.ErrNoRows) {
			// 3. if does not exists, then we create it
			log.Println("NO ROWS FOUND → creating settings")
			return s.repo.CreateOrganizationFullInfo(ctx, o)
		}
		log.Printf("DB ERROR: %v\n", err)
		return errors.NewDatabaseError(err)
	}

	_ = existing
	return s.repo.UpdateOrganizationFullInfo(ctx, o)
}

func (s *Service) GetDepartmentInfo(
	ctx context.Context,
	departmentId string,
) (*organization_dto.DeparmentInfoDTO, error) {

	userID := auth.GetUserIDFromContext(ctx)

	// 1. Validate user
	if userID == "" {
		return nil, errors.NewUnauthorizedError("Unauthorized")
	}

	// 2. Validate input
	if departmentId == "" {
		return nil, errors.NewValidationError("departmentId is required")
	}

	// 3. Get organization
	dept, err := s.repo.GetDepartmentById(ctx, departmentId)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	return dept, nil
}

func (s *Service) NewDepartment(
	ctx context.Context,
	o organization_dto.DeparmentInfoDTO,
) error {

	userID := auth.GetUserIDFromContext(ctx)

	// 1. validation user
	if userID == "" {
		return errors.NewUnauthorizedError("Unauthorized")
	}

	// 1. validations
	if o.OrgId == "" {
		return errors.NewValidationError("OrgId is required")
	}
	if o.DeptName == "" {
		return errors.NewValidationError("Name is required")
	}
	if o.ManagerUserId == "" {
		return errors.NewValidationError("ManagerUserId is required")
	}

	// 2. create organization
	if err := s.repo.CreateDepartment(ctx, o); err != nil {
		return errors.NewDatabaseError(err)
	}

	return nil
}
