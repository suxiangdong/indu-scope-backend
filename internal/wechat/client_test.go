package wechat

import (
	"encoding/json"
	"testing"
)

func TestPhoneNumberResponseDecodesBothCountryCodeFormats(t *testing.T) {
	for _, countryCode := range []string{`86`, `"86"`} {
		var response phoneNumberResponse
		payload := []byte(`{
		"errcode": 0,
		"errmsg": "ok",
		"phone_info": {
			"phoneNumber": "+86 13800138000",
			"purePhoneNumber": "13800138000",
			"countryCode": ` + countryCode + `
		}
	}`)
		if err := json.Unmarshal(payload, &response); err != nil {
			t.Fatalf("decode phone response (%s): %v", countryCode, err)
		}
		if response.PhoneInfo.PurePhone != "13800138000" {
			t.Fatalf("unexpected phone number: %q", response.PhoneInfo.PurePhone)
		}
	}
}
