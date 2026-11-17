package service

import (
	"context"

	procurementRequirementModel "github.com/suxiangdong/indu-scope-backend/internal/model/procurementrequirement"
)

type IProcurementRequirements interface {
	FormOptions(ctx context.Context, categoryID *uint64) (procurementRequirementModel.FormOptions, error)
	List(ctx context.Context, filter procurementRequirementModel.ListFilter) (procurementRequirementModel.ListResult, error)
	AdminList(ctx context.Context, filter procurementRequirementModel.AdminListFilter) (procurementRequirementModel.ListResult, error)
	AdminDetail(ctx context.Context, id uint64) (procurementRequirementModel.Detail, error)
	Review(ctx context.Context, adminUserID, id uint64, reviewStatus uint8, reason string) (procurementRequirementModel.Detail, error)
	Detail(ctx context.Context, userID uint64, role uint8, id uint64) (procurementRequirementModel.Detail, error)
	Create(ctx context.Context, input procurementRequirementModel.CreateInput) (procurementRequirementModel.RequirementItem, error)
	UpdateStatus(ctx context.Context, userID, id uint64, status uint8) (procurementRequirementModel.RequirementItem, error)
	CreateResponse(ctx context.Context, input procurementRequirementModel.ResponseInput) (procurementRequirementModel.ResponseItem, error)
	UpdateResponseStatus(ctx context.Context, userID, responseID uint64, status uint8) (procurementRequirementModel.ResponseItem, error)
	UploadFile(ctx context.Context, input procurementRequirementModel.FileUploadInput) (procurementRequirementModel.FileUploadResult, error)
}

var localProcurementRequirements IProcurementRequirements

func ProcurementRequirements() IProcurementRequirements {
	if localProcurementRequirements == nil {
		panic("implement not found for interface IProcurementRequirements, forgot register?")
	}
	return localProcurementRequirements
}

func RegisterProcurementRequirements(i IProcurementRequirements) {
	localProcurementRequirements = i
}
