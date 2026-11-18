// Package productstatus owns the review and publication transitions shared by
// administrator and supplier product endpoints.
package productstatus

import (
	"database/sql"
	"errors"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
)

const (
	PendingReview uint8 = 0
	Online        uint8 = 1
	Offline       uint8 = 2
)

// Change applies a publication change to the caller's product/owner scope.
// Returning a product to review is allowed, but leaving review requires Approve.
func Change(model *gdb.Model, status uint8) error {
	if status > Offline {
		return &consts.InvalidParameterError{Message: "status must be 0, 1 or 2"}
	}
	current, err := load(model)
	if err != nil {
		return err
	}
	if current == PendingReview && status != PendingReview {
		return consts.ErrProductPendingReview
	}
	if current == status {
		return nil
	}
	return compareAndSet(model, current, status)
}

// Approve records a successful review without publishing the product. The
// supplier or administrator can subsequently put the approved product online.
func Approve(model *gdb.Model) error {
	current, err := load(model)
	if err != nil {
		return err
	}
	if current != PendingReview {
		return consts.ErrProductAlreadyReviewed
	}
	return compareAndSet(model, PendingReview, Offline)
}

func load(model *gdb.Model) (uint8, error) {
	var current struct {
		ID     uint64
		Status uint8
	}
	err := model.Clone().Fields("id", "status").Scan(&current)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && current.ID == 0) {
		return 0, consts.ErrProductNotFound
	}
	if err != nil {
		return 0, gerror.Wrap(err, "query product status")
	}
	return current.Status, nil
}

func compareAndSet(model *gdb.Model, current, next uint8) error {
	// Keep the observed status in the UPDATE predicate: an intervening return
	// to review must never be overwritten by a stale up/down request.
	result, err := model.Clone().Where("status", current).Data(g.Map{"status": next}).Update()
	if err != nil {
		return gerror.Wrap(err, "update product status")
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return gerror.Wrap(err, "check updated product status")
	}
	if affected == 0 {
		return consts.ErrProductStatusChanged
	}
	return nil
}
