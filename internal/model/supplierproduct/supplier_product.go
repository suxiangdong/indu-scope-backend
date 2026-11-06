package supplierproduct

type CategoryItem struct {
	ID        uint64
	ParentID  uint64
	Name      string
	SortOrder uint
}

type AttributeOptionItem struct {
	ID   uint64
	Code string
	Name string
}

type AttributeItem struct {
	ID               uint64
	Code             string
	Name             string
	ValueType        uint
	InputType        string
	Unit             string
	Placeholder      string
	HelpText         string
	IsRequired       uint8
	SortOrder        uint
	GroupDisplayName string
	Options          []AttributeOptionItem
}

type FormOptions struct {
	Categories []CategoryItem
	Attributes []AttributeItem
}

type AttributeValueInput struct {
	AttributeID uint64
	Value       string
	OptionID    uint64
	OptionIDs   []uint64
}

type CreateInput struct {
	SupplierUserID  uint64
	CategoryID      uint64
	Name            string
	DescriptionHTML string
	ConditionCode   string
	StockQuantity   float64
	StockUnitCode   string
	Attributes      []AttributeValueInput
}

type ProductItem struct {
	ID              uint64
	SupplierUserID  uint64
	CategoryID      uint64
	Name            string
	DescriptionHTML string
	ConditionCode   string
	StockQuantity   float64
	StockUnitCode   string
	Status          uint
}
