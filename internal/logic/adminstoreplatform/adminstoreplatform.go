package adminstoreplatform

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	"github.com/suxiangdong/indu-scope-backend/internal/model/do"
	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
	storePlatformModel "github.com/suxiangdong/indu-scope-backend/internal/model/storeplatform"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

type sAdminStorePlatform struct{}

var _ service.IAdminStorePlatform = (*sAdminStorePlatform)(nil)

func New() service.IAdminStorePlatform { return &sAdminStorePlatform{} }

func init() { service.RegisterAdminStorePlatform(New()) }

func (s *sAdminStorePlatform) List(ctx context.Context) (storePlatformModel.ListResult, error) {
	cols := dao.StorePlatforms.Columns()
	var records []*entity.StorePlatforms
	if err := dao.StorePlatforms.Ctx(ctx).Fields(cols.Id, cols.Name, cols.SortOrder, cols.Status, cols.CreatedAt, cols.UpdatedAt).
		OrderAsc(cols.SortOrder).OrderAsc(cols.Id).Scan(&records); err != nil {
		return storePlatformModel.ListResult{}, gerror.Wrap(err, "list store platforms")
	}
	items := make([]storePlatformModel.Item, 0, len(records))
	for _, record := range records {
		if record != nil {
			items = append(items, toItem(record))
		}
	}
	return storePlatformModel.ListResult{Items: items}, nil
}

func (s *sAdminStorePlatform) Create(ctx context.Context, input storePlatformModel.Input) (storePlatformModel.Item, error) {
	input = normalizeInput(input)
	if err := validateInput(input); err != nil {
		return storePlatformModel.Item{}, err
	}
	if count, err := dao.StorePlatforms.Ctx(ctx).Where(do.StorePlatforms{Name: input.Name}).Count(); err != nil {
		return storePlatformModel.Item{}, gerror.Wrap(err, "check store platform")
	} else if count > 0 {
		return storePlatformModel.Item{}, consts.ErrStorePlatformAlreadyExists
	}
	id, err := dao.StorePlatforms.Ctx(ctx).Data(g.Map{
		"name": input.Name, "sort_order": input.SortOrder, "status": input.Status,
	}).InsertAndGetId()
	if err != nil {
		return storePlatformModel.Item{}, gerror.Wrap(err, "create store platform")
	}
	return s.loadByID(ctx, uint64(id))
}

func (s *sAdminStorePlatform) Update(ctx context.Context, id uint64, input storePlatformModel.Input) (storePlatformModel.Item, error) {
	if id == 0 {
		return storePlatformModel.Item{}, &consts.InvalidParameterError{Message: "id is required"}
	}
	input = normalizeInput(input)
	if err := validateInput(input); err != nil {
		return storePlatformModel.Item{}, err
	}
	if count, err := dao.StorePlatforms.Ctx(ctx).Where(do.StorePlatforms{Id: id}).Count(); err != nil {
		return storePlatformModel.Item{}, gerror.Wrap(err, "query store platform")
	} else if count == 0 {
		return storePlatformModel.Item{}, consts.ErrStorePlatformNotFound
	}
	cols := dao.StorePlatforms.Columns()
	if count, err := dao.StorePlatforms.Ctx(ctx).Where(do.StorePlatforms{Name: input.Name}).Wheref("%s <> ?", cols.Id, id).Count(); err != nil {
		return storePlatformModel.Item{}, gerror.Wrap(err, "check store platform name")
	} else if count > 0 {
		return storePlatformModel.Item{}, consts.ErrStorePlatformAlreadyExists
	}
	if _, err := dao.StorePlatforms.Ctx(ctx).Where(do.StorePlatforms{Id: id}).Data(g.Map{
		"name": input.Name, "sort_order": input.SortOrder, "status": input.Status,
	}).Update(); err != nil {
		return storePlatformModel.Item{}, gerror.Wrap(err, "update store platform")
	}
	return s.loadByID(ctx, id)
}

func (s *sAdminStorePlatform) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return &consts.InvalidParameterError{Message: "id is required"}
	}
	if count, err := dao.StorePlatforms.Ctx(ctx).Where(do.StorePlatforms{Id: id}).Count(); err != nil {
		return gerror.Wrap(err, "query store platform")
	} else if count == 0 {
		return consts.ErrStorePlatformNotFound
	}
	if count, err := dao.PersonalVerifications.Ctx(ctx).Where(do.PersonalVerifications{StorePlatformId: id}).Count(); err != nil {
		return gerror.Wrap(err, "check store platform usage")
	} else if count > 0 {
		return consts.ErrStorePlatformInUse
	}
	if _, err := dao.StorePlatforms.Ctx(ctx).Where(do.StorePlatforms{Id: id}).Delete(); err != nil {
		return gerror.Wrap(err, "delete store platform")
	}
	return nil
}

func (s *sAdminStorePlatform) loadByID(ctx context.Context, id uint64) (storePlatformModel.Item, error) {
	cols := dao.StorePlatforms.Columns()
	var record *entity.StorePlatforms
	if err := dao.StorePlatforms.Ctx(ctx).Fields(cols.Id, cols.Name, cols.SortOrder, cols.Status, cols.CreatedAt, cols.UpdatedAt).
		Where(do.StorePlatforms{Id: id}).Scan(&record); err != nil {
		return storePlatformModel.Item{}, gerror.Wrap(err, "query store platform")
	}
	if record == nil {
		return storePlatformModel.Item{}, consts.ErrStorePlatformNotFound
	}
	return toItem(record), nil
}

func normalizeInput(input storePlatformModel.Input) storePlatformModel.Input {
	input.Name = strings.TrimSpace(input.Name)
	return input
}

func validateInput(input storePlatformModel.Input) error {
	if input.Name == "" || len([]rune(input.Name)) > 100 {
		return &consts.InvalidParameterError{Message: "name must contain 1 to 100 characters"}
	}
	if input.Status > 1 {
		return &consts.InvalidParameterError{Message: "status must be 0 or 1"}
	}
	return nil
}

func toItem(record *entity.StorePlatforms) storePlatformModel.Item {
	return storePlatformModel.Item{ID: record.Id, Name: record.Name, SortOrder: record.SortOrder, Status: uint8(record.Status), CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt}
}
