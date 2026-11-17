package appcontent

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	appcontentModel "github.com/suxiangdong/indu-scope-backend/internal/model/appcontent"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

type sAppContent struct{}

var _ service.IAppContent = (*sAppContent)(nil)

func New() service.IAppContent { return &sAppContent{} }

func init() { service.RegisterAppContent(New()) }

func (s *sAppContent) List(ctx context.Context) (appcontentModel.ListResult, error) {
	var records []appcontentModel.PageContent
	if err := dao.AppPageContents.Ctx(ctx).OrderAsc("id").Scan(&records); err != nil {
		return appcontentModel.ListResult{}, gerror.Wrap(err, "list app page contents")
	}
	return appcontentModel.ListResult{Items: records}, nil
}

func (s *sAppContent) Get(ctx context.Context, pageKey string) (appcontentModel.PageContent, error) {
	pageKey = strings.TrimSpace(pageKey)
	if !validPageKey(pageKey) {
		return appcontentModel.PageContent{}, consts.ErrAppPageNotFound
	}
	var record appcontentModel.PageContent
	if err := dao.AppPageContents.Ctx(ctx).Cache(gdb.CacheOption{}).Where("page_key", pageKey).Scan(&record); err != nil {
		return appcontentModel.PageContent{}, gerror.Wrap(err, "get app page content")
	}
	if record.ID == 0 || record.Status != 1 {
		return appcontentModel.PageContent{}, consts.ErrAppPageNotFound
	}
	return record, nil
}

func (s *sAppContent) Update(ctx context.Context, input appcontentModel.UpdateInput) (appcontentModel.PageContent, error) {
	input.PageKey = strings.TrimSpace(input.PageKey)
	input.Title = strings.TrimSpace(input.Title)
	input.Summary = strings.TrimSpace(input.Summary)
	input.Body = strings.TrimSpace(input.Body)
	input.ContactPhone = strings.TrimSpace(input.ContactPhone)
	input.ContactEmail = strings.TrimSpace(input.ContactEmail)
	input.ServiceHours = strings.TrimSpace(input.ServiceHours)
	input.Address = strings.TrimSpace(input.Address)
	input.WechatAccount = strings.TrimSpace(input.WechatAccount)
	if !validPageKey(input.PageKey) {
		return appcontentModel.PageContent{}, &consts.InvalidParameterError{Message: "page_key is invalid"}
	}
	if input.Title == "" || len([]rune(input.Title)) > 100 {
		return appcontentModel.PageContent{}, &consts.InvalidParameterError{Message: "title is required and must be no longer than 100 characters"}
	}
	if input.Body == "" {
		return appcontentModel.PageContent{}, &consts.InvalidParameterError{Message: "body is required"}
	}
	if len([]rune(input.Summary)) > 500 || len([]rune(input.Body)) > 20000 {
		return appcontentModel.PageContent{}, &consts.InvalidParameterError{Message: "content length is invalid"}
	}
	if input.Status > 1 {
		return appcontentModel.PageContent{}, &consts.InvalidParameterError{Message: "status must be 0 or 1"}
	}
	result, err := dao.AppPageContents.Ctx(ctx).Cache(gdb.CacheOption{}).Where("page_key", input.PageKey).Data(g.Map{
		"title": input.Title, "summary": nullable(input.Summary), "body": input.Body,
		"contact_phone": nullable(input.ContactPhone), "contact_email": nullable(input.ContactEmail),
		"service_hours": nullable(input.ServiceHours), "address": nullable(input.Address),
		"wechat_account": nullable(input.WechatAccount), "status": input.Status, "updated_by": input.AdminUserID,
	}).Update()
	if err != nil {
		return appcontentModel.PageContent{}, gerror.Wrap(err, "update app page content")
	}
	if affected, err := result.RowsAffected(); err != nil {
		return appcontentModel.PageContent{}, gerror.Wrap(err, "check app page content update")
	} else if affected == 0 {
		return appcontentModel.PageContent{}, consts.ErrAppPageNotFound
	}
	return s.getAdmin(ctx, input.PageKey)
}

func (s *sAppContent) getAdmin(ctx context.Context, pageKey string) (appcontentModel.PageContent, error) {
	var record appcontentModel.PageContent
	if err := dao.AppPageContents.Ctx(ctx).Cache(gdb.CacheOption{}).Where("page_key", pageKey).Scan(&record); err != nil {
		return appcontentModel.PageContent{}, gerror.Wrap(err, "get updated app page content")
	}
	if record.ID == 0 {
		return appcontentModel.PageContent{}, consts.ErrAppPageNotFound
	}
	return record, nil
}

func validPageKey(key string) bool {
	return key == appcontentModel.PageKeyAboutUs || key == appcontentModel.PageKeyContact
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}
