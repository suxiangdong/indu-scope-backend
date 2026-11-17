package favorites

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	"github.com/suxiangdong/indu-scope-backend/internal/testutil/dbmock"
)

func expectToggleRead(mock sqlmock.Sqlmock, favoriteID uint64, status uint8) {
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT (.+) FROM `users`(.+)FOR UPDATE").WithArgs(uint64(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	favorites := sqlmock.NewRows([]string{"id", "product_id"})
	if favoriteID != 0 {
		favorites.AddRow(favoriteID, 10)
	}
	mock.ExpectQuery("SELECT `id`,`product_id` FROM `product_favorites` WHERE (.+) LIMIT 1$").WillReturnRows(favorites)
	mock.ExpectQuery("SELECT (.+) FROM (.+)products(.+)").
		WillReturnRows(sqlmock.NewRows([]string{"id", "status", "supplier_status", "category_status", "unit_status"}).AddRow(10, status, 1, 1, 1))
}

func TestToggleUsesStableOwnerLockAndDoesNotLoseSecondToggle(t *testing.T) {
	db, mock := dbmock.New(t, "id user_id product_id status supplier_user_id category_id stock_unit_id")
	expectToggleRead(mock, 0, 1)
	mock.ExpectExec("INSERT INTO `product_favorites`").WillReturnResult(sqlmock.NewResult(8, 1))
	mock.ExpectCommit()
	state, err := toggleFavorite(context.Background(), db, 2, 10)
	if err != nil || !state.Favorited || state.FavoriteID != 8 {
		t.Fatalf("first toggle: %+v %v", state, err)
	}
	// The next lock holder must read the committed favorite and remove it;
	// there is no Count/Scan window or duplicate-key upsert to collapse toggles.
	expectToggleRead(mock, 8, 1)
	mock.ExpectExec("DELETE FROM `product_favorites`").WithArgs(uint64(8)).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	state, err = toggleFavorite(context.Background(), db, 2, 10)
	if err != nil || state.Favorited {
		t.Fatalf("second toggle: %+v %v", state, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUnavailableProductDoesNotCreateFavorite(t *testing.T) {
	db, mock := dbmock.New(t, "id user_id product_id status supplier_user_id category_id stock_unit_id")
	expectToggleRead(mock, 0, 2)
	mock.ExpectRollback()
	if _, err := toggleFavorite(context.Background(), db, 2, 10); !errors.Is(err, consts.ErrProductUnavailable) {
		t.Fatalf("unavailable product: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUnavailableProductCanStillBeUnfavorited(t *testing.T) {
	db, mock := dbmock.New(t, "id user_id product_id status supplier_user_id category_id stock_unit_id")
	expectToggleRead(mock, 8, 2)
	mock.ExpectExec("DELETE FROM `product_favorites`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	state, err := toggleFavorite(context.Background(), db, 2, 10)
	if err != nil || state.Favorited {
		t.Fatalf("remove unavailable favorite: %+v %v", state, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
