package organization_dto

type OrganizationInfoSmallDTO struct {
	OrgId       string `json:"orgId" db:"org_id"`
	OrgName     string `json:"orgName" db:"org_name"`
	OwnerUserId string `json:"ownerUserId" db:"owner_user_id"`
	Status      bool   `json:"status" db:"status"`
	CreatedAt   string `json:"createdAt" db:"created_at"`
	UpdatedAt   string `json:"updatedAt" db:"updated_at"`
}

type OrganizationFullInfoDTO struct {
	OrgInfoId   string `json:"orgInfoId" db:"org_info_id"`
	OrgId       string `json:"orgId" db:"org_id"`
	LegalName   string `json:"legalName" db:"legal_name"`
	TaxId       string `json:"taxId" db:"tax_id"`
	Industry    string `json:"industry" db:"industry"`
	CompanySize string `json:"companySize" db:"company_size"`
	Address     string `json:"address" db:"address"`
	City        string `json:"city" db:"city"`
	State       string `json:"state" db:"state"`
	Country     string `json:"country" db:"country"`
	PostalCode  string `json:"postalCode" db:"postal_code"`
	Phone       string `json:"phone" db:"phone"`
	Website     string `json:"website" db:"website"`
	LogoUrl     string `json:"logoUrl" db:"logo_url"`
}

type DeparmentInfoDTO struct {
	DeptId        string `json:"deptId" db:"dept_id"`
	OrgId         string `json:"orgId" db:"org_id"`
	DeptName      string `json:"deptName" db:"dept_name"`
	Description   string `json:"description" db:"description"`
	ParentDeptId  string `json:"parentDeptId" db:"parent_dept_id"`
	IsActive      bool   `json:"isActive" db:"is_active"`
	ManagerUserId string `json:"managerUserId" db:"manager_user_id"`
	CreatedAt     string `json:"createdAt" db:"created_at"`
}
