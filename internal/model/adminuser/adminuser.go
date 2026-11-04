package adminuser

import "time"

// AdminUserItem is the administrator data exposed by management APIs.
type AdminUserItem struct {
	ID          int64
	Username    string
	Name        string
	Phone       string
	Status      uint8
	LastLoginAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// AdminUserListResult contains all administrator accounts.
type AdminUserListResult struct {
	Items []AdminUserItem
}
