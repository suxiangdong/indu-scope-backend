package v1

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/frame/g"
)

func TestRefundRequestValidation(t *testing.T) {
	for _, test := range []struct {
		name    string
		input   any
		wantErr bool
	}{
		{"unfiltered refund list", ListRefundsReq{}, false},
		{"processing filter", ListRefundsReq{Status: "processing"}, false},
		{"invalid filter", ListRefundsReq{Status: "fake_success"}, true},
		{"missing refund order", CreateRefundReq{Reason: "用户申请退款"}, true},
		{"missing reason", CreateRefundReq{ID: 1}, true},
		{"valid refund request", CreateRefundReq{ID: 1, Reason: "用户申请退款"}, false},
		{"missing refund identity", RetryRefundReq{}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := g.Validator().Data(test.input).Run(context.Background()); (err != nil) != test.wantErr {
				t.Fatalf("request validation err=%v, wantErr=%t", err, test.wantErr)
			}
		})
	}
}
