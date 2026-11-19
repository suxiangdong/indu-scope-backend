package supplierproducts

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	v1 "github.com/suxiangdong/indu-scope-backend/api/supplierproducts/v1"
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic/bizctx"
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic/supplierproducts"
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic/userprofile"
	_ "github.com/suxiangdong/indu-scope-backend/internal/logic/vip"
	authModel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
	bizctxModel "github.com/suxiangdong/indu-scope-backend/internal/model/bizctx"
	supplierProductModel "github.com/suxiangdong/indu-scope-backend/internal/model/supplierproduct"
	userprofileModel "github.com/suxiangdong/indu-scope-backend/internal/model/userprofile"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

const companyName = "仅会员可见的供应商企业"

type companySupplierService struct{ service.ISupplierProducts }

func ownProduct() supplierProductModel.ProductItem {
	return supplierProductModel.ProductItem{ID: 10, SupplierUserID: 2, Name: "测试产品", Status: 1,
		Contacts: []userprofileModel.Contact{{Name: "自己的联系人", Phone: "13800000000"}}}
}

func (companySupplierService) List(_ context.Context, userID uint64, _, _ uint) (supplierProductModel.ListResult, error) {
	if userID != 2 {
		return supplierProductModel.ListResult{}, errors.New("unexpected supplier")
	}
	return supplierProductModel.ListResult{Items: []supplierProductModel.ProductItem{ownProduct()}, Total: 1}, nil
}

func (companySupplierService) Detail(_ context.Context, userID, productID uint64) (supplierProductModel.ProductItem, error) {
	if userID != 2 || productID != 10 {
		return supplierProductModel.ProductItem{}, errors.New("unexpected own-product request")
	}
	return ownProduct(), nil
}

func (companySupplierService) Create(_ context.Context, _ supplierProductModel.CreateInput) (supplierProductModel.ProductItem, error) {
	return ownProduct(), nil
}

func (companySupplierService) UpdateStatus(_ context.Context, _, _ uint64, status uint8) (supplierProductModel.ProductItem, error) {
	item := ownProduct()
	item.Status = uint(status)
	return item, nil
}

func (companySupplierService) UpdateStock(_ context.Context, input supplierProductModel.UpdateStockInput) (supplierProductModel.ProductItem, error) {
	item := ownProduct()
	item.StockQuantity = input.StockQuantity
	return item, nil
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

type companyProfileService struct {
	service.IUserProfile
	calls int
	err   error
}

func (s *companyProfileService) Profile(_ context.Context, userID uint64) (userprofileModel.Profile, error) {
	s.calls++
	if userID != 2 {
		return userprofileModel.Profile{}, errors.New("unexpected company profile owner")
	}
	return userprofileModel.Profile{ID: 2, CompanyName: companyName, Introduction: "不属于企业摘要的资料"}, s.err
}

func companyFixture(t *testing.T, vip *companyVipService, profile *companyProfileService) context.Context {
	t.Helper()
	previousProducts, previousVip, previousProfile := service.SupplierProducts(), service.Vip(), service.UserProfile()
	service.RegisterSupplierProducts(companySupplierService{})
	service.RegisterVip(vip)
	service.RegisterUserProfile(profile)
	t.Cleanup(func() {
		service.RegisterSupplierProducts(previousProducts)
		service.RegisterVip(previousVip)
		service.RegisterUserProfile(previousProfile)
	})
	return service.Bizctx().Set(context.Background(), &bizctxModel.Context{User: &authModel.User{ID: 2, Role: 2, Status: 1, CompanyName: companyName}})
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

func assertNoCompanyFields(t *testing.T, fields map[string]any) {
	t.Helper()
	for _, key := range []string{"company_info", "company_name", "supplier_company"} {
		if _, exists := fields[key]; exists {
			t.Fatalf("ordinary product response exposes %s", key)
		}
	}
}

func TestSupplierListAndMutationResponsesDoNotContainCompanyCards(t *testing.T) {
	vip := &companyVipService{subscription: &vipModel.Subscription{Status: "active"}}
	profile := &companyProfileService{}
	ctx := companyFixture(t, vip, profile)
	controller := &ControllerV1{}
	list, err := controller.List(ctx, &v1.ListReq{})
	if err != nil {
		t.Fatal(err)
	}
	assertNoCompanyFields(t, companyJSON(t, list)["items"].([]any)[0].(map[string]any))
	created, err := controller.Create(ctx, &v1.CreateReq{CategoryID: 3, Name: "测试产品", StockUnitID: 4})
	if err != nil {
		t.Fatal(err)
	}
	assertNoCompanyFields(t, companyJSON(t, created))
	updated, err := controller.UpdateStatus(ctx, &v1.UpdateStatusReq{ID: 10, Status: 2})
	if err != nil {
		t.Fatal(err)
	}
	assertNoCompanyFields(t, companyJSON(t, updated))
	stock, err := controller.UpdateStock(ctx, &v1.UpdateStockReq{ID: 10, StockQuantity: 12})
	if err != nil {
		t.Fatal(err)
	}
	assertNoCompanyFields(t, companyJSON(t, stock))
	if vip.calls != 0 || profile.calls != 0 {
		t.Fatalf("ordinary product responses loaded company membership/profile: %d/%d", vip.calls, profile.calls)
	}
}

func TestSupplierDetailCompanyIsMemberOnlyAndOwnContactsRemainAvailable(t *testing.T) {
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
			profile := &companyProfileService{}
			response, err := (&ControllerV1{}).Detail(companyFixture(t, vip, profile), &v1.DetailReq{ID: 10})
			if err != nil {
				t.Fatal(err)
			}
			if response.CompanyLocked != !test.allowed {
				t.Fatalf("company lock=%t, allowed=%t", response.CompanyLocked, test.allowed)
			}
			fields := companyJSON(t, response)
			info, exists := fields["company_info"]
			if exists != test.allowed {
				t.Fatalf("company_info present=%t, allowed=%t", exists, test.allowed)
			}
			if test.allowed {
				summary := info.(map[string]any)
				if len(summary) != 2 || summary["id"] != float64(2) || summary["company_name"] != companyName {
					t.Fatalf("unexpected own-company summary: %+v", summary)
				}
			}
			if len(response.SupplierContacts) != 1 || response.SupplierContacts[0].Phone != "13800000000" {
				t.Fatal("the supplier's own contacts changed with membership")
			}
			wantProfileCalls := 0
			if test.allowed {
				wantProfileCalls = 1
			}
			if vip.calls != 1 || profile.calls != wantProfileCalls {
				t.Fatalf("membership/profile reads=%d/%d, want=1/%d", vip.calls, profile.calls, wantProfileCalls)
			}
		})
	}
}

func TestSupplierDetailDoesNotReturnCompanyWhenMembershipLookupFails(t *testing.T) {
	vip := &companyVipService{err: errors.New("membership lookup unavailable")}
	profile := &companyProfileService{}
	response, err := (&ControllerV1{}).Detail(companyFixture(t, vip, profile), &v1.DetailReq{ID: 10})
	if err == nil || response != nil || vip.calls != 1 || profile.calls != 0 {
		t.Fatalf("membership error must fail closed: response=%+v err=%v calls=%d/%d", response, err, vip.calls, profile.calls)
	}
}

func TestSupplierDetailDoesNotReturnPartialCompanyWhenProfileLookupFails(t *testing.T) {
	vip := &companyVipService{subscription: &vipModel.Subscription{Status: "active"}}
	profile := &companyProfileService{err: errors.New("company profile unavailable")}
	response, err := (&ControllerV1{}).Detail(companyFixture(t, vip, profile), &v1.DetailReq{ID: 10})
	if err == nil || response != nil || profile.calls != 1 {
		t.Fatalf("profile error must not return a partial summary: response=%+v err=%v", response, err)
	}
}
