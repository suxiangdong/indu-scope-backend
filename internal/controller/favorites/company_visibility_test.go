package favorites

import (
	"encoding/json"
	"strings"
	"testing"

	favoriteModel "github.com/suxiangdong/indu-scope-backend/internal/model/favorite"
)

func TestFavoriteProductListDoesNotExposeCompanyName(t *testing.T) {
	item := toAPI(favoriteModel.ProductItem{ID: 1, ProductID: 20, SupplierUserID: 30, SupplierCompany: "PRIVATE_COMPANY_NAME", ProductName: "产品"})
	encoded, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "supplier_company") || strings.Contains(string(encoded), "PRIVATE_COMPANY_NAME") {
		t.Fatalf("favorite list leaked company identity: %s", encoded)
	}
}
