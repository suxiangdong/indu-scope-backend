package adminattributes

import (
	"context"
	"sort"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/dao"
	attributeModel "github.com/suxiangdong/indu-scope-backend/internal/model/attribute"
)

func (s *sAdminAttributes) Delete(ctx context.Context, id uint64) error {
	if id == 0 {
		return &consts.InvalidParameterError{Message: "id is required"}
	}
	return dao.Attributes.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := lockAttributeTx(ctx, tx, id); err != nil {
			return err
		}
		return s.deleteAttribute(ctx, id)
	})
}

func (s *sAdminAttributes) UpdateGroup(ctx context.Context, id uint64, input attributeModel.AttributeGroupInput) (item attributeModel.AttributeGroupItem, err error) {
	if id == 0 {
		return item, &consts.InvalidParameterError{Message: "id is required"}
	}
	err = dao.Attributes.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := lockAttributeGroupTx(ctx, tx, id); err != nil {
			return err
		}
		var err error
		item, err = s.updateGroup(ctx, id, input)
		return err
	})
	return item, err
}

func (s *sAdminAttributes) DeleteGroup(ctx context.Context, id uint64) error {
	if id == 0 {
		return &consts.InvalidParameterError{Message: "id is required"}
	}
	return dao.Attributes.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := lockAttributeGroupTx(ctx, tx, id); err != nil {
			return err
		}
		return s.deleteGroup(ctx, id)
	})
}

func (s *sAdminAttributes) SaveGroupAttribute(ctx context.Context, input attributeModel.AttributeGroupAttributeInput) (item attributeModel.AttributeGroupAttributeItem, err error) {
	if input.AttributeID == 0 || input.GroupID == 0 {
		return item, &consts.InvalidParameterError{Message: "group_id and attribute_id are required"}
	}
	err = dao.Attributes.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := lockAttributeTx(ctx, tx, input.AttributeID); err != nil {
			return err
		}
		if err := lockAttributeGroupTx(ctx, tx, input.GroupID); err != nil {
			return err
		}
		var err error
		item, err = s.saveGroupAttribute(ctx, input)
		return err
	})
	return item, err
}

func (s *sAdminAttributes) DeleteGroupAttribute(ctx context.Context, groupID, attributeID uint64) error {
	if attributeID == 0 || groupID == 0 {
		return &consts.InvalidParameterError{Message: "group_id and attribute_id are required"}
	}
	return dao.Attributes.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := lockAttributeTx(ctx, tx, attributeID); err != nil {
			return err
		}
		if err := lockAttributeGroupTx(ctx, tx, groupID); err != nil {
			return err
		}
		return s.deleteGroupAttribute(ctx, groupID, attributeID)
	})
}

func (s *sAdminAttributes) BindCategoryGroup(ctx context.Context, input attributeModel.CategoryGroupBindingInput) (item attributeModel.CategoryGroupBindingItem, err error) {
	if input.CategoryID == 0 || input.GroupID == 0 {
		return item, &consts.InvalidParameterError{Message: "category_id and group_id are required"}
	}
	err = dao.Attributes.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := lockAttributeGroupTx(ctx, tx, input.GroupID); err != nil {
			return err
		}
		var err error
		item, err = s.bindCategoryGroup(ctx, input)
		return err
	})
	return item, err
}

func (s *sAdminAttributes) UnbindCategoryGroup(ctx context.Context, categoryID, groupID uint64) error {
	if categoryID == 0 || groupID == 0 {
		return &consts.InvalidParameterError{Message: "category_id and group_id are required"}
	}
	return dao.Attributes.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if err := lockAttributeGroupTx(ctx, tx, groupID); err != nil {
			return err
		}
		return s.unbindCategoryGroup(ctx, categoryID, groupID)
	})
}

func createAttributeTx(ctx context.Context, tx gdb.TX, input attributeModel.AttributeInput, code string) (uint64, error) {
	id, err := tx.Model("attributes").Ctx(ctx).Data(attributeData(input, code)).InsertAndGetId()
	if err != nil {
		if isDuplicateKey(err) {
			return 0, consts.ErrAttributeAlreadyExists
		}
		return 0, gerror.Wrap(err, "create attribute")
	}
	if err := syncAttributeGroupsTx(ctx, tx, uint64(id), input.GroupIDs); err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func lockAttributeTx(ctx context.Context, tx gdb.TX, id uint64) error {
	var row *struct{ ID uint64 }
	if err := tx.Model("attributes").Ctx(ctx).Fields("id").Where("id", id).LockUpdate().Scan(&row); err != nil {
		return gerror.Wrap(err, "lock attribute")
	}
	if row == nil {
		return consts.ErrAttributeNotFound
	}
	return nil
}

func updateAttributeTx(ctx context.Context, tx gdb.TX, id uint64, input attributeModel.AttributeInput) error {
	// Serialize updates before reading or replacing this attribute's memberships.
	if err := lockAttributeTx(ctx, tx, id); err != nil {
		return err
	}
	if input.GroupIDs != nil {
		if err := syncAttributeGroupsTx(ctx, tx, id, input.GroupIDs); err != nil {
			return err
		}
	}
	if _, err := tx.Model("attributes").Ctx(ctx).Where("id", id).Data(attributeData(input, "")).Update(); err != nil {
		if isDuplicateKey(err) {
			return consts.ErrAttributeAlreadyExists
		}
		return gerror.Wrap(err, "update attribute")
	}
	return nil
}

func lockAttributeGroupTx(ctx context.Context, tx gdb.TX, id uint64) error {
	var row *struct{ ID uint64 }
	if err := tx.Model("attribute_groups").Ctx(ctx).Fields("id").Where("id", id).LockUpdate().Scan(&row); err != nil {
		return gerror.Wrap(err, "lock attribute group")
	}
	if row == nil {
		return consts.ErrAttributeGroupNotFound
	}
	return nil
}

func syncAttributeGroupsTx(ctx context.Context, tx gdb.TX, attributeID uint64, groupIDs []uint64) error {
	requested := make(map[uint64]struct{}, len(groupIDs))
	for _, id := range groupIDs {
		if id == 0 {
			return &consts.InvalidParameterError{Message: "group_ids must contain positive IDs"}
		}
		if _, duplicate := requested[id]; duplicate {
			return &consts.InvalidParameterError{Message: "group_ids must not contain duplicates"}
		}
		requested[id] = struct{}{}
	}
	var current []struct {
		ID      uint64
		GroupID uint64 `orm:"group_id"`
	}
	// Every membership writer holds the attribute row. A plain read avoids
	// locking gaps for attributes that have no group memberships yet.
	if err := tx.Model("attribute_group_attributes").Ctx(ctx).Fields("id", "group_id").Where("attribute_id", attributeID).Scan(&current); err != nil {
		return gerror.Wrap(err, "list attribute groups for update")
	}
	all := make(map[uint64]struct{}, len(requested)+len(current))
	for id := range requested {
		all[id] = struct{}{}
	}
	for _, row := range current {
		all[row.GroupID] = struct{}{}
	}
	ids := make([]uint64, 0, len(all))
	for id := range all {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		if err := lockAttributeGroupTx(ctx, tx, id); err != nil {
			return err
		}
	}
	for _, row := range current {
		if _, keep := requested[row.GroupID]; keep {
			continue
		}
		var bindings []struct{ ID uint64 }
		// The group lock serializes category binding changes. Use a current read
		// because another binder may have committed while we waited for that lock.
		if err := tx.Model("category_attribute_group_bindings").Ctx(ctx).Fields("id").Where("group_id", row.GroupID).Limit(1).LockUpdate().Scan(&bindings); err != nil {
			return gerror.Wrap(err, "check removed attribute group bindings")
		}
		if len(bindings) > 0 {
			return consts.ErrAttributeGroupInUse
		}
	}
	existing := make(map[uint64]uint64, len(current))
	for _, row := range current {
		existing[row.GroupID] = row.ID
		if _, keep := requested[row.GroupID]; !keep {
			if _, err := tx.Model("attribute_group_attributes").Ctx(ctx).Where("id", row.ID).Delete(); err != nil {
				return gerror.Wrap(err, "remove attribute group")
			}
		}
	}
	for index, id := range groupIDs {
		if rowID, keep := existing[id]; keep {
			if _, err := tx.Model("attribute_group_attributes").Ctx(ctx).Where("id", rowID).Data(g.Map{"sort_order": index}).Update(); err != nil {
				return gerror.Wrap(err, "sort attribute group")
			}
			continue
		}
		if _, err := tx.Model("attribute_group_attributes").Ctx(ctx).Data(g.Map{"group_id": id, "attribute_id": attributeID, "sort_order": index}).Insert(); err != nil {
			return gerror.Wrap(err, "save attribute group")
		}
	}
	return nil
}
