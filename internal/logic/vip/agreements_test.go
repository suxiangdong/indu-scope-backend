package vip

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
)

func TestLatestAgreementMissingOrDisabledDoesNotUseHistory(t *testing.T) {
	for _, state := range []string{"empty", "no rows error", "disabled latest"} {
		t.Run(state, func(t *testing.T) {
			s, mock := paymentServiceFixture(t)
			query := mock.ExpectQuery("SELECT (.+) FROM `vip_agreements` WHERE (.+)agreement_key(.+)ORDER BY (.+)version. DESC,(.+)id. DESC LIMIT 1").WithArgs("privacy")
			rows := sqlmock.NewRows([]string{"id", "agreement_key", "status", "version"})
			switch state {
			case "no rows error":
				query.WillReturnError(sql.ErrNoRows)
			case "disabled latest":
				query.WillReturnRows(rows.AddRow(9, "privacy", 0, 3))
			default:
				query.WillReturnRows(rows)
			}
			item, err := s.latestAgreement(context.Background(), "privacy", true)
			if err != nil || item != nil {
				t.Fatalf("item=%+v err=%v", item, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAgreementPublicationSerializesFirstAndLaterVersions(t *testing.T) {
	for _, latestVersion := range []uint{0, 4} {
		t.Run(fmt.Sprint(latestVersion), func(t *testing.T) {
			s, mock := paymentServiceFixture(t)
			mock.ExpectBegin()
			mock.ExpectQuery("SELECT `id` FROM `vip_agreements` ORDER BY `id` ASC LIMIT 1 FOR UPDATE").
				WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			latest := mock.ExpectQuery("SELECT (.+) FROM `vip_agreements` WHERE (.+)agreement_key(.+)FOR UPDATE").WithArgs("privacy")
			if latestVersion == 0 {
				latest.WillReturnError(sql.ErrNoRows)
			} else {
				latest.WillReturnRows(sqlmock.NewRows([]string{"id", "agreement_key", "version"}).AddRow(7, "privacy", latestVersion))
			}
			mock.ExpectExec("INSERT INTO `vip_agreements`(.+)version(.+)").WillReturnResult(sqlmock.NewResult(9, 1))
			mock.ExpectQuery("SELECT (.+) FROM `vip_agreements` WHERE (.+)id(.+)").WithArgs(int64(9)).
				WillReturnRows(sqlmock.NewRows([]string{"id", "agreement_key", "title", "content", "status", "version"}).
					AddRow(9, "privacy", "隐私政策", "完整正文", 1, latestVersion+1))
			mock.ExpectCommit()
			item, err := s.UpdateAgreement(context.Background(), vipModel.AgreementInput{Key: "privacy", Title: "隐私政策", Content: "完整正文", Status: 1})
			if err != nil || item.ID != 9 || item.Version != latestVersion+1 {
				t.Fatalf("published=%+v err=%v", item, err)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
