package productcatalog

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	v1 "github.com/suxiangdong/indu-scope-backend/api/products/v1"
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic/bizctx"
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic/productcatalog"
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic/vip"
	authModel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
	bizctxModel "github.com/suxiangdong/indu-scope-backend/internal/model/bizctx"
	productCatalogModel "github.com/suxiangdong/indu-scope-backend/internal/model/productcatalog"
	userprofileModel "github.com/suxiangdong/indu-scope-backend/internal/model/userprofile"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

const companyName = "仅会员可见的企业名称"

type companyCatalogService struct{ service.IProductCatalog }

func (companyCatalogService) List(_ context.Context, buyerID uint64, _ productCatalogModel.ListFilter) (productCatalogModel.ListResult, error) {
	if buyerID != 2 {
		return productCatalogModel.ListResult{}, errors.New("unexpected buyer")
	}
	return productCatalogModel.ListResult{Items: []productCatalogModel.ProductItem{companyProduct()}, Total: 1}, nil
}

func (companyCatalogService) Detail(_ context.Context, buyerID, productID uint64) (productCatalogModel.ProductDetail, error) {
	if buyerID != 2 || productID != 10 {
		return productCatalogModel.ProductDetail{}, errors.New("unexpected product request")
	}
	return productCatalogModel.ProductDetail{ProductItem: companyProduct(),
		SupplierContacts: []userprofileModel.Contact{{Name: "联系人", Position: "经理", Phone: "13800000000"}}}, nil
}

func companyProduct() productCatalogModel.ProductItem {
	return productCatalogModel.ProductItem{ID: 10, SupplierUserID: 99, SupplierCompany: companyName, Name: "测试产品", Status: 1}
}

type companyVipService struct {
	service.IVip
	subscription *vipModel.Subscription
	err          error
	calls        int
}

func (s *companyVipService) GetSubscription(_ context.Context, userID uint64) (*vipModel.Subscription, error) {
	s.calls++
	if userID != 2 {
		return nil, errors.New("unexpected membership owner")
	}
	return s.subscription, s.err
}

func companyFixture(t *testing.T, vip *companyVipService) context.Context {
	t.Helper()
	previousCatalog, previousVip := service.ProductCatalog(), service.Vip()
	service.RegisterProductCatalog(companyCatalogService{})
	service.RegisterVip(vip)
	t.Cleanup(func() {
		service.RegisterProductCatalog(previousCatalog)
		service.RegisterVip(previousVip)
	})
	return service.Bizctx().Set(context.Background(), &bizctxModel.Context{User: &authModel.User{ID: 2, Role: 1, Status: 1}})
}

func companyJSON(t *testing.T, response any) map[string]any {
	t.Helper()
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatal(err)
	}
	return fields
}

func TestProductListDoesNotExposeCompanyEvenToMembers(t *testing.T) {
	for _, member := range []bool{false, true} {
		t.Run(map[bool]string{false: "nonmember", true: "member"}[member], func(t *testing.T) {
			vip := &companyVipService{}
			if member {
				vip.subscription = &vipModel.Subscription{Status: "active"}
			}
			ctx := companyFixture(t, vip)
			response, err := (&ControllerV1{}).List(ctx, &v1.ListReq{})
			if err != nil || len(response.Items) != 1 {
				t.Fatalf("list response=%+v err=%v", response, err)
			}
			item := companyJSON(t, response)["items"].([]any)[0].(map[string]any)
			for _, key := range []string{"supplier_company", "company_name", "company_info"} {
				if _, exists := item[key]; exists {
					t.Fatalf("product list exposes %s", key)
				}
			}
			if vip.calls != 0 {
				t.Fatal("ordinary product listings should not load membership")
			}
		})
	}
}

func TestProductDetailCompanyAndPhoneUseOneMembershipDecision(t *testing.T) {
	future, past := time.Now().Add(time.Hour), time.Now().Add(-time.Hour)
	for _, test := range []struct {
		name         string
		subscription *vipModel.Subscription
		allowed      bool
	}{
		{name: "nonmember"},
		{name: "expired", subscription: &vipModel.Subscription{Status: "active", EndAt: &past}},
		{name: "active", subscription: &vipModel.Subscription{Status: "active", EndAt: &future}, allowed: true},
		{name: "lifetime", subscription: &vipModel.Subscription{Status: "active"}, allowed: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			vip := &companyVipService{subscription: test.subscription}
			response, err := (&ControllerV1{}).Detail(companyFixture(t, vip), &v1.DetailReq{ID: 10})
			if err != nil {
				t.Fatal(err)
			}
			if response.CompanyLocked != !test.allowed || response.ContactLocked != !test.allowed {
				t.Fatalf("inconsistent company/contact locks: %+v", response)
			}
			fields := companyJSON(t, response)
			if _, exists := fields["supplier_company"]; exists {
				t.Fatal("legacy supplier_company bypasses company lock")
			}
			info, exists := fields["company_info"]
			if exists != test.allowed {
				t.Fatalf("company_info present=%t, allowed=%t", exists, test.allowed)
			}
			if test.allowed {
				summary := info.(map[string]any)
				if summary["id"] != float64(99) || summary["company_name"] != companyName {
					t.Fatalf("unexpected company summary: %+v", summary)
				}
			}
			wantPhone := ""
			if test.allowed {
				wantPhone = "13800000000"
			}
			if len(response.SupplierContacts) != 1 || response.SupplierContacts[0].Phone != wantPhone {
				t.Fatalf("unexpected contacts: %+v", response.SupplierContacts)
			}
			if vip.calls != 1 {
				t.Fatalf("membership was loaded %d times, want 1", vip.calls)
			}
		})
	}
}

func TestProductDetailDoesNotReturnDataWhenMembershipLookupFails(t *testing.T) {
	vip := &companyVipService{err: errors.New("membership lookup unavailable")}
	response, err := (&ControllerV1{}).Detail(companyFixture(t, vip), &v1.DetailReq{ID: 10})
	if err == nil || response != nil || vip.calls != 1 {
		t.Fatalf("membership error must fail closed: response=%+v err=%v calls=%d", response, err, vip.calls)
	}
}
