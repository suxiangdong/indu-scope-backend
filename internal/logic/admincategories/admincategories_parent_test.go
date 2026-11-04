package admincategories

import (
	"testing"

	"github.com/suxiangdong/indu-scope-backend/internal/model/entity"
)

func TestNormalizeParentIDUsesZeroForRoot(t *testing.T) {
	parentID := normalizeParentID(nil)
	if parentID == nil || *parentID != 0 {
		t.Fatalf("normalizeParentID(nil) = %v, want pointer to 0", parentID)
	}

	zero := uint64(0)
	parentID = normalizeParentID(&zero)
	if parentID == nil || *parentID != 0 {
		t.Fatalf("normalizeParentID(0) = %v, want pointer to 0", parentID)
	}
}

func TestToCategoryItemIncludesZeroParentID(t *testing.T) {
	item := toCategoryItem(&entity.Categories{Id: 1, Name: "一级类目"})
	if item.ParentID == nil || *item.ParentID != 0 {
		t.Fatalf("toCategoryItem root ParentID = %v, want pointer to 0", item.ParentID)
	}
}
