package payment

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
	"github.com/suxiangdong/indu-scope-backend/internal/wechatpay"
)

// Notify is public: authentication comes from WeChat Pay's signature, not a
// customer session. Keep its response outside the normal API envelope.
func Notify(r *ghttp.Request) {
	handleNotification(r, service.Vip().HandleWechatPaymentNotification)
}

func RefundNotify(r *ghttp.Request) {
	handleNotification(r, service.Vip().HandleWechatRefundNotification)
}

func handleNotification(r *ghttp.Request, handler func(context.Context, *http.Request) error) {
	ctx, cancel := context.WithTimeout(r.Context(), 4*time.Second)
	defer cancel()
	if err := handler(ctx, r.Request); err != nil {
		g.Log().Errorf(ctx, "wechatpay notification rejected: %v", err)
		r.Response.Status = http.StatusBadRequest
		if errors.Is(err, wechatpay.ErrDisabled) {
			r.Response.Status = http.StatusServiceUnavailable
		}
		r.Response.WriteJson(g.Map{"code": "FAIL", "message": "通知暂未处理，请重试"})
		return
	}
	r.Response.WriteStatus(http.StatusNoContent)
}
