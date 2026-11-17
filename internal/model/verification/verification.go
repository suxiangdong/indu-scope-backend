package verification

import "io"

const (
	ReviewStatusPending  uint8 = 0
	ReviewStatusApproved uint8 = 1
	ReviewStatusRejected uint8 = 2
)

// Summary is the authenticated user's verification status and submitted data.
type Summary struct {
	Role         uint8
	UserType     uint8
	Status       string
	ReviewReason string
	Personal     *PersonalInfo
	Company      *CompanyInfo
}

type PersonalInfo struct {
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
	VerifiedAt        string
	ReviewStatus      uint8
	ReviewReason      string
	ReviewedAt        string
}

type CompanyInfo struct {
	CompanyName                string
	UnifiedSocialCreditCode    string
	BusinessLicenseEndpoint    string
	BusinessLicenseBucket      string
	BusinessLicensePath        string
	BusinessLicenseFileName    string
	BusinessLicenseFileSize    int64
	BusinessLicenseContentType string
	RegisteredAddress          string
	Introduction               string
	StorePlatformID            uint64
	StorePlatformName          string
	StoreURL                   string
	Offices                    []string
	Factories                  []string
	Contacts                   []ContactInfo
	VerifiedAt                 string
	ReviewStatus               uint8
	ReviewReason               string
	ReviewedAt                 string
}

type ContactInfo struct {
	Name     string
	Position string
	Phone    string
}

type PersonalInput struct {
	Role            uint8
	RealName        string
	IDNumber        string
	StorePlatformID uint64
	StoreURL        string
	IDFrontEndpoint string
	IDFrontBucket   string
	IDFrontPath     string
	IDBackEndpoint  string
	IDBackBucket    string
	IDBackPath      string
}

type CompanyInput struct {
	Role                    uint8
	CompanyName             string
	UnifiedSocialCreditCode string
	BusinessLicenseEndpoint string
	BusinessLicenseBucket   string
	BusinessLicensePath     string
	RegisteredAddress       string
	Introduction            string
	StorePlatformID         uint64
	StoreURL                string
	Offices                 []string
	Factories               []string
	Contacts                []ContactInput
}

// BusinessLicenseUploadInput contains the authenticated upload stream and
// metadata received from the multipart request.
type BusinessLicenseUploadInput struct {
	Reader      io.Reader
	Filename    string
	ContentType string
	Size        int64
}

// BusinessLicenseInfo is the persisted result of an OSS upload. The object is
// identified by its endpoint, bucket, and path, plus safe display metadata.
type BusinessLicenseInfo struct {
	Endpoint    string
	Bucket      string
	Path        string
	Filename    string
	ContentType string
	Size        int64
}

// BusinessLicenseOCR contains safe fields returned by business-license OCR.
type BusinessLicenseOCR struct {
	CompanyName             string
	UnifiedSocialCreditCode string
	BusinessAddress         string
}

type IDCardUploadInput struct {
	Side        string
	Reader      io.Reader
	Filename    string
	ContentType string
	Size        int64
}

type IDCardUploadInfo struct {
	Side        string
	Endpoint    string
	Bucket      string
	Path        string
	Filename    string
	ContentType string
	Size        int64
}

type ContactInput struct {
	Name     string
	Position string
	Phone    string
}
