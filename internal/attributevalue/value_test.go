package attributevalue

import (
	"testing"

	supplierProductModel "github.com/suxiangdong/indu-scope-backend/internal/model/supplierproduct"
)

func TestNonOptionAttributesRejectOptionIDsBeforeEncoding(t *testing.T) {
	for _, value := range []supplierProductModel.AttributeValueInput{
		{OptionID: 20},
		{Value: "legitimate text", OptionIDs: []uint64{20}},
	} {
		definition := supplierProductModel.AttributeItem{Name: "text attribute", ValueType: 1, InputType: "text"}
		if err := Validate(definition, value); err == nil {
			t.Fatal("a foreign option was accepted on a non-option attribute")
		}
		if data, _, err := Data(definition, value); err == nil || data != nil {
			t.Fatalf("invalid option encoded into database fields: %+v %v", data, err)
		}
	}
}

func TestOptionsAreOwnedByDefinitionAndDeduplicated(t *testing.T) {
	definition := supplierProductModel.AttributeItem{Name: "selection", ValueType: 1, InputType: "checkbox", Options: []supplierProductModel.AttributeOptionItem{{ID: 20}, {ID: 21}}}
	data, ids, err := Data(definition, supplierProductModel.AttributeValueInput{OptionID: 20, OptionIDs: []uint64{20, 0, 21}})
	if err != nil || len(ids) != 2 || ids[0] != 20 || ids[1] != 21 || data["value_option_id"] != uint64(20) {
		t.Fatalf("valid options: %+v %v %v", data, ids, err)
	}
	if _, _, err := Data(definition, supplierProductModel.AttributeValueInput{OptionID: 30}); err == nil {
		t.Fatal("another definition's option was accepted")
	}
}

func TestAttributeTypesRejectInvalidValues(t *testing.T) {
	for _, test := range []struct {
		name  string
		kind  uint
		value string
	}{
		{"integer infinity", 2, "+Inf"},
		{"integer nan", 2, "NaN"},
		{"integer fraction", 2, "1.5"},
		{"decimal infinity", 3, "-Inf"},
		{"decimal overflow", 3, "1000000000000"},
		{"boolean", 4, "unknown"},
		{"date", 5, "2026-02-30"},
		{"json", 6, "{broken"},
	} {
		t.Run(test.name, func(t *testing.T) {
			definition := supplierProductModel.AttributeItem{Name: test.name, ValueType: test.kind}
			if _, _, err := Data(definition, supplierProductModel.AttributeValueInput{Value: test.value}); err == nil {
				t.Fatal("invalid value was accepted")
			}
		})
	}
}

func TestAttributeDataPreservesNumbersBooleansDatesAndJSON(t *testing.T) {
	for _, test := range []struct {
		kind  uint
		value string
		field string
		want  any
	}{
		{2, "12", "value_number", float64(12)},
		{3, "12.125", "value_number", 12.125},
		{4, "false", "value_bool", 0},
		{5, "2026-11-17", "value_text", "2026-11-17"},
		{6, `{"a":true}`, "value_json", `{"a":true}`},
	} {
		data, _, err := Data(supplierProductModel.AttributeItem{ValueType: test.kind}, supplierProductModel.AttributeValueInput{Value: test.value})
		if err != nil || data[test.field] != test.want {
			t.Fatalf("kind %d: %+v %v", test.kind, data, err)
		}
	}
	data, _, err := Data(supplierProductModel.AttributeItem{ValueType: 4}, supplierProductModel.AttributeValueInput{})
	if err != nil || data["value_bool"] != nil {
		t.Fatalf("absent boolean must stay null: %+v %v", data, err)
	}
}
