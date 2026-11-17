package vip

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
)

type agreementRow struct {
	ID        uint64    `orm:"id"`
	Key       string    `orm:"agreement_key"`
	Title     string    `orm:"title"`
	Content   string    `orm:"content"`
	Status    uint8     `orm:"status"`
	Version   uint      `orm:"version"`
	UpdatedAt time.Time `orm:"updated_at"`
}

// ListAgreements returns only the newest version for each supported agreement
// key. Historical rows remain in vip_agreements for order auditing and future
// administration, but are never shown as current agreements.
func (s *sVip) ListAgreements(ctx context.Context, activeOnly bool) ([]vipModel.Agreement, error) {
	items := make([]vipModel.Agreement, 0, 2)
	for _, key := range []string{vipModel.AgreementKeyService, vipModel.AgreementKeyAutoRenew} {
		item, err := s.latestAgreement(ctx, key, activeOnly)
		if err != nil {
			return nil, err
		}
		if item != nil {
			items = append(items, *item)
		}
	}
	return items, nil
}

func (s *sVip) latestAgreement(ctx context.Context, key string, activeOnly bool) (*vipModel.Agreement, error) {
	return latestAgreement(s.paymentDB().Model("vip_agreements").Ctx(ctx), key, activeOnly)
}

func latestAgreement(query *gdb.Model, key string, activeOnly bool) (*vipModel.Agreement, error) {
	var row agreementRow
	query = query.Where("agreement_key", key).OrderDesc("version").OrderDesc("id").Limit(1)
	if err := query.Scan(&row); err != nil {
		return nil, gerror.Wrap(err, "get latest vip agreement")
	}
	if row.ID == 0 || (activeOnly && row.Status != 1) {
		return nil, nil
	}
	item := row.agreement()
	return &item, nil
}

// UpdateAgreement creates a new immutable agreement version. The previous
// version is deliberately never updated or deleted.
func (s *sVip) UpdateAgreement(ctx context.Context, input vipModel.AgreementInput) (vipModel.Agreement, error) {
	input.Key = strings.TrimSpace(input.Key)
	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(input.Content)
	if input.Key != vipModel.AgreementKeyService && input.Key != vipModel.AgreementKeyAutoRenew {
		return vipModel.Agreement{}, &consts.InvalidParameterError{Message: "agreement key is invalid"}
	}
	if input.Title == "" || len([]rune(input.Title)) > 100 {
		return vipModel.Agreement{}, &consts.InvalidParameterError{Message: "agreement title is required and must be no longer than 100 characters"}
	}
	if input.Content == "" || len([]rune(input.Content)) > 50000 {
		return vipModel.Agreement{}, &consts.InvalidParameterError{Message: "agreement content is required and must be no longer than 50000 characters"}
	}
	if input.Status > 1 {
		return vipModel.Agreement{}, &consts.InvalidParameterError{Message: "agreement status must be 0 or 1"}
	}

	var result vipModel.Agreement
	err := g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		var latest agreementRow
		if err := tx.Model("vip_agreements").Where("agreement_key", input.Key).OrderDesc("version").OrderDesc("id").Limit(1).LockUpdate().Scan(&latest); err != nil {
			return gerror.Wrap(err, "get current vip agreement version")
		}
		if latest.ID == 0 {
			return gerror.New("vip agreement not found")
		}
		id, err := tx.Model("vip_agreements").Data(g.Map{
			"agreement_key": input.Key,
			"title":         input.Title,
			"content":       input.Content,
			"status":        input.Status,
			"version":       latest.Version + 1,
			"updated_by":    nullableAgreementAdmin(input.AdminUserID),
		}).InsertAndGetId()
		if err != nil {
			return gerror.Wrap(err, "create vip agreement version")
		}
		var created agreementRow
		if err := tx.Model("vip_agreements").Where("id", id).Scan(&created); err != nil {
			return gerror.Wrap(err, "get created vip agreement version")
		}
		result = created.agreement()
		return nil
	})
	if err != nil {
		return vipModel.Agreement{}, err
	}
	return result, nil
}

func (r agreementRow) agreement() vipModel.Agreement {
	return vipModel.Agreement{ID: r.ID, Key: r.Key, Title: r.Title, Content: r.Content, Status: r.Status, Version: r.Version, UpdatedAt: r.UpdatedAt}
}

func nullableAgreementAdmin(id uint64) any {
	if id == 0 {
		return nil
	}
	return id
}
