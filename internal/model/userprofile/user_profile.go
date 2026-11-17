package userprofile

type Contact struct {
	Name     string
	Position string
	Phone    string
}

type Profile struct {
	ID                uint64
	CompanyName       string
	Role              uint8
	Introduction      string
	RegisteredAddress string
	Offices           []string
	Factories         []string
	Contacts          []Contact
}
