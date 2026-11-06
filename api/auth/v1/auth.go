package v1

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

type RegisterReq struct {
	g.Meta   `path:"/api/v1/auth/register" tags:"Auth" method:"post" summary:"Register with phone and password"`
	Phone    string `json:"phone" v:"required" dc:"Phone number"`
	Password string `json:"password" v:"required" dc:"Password, at least 8 characters"`
	UserType uint8  `json:"user_type" v:"required" dc:"1 personal, 2 company"`
}

type LoginReq struct {
	g.Meta   `path:"/api/v1/auth/login" tags:"Auth" method:"post" summary:"Login with phone and password"`
	Phone    string `json:"phone" v:"required" dc:"Phone number"`
	Password string `json:"password" v:"required" dc:"Password"`
}

type UserInfo struct {
	ID       int64  `json:"id" dc:"User ID"`
	Phone    string `json:"phone" dc:"Phone number"`
	Role     uint8  `json:"role" dc:"0 pending selection, 1 buyer, 2 supplier"`
	UserType uint8  `json:"user_type" dc:"1 personal, 2 company"`
	Status   uint8  `json:"status" dc:"0 unverified, 1 verified, 2 disabled"`
}

type AuthData struct {
	AccessToken string   `json:"access_token" dc:"JWT access token"`
	TokenType   string   `json:"token_type" example:"Bearer"`
	User        UserInfo `json:"user"`
}

type RegisterRes struct {
	g.Meta `mime:"application/json"`
	AuthData
}

type LoginRes struct {
	g.Meta `mime:"application/json"`
	AuthData
}

type WechatPhoneLoginReq struct {
	g.Meta    `path:"/api/v1/auth/wechat/phone-login" tags:"Auth" method:"post" summary:"Login with WeChat login and phone authorization codes"`
	LoginCode string `json:"login_code" v:"required" dc:"Temporary code returned by wx.login"`
	PhoneCode string `json:"phone_code" v:"required" dc:"Code returned by the getPhoneNumber component"`
}

type WechatPhoneLoginRes struct {
	g.Meta `mime:"application/json"`
	AuthData
}

type StorePlatformInfo struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type ListStorePlatformsReq struct {
	g.Meta `path:"/api/v1/auth/verification/store-platforms" tags:"Auth Verification" method:"get" summary:"List active store platforms"`
}

type ListStorePlatformsRes struct {
	g.Meta `mime:"application/json"`
	Items  []StorePlatformInfo `json:"items"`
}

type MeReq struct {
	g.Meta `path:"/api/v1/auth/me" tags:"Auth" method:"get" summary:"Get current user"`
}

type MeRes struct {
	g.Meta `mime:"application/json"`
	UserInfo
}

type LogoutReq struct {
	g.Meta `path:"/api/v1/auth/logout" tags:"Auth" method:"post" summary:"Logout"`
}

type LogoutRes struct {
	g.Meta `mime:"application/json"`
}

type VerificationInfo struct {
	Role         uint8                     `json:"role" dc:"0 pending selection, 1 buyer, 2 supplier"`
	UserType     uint8                     `json:"user_type" dc:"1 personal, 2 company"`
	Status       string                    `json:"status" dc:"unsubmitted, pending, rejected, verified"`
	ReviewReason string                    `json:"review_reason,omitempty"`
	Personal     *PersonalVerificationInfo `json:"personal,omitempty"`
	Company      *CompanyVerificationInfo  `json:"company,omitempty"`
}

type PersonalVerificationInfo struct {
	RealName          string `json:"real_name"`
	IDNumberMasked    string `json:"id_number_masked,omitempty"`
	StorePlatformID   uint64 `json:"store_platform_id,omitempty"`
	StorePlatformName string `json:"store_platform_name,omitempty"`
	StoreURL          string `json:"store_url,omitempty"`
	IDFrontEndpoint   string `json:"id_front_endpoint,omitempty"`
	IDFrontBucket     string `json:"id_front_bucket,omitempty"`
	IDFrontPath       string `json:"id_front_path,omitempty"`
	IDBackEndpoint    string `json:"id_back_endpoint,omitempty"`
	IDBackBucket      string `json:"id_back_bucket,omitempty"`
	IDBackPath        string `json:"id_back_path,omitempty"`
	VerifiedAt        string `json:"verified_at,omitempty"`
	ReviewStatus      uint8  `json:"review_status"`
	ReviewReason      string `json:"review_reason,omitempty"`
	ReviewedAt        string `json:"reviewed_at,omitempty"`
}

type CompanyVerificationInfo struct {
	CompanyName                string               `json:"company_name"`
	UnifiedSocialCreditCode    string               `json:"unified_social_credit_code"`
	BusinessLicenseEndpoint    string               `json:"business_license_endpoint,omitempty"`
	BusinessLicenseBucket      string               `json:"business_license_bucket,omitempty"`
	BusinessLicensePath        string               `json:"business_license_path,omitempty"`
	BusinessLicenseFileName    string               `json:"business_license_file_name,omitempty"`
	BusinessLicenseFileSize    int64                `json:"business_license_file_size,omitempty"`
	BusinessLicenseContentType string               `json:"business_license_content_type,omitempty"`
	RegisteredAddress          string               `json:"registered_address"`
	Introduction               string               `json:"introduction,omitempty"`
	Offices                    []string             `json:"offices"`
	Factories                  []string             `json:"factories"`
	Contacts                   []CompanyContactInfo `json:"contacts"`
	VerifiedAt                 string               `json:"verified_at,omitempty"`
	ReviewStatus               uint8                `json:"review_status"`
	ReviewReason               string               `json:"review_reason,omitempty"`
	ReviewedAt                 string               `json:"reviewed_at,omitempty"`
}

type CompanyContactInfo struct {
	Name     string `json:"name"`
	Position string `json:"position,omitempty"`
	Phone    string `json:"phone"`
}

type GetVerificationReq struct {
	g.Meta `path:"/api/v1/auth/verification" tags:"Auth Verification" method:"get" summary:"Get current user verification"`
}

type GetVerificationRes struct {
	g.Meta `mime:"application/json"`
	VerificationInfo
}

type SubmitPersonalVerificationReq struct {
	g.Meta          `path:"/api/v1/auth/verification/personal" tags:"Auth Verification" method:"post" summary:"Submit personal verification"`
	RealName        string `json:"real_name" v:"required" dc:"Real name"`
	Role            uint8  `json:"role" dc:"Role is required only when current role is 0"`
	IDNumber        string `json:"id_number" v:"required" dc:"Identity document number"`
	StorePlatformID uint64 `json:"store_platform_id" v:"required" dc:"Store platform ID"`
	StoreURL        string `json:"store_url" v:"required" dc:"Store URL"`
	IDFrontEndpoint string `json:"id_front_endpoint" dc:"Identity document front OSS endpoint"`
	IDFrontBucket   string `json:"id_front_bucket" dc:"Identity document front OSS bucket"`
	IDFrontPath     string `json:"id_front_path" dc:"Identity document front OSS path"`
	IDBackEndpoint  string `json:"id_back_endpoint" dc:"Identity document back OSS endpoint"`
	IDBackBucket    string `json:"id_back_bucket" dc:"Identity document back OSS bucket"`
	IDBackPath      string `json:"id_back_path" dc:"Identity document back OSS path"`
}

type SubmitPersonalVerificationRes struct {
	g.Meta `mime:"application/json"`
	VerificationInfo
}

type SubmitCompanyVerificationReq struct {
	g.Meta                  `path:"/api/v1/auth/verification/company" tags:"Auth Verification" method:"post" summary:"Submit company verification"`
	CompanyName             string               `json:"company_name" v:"required" dc:"Company name"`
	Role                    uint8                `json:"role" dc:"Role is required only when current role is 0"`
	UnifiedSocialCreditCode string               `json:"unified_social_credit_code" v:"required" dc:"Unified social credit code"`
	BusinessLicenseEndpoint string               `json:"business_license_endpoint" v:"required" dc:"Business license OSS endpoint"`
	BusinessLicenseBucket   string               `json:"business_license_bucket" v:"required" dc:"Business license OSS bucket"`
	BusinessLicensePath     string               `json:"business_license_path" v:"required" dc:"Business license OSS path"`
	RegisteredAddress       string               `json:"registered_address" v:"required" dc:"Registered address"`
	Introduction            string               `json:"introduction" dc:"Company introduction"`
	Offices                 []string             `json:"offices" dc:"Office addresses"`
	Factories               []string             `json:"factories" dc:"Factory addresses"`
	Contacts                []CompanyContactInfo `json:"contacts" v:"required" dc:"Company contacts"`
}

type SubmitCompanyVerificationRes struct {
	g.Meta `mime:"application/json"`
	VerificationInfo
}

type UploadBusinessLicenseReq struct {
	g.Meta `path:"/api/v1/auth/verification/business-license" tags:"Auth Verification" method:"post" summary:"Upload business license"`
	File   *ghttp.UploadFile `json:"file" type:"file" v:"required" dc:"Business license image or PDF"`
}

type BusinessLicenseUploadInfo struct {
	Endpoint    string `json:"endpoint"`
	Bucket      string `json:"bucket"`
	Path        string `json:"path"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

type UploadBusinessLicenseRes struct {
	g.Meta `mime:"application/json"`
	BusinessLicenseUploadInfo
}

type UploadIDCardReq struct {
	g.Meta `path:"/api/v1/auth/verification/id-card" tags:"Auth Verification" method:"post" summary:"Upload identity document"`
	Side   string            `json:"side" v:"required" dc:"front or back"`
	File   *ghttp.UploadFile `json:"file" type:"file" v:"required" dc:"Identity document image"`
}

type IDCardUploadInfo struct {
	Side        string `json:"side"`
	Endpoint    string `json:"endpoint"`
	Bucket      string `json:"bucket"`
	Path        string `json:"path"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

type UploadIDCardRes struct {
	g.Meta `mime:"application/json"`
	IDCardUploadInfo
}
