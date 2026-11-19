package procurementrequirements

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	v1 "github.com/suxiangdong/indu-scope-backend/api/procurementrequirements/v1"
	logicbizctx "github.com/suxiangdong/indu-scope-backend/internal/logic/bizctx"
	authModel "github.com/suxiangdong/indu-scope-backend/internal/model/auth"
	bizctxModel "github.com/suxiangdong/indu-scope-backend/internal/model/bizctx"
	requirementModel "github.com/suxiangdong/indu-scope-backend/internal/model/procurementrequirement"
	userprofileModel "github.com/suxiangdong/indu-scope-backend/internal/model/userprofile"
	vipModel "github.com/suxiangdong/indu-scope-backend/internal/model/vip"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

type companyRequirementsStub struct {
	service.IProcurementRequirements
	detail        requirementModel.Detail
	detailCalls   int
	listCalls     int
	responseCalls int
}

func (s *companyRequirementsStub) Detail(context.Context, uint64, uint8, uint64) (requirementModel.Detail, error) {
	s.detailCalls++
	return s.detail, nil
}

func (s *companyRequirementsStub) List(context.Context, requirementModel.ListFilter) (requirementModel.ListResult, error) {
	s.listCalls++
	return requirementModel.ListResult{Items: []requirementModel.RequirementItem{s.detail.Requirement}, Total: 1, Page: 1, PageSize: 20}, nil
}

func (s *companyRequirementsStub) CreateResponse(context.Context, requirementModel.ResponseInput) (requirementModel.ResponseItem, error) {
	s.responseCalls++
	return s.detail.Responses[0], nil
}

func (s *companyRequirementsStub) UpdateResponseStatus(context.Context, uint64, uint64, uint8) (requirementModel.ResponseItem, error) {
	s.responseCalls++
	return s.detail.Responses[0], nil
}

type companyMembershipStub struct {
	service.IVip
	subscription *vipModel.Subscription
	err          error
	viewers      []uint64
}

func (s *companyMembershipStub) GetSubscription(_ context.Context, userID uint64) (*vipModel.Subscription, error) {
	s.viewers = append(s.viewers, userID)
	return s.subscription, s.err
}

// A controller package may run without any logic implementations registered.
// Preserve an existing implementation when present, or restore nil afterward.
func optionalRegisteredService[T any](getter func() T) (previous T) {
	defer func() { _ = recover() }()
	return getter()
}

func companyControllerFixture(t *testing.T, subscription *vipModel.Subscription, membershipErr error) (*companyRequirementsStub, *companyMembershipStub) {
	t.Helper()
	previousRequirements := optionalRegisteredService(service.ProcurementRequirements)
	previousVIP := optionalRegisteredService(service.Vip)
	quantity, leadTime := 8.0, uint(3)
	requirements := &companyRequirementsStub{detail: requirementModel.Detail{
		Requirement: requirementModel.RequirementItem{
			ID: 10, BuyerUserID: 2, BuyerCompany: "采购企业秘密甲", Name: "采购离心风机", CategoryID: 3,
			ContactName: "需求联系人", ContactPhone: "13800138000", ReviewStatus: 1, Status: 1,
		},
		Responses: []requirementModel.ResponseItem{
			{ID: 20, RequirementID: 10, SupplierUserID: 3, SupplierCompany: "供货企业秘密乙", Message: "可供现货", AvailableQuantity: &quantity, LeadTimeDays: &leadTime, Status: 1},
			{ID: 21, RequirementID: 10, SupplierUserID: 4, SupplierCompany: "供货企业秘密丙", Message: "可按期交付", Status: 0},
		},
		BuyerContacts: []userprofileModel.Contact{{Name: "企业联系人", Position: "采购", Phone: "13900139000"}},
	}}
	membership := &companyMembershipStub{subscription: subscription, err: membershipErr}
	service.RegisterProcurementRequirements(requirements)
	service.RegisterVip(membership)
	t.Cleanup(func() {
		service.RegisterProcurementRequirements(previousRequirements)
		service.RegisterVip(previousVIP)
	})
	return requirements, membership
}

func companyRequestContext(userID int64, role uint8, status uint8) context.Context {
	return logicbizctx.New().Set(context.Background(), &bizctxModel.Context{
		User: &authModel.User{ID: userID, Role: role, Status: status},
	})
}

func companyJSON(t *testing.T, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func assertNoLegacyCompanyFields(t *testing.T, encoded string) {
	t.Helper()
	for _, key := range []string{`"buyer_company"`, `"supplier_company"`} {
		if strings.Contains(encoded, key) {
			t.Fatalf("legacy company field leaked in response: %s", encoded)
		}
	}
}

func assertNoCompanyNames(t *testing.T, encoded string) {
	t.Helper()
	for _, name := range []string{"采购企业秘密甲", "供货企业秘密乙", "供货企业秘密丙"} {
		if strings.Contains(encoded, name) {
			t.Fatalf("company name leaked in response: %s", encoded)
		}
	}
}

func TestRequirementDetailGatesAllCompanySummariesWithOneMembershipLookup(t *testing.T) {
	future, past := time.Now().Add(time.Hour), time.Now().Add(-time.Hour)
	for _, test := range []struct {
		name         string
		viewer       int64
		role         uint8
		subscription *vipModel.Subscription
		allowed      bool
	}{
		{name: "non-member supplier", viewer: 9, role: 2},
		{name: "own requirement has no membership exception", viewer: 2, role: 1},
		{name: "expired membership", viewer: 2, role: 1, subscription: &vipModel.Subscription{Status: "active", EndAt: &past}},
		{name: "active member", viewer: 9, role: 2, subscription: &vipModel.Subscription{Status: "active", EndAt: &future}, allowed: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			requirements, membership := companyControllerFixture(t, test.subscription, nil)
			res, err := (&ControllerV1{}).Detail(companyRequestContext(test.viewer, test.role, 1), &v1.DetailReq{ID: 10})
			if err != nil || res == nil {
				t.Fatalf("detail result=%+v err=%v", res, err)
			}
			if requirements.detailCalls != 1 || len(membership.viewers) != 1 || membership.viewers[0] != uint64(test.viewer) {
				t.Fatalf("detail calls=%d membership viewers=%v", requirements.detailCalls, membership.viewers)
			}
			if res.CompanyLocked != !test.allowed || (res.CompanyInfo != nil) != test.allowed {
				t.Fatalf("buyer company gate: locked=%t info=%+v", res.CompanyLocked, res.CompanyInfo)
			}
			if test.allowed && (res.CompanyInfo.ID != 2 || res.CompanyInfo.CompanyName != requirements.detail.Requirement.BuyerCompany) {
				t.Fatalf("wrong buyer summary: %+v", res.CompanyInfo)
			}
			if len(res.Responses) != 2 {
				t.Fatalf("response count=%d", len(res.Responses))
			}
			for index, response := range res.Responses {
				if response.CompanyLocked != !test.allowed || (response.CompanyInfo != nil) != test.allowed {
					t.Fatalf("supplier company gate: %+v", response)
				}
				if test.allowed && (response.CompanyInfo.ID != requirements.detail.Responses[index].SupplierUserID ||
					response.CompanyInfo.CompanyName != requirements.detail.Responses[index].SupplierCompany) {
					t.Fatalf("wrong supplier summary: %+v", response.CompanyInfo)
				}
			}
			// This change only gates company information, preserving existing
			// requirement contacts, response content and review/publication state.
			if res.Requirement.ContactName != "需求联系人" || res.Requirement.ContactPhone != "13800138000" ||
				res.Requirement.ReviewStatus != 1 || res.Requirement.Status != 1 ||
				len(res.BuyerContacts) != 1 || res.BuyerContacts[0].Phone != "13900139000" ||
				res.Responses[0].Message != "可供现货" || res.Responses[0].AvailableQuantity == nil ||
				*res.Responses[0].AvailableQuantity != 8 || res.Responses[0].LeadTimeDays == nil || *res.Responses[0].LeadTimeDays != 3 {
				t.Fatalf("unrelated requirement/response fields changed: %+v", res)
			}
			encoded := companyJSON(t, res)
			assertNoLegacyCompanyFields(t, encoded)
			if !test.allowed {
				assertNoCompanyNames(t, encoded)
				if strings.Contains(encoded, `"company_info"`) {
					t.Fatalf("locked response must omit company_info: %s", encoded)
				}
			}
		})
	}
}

func TestRequirementDetailFailsClosedWhenMembershipLookupFails(t *testing.T) {
	_, membership := companyControllerFixture(t, nil, errors.New("isolated membership lookup failure"))
	res, err := (&ControllerV1{}).Detail(companyRequestContext(2, 1, 1), &v1.DetailReq{ID: 10})
	if err == nil || res != nil || len(membership.viewers) != 1 {
		t.Fatalf("membership failure returned data: result=%+v err=%v viewers=%v", res, err, membership.viewers)
	}
}

func TestRequirementListOmitsCompanyNamesWithoutCheckingMembership(t *testing.T) {
	_, membership := companyControllerFixture(t, nil, nil)
	res, err := (&ControllerV1{}).List(companyRequestContext(2, 1, 1), &v1.ListReq{})
	if err != nil || res == nil || len(res.Items) != 1 {
		t.Fatalf("list result=%+v err=%v", res, err)
	}
	if len(membership.viewers) != 0 || res.Items[0].ContactName != "需求联系人" || res.Items[0].ContactPhone != "13800138000" {
		t.Fatalf("list made membership calls or changed contacts: %+v viewers=%v", res, membership.viewers)
	}
	encoded := companyJSON(t, res)
	assertNoLegacyCompanyFields(t, encoded)
	assertNoCompanyNames(t, encoded)
}

func TestResponseMutationResultsKeepCompanyInformationLocked(t *testing.T) {
	future := time.Now().Add(time.Hour)
	for _, action := range []string{"create", "update status"} {
		t.Run(action, func(t *testing.T) {
			requirements, membership := companyControllerFixture(t, &vipModel.Subscription{Status: "active", EndAt: &future}, nil)
			controller := &ControllerV1{}
			var (
				info    v1.RequirementResponseInfo
				payload any
			)
			if action == "create" {
				res, err := controller.CreateResponse(companyRequestContext(3, 2, 1), &v1.CreateResponseReq{RequirementID: 10, Message: "可供现货"})
				if err != nil || res == nil {
					t.Fatalf("create response=%+v err=%v", res, err)
				}
				info, payload = res.RequirementResponseInfo, res
			} else {
				res, err := controller.UpdateResponseStatus(companyRequestContext(2, 1, 1), &v1.UpdateResponseStatusReq{ID: 20, Status: 1})
				if err != nil || res == nil {
					t.Fatalf("update response=%+v err=%v", res, err)
				}
				info, payload = res.RequirementResponseInfo, res
			}
			if !info.CompanyLocked || info.CompanyInfo != nil || info.SupplierCompany != "" || len(membership.viewers) != 0 || requirements.responseCalls != 1 {
				t.Fatalf("mutation leaked company or bypassed existing service: %+v viewers=%v calls=%d", info, membership.viewers, requirements.responseCalls)
			}
			encoded := companyJSON(t, payload)
			assertNoLegacyCompanyFields(t, encoded)
			assertNoCompanyNames(t, encoded)
		})
	}
}

func TestRequirementDetailRetainsAuthenticationAndRequestGuards(t *testing.T) {
	for _, test := range []struct {
		name   string
		role   uint8
		status uint8
		req    *v1.DetailReq
	}{
		{name: "unverified", role: 1, status: 0, req: &v1.DetailReq{ID: 10}},
		{name: "unsupported role", role: 3, status: 1, req: &v1.DetailReq{ID: 10}},
		{name: "missing request", role: 1, status: 1},
		{name: "missing id", role: 1, status: 1, req: &v1.DetailReq{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			requirements, membership := companyControllerFixture(t, nil, nil)
			res, err := (&ControllerV1{}).Detail(companyRequestContext(2, test.role, test.status), test.req)
			if err == nil || res != nil || requirements.detailCalls != 0 || len(membership.viewers) != 0 {
				t.Fatalf("request guard changed: result=%+v err=%v calls=%d viewers=%v", res, err, requirements.detailCalls, membership.viewers)
			}
		})
	}
}
