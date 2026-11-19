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
	"github.com/suxiangdong/indu-scope-backend/internal/selectcache"
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
	if current.Status == PendingReview && status != PendingReview {
		return consts.ErrProductPendingReview
	}
	if current.Status == status {
		return nil
	}
	return updateStatus(model, current.ID, status)
}

// Approve records a successful review without publishing the product. The
// supplier or administrator can subsequently put the approved product online.
func Approve(model *gdb.Model) error {
	current, err := load(model)
	if err != nil {
		return err
	}
	if current.Status != PendingReview {
		return consts.ErrProductAlreadyReviewed
	}
	return updateStatus(model, current.ID, Offline)
}

type productState struct {
	ID     uint64
	Status uint8
}

// The caller supplies an uncached model so review and owner checks always read
// the database. A partial status query must never populate the full-row cache.
func load(model *gdb.Model) (productState, error) {
	var current productState
	err := model.Clone().Fields("id", "status").Scan(&current)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && current.ID == 0) {
		return productState{}, consts.ErrProductNotFound
	}
	if err != nil {
		return productState{}, gerror.Wrap(err, "query product status")
	}
	return current, nil
}

func updateStatus(model *gdb.Model, id uint64, next uint8) error {
	// Keep the caller's ID and ownership scope without adding a status predicate.
	result, err := model.Clone().Cache(selectcache.Invalidate("products", id)).Data(g.Map{"status": next}).Update()
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
