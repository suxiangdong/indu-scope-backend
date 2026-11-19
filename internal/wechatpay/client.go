// Package wechatpay integrates ordinary mini-program payments using WeChat
// Pay's public key mode. Platform certificates are intentionally not used.
package wechatpay

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

var (
	ErrDisabled       = errors.New("微信支付暂未启用")
	ErrInvalidOrderNo = errors.New("WeChat Pay out_trade_no must contain 6 to 32 allowed ASCII characters")
)

// ValidateOrderNo applies the JSAPI contract before an invalid merchant order
// number can reach prepay, query or close: https://pay.weixin.qq.com/doc/v3/merchant/4012791897
func ValidateOrderNo(orderNo string) error {
	if len(orderNo) < 6 || len(orderNo) > 32 {
		return ErrInvalidOrderNo
	}
	for i := 0; i < len(orderNo); i++ {
		c := orderNo[i]
		if (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			c == '_' || c == '-' || c == '|' || c == '*' {
			continue
		}
		return ErrInvalidOrderNo
	}
	return nil
}

type Config struct {
	Enabled                                                          bool
	AppID, MchID, MerchantSerialNo, MerchantPrivateKeyPath           string
	APIv3Key, PublicKeyID, PublicKeyPath, NotifyURL, RefundNotifyURL string
}

func ReadConfig(ctx context.Context) Config {
	get := func(name string) string {
		if value, ok := os.LookupEnv("WECHAT_PAY_" + strings.ToUpper(name)); ok {
			return strings.TrimSpace(value)
		}
		value, _ := g.Cfg().Get(ctx, "wechat_pay."+name)
		if value == nil {
			return ""
		}
		return strings.TrimSpace(value.String())
	}
	enabled, _ := strconv.ParseBool(get("enabled"))
	appID := get("app_id")
	if appID == "" {
		if value, _ := g.Cfg().Get(ctx, "wechat.app_id"); value != nil {
			appID = strings.TrimSpace(value.String())
		}
	}
	return Config{Enabled: enabled, AppID: appID, MchID: get("mch_id"), MerchantSerialNo: get("merchant_serial_no"),
		MerchantPrivateKeyPath: get("merchant_private_key_path"), APIv3Key: get("api_v3_key"),
		PublicKeyID: get("public_key_id"), PublicKeyPath: get("public_key_path"), NotifyURL: get("notify_url"), RefundNotifyURL: get("refund_notify_url")}
}

func (c Config) validate() error {
	if !c.Enabled {
		return ErrDisabled
	}
	for _, field := range []struct{ name, value string }{
		{"app_id", c.AppID}, {"mch_id", c.MchID}, {"merchant_serial_no", c.MerchantSerialNo},
		{"merchant_private_key_path", c.MerchantPrivateKeyPath}, {"public_key_path", c.PublicKeyPath},
	} {
		if field.value == "" {
			return fmt.Errorf("wechat_pay.%s is required", field.name)
		}
	}
	if len(c.APIv3Key) != 32 {
		return errors.New("wechat_pay.api_v3_key must contain 32 bytes")
	}
	if !strings.HasPrefix(c.PublicKeyID, "PUB_KEY_ID_") || len(c.PublicKeyID) <= len("PUB_KEY_ID_") {
		return errors.New("wechat_pay.public_key_id must be a WeChat Pay PUB_KEY_ID")
	}
	u, err := url.Parse(c.NotifyURL)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil || u.Path == "" {
		return errors.New("wechat_pay.notify_url must be a public HTTPS callback URL without query parameters")
	}
	if u.Hostname() == "localhost" || strings.Contains(c.NotifyURL, "你的域名") {
		return errors.New("wechat_pay.notify_url must be configured with your public callback domain")
	}
	return nil
}

type Client struct {
	Config   Config
	sdk      *core.Client
	service  jsapi.JsapiApiService
	notifier *notify.Handler
}

func Load(ctx context.Context) (*Client, error) { return newClient(ctx, ReadConfig(ctx)) }

func newClient(ctx context.Context, cfg Config, options ...core.ClientOption) (*Client, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	privateKey, err := utils.LoadPrivateKeyWithPath(cfg.MerchantPrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load merchant private key: %w", err)
	}
	publicKey, err := utils.LoadPublicKeyWithPath(cfg.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load WeChat Pay public key: %w", err)
	}
	opts := []core.ClientOption{
		option.WithWechatPayPublicKeyAuthCipher(cfg.MchID, cfg.MerchantSerialNo, privateKey, cfg.PublicKeyID, publicKey),
		option.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
	}
	sdk, err := core.NewClient(ctx, append(opts, options...)...)
	if err != nil {
		return nil, err
	}
	handler, err := notify.NewRSANotifyHandler(cfg.APIv3Key, verifiers.NewSHA256WithRSAPubkeyVerifier(cfg.PublicKeyID, *publicKey))
	if err != nil {
		return nil, err
	}
	return &Client{Config: cfg, sdk: sdk, service: jsapi.JsapiApiService{Client: sdk}, notifier: handler}, nil
}

type PrepayInput struct {
	OrderNo, Description, OpenID, Currency string
	AmountFen                              int64
	ExpiresAt                              time.Time
}

type Payment struct {
	PrepayID  string `json:"-"`
	TimeStamp string `json:"time_stamp"`
	NonceStr  string `json:"nonce_str"`
	Package   string `json:"package"`
	SignType  string `json:"sign_type"`
	PaySign   string `json:"pay_sign"`
}

func (c *Client) Prepay(ctx context.Context, in PrepayInput) (*Payment, error) {
	if err := ValidateOrderNo(in.OrderNo); err != nil {
		return nil, err
	}
	if in.AmountFen <= 0 || in.OpenID == "" || in.OrderNo == "" || !in.ExpiresAt.After(time.Now()) {
		return nil, errors.New("invalid WeChat Pay order")
	}
	response, result, err := c.service.PrepayWithRequestPayment(ctx, jsapi.PrepayRequest{
		Appid: core.String(c.Config.AppID), Mchid: core.String(c.Config.MchID),
		Description: core.String(shortDescription(in.Description)), OutTradeNo: core.String(in.OrderNo),
		TimeExpire: &in.ExpiresAt, NotifyUrl: core.String(c.Config.NotifyURL),
		Amount: &jsapi.Amount{Total: core.Int64(in.AmountFen), Currency: core.String(in.Currency)},
		Payer:  &jsapi.Payer{Openid: core.String(in.OpenID)},
	})
	logRequestID(ctx, "prepay", result)
	if err != nil {
		return nil, err
	}
	if response == nil || response.PrepayId == nil || *response.PrepayId == "" {
		return nil, errors.New("missing prepay_id")
	}
	return &Payment{PrepayID: *response.PrepayId, TimeStamp: *response.TimeStamp, NonceStr: *response.NonceStr,
		Package: *response.Package, SignType: *response.SignType, PaySign: *response.PaySign}, nil
}

func (c *Client) SignPayment(ctx context.Context, prepayID string) (*Payment, error) {
	nonce, err := utils.GenerateNonce()
	if err != nil {
		return nil, err
	}
	p := &Payment{PrepayID: prepayID, TimeStamp: strconv.FormatInt(time.Now().Unix(), 10), NonceStr: nonce, Package: "prepay_id=" + prepayID, SignType: "RSA"}
	signature, err := c.sdk.Sign(ctx, fmt.Sprintf("%s\n%s\n%s\n%s\n", c.Config.AppID, p.TimeStamp, p.NonceStr, p.Package))
	if err != nil {
		return nil, err
	}
	p.PaySign = signature.Signature
	return p, nil
}

func (c *Client) Query(ctx context.Context, orderNo string) (*payments.Transaction, error) {
	if err := ValidateOrderNo(orderNo); err != nil {
		return nil, err
	}
	response, result, err := c.service.QueryOrderByOutTradeNo(ctx, jsapi.QueryOrderByOutTradeNoRequest{Mchid: core.String(c.Config.MchID), OutTradeNo: core.String(orderNo)})
	logRequestID(ctx, "query", result)
	return response, err
}

func (c *Client) Close(ctx context.Context, orderNo string) error {
	if err := ValidateOrderNo(orderNo); err != nil {
		return err
	}
	result, err := c.service.CloseOrder(ctx, jsapi.CloseOrderRequest{Mchid: core.String(c.Config.MchID), OutTradeNo: core.String(orderNo)})
	logRequestID(ctx, "close", result)
	return err
}

type Notification struct {
	ID          string
	Transaction *payments.Transaction
	Payload     []byte
}

func (c *Client) ParseNotification(ctx context.Context, req *http.Request) (*Notification, error) {
	transaction := new(payments.Transaction)
	event, body, err := c.parseNotificationResource(ctx, req, transaction)
	if err != nil {
		return nil, err
	}
	if event.EventType != "TRANSACTION.SUCCESS" || event.Resource.OriginalType != "transaction" {
		return nil, errors.New("unsupported WeChat Pay event")
	}
	return &Notification{ID: event.ID, Transaction: transaction, Payload: body}, nil
}

func (c *Client) parseNotificationResource(ctx context.Context, req *http.Request, target any) (*notify.Request, []byte, error) {
	// Bound the body and validate the AEAD nonce length before the SDK calls
	// cipher.GCM.Open, which otherwise panics on an invalid nonce length.
	body, err := io.ReadAll(io.LimitReader(req.Body, (1<<20)+1))
	if err != nil || len(body) > 1<<20 {
		return nil, nil, errors.New("invalid notification size")
	}
	var shape struct {
		Resource struct {
			Nonce string `json:"nonce"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(body, &shape); err != nil || len(shape.Resource.Nonce) != 12 {
		return nil, nil, errors.New("invalid notification resource")
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	event, err := c.notifier.ParseNotifyRequest(ctx, req, target)
	if err != nil {
		return nil, nil, errors.New("WeChat Pay notification verification or decryption failed")
	}
	if event.ID == "" || event.ResourceType != "encrypt-resource" {
		return nil, nil, errors.New("unsupported WeChat Pay event")
	}
	return event, body, nil
}

func logRequestID(ctx context.Context, operation string, result *core.APIResult) {
	if result != nil && result.Response != nil {
		glog.Infof(ctx, "wechatpay %s HTTP %d request_id=%s", operation, result.Response.StatusCode, result.Response.Header.Get("Request-ID"))
	}
}

func shortDescription(s string) string {
	for len(s) > 127 {
		r := []rune(s)
		s = string(r[:len(r)-1])
	}
	return s
}
