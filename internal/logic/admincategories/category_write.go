package admincategories

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
)

func lockCategoryTree(ctx context.Context, tx gdb.TX) error {
	var row *struct {
		Key string `orm:"lock_key"`
	}
	if err := tx.Model("backend_write_locks").Ctx(ctx).Fields("lock_key").Where("lock_key", "category_tree").LockUpdate().Scan(&row); err != nil {
		return gerror.Wrap(err, "lock category tree")
	}
	if row == nil {
		return gerror.New("category tree lock is missing; apply database migrations")
	}
	return nil
}

func createCategoryTx(ctx context.Context, tx gdb.TX, parentID *uint64, name string, sortOrder, status uint) (uint64, error) {
	// One durable row serializes structural mutations across server instances,
	// including the first insert into an otherwise empty tree.
	if err := lockCategoryTree(ctx, tx); err != nil {
		return 0, err
	}
	if err := validateParentTx(ctx, tx, 0, parentID); err != nil {
		return 0, err
	}
	if err := ensureSiblingNameAvailableTx(ctx, tx, 0, parentID, name); err != nil {
		return 0, err
	}
	id, err := tx.Model("categories").Ctx(ctx).Data(g.Map{"parent_id": parentValue(parentID), "name": name, "sort_order": sortOrder, "status": status}).InsertAndGetId()
	if err != nil {
		return 0, gerror.Wrap(err, "create category")
	}
	return uint64(id), nil
}

func updateCategoryTx(ctx context.Context, tx gdb.TX, id uint64, parentID *uint64, name string, sortOrder, status uint) error {
	if err := lockCategoryTree(ctx, tx); err != nil {
		return err
	}
	if err := requireCategoryTx(ctx, tx, id); err != nil {
		return err
	}
	if err := validateParentTx(ctx, tx, id, parentID); err != nil {
		return err
	}
	if err := ensureSiblingNameAvailableTx(ctx, tx, id, parentID, name); err != nil {
		return err
	}
	_, err := tx.Model("categories").Ctx(ctx).Where("id", id).Data(g.Map{"parent_id": parentValue(parentID), "name": name, "sort_order": sortOrder, "status": status}).Update()
	return gerror.Wrap(err, "update category")
}

func requireCategoryTx(ctx context.Context, tx gdb.TX, id uint64) error {
	var row *struct{ ID uint64 }
	if err := tx.Model("categories").Ctx(ctx).Fields("id").Where("id", id).Scan(&row); err != nil {
		return gerror.Wrap(err, "query category")
	}
	if row == nil {
		return consts.ErrCategoryNotFound
	}
	return nil
}

func deleteCategoryTx(ctx context.Context, tx gdb.TX, id uint64) error {
	if err := lockCategoryTree(ctx, tx); err != nil {
		return err
	}
	if err := requireCategoryTx(ctx, tx, id); err != nil {
		return err
	}
	checks := []struct {
		table, column string
		err           error
	}{
		{"categories", "parent_id", consts.ErrCategoryHasChildren},
		{"products", "category_id", consts.ErrCategoryHasProducts},
		{"procurement_requirements", "category_id", consts.ErrCategoryHasRequirements},
	}
	for _, check := range checks {
		count, err := tx.Model(check.table).Ctx(ctx).Where(check.column, id).Count()
		if err != nil {
			return gerror.Wrap(err, "check category references")
		}
		if count > 0 {
			return check.err
		}
	}
	_, err := tx.Model("categories").Ctx(ctx).Where("id", id).Delete()
	return gerror.Wrap(err, "delete category")
}
