package admincustomeruser

import "time"

type ListFilter struct {
	Phone    string
	Role     *uint8
	UserType *uint8
	Status   *uint8
}

type UserItem struct {
	ID                 int64
	Phone              string
	Role               uint8
	UserType           uint8
	Status             uint8
	VerificationType   string
	VerificationStatus string
	WechatBound        bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type ListResult struct {
	Items []UserItem
}
