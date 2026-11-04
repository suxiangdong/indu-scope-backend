package attribute

import "time"

// AttributeItem is an attribute definition exposed by management APIs.
type AttributeItem struct {
	ID              uint64
	Code            string
	Name            string
	ValueType       uint8
	InputType       string
	Unit            string
	Placeholder     string
	ValidationRules string
	HelpText        string
	Status          uint8
	OptionCount     uint
	GroupCount      uint
	GroupIDs        []uint64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type AttributeInput struct {
	Code            string
	Name            string
	ValueType       uint
	InputType       string
	Unit            string
	Placeholder     string
	ValidationRules string
	HelpText        string
	Status          uint
	GroupIDs        []uint64
}

type AttributeListResult struct {
	Items []AttributeItem
}

type AttributeOptionItem struct {
	ID          uint64
	AttributeID uint64
	Code        string
	Name        string
	SortOrder   uint
	Status      uint8
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AttributeOptionInput struct {
	Code      string
	Name      string
	SortOrder uint
	Status    uint
}

type AttributeOptionListResult struct {
	Items []AttributeOptionItem
}

type CategoryAttributeItem struct {
	CategoryID    uint64
	CategoryName  string
	AttributeID   uint64
	AttributeCode string
	AttributeName string
	GroupID       *uint64
	GroupCode     string
	IsRequired    uint8
	IsFilterable  uint8
	IsSearchable  uint8
	SortOrder     uint
	GroupName     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CategoryAttributeInput struct {
	CategoryID   uint64
	AttributeID  uint64
	GroupID      *uint64
	IsRequired   uint
	IsFilterable uint
	IsSearchable uint
	SortOrder    uint
	GroupName    string
}

type CategoryAttributeListResult struct {
	Items []CategoryAttributeItem
}

type AttributeGroupItem struct {
	ID             uint64
	Code           string
	Name           string
	SortOrder      uint
	Status         uint8
	AttributeCount uint
	CategoryCount  uint
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type AttributeGroupInput struct {
	Code      string
	Name      string
	SortOrder uint
	Status    uint
}

type AttributeGroupListResult struct {
	Items []AttributeGroupItem
}

type AttributeGroupAttributeItem struct {
	GroupID     uint64
	AttributeID uint64
	Code        string
	Name        string
	ValueType   uint8
	InputType   string
	SortOrder   uint
	Status      uint8
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type AttributeGroupAttributeInput struct {
	GroupID     uint64
	AttributeID uint64
	SortOrder   uint
}

type AttributeGroupAttributeListResult struct {
	Items []AttributeGroupAttributeItem
}

type CategoryGroupBindingItem struct {
	CategoryID     uint64
	CategoryName   string
	GroupID        uint64
	GroupCode      string
	GroupName      string
	SortOrder      uint
	Status         uint8
	AttributeCount uint
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CategoryGroupBindingInput struct {
	CategoryID uint64
	GroupID    uint64
	SortOrder  uint
	Status     uint
}

type CategoryGroupBindingListResult struct {
	Items []CategoryGroupBindingItem
}
