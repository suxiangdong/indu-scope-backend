package adminattributes

import (
	"testing"

	attributeModel "github.com/suxiangdong/indu-scope-backend/internal/model/attribute"
)

func TestValidateAttributeInput(t *testing.T) {
	valid := attributeModel.AttributeInput{
		Name: "额定电压", ValueType: 3, InputType: "number", Status: 1,
		ValidationRules: `{"min":0,"max":1000}`,
	}
	if err := validateAttributeInput(valid); err != nil {
		t.Fatalf("valid attribute rejected: %v", err)
	}

	invalidJSON := valid
	invalidJSON.ValidationRules = "not-json"
	if err := validateAttributeInput(invalidJSON); err == nil {
		t.Fatal("invalid validation JSON was accepted")
	}

	invalidType := valid
	invalidType.ValueType = 7
	if err := validateAttributeInput(invalidType); err == nil {
		t.Fatal("invalid value type was accepted")
	}
}

func TestValidateOptionInput(t *testing.T) {
	if err := validateOptionInput(attributeModel.AttributeOptionInput{
		Name: "高压", Status: 1,
	}); err != nil {
		t.Fatalf("valid option rejected: %v", err)
	}
}

func TestGenerateCode(t *testing.T) {
	first, err := generateCode("attr")
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}
	second, err := generateCode("attr")
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}
	if first == second || len(first) > 20 || len(second) > 20 {
		t.Fatalf("generated codes are not unique or too long: %q, %q", first, second)
	}
}
