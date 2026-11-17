package adminattributes

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	attributeModel "github.com/suxiangdong/indu-scope-backend/internal/model/attribute"
	"github.com/suxiangdong/indu-scope-backend/internal/testutil/dbmock"
)

const attributeTestColumns = "id code name value_type input_type unit placeholder validation_rules help_text status group_id attribute_id sort_order category_id"

func TestAttributeGroupFailureRollsBackCreation(t *testing.T) {
	db, mock := dbmock.New(t, attributeTestColumns)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `attributes`").WillReturnResult(sqlmock.NewResult(10, 1))
	mock.ExpectQuery("SELECT (.+) FROM `attribute_group_attributes`").WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}))
	mock.ExpectQuery("SELECT (.+) FROM `attribute_groups`(.+)FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
	mock.ExpectExec("INSERT INTO `attribute_group_attributes`").WillReturnError(errors.New("write interrupted"))
	mock.ExpectRollback()
	err := db.Transaction(context.Background(), func(ctx context.Context, tx gdb.TX) error {
		_, err := createAttributeTx(ctx, tx, attributeModel.AttributeInput{Name: "电压", ValueType: 3, InputType: "number", Status: 1, GroupIDs: []uint64{5}}, "attr_test")
		return err
	})
	if err == nil {
		t.Fatal("expected group write failure")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAttributeUpdateCannotPartiallyApplyWhenGroupIsInUse(t *testing.T) {
	db, mock := dbmock.New(t, attributeTestColumns)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `attributes`(.+)FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectQuery("SELECT (.+) FROM `attribute_group_attributes`").WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(20, 5))
	mock.ExpectQuery("SELECT (.+) FROM `attribute_groups`(.+)FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(5))
	mock.ExpectQuery("SELECT (.+) FROM `category_attribute_group_bindings`(.+)FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectRollback()
	err := db.Transaction(context.Background(), func(ctx context.Context, tx gdb.TX) error {
		return updateAttributeTx(ctx, tx, 10, attributeModel.AttributeInput{Name: "新名称", ValueType: 3, InputType: "number", GroupIDs: []uint64{}})
	})
	if !errors.Is(err, consts.ErrAttributeGroupInUse) {
		t.Fatalf("expected in-use rejection, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestAttributeReplacementFailureRollsBackRemovedMembership(t *testing.T) {
	db, mock := dbmock.New(t, attributeTestColumns)
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `attributes`(.+)FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(10))
	mock.ExpectQuery("SELECT (.+) FROM `attribute_group_attributes`").WillReturnRows(sqlmock.NewRows([]string{"id", "group_id"}).AddRow(20, 5))
	for _, id := range []int{5, 6} {
		mock.ExpectQuery("SELECT (.+) FROM `attribute_groups`(.+)FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(id))
	}
	mock.ExpectQuery("SELECT (.+) FROM `category_attribute_group_bindings`(.+)FOR UPDATE").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec("DELETE FROM `attribute_group_attributes`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO `attribute_group_attributes`").WillReturnError(errors.New("new membership write failed"))
	mock.ExpectRollback()
	err := db.Transaction(context.Background(), func(ctx context.Context, tx gdb.TX) error {
		return updateAttributeTx(ctx, tx, 10, attributeModel.AttributeInput{Name: "新名称", ValueType: 3, InputType: "number", GroupIDs: []uint64{6}})
	})
	if err == nil {
		t.Fatal("expected replacement failure")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
