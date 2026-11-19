package wechatpay

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestMerchantOrderNumberConstraints(t *testing.T) {
	for _, orderNo := range []string{"ABC123", strings.Repeat("A", 32), "VIPabc_123-XYZ|*"} {
		if err := ValidateOrderNo(orderNo); err != nil {
			t.Fatalf("valid merchant order number rejected: %v", err)
		}
	}
	for _, orderNo := range []string{"", "VIP12", strings.Repeat("A", 33),
		"VIP20261119102030" + strings.Repeat("a", 24), "VIP/123", "VIP 123", "VIP会员订单"} {
		t.Run(orderNo, func(t *testing.T) {
			if !errors.Is(ValidateOrderNo(orderNo), ErrInvalidOrderNo) {
				t.Fatal("invalid merchant order number accepted")
			}
			// An uninitialized client cannot make an SDK or network request.
			// All three entry points must reject the identifier before using it.
			client := &Client{}
			_, err := client.Prepay(context.Background(), PrepayInput{OrderNo: orderNo,
				OpenID: "test_openid", AmountFen: 1, ExpiresAt: time.Now().Add(time.Hour)})
			if !errors.Is(err, ErrInvalidOrderNo) {
				t.Fatalf("prepay error = %v", err)
			}
			if _, err := client.Query(context.Background(), orderNo); !errors.Is(err, ErrInvalidOrderNo) {
				t.Fatalf("query error = %v", err)
			}
			if err := client.Close(context.Background(), orderNo); !errors.Is(err, ErrInvalidOrderNo) {
				t.Fatalf("close error = %v", err)
			}
		})
	}
}
