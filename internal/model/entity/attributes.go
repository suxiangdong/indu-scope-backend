// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Attributes is the golang structure for table attributes.
type Attributes struct {
	Id              uint64    `json:"id"              orm:"id"               description:""` //
	Code            string    `json:"code"            orm:"code"             description:""` //
	Name            string    `json:"name"            orm:"name"             description:""` //
	ValueType       uint      `json:"valueType"       orm:"value_type"       description:""` //
	InputType       string    `json:"inputType"       orm:"input_type"       description:""` //
	Unit            string    `json:"unit"            orm:"unit"             description:""` //
	Placeholder     string    `json:"placeholder"     orm:"placeholder"      description:""` //
	ValidationRules string    `json:"validationRules" orm:"validation_rules" description:""` //
	HelpText        string    `json:"helpText"        orm:"help_text"        description:""` //
	Status          uint      `json:"status"          orm:"status"           description:""` //
	CreatedAt       time.Time `json:"createdAt"       orm:"created_at"       description:""` //
	UpdatedAt       time.Time `json:"updatedAt"       orm:"updated_at"       description:""` //
}
