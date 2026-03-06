package organization_repository

import (
	"context"
	organization_dto "core/project/core/internal/organization/dtos"
	"database/sql"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error // Para listas
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row              // Para RETURNING
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetOrganizationById(
	ctx context.Context,
	organizationId string,
) (*organization_dto.OrganizationInfoSmallDTO, error) {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		SELECT 
			org_id,
			org_name,
			owner_user_id,
			status,
			created_at,
			updated_at
		FROM organization_core_tbl
		WHERE org_id = $1
	`

	var settings organization_dto.OrganizationInfoSmallDTO

	err := r.db.GetContext(ctx, &settings, query, organizationId)
	if err != nil {
		log.Printf("GET ERROR: %v\n", err)
		return nil, err
	}

	return &settings, nil
}

func (r *Repository) CreateOrganization(ctx context.Context, db DBTX, o organization_dto.OrganizationInfoSmallDTO) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
        INSERT INTO organization_core_tbl (org_name, org_slug, owner_user_id, plan_id, status)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING org_id
    `

	var orgID string
	// Usamos QueryRowContext para obtener el valor del RETURNING
	err := db.QueryRowContext(ctx, query, o.OrgName, o.OrgSlug, o.OwnerUserId, o.PlanId, o.Status).Scan(&orgID)

	if err != nil {
		log.Printf("INSERT ORG ERROR: %v\n", err)
		return "", err
	}

	return orgID, nil
}

func (r *Repository) GetOrganizationFullInfoById(
	ctx context.Context,
	organizationInfoId string,
) (*organization_dto.OrganizationFullInfoDTO, error) {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		SELECT 
			org_info_id,
			org_id,
			legal_name,
			tax_id,
			industry,
			company_size,
			address,
			city,
			state,
			country,
			postal_code,
			phone,
			website,
			logo_url
		FROM organization_info_core_tbl
		WHERE org_id = $1
	`

	var settings organization_dto.OrganizationFullInfoDTO

	err := r.db.GetContext(ctx, &settings, query, organizationInfoId)
	if err != nil {
		log.Printf("GET ERROR: %v\n", err)
		return nil, err
	}

	return &settings, nil
}

func (r *Repository) CreateOrganizationFullInfo(
	ctx context.Context,
	o organization_dto.OrganizationFullInfoDTO,
) error {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO organization_info_core_tbl
		(
			org_id,
			legal_name,
			tax_id,
			industry,
			company_size,
			address,
			city,
			state,
			country,
			postal_code,
			phone,
			website,
			logo_url
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		o.OrgId,
		o.LegalName,
		o.TaxId,
		o.Industry,
		o.CompanySize,
		o.Address,
		o.City,
		o.State,
		o.Country,
		o.PostalCode,
		o.Phone,
		o.Website,
		o.LogoUrl,
	)

	if err != nil {
		log.Printf("INSERT ERROR: %v\n", err)
	}

	return err
}

func (r *Repository) UpdateOrganizationFullInfo(
	ctx context.Context,
	o organization_dto.OrganizationFullInfoDTO,
) error {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		UPDATE organization_info_core_tbl
		SET
			legal_name = $1,
			tax_id = $2,
			industry = $3,
			company_size = $4,
			address = $5,
			city = $6,
			state = $7,
			country = $8,
			postal_code = $9,
			phone = $10,
			website = $11,
			logo_url = $12
		WHERE org_id = $13
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		o.LegalName,
		o.TaxId,
		o.Industry,
		o.CompanySize,
		o.Address,
		o.City,
		o.State,
		o.Country,
		o.PostalCode,
		o.Phone,
		o.Website,
		o.LogoUrl,
		o.OrgId,
	)

	if err != nil {
		log.Printf("INSERT ERROR: %v\n", err)
	}

	return err
}

// departments
func (r *Repository) GetDepartmentById(
	ctx context.Context,
	organizationId string,
) (*organization_dto.DeparmentInfoDTO, error) {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		SELECT 
			dept_id,
			org_id,
			dept_name,
			description,
			parent_dept_id,
			is_active,
			manager_user_id,
			created_at
		FROM department_core_tbl
		WHERE org_id = $1
	`

	var settings organization_dto.DeparmentInfoDTO

	err := r.db.GetContext(ctx, &settings, query, organizationId)
	if err != nil {
		log.Printf("GET ERROR: %v\n", err)
		return nil, err
	}

	return &settings, nil
}

func (r *Repository) CreateDepartment(
	ctx context.Context,
	d organization_dto.DeparmentInfoDTO,
) error {

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	query := `
		INSERT INTO department_core_tbl
		(
			org_id,
			dept_name,
			description,
			parent_dept_id,
			is_active,
			manager_user_id,
			now()
		)
		VALUES ($1,$2,$3,$4,$5,$6)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		d.OrgId,
		d.DeptName,
		d.Description,
		d.ParentDeptId,
		d.IsActive,
		d.ManagerUserId,
	)

	if err != nil {
		log.Printf("INSERT ERROR: %v\n", err)
	}

	return err
}
