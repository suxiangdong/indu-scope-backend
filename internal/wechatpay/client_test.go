package wechatpay

import (
	"bytes"
	"context"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func paymentFixture(t *testing.T) (Config, *rsa.PrivateKey, *rsa.PrivateKey) {
	t.Helper()
	merchant, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	platform, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	privateDER, err := x509.MarshalPKCS8PrivateKey(merchant)
	if err != nil {
		t.Fatal(err)
	}
	publicDER, err := x509.MarshalPKIXPublicKey(&platform.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	privatePath, publicPath := filepath.Join(dir, "merchant.pem"), filepath.Join(dir, "pub_key.pem")
	if err := os.WriteFile(privatePath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privateDER}), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(publicPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: publicDER}), 0600); err != nil {
		t.Fatal(err)
	}
	return Config{Enabled: true, AppID: "wx_test", MchID: "1005166565", MerchantSerialNo: "ABC123", MerchantPrivateKeyPath: privatePath,
		APIv3Key: "0123456789abcdef0123456789abcdef", PublicKeyID: "PUB_KEY_ID_TEST", PublicKeyPath: publicPath, NotifyURL: "https://api.example.com/api/v1/payment/wechat/notify"}, merchant, platform
}

func signTest(t *testing.T, key *rsa.PrivateKey, message string) string {
	t.Helper()
	digest := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(signature)
}

func signedHeaders(t *testing.T, key *rsa.PrivateKey, keyID string, body []byte, ts string) http.Header {
	t.Helper()
	h := http.Header{}
	nonce := "test-response-nonce"
	h.Set("Wechatpay-Serial", keyID)
	h.Set("Wechatpay-Timestamp", ts)
	h.Set("Wechatpay-Nonce", nonce)
	h.Set("Wechatpay-Signature", signTest(t, key, ts+"\n"+nonce+"\n"+string(body)+"\n"))
	h.Set("Content-Type", "application/json")
	h.Set("Request-ID", "fixture-request-id")
	return h
}

func TestPublicKeyPrepayAndQuery(t *testing.T) {
	cfg, merchant, platform := paymentFixture(t)
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Host != "api.mch.weixin.qq.com" {
			t.Fatalf("unexpected host %s", req.URL.Host)
		}
		if req.Header.Get("Wechatpay-Serial") != cfg.PublicKeyID {
			t.Fatal("SDK did not select public key response mode")
		}
		var body []byte
		if req.Body != nil {
			body, _ = io.ReadAll(req.Body)
		}
		fields := map[string]string{}
		for _, m := range regexp.MustCompile(`(\w+)="([^"]+)"`).FindAllStringSubmatch(req.Header.Get("Authorization"), -1) {
			fields[m[1]] = m[2]
		}
		if fields["mchid"] != cfg.MchID || fields["serial_no"] != cfg.MerchantSerialNo {
			t.Fatal("merchant authorization identity mismatch")
		}
		verifyTestSignature(t, &merchant.PublicKey, req.Method+"\n"+req.URL.RequestURI()+"\n"+fields["timestamp"]+"\n"+fields["nonce_str"]+"\n"+string(body)+"\n", fields["signature"])
		responseBody := []byte(`{"prepay_id":"wx_prepay_test"}`)
		if req.Method == http.MethodPost {
			var payload struct {
				AppID  string `json:"appid"`
				Amount struct {
					Total int64 `json:"total"`
				} `json:"amount"`
				TimeExpire string `json:"time_expire"`
			}
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Fatal(err)
			}
			if payload.AppID != cfg.AppID || payload.Amount.Total != 1 || payload.TimeExpire == "" {
				t.Fatalf("unexpected prepay payload %s", body)
			}
		} else {
			if req.URL.Query().Get("mchid") != cfg.MchID {
				t.Fatal("query merchant missing")
			}
			responseBody = []byte(`{"appid":"wx_test","mchid":"1005166565","out_trade_no":"VIPTEST","trade_state":"SUCCESS"}`)
		}
		return &http.Response{StatusCode: 200, Header: signedHeaders(t, platform, cfg.PublicKeyID, responseBody, strconv.FormatInt(time.Now().Unix(), 10)), Body: io.NopCloser(bytes.NewReader(responseBody)), Request: req}, nil
	})
	client, err := newClient(context.Background(), cfg, option.WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatal(err)
	}
	p, err := client.Prepay(context.Background(), PrepayInput{OrderNo: "VIPTEST", Description: "会员", OpenID: "test_openid", Currency: "CNY", AmountFen: 1, ExpiresAt: time.Now().Add(30 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if p.PrepayID != "wx_prepay_test" || p.SignType != "RSA" {
		t.Fatal("invalid mini-program payment result")
	}
	verifyTestSignature(t, &merchant.PublicKey, fmt.Sprintf("%s\n%s\n%s\n%s\n", cfg.AppID, p.TimeStamp, p.NonceStr, p.Package), p.PaySign)
	tx, err := client.Query(context.Background(), "VIPTEST")
	if err != nil {
		t.Fatal(err)
	}
	if tx.TradeState == nil || *tx.TradeState != "SUCCESS" {
		t.Fatal("missing verified query response")
	}
}

func verifyTestSignature(t *testing.T, key *rsa.PublicKey, message, signature string) {
	t.Helper()
	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256([]byte(message))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], sig); err != nil {
		t.Fatal(err)
	}
}

func TestPrepayRejectsUntrustedResponse(t *testing.T) {
	cfg, _, platform := paymentFixture(t)
	for _, test := range []string{"tampered body", "wrong key id", "expired timestamp", "signature probe"} {
		t.Run(test, func(t *testing.T) {
			transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
				body := []byte(`{"prepay_id":"original"}`)
				headers := signedHeaders(t, platform, cfg.PublicKeyID, body, strconv.FormatInt(time.Now().Unix(), 10))
				switch test {
				case "tampered body":
					body = []byte(`{"prepay_id":"attacker"}`)
				case "wrong key id":
					headers.Set("Wechatpay-Serial", "PUB_KEY_ID_OTHER")
				case "expired timestamp":
					headers = signedHeaders(t, platform, cfg.PublicKeyID, body, strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10))
				case "signature probe":
					headers.Set("Wechatpay-Signature", "WECHATPAY/SIGNTEST/probe")
				}
				return &http.Response{StatusCode: 200, Header: headers, Body: io.NopCloser(bytes.NewReader(body)), Request: req}, nil
			})
			client, err := newClient(context.Background(), cfg, option.WithHTTPClient(&http.Client{Transport: transport}))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := client.Prepay(context.Background(), PrepayInput{OrderNo: "VIPTEST", Description: "会员", OpenID: "openid", Currency: "CNY", AmountFen: 1, ExpiresAt: time.Now().Add(time.Hour)}); err == nil {
				t.Fatal("untrusted response was accepted")
			}
		})
	}
}

func callbackBody(t *testing.T, cfg Config) []byte {
	t.Helper()
	block, err := aes.NewCipher([]byte(cfg.APIv3Key))
	if err != nil {
		t.Fatal(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatal(err)
	}
	nonce := "0123456789ab"
	associated := "transaction"
	plain := []byte(`{"appid":"wx_test","mchid":"1005166565","out_trade_no":"VIPTEST","trade_state":"SUCCESS","transaction_id":"WXTEST"}`)
	encrypted := gcm.Seal(nil, []byte(nonce), plain, []byte(associated))
	body, err := json.Marshal(map[string]any{"id": "event-test", "event_type": "TRANSACTION.SUCCESS", "resource_type": "encrypt-resource", "resource": map[string]string{
		"original_type": "transaction", "algorithm": "AEAD_AES_256_GCM", "nonce": nonce, "associated_data": associated, "ciphertext": base64.StdEncoding.EncodeToString(encrypted)}})
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func TestPublicKeyNotificationVerification(t *testing.T) {
	cfg, _, platform := paymentFixture(t)
	client, err := newClient(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	body := callbackBody(t, cfg)
	for _, test := range []string{"valid", "tampered", "wrong key id", "expired", "bad nonce", "wrong ciphertext"} {
		t.Run(test, func(t *testing.T) {
			payload := append([]byte{}, body...)
			if test == "bad nonce" {
				payload = bytes.Replace(payload, []byte("0123456789ab"), []byte("short"), 1)
			}
			if test == "wrong ciphertext" {
				var envelope map[string]any
				_ = json.Unmarshal(payload, &envelope)
				envelope["resource"].(map[string]any)["ciphertext"] = "AAAAAAAAAAAAAAAAAAAAAA=="
				payload, _ = json.Marshal(envelope)
			}
			headers := signedHeaders(t, platform, cfg.PublicKeyID, payload, strconv.FormatInt(time.Now().Unix(), 10))
			switch test {
			case "tampered":
				payload = bytes.Replace(payload, []byte("event-test"), []byte("event-evil"), 1)
			case "wrong key id":
				headers.Set("Wechatpay-Serial", "PUB_KEY_ID_OTHER")
			case "expired":
				headers = signedHeaders(t, platform, cfg.PublicKeyID, payload, strconv.FormatInt(time.Now().Add(-10*time.Minute).Unix(), 10))
			}
			req := httptest.NewRequest(http.MethodPost, cfg.NotifyURL, bytes.NewReader(payload))
			req.Header = headers
			n, err := client.ParseNotification(context.Background(), req)
			if test == "valid" {
				if err != nil {
					t.Fatal(err)
				}
				if n.ID != "event-test" || *n.Transaction.OutTradeNo != "VIPTEST" {
					t.Fatal("unexpected notification")
				}
			} else if err == nil {
				t.Fatal("invalid notification accepted")
			}
		})
	}
}

func TestConfigRejectsMissingPublicKeyMode(t *testing.T) {
	cfg, _, _ := paymentFixture(t)
	for _, test := range []string{"platform serial as public id", "short api key", "insecure callback", "callback query"} {
		t.Run(test, func(t *testing.T) {
			c := cfg
			switch test {
			case "platform serial as public id":
				c.PublicKeyID = "ABC123"
			case "short api key":
				c.APIv3Key = "short"
			case "insecure callback":
				c.NotifyURL = "http://api.example.com/notify"
			case "callback query":
				c.NotifyURL += "?token=1"
			}
			if err := c.validate(); err == nil {
				t.Fatal("invalid configuration accepted")
			}
		})
	}
	if !strings.Contains(shortDescription(strings.Repeat("会", 100)), "会") || len(shortDescription(strings.Repeat("会", 100))) > 127 {
		t.Fatal("description limit is not enforced")
	}
}
