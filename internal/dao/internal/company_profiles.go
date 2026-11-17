// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CompanyProfilesDao is the data access object for the table company_profiles.
type CompanyProfilesDao struct {
	table    string                 // table is the underlying table name of the DAO.
	group    string                 // group is the database configuration group name of the current DAO.
	columns  CompanyProfilesColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler     // handlers for customized model modification.
}

// CompanyProfilesColumns defines and stores column names for the table company_profiles.
type CompanyProfilesColumns struct {
	Id               string //
	UserId           string //
	IntroductionHtml string //
	StorePlatformId  string //
	StoreUrl         string //
	Offices          string //
	Factories        string //
	CreatedAt        string //
	UpdatedAt        string //
}

// companyProfilesColumns holds the columns for the table company_profiles.
var companyProfilesColumns = CompanyProfilesColumns{
	Id:               "id",
	UserId:           "user_id",
	IntroductionHtml: "introduction_html",
	StorePlatformId:  "store_platform_id",
	StoreUrl:         "store_url",
	Offices:          "offices",
	Factories:        "factories",
	CreatedAt:        "created_at",
	UpdatedAt:        "updated_at",
}

// NewCompanyProfilesDao creates and returns a new DAO object for table data access.
func NewCompanyProfilesDao(handlers ...gdb.ModelHandler) *CompanyProfilesDao {
	return &CompanyProfilesDao{
		group:    "default",
		table:    "company_profiles",
		columns:  companyProfilesColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CompanyProfilesDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CompanyProfilesDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CompanyProfilesDao) Columns() CompanyProfilesColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CompanyProfilesDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CompanyProfilesDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *CompanyProfilesDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
