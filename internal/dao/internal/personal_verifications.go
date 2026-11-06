// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// PersonalVerificationsDao is the data access object for the table personal_verifications.
type PersonalVerificationsDao struct {
	table    string                       // table is the underlying table name of the DAO.
	group    string                       // group is the database configuration group name of the current DAO.
	columns  PersonalVerificationsColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler           // handlers for customized model modification.
}

// PersonalVerificationsColumns defines and stores column names for the table personal_verifications.
type PersonalVerificationsColumns struct {
	Id                string //
	UserId            string //
	RealName          string //
	IdNumber          string //
	IdNumberEncrypted string //
	StorePlatformId   string //
	StoreUrl          string //
	IdFrontBucket     string //
	IdFrontEndpoint   string //
	IdFrontPath       string //
	IdBackBucket      string //
	IdBackEndpoint    string //
	IdBackPath        string //
	VerifiedAt        string //
	ReviewStatus      string //
	ReviewReason      string //
	ReviewedBy        string //
	ReviewedAt        string //
	CreatedAt         string //
	UpdatedAt         string //
}

// personalVerificationsColumns holds the columns for the table personal_verifications.
var personalVerificationsColumns = PersonalVerificationsColumns{
	Id:                "id",
	UserId:            "user_id",
	RealName:          "real_name",
	IdNumber:          "id_number",
	IdNumberEncrypted: "id_number_encrypted",
	StorePlatformId:   "store_platform_id",
	StoreUrl:          "store_url",
	IdFrontBucket:     "id_front_bucket",
	IdFrontEndpoint:   "id_front_endpoint",
	IdFrontPath:       "id_front_path",
	IdBackBucket:      "id_back_bucket",
	IdBackEndpoint:    "id_back_endpoint",
	IdBackPath:        "id_back_path",
	VerifiedAt:        "verified_at",
	ReviewStatus:      "review_status",
	ReviewReason:      "review_reason",
	ReviewedBy:        "reviewed_by",
	ReviewedAt:        "reviewed_at",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
}

// NewPersonalVerificationsDao creates and returns a new DAO object for table data access.
func NewPersonalVerificationsDao(handlers ...gdb.ModelHandler) *PersonalVerificationsDao {
	return &PersonalVerificationsDao{
		group:    "default",
		table:    "personal_verifications",
		columns:  personalVerificationsColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *PersonalVerificationsDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *PersonalVerificationsDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *PersonalVerificationsDao) Columns() PersonalVerificationsColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *PersonalVerificationsDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *PersonalVerificationsDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *PersonalVerificationsDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
