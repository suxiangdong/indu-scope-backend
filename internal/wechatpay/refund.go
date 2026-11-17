package wechatpay

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/refunddomestic"
)

type RefundInput struct {
	OrderNo, TransactionID, RefundNo, Reason, Currency string
	TotalFen, RefundFen                                int64
}

type RefundResult struct {
	MchID, OrderNo, TransactionID, RefundNo, RefundID, Status, Currency string
	TotalFen, RefundFen                                                 int64
	PayerRefundFen                                                      *int64
	ReceivedAccount                                                     string
	SuccessAt                                                           *time.Time
}

func (c *Client) RefundCallbackURL() (string, error) {
	address := c.Config.RefundNotifyURL
	if address == "" {
		const paymentPath = "/api/v1/payment/wechat/notify"
		if !strings.HasSuffix(c.Config.NotifyURL, paymentPath) {
			return "", errors.New("请配置 wechat_pay.refund_notify_url")
		}
		address = strings.TrimSuffix(c.Config.NotifyURL, paymentPath) + "/api/v1/payment/wechat/refund/notify"
	}
	u, err := url.Parse(address)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path == "" || u.Hostname() == "localhost" || strings.Contains(address, "你的域名") {
		return "", errors.New("退款回调地址必须为不带查询参数的公网 HTTPS 地址")
	}
	if ip := net.ParseIP(u.Hostname()); ip != nil && (!ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback()) {
		return "", errors.New("退款回调地址不能是本机或内网地址")
	}
	return address, nil
}

func (c *Client) CreateRefund(ctx context.Context, in RefundInput) (*RefundResult, error) {
	if in.RefundNo == "" || in.OrderNo == "" || in.TransactionID == "" || in.TotalFen <= 0 || in.RefundFen != in.TotalFen || in.Currency != "CNY" || len(in.Reason) > 80 {
		return nil, errors.New("invalid full refund request")
	}
	callback, err := c.RefundCallbackURL()
	if err != nil {
		return nil, err
	}
	service := refunddomestic.RefundsApiService{Client: c.sdk}
	response, result, err := service.Create(ctx, refunddomestic.CreateRequest{
		TransactionId: core.String(in.TransactionID), OutRefundNo: core.String(in.RefundNo),
		Reason: core.String(in.Reason), NotifyUrl: core.String(callback),
		Amount: &refunddomestic.AmountReq{Total: core.Int64(in.TotalFen), Refund: core.Int64(in.RefundFen), Currency: core.String(in.Currency)},
	})
	logRequestID(ctx, "refund-create", result)
	if err != nil {
		return nil, err
	}
	return c.refundResult(response)
}

func (c *Client) QueryRefund(ctx context.Context, refundNo string) (*RefundResult, error) {
	service := refunddomestic.RefundsApiService{Client: c.sdk}
	response, result, err := service.QueryByOutRefundNo(ctx, refunddomestic.QueryByOutRefundNoRequest{OutRefundNo: core.String(refundNo)})
	logRequestID(ctx, "refund-query", result)
	if err != nil {
		return nil, err
	}
	return c.refundResult(response)
}

func (c *Client) refundResult(r *refunddomestic.Refund) (*RefundResult, error) {
	if r == nil || r.Status == nil || r.Amount == nil || r.Amount.Total == nil || r.Amount.Refund == nil || r.Amount.Currency == nil {
		return nil, errors.New("incomplete WeChat refund response")
	}
	return &RefundResult{MchID: c.Config.MchID, OrderNo: refundText(r.OutTradeNo), TransactionID: refundText(r.TransactionId),
		RefundNo: refundText(r.OutRefundNo), RefundID: refundText(r.RefundId), Status: string(*r.Status),
		TotalFen: *r.Amount.Total, RefundFen: *r.Amount.Refund, Currency: *r.Amount.Currency,
		PayerRefundFen: r.Amount.PayerRefund, ReceivedAccount: refundText(r.UserReceivedAccount), SuccessAt: r.SuccessTime}, nil
}

type RefundNotification struct {
	ID, EventType string
	Result        *RefundResult
	Payload       []byte
}

func (c *Client) ParseRefundNotification(ctx context.Context, req *http.Request) (*RefundNotification, error) {
	var resource struct {
		MchID           string     `json:"mchid"`
		OrderNo         string     `json:"out_trade_no"`
		TransactionID   string     `json:"transaction_id"`
		RefundNo        string     `json:"out_refund_no"`
		RefundID        string     `json:"refund_id"`
		Status          string     `json:"refund_status"`
		SuccessTime     *time.Time `json:"success_time"`
		ReceivedAccount string     `json:"user_received_account"`
		Amount          struct {
			Total       *int64 `json:"total"`
			Refund      *int64 `json:"refund"`
			PayerRefund *int64 `json:"payer_refund"`
		} `json:"amount"`
	}
	event, body, err := c.parseNotificationResource(ctx, req, &resource)
	if err != nil {
		return nil, err
	}
	if resource.MchID != c.Config.MchID || event.EventType != "REFUND."+resource.Status ||
		(resource.Status != "SUCCESS" && resource.Status != "CLOSED" && resource.Status != "ABNORMAL") ||
		event.Resource.OriginalType != "refund" || resource.Amount.Total == nil || resource.Amount.Refund == nil {
		return nil, errors.New("invalid WeChat refund notification identity or event")
	}
	return &RefundNotification{ID: event.ID, EventType: event.EventType, Payload: body, Result: &RefundResult{
		MchID: resource.MchID, OrderNo: resource.OrderNo, TransactionID: resource.TransactionID,
		RefundNo: resource.RefundNo, RefundID: resource.RefundID, Status: resource.Status,
		TotalFen: *resource.Amount.Total, RefundFen: *resource.Amount.Refund, Currency: "CNY",
		PayerRefundFen: resource.Amount.PayerRefund, ReceivedAccount: resource.ReceivedAccount, SuccessAt: resource.SuccessTime,
	}}, nil
}

func refundText(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
