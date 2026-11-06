package adminverification

import "time"

const (
	TypePersonal = "personal"
	TypeCompany  = "company"

	ActionApprove = "approve"
	ActionReject  = "reject"
)

type ListFilter struct {
	Type         string
	ReviewStatus *uint8
}

type ListResult struct {
	Items []ListItem
}

type ListItem struct {
	ID           uint64
	UserID       uint64
	Type         string
	Phone        string
	UserType     uint8
	DisplayName  string
	ReviewStatus uint8
	ReviewReason string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	ReviewedAt   *time.Time
}

type Detail struct {
	ListItem
	UserStatus uint8
	Personal   *PersonalDetail
	Company    *CompanyDetail
}

type PersonalDetail struct {
	RealName          string
	IDNumberMasked    string
	StorePlatformID   uint64
	StorePlatformName string
	StoreURL          string
	IDFrontEndpoint   string
	IDFrontBucket     string
	IDFrontPath       string
	IDBackEndpoint    string
	IDBackBucket      string
	IDBackPath        string
	VerifiedAt        *time.Time
}

type CompanyDetail struct {
	CompanyName                string
	UnifiedSocialCreditCode    string
	BusinessLicenseEndpoint    string
	BusinessLicenseBucket      string
	BusinessLicensePath        string
	BusinessLicenseURL         string
	BusinessLicenseFileName    string
	BusinessLicenseFileSize    int64
	BusinessLicenseContentType string
	RegisteredAddress          string
	Introduction               string
	Offices                    []string
	Factories                  []string
	Contacts                   []Contact
	VerifiedAt                 *time.Time
}

type Contact struct {
	Name     string
	Position string
	Phone    string
}

type ReviewInput struct {
	ID     uint64
	Type   string
	Action string
	Reason string
}
