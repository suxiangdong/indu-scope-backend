package procurementrequirements

import (
	"context"

	v1 "github.com/suxiangdong/indu-scope-backend/api/procurementrequirements/v1"
)

type IProcurementRequirementsV1 interface {
	FormOptions(ctx context.Context, req *v1.FormOptionsReq) (res *v1.FormOptionsRes, err error)
	List(ctx context.Context, req *v1.ListReq) (res *v1.ListRes, err error)
	Detail(ctx context.Context, req *v1.DetailReq) (res *v1.DetailRes, err error)
	Create(ctx context.Context, req *v1.CreateReq) (res *v1.CreateRes, err error)
	UpdateStatus(ctx context.Context, req *v1.UpdateStatusReq) (res *v1.UpdateStatusRes, err error)
	UploadFile(ctx context.Context, req *v1.UploadFileReq) (res *v1.UploadFileRes, err error)
	CreateResponse(ctx context.Context, req *v1.CreateResponseReq) (res *v1.CreateResponseRes, err error)
	UpdateResponseStatus(ctx context.Context, req *v1.UpdateResponseStatusReq) (res *v1.UpdateResponseStatusRes, err error)
}
