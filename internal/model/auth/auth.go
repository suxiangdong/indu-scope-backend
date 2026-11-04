package auth

// User is the authenticated user data returned by the auth service.
type User struct {
	ID       int64
	Phone    string
	Role     uint8
	UserType uint8
	Status   uint8
}

// AdminUser is the authenticated administrator data returned by the auth service.
type AdminUser struct {
	ID       int64
	Username string
	Name     string
	Phone    string
	Status   uint8
}

type AuthResult struct {
	AccessToken string
	User        User
}

type AdminAuthResult struct {
	AccessToken string
	User        AdminUser
}
