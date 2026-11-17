// Package attributevalue validates and encodes the dynamic attributes shared
// by product and procurement requirement publishing.
package attributevalue

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	supplierProductModel "github.com/suxiangdong/indu-scope-backend/internal/model/supplierproduct"
)

func optionInput(definition supplierProductModel.AttributeItem) bool {
	inputType := strings.ToLower(definition.InputType)
	return len(definition.Options) > 0 || strings.Contains(inputType, "select") || strings.Contains(inputType, "radio") || strings.Contains(inputType, "checkbox")
}

func optionIDs(value supplierProductModel.AttributeValueInput) []uint64 {
	ids := append([]uint64{value.OptionID}, value.OptionIDs...)
	result := make([]uint64, 0, len(ids))
	seen := make(map[uint64]bool, len(ids))
	for _, id := range ids {
		if id != 0 && !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	return result
}

func invalid(message, name string) error {
	return &consts.InvalidParameterError{Message: message + "：" + name}
}

func Validate(definition supplierProductModel.AttributeItem, value supplierProductModel.AttributeValueInput) error {
	ids := optionIDs(value)
	if optionInput(definition) {
		if len(ids) == 0 {
			return invalid("请选择属性", definition.Name)
		}
		allowed := make(map[uint64]bool, len(definition.Options))
		for _, option := range definition.Options {
			allowed[option.ID] = true
		}
		for _, id := range ids {
			if !allowed[id] {
				return invalid("属性选项无效", definition.Name)
			}
		}
		return nil
	}
	if len(ids) != 0 {
		return invalid("此属性不接受选项", definition.Name)
	}
	if value.Value == "" {
		return nil
	}
	switch definition.ValueType {
	case 2, 3:
		number, err := strconv.ParseFloat(value.Value, 64)
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) || math.Abs(number) >= 1e12 {
			return invalid("属性数字无效或超出范围", definition.Name)
		}
		if definition.ValueType == 2 && math.Trunc(number) != number {
			return invalid("属性必须是整数", definition.Name)
		}
	case 4:
		if _, err := parseBool(value.Value); err != nil {
			return invalid("属性必须是布尔值", definition.Name)
		}
	case 5:
		if _, err := time.Parse("2006-01-02", value.Value); err != nil {
			return invalid("属性日期格式必须为 YYYY-MM-DD", definition.Name)
		}
	case 6:
		if !json.Valid([]byte(value.Value)) {
			return invalid("属性必须是有效的 JSON", definition.Name)
		}
	default:
		if len([]rune(value.Value)) > 1000 {
			return invalid("属性文本不能超过1000个字符", definition.Name)
		}
	}
	return nil
}

// Data validates again so callers cannot encode an option from a different
// attribute or silently turn invalid numeric input into a null database value.
func Data(definition supplierProductModel.AttributeItem, value supplierProductModel.AttributeValueInput) (g.Map, []uint64, error) {
	if err := Validate(definition, value); err != nil {
		return nil, nil, err
	}
	data := g.Map{"value_number": nil, "value_text": nil, "value_bool": nil, "value_option_id": nil, "value_json": nil}
	if optionInput(definition) {
		ids := optionIDs(value)
		data["value_option_id"] = ids[0]
		return data, ids, nil
	}
	if value.Value == "" {
		return data, nil, nil
	}
	switch definition.ValueType {
	case 2, 3:
		number, _ := strconv.ParseFloat(value.Value, 64)
		data["value_number"] = number
	case 4:
		value, _ := parseBool(value.Value)
		if value {
			data["value_bool"] = 1
		} else {
			data["value_bool"] = 0
		}
	case 6:
		data["value_json"] = value.Value
	default:
		data["value_text"] = value.Value
	}
	return data, nil, nil
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "是":
		return true, nil
	case "0", "false", "no", "否":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value")
	}
}
