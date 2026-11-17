package wechatpay

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
)

func TestSignedRefundCreateAndQuery(t *testing.T) {
	cfg, merchant, platform := paymentFixture(t)
	for _, trusted := range []bool{true, false} {
		t.Run(strconv.FormatBool(trusted), func(t *testing.T) {
			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				var body []byte
				if req.Body != nil {
					body, _ = io.ReadAll(req.Body)
				}
				if req.Header.Get("Wechatpay-Serial") != cfg.PublicKeyID {
					t.Fatal("refund did not request public-key verification")
				}
				fields := map[string]string{}
				for _, m := range regexp.MustCompile(`(\w+)="([^"]+)"`).FindAllStringSubmatch(req.Header.Get("Authorization"), -1) {
					fields[m[1]] = m[2]
				}
				verifyTestSignature(t, &merchant.PublicKey, req.Method+"\n"+req.URL.RequestURI()+"\n"+fields["timestamp"]+"\n"+fields["nonce_str"]+"\n"+string(body)+"\n", fields["signature"])
				if req.Method == "POST" {
					var payload struct {
						RefundNo      string                        `json:"out_refund_no"`
						TransactionID string                        `json:"transaction_id"`
						NotifyURL     string                        `json:"notify_url"`
						Amount        struct{ Total, Refund int64 } `json:"amount"`
					}
					if err := json.Unmarshal(body, &payload); err != nil {
						t.Fatal(err)
					}
					if payload.RefundNo != "RFTEST" || payload.TransactionID != "WXTRANSACTION" || payload.Amount.Total != 20000 || payload.Amount.Refund != 20000 || payload.NotifyURL != "https://api.example.com/api/v1/payment/wechat/refund/notify" {
						t.Fatal("incorrect frozen refund request")
					}
				} else if req.URL.Path != "/v3/refund/domestic/refunds/RFTEST" {
					t.Fatal("wrong refund query path")
				}
				response := []byte(`{"refund_id":"WXREFUND","out_refund_no":"RFTEST","transaction_id":"WXTRANSACTION","out_trade_no":"VIPTEST","status":"PROCESSING","amount":{"total":20000,"refund":20000,"currency":"CNY"}}`)
				headers := signedHeaders(t, platform, cfg.PublicKeyID, response, strconv.FormatInt(time.Now().Unix(), 10))
				if !trusted {
					response = bytes.ReplaceAll(response, []byte("20000"), []byte("1"))
				}
				return &http.Response{StatusCode: 200, Header: headers, Body: io.NopCloser(bytes.NewReader(response)), Request: req}, nil
			})
			client, err := newClient(context.Background(), cfg, option.WithHTTPClient(&http.Client{Transport: transport}))
			if err != nil {
				t.Fatal(err)
			}
			for i := 0; i < 2; i++ {
				result, err := client.CreateRefund(context.Background(), RefundInput{OrderNo: "VIPTEST", TransactionID: "WXTRANSACTION", RefundNo: "RFTEST", Reason: "测试退款", Currency: "CNY", TotalFen: 20000, RefundFen: 20000})
				if trusted && (err != nil || result.Status != "PROCESSING") {
					t.Fatalf("refund result=%v err=%v", result, err)
				}
				if !trusted && err == nil {
					t.Fatal("tampered refund response accepted")
				}
			}
			_, err = client.QueryRefund(context.Background(), "RFTEST")
			if (err == nil) != trusted {
				t.Fatalf("query trust check err=%v", err)
			}
		})
	}
}

func TestRefundNotificationVerification(t *testing.T) {
	cfg, _, key := paymentFixture(t)
	client, err := newClient(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []string{"valid", "wrong merchant", "wrong amount event", "tampered", "bad nonce"} {
		t.Run(test, func(t *testing.T) {
			merchant := cfg.MchID
			if test == "wrong merchant" {
				merchant = "different"
			}
			plain := []byte(`{"mchid":"` + merchant + `","out_trade_no":"VIPTEST","transaction_id":"WXTRANSACTION","out_refund_no":"RFTEST","refund_id":"WXREFUND","refund_status":"SUCCESS","success_time":"2026-11-17T10:00:00+08:00","amount":{"total":20000,"refund":20000,"payer_refund":20000}}`)
			block, _ := aes.NewCipher([]byte(cfg.APIv3Key))
			gcm, _ := cipher.NewGCM(block)
			nonce := "0123456789ab"
			ciphertext := gcm.Seal(nil, []byte(nonce), plain, []byte("refund"))
			if test == "bad nonce" {
				nonce = "bad"
			}
			eventType := "REFUND.SUCCESS"
			if test == "wrong amount event" {
				eventType = "REFUND.CLOSED"
			}
			body, _ := json.Marshal(map[string]any{"id": "refund-event", "event_type": eventType, "resource_type": "encrypt-resource", "resource": map[string]string{"algorithm": "AEAD_AES_256_GCM", "original_type": "refund", "associated_data": "refund", "nonce": nonce, "ciphertext": base64.StdEncoding.EncodeToString(ciphertext)}})
			headers := signedHeaders(t, key, cfg.PublicKeyID, body, strconv.FormatInt(time.Now().Unix(), 10))
			if test == "tampered" {
				body = bytes.ReplaceAll(body, []byte("refund-event"), []byte("altered-event"))
			}
			req := httptest.NewRequest("POST", "https://api.example.com/api/v1/payment/wechat/refund/notify", bytes.NewReader(body))
			req.Header = headers
			n, err := client.ParseRefundNotification(context.Background(), req)
			if test == "valid" {
				if err != nil || n.Result.RefundFen != 20000 || n.Result.Status != "SUCCESS" || n.Result.SuccessAt == nil {
					t.Fatalf("callback result=%v err=%v", n, err)
				}
			} else if err == nil {
				t.Fatal("invalid refund callback accepted")
			}
		})
	}
}

func TestRefundCallbackURLAndFullAmount(t *testing.T) {
	cfg, _, _ := paymentFixture(t)
	client, err := newClient(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	client.Config.NotifyURL = "https://api.example.com/prefix/api/v1/payment/wechat/notify"
	url, err := client.RefundCallbackURL()
	if err != nil || url != "https://api.example.com/prefix/api/v1/payment/wechat/refund/notify" {
		t.Fatal("refund callback prefix lost")
	}
	client.Config.RefundNotifyURL = "http://insecure.example.com/notify"
	if _, err := client.RefundCallbackURL(); err == nil {
		t.Fatal("insecure refund URL accepted")
	}
	if _, err := client.CreateRefund(context.Background(), RefundInput{RefundNo: "RF", OrderNo: "VIP", TransactionID: "WX", TotalFen: 20000, RefundFen: 1, Currency: "CNY"}); err == nil {
		t.Fatal("partial refund unexpectedly accepted")
	}
	client.Config.RefundNotifyURL = ""
	client.Config.NotifyURL = "https://api.example.com/custom"
	if _, err := client.RefundCallbackURL(); err == nil || !strings.Contains(err.Error(), "refund_notify_url") {
		t.Fatal("custom callback URL should require explicit refund URL")
	}
}
