# VIP 会员与支付设计

## 已实现范围

- `vip_plans`：后台维护套餐名称、周期、有效天数、现价、原价、套餐类型（普通套餐或自动续费套餐）和启停状态；价格允许为 0，用于免费期；套餐编码由后端自动生成且不可编辑，有效天数在 1–36500 天之间，与可结算订单的范围一致；历史异常天数的订单不能继续发起支付。
- `vip_subscriptions`：记录用户会员状态、有效期、自动续费和微信周期扣款签约号，并固化已购买权益的套餐类型、币种；后续编辑套餐不改变当前会员类型。
- `payment_orders`：记录会员订单金额、支付状态、商户订单号、微信交易号、预支付信息及本单会员权益起止区间。
- `payment_refunds`：保存整单退款申请、退款金额快照、操作管理员、退款原因、微信退款单号、退款状态及权益回收结果。
- `payment_events`：保存经公钥验签、APIv3 密钥解密后的支付事件处理记录；回调保存加密原始报文，主动查单记录查单结果。处理失败保留 `failed` 记录供排查，重试不会覆盖已完成事件。
- `vip_agreements`：维护会员服务协议和自动续费协议的标题、正文、发布状态与版本号；每次编辑新增一条版本记录，当前接口只返回每类协议最新版本。
- `payment_orders`：记录下单时的套餐快照（编码、名称、说明、类型、周期、有效天数、自动续费开关、价格和币种），并记录会员服务协议版本；自动续费套餐同时记录自动续费协议版本，保证订单可追溯。当前订单只有一个套餐，因此扩展订单表比新增订单明细表更合适。
- 微信支付通过官方 `wechatpay-go v0.2.21` SDK 接入公钥模式：商户私钥签名、公钥及公钥 ID 验证应答/回调、APIv3 密钥解密回调；不依赖平台证书或平台证书下载器。
- 小程序支持 `requestPayment`、继续原订单、关闭未支付订单、查询支付状态；前端只在后端确认 `paid` 后提示会员生效，网络失败重试使用原幂等键。
- 普通付费套餐支付成功后开通会员，续费从当前到期时间顺延；按订单有效天数快照入账，修改套餐不会改变已下单订单。
- `payment_orders` 固化付款 AppID、商户号和 OpenID。验签成功后仍逐一核对订单号、金额、币种、付款人和交易类型；数据库为微信交易号建立唯一索引。
- 入账事务按用户、订单顺序加锁，订单状态防止回调与主动查单重复延长权益；创建免费会员使用同一用户锁。预支付和关单也按同一顺序加锁，避免关单后才创建微信订单。
- 管理后台新增“会员套餐”和“会员协议”页面，可配置价格、周期、协议正文和发布状态。
- 需求发布、需求附件上传和供应商响应由 `RequireActiveVip` 路由中间件保护；无有效会员返回 HTTP 403，需求浏览和已有需求处理不受影响。
- 0 元套餐创建订单后直接写入 `paid` 订单和 `active` 会员订阅，不调用微信支付。
- 已有未到期的普通 VIP 可从同币种的全部已上架普通付费套餐中选择续费周期，卡片、价格、原价、协议及底部按钮复用首次开通 UI；不允许免费套餐续费、连续包月续费或长期会员重复开通。会员没有等级设置，所有套餐提供相同会员资格，仅价格、周期和套餐类型不同。支付成功后保留剩余有效期，按订单有效天数快照顺延，并将当前订阅关联到本次选择的套餐。下单与每次生成支付参数都会检查购买资格；支付通知与主动查单只核验真实支付身份、金额及订单快照，不会因当前会员状态或目录变化拒绝已成功扣款的订单。

## 接口

用户端：

```text
GET  /api/v1/vip/plans
GET  /api/v1/vip/agreements
GET  /api/v1/vip/me
POST /api/v1/vip/orders
GET  /api/v1/vip/orders/{id}
POST /api/v1/vip/orders/{id}/pay
POST /api/v1/vip/orders/{id}/close
POST /api/v1/vip/subscription/cancel-auto-renew
```

支付通知（不使用登录校验，也不包装通用 API 响应）：

```text
POST /api/v1/payment/wechat/notify
POST /api/v1/payment/wechat/refund/notify
```

上述查单、续付、关单接口只允许当前登录用户操作自己的订单。回调必须公钥验签通过并成功提交入账事务后才返回 HTTP 204；其他情况返回失败，等待微信重试。 支付通知由主后端应用直接接收，`notify_url` 需指向主后端公开的 HTTPS 回调地址，反向代理转发至主后端的该路由。

管理端：

```text
GET  /api/v1/admin/vip/plans
GET  /api/v1/admin/vip/agreements
POST /api/v1/admin/vip/agreements/update
POST /api/v1/admin/vip/plans/create
POST /api/v1/admin/vip/plans/update
POST /api/v1/admin/vip/plans/status
GET  /api/v1/admin/vip/orders
POST /api/v1/admin/vip/orders/{id}/refund
GET  /api/v1/admin/vip/refunds
GET  /api/v1/admin/vip/refunds/{id}
POST /api/v1/admin/vip/refunds/{id}/retry
```

## 订单状态

```text
pending -> paying -> paid
pending -> closed
paying  -> closed
paying  -> failed
paid    -> refunded
```

会员状态：`pending`、`active`、`paused`、`canceled`、`expired`。

## 微信支付接入边界

当前接入范围为普通小程序支付和普通套餐续费。连续包月需要独立的委托代扣权限、模板及用户签约，不使用普通支付冒充签约。未实现委托代扣签约、实际周期扣款、部分退款、账单对账或后台定时扫单；查询订单可主动补偿未收到回调的支付成功结果。暂未启用支付时返回 `payment_ready=false`，0 元套餐继续直接开通。

真实支付参数仅写入被 Git 忽略的 `manifest/config/config.yaml` 或后端环境变量，参考 `config.example.yaml`：

```yaml
wechat_pay:
  enabled: false
  mch_id: "1005166565"
  merchant_serial_no: "商户API证书序列号"
  merchant_private_key_path: "/run/secrets/apiclient_key.pem"
  api_v3_key: "本地填写32位APIv3密钥"
  public_key_id: "PUB_KEY_ID_商户平台显示的真实ID"
  public_key_path: "/run/secrets/pub_key.pem"
  notify_url: "https://api.example.com/api/v1/payment/wechat/notify"
```

每个参数可通过 `WECHAT_PAY_` 加大写字段名覆盖，例如 `WECHAT_PAY_API_V3_KEY`、`WECHAT_PAY_PUBLIC_KEY_ID`。AppID 默认取 `wechat.app_id`；与微信登录、openid 获取时的 AppID 必须一致。

启用步骤：

1. 在商户平台确认 JSAPI 支付权限及小程序 AppID 绑定。
2. 超级管理员或安全联系人进入「账户中心 → API安全 → 微信支付公钥 → 管理公钥」，下载 `pub_key.pem` 并复制其 `PUB_KEY_ID_...`。它是微信支付公钥，不是商户公钥、商户API证书或平台证书。
3. 将商户私钥和微信支付公钥部署到后端可读、非公开目录。填写 APIv3 密钥和上述配置；无需运行平台证书下载工具。
4. 先部署公钥模式代码，再按商户平台当前提示确认公钥模式状态/开启公钥切换；如商户原来使用证书，需要按官方迁移流程处理灰度。
5. 将真实公网 HTTPS 回调域名配置到 `notify_url`，保留原始 HTTP 报文和 `Wechatpay-*` 头；回调不需要用户 token，不可使用局域网地址。
6. 执行新增迁移 `000042`（付款身份快照和微信交易号唯一索引），然后在配置齐全后开启 `enabled`。真实小额支付应由用户在真机确认，本文开发测试不发起真实扣款。

数据库迁移：

```bash
go run . migrate up
```

迁移 `000044` 为订阅增加 `plan_type_snapshot`、`currency_snapshot`，优先用关联的最近已支付订单回填，缺少历史订单时根据已有签约状态或当前目录冻结。历史等级列保留供兼容和审计，API、后台、小程序与结算均不再使用。应先执行迁移再部署依赖新字段的代码。

下单将用户行锁、幂等键查询、资格检查及订单写入放在同一事务；0 元套餐的会员激活也在该事务内。并发请求等待用户锁后，先返回同一幂等键的原订单，再读取当前套餐/协议或判断免费资格，避免重复请求报唯一键冲突或“已有会员”。用户锁在该事务首次一致性读之前取得；用户锁保护的幂等键、订阅、退款申请范围及权益区间查询使用普通一致性读，避免锁住不存在的子记录范围而导致不同用户互相等待 InnoDB 索引间隙锁。

会员查询使用一次查询，将无记录（包括到期/退款边界）视为正常无会员结果；多条有效权益优先返回长期会员，其次返回到期时间最晚的会员。已验证的成功扣款优先顺延普通、有限期、同币种订阅；若当前只有长期或委托代扣等权益，则为本次订单创建独立普通有限期订阅，保留原有权益和签约信息，防止迟到通知导致已扣款未入账。

验证：`go test ./...` 包含本地生成的测试 RSA 公私钥/AES-GCM 签名报文测试（不使用商户真实密钥）和隔离 SQLMock 的入账事务测试，覆盖错误公钥 ID、过期/篡改/探测签名、无效密文及 nonce、金额与付款人不匹配、重复通知和续费期限，以及目录变更后的支付结算、特殊会员的迟到支付独立入账、免费/付费下单幂等恢复、免费订单写入失败回滚、到期/退款边界无会员查询和长期权益优先级。

官方参考：[签名与验签总述](https://pay.wechatpay.cn/doc/v3/merchant/4012365342)、[公钥验签](https://pay.wechatpay.cn/doc/v3/merchant/4013053249)、[小程序下单](https://pay.wechatpay.cn/doc/v3/merchant/4012791897)、[微信支付 Go SDK](https://github.com/wechatpay-apiv3/wechatpay-go)。


## 整单退款

后台「业务运营 → 会员订单」展示套餐快照、协议版本和付款信息；已支付的普通微信付费订单可填写原因并确认整单原路退款。「退款管理」提供订单号/退款单号/状态筛选、详情、主动查询微信结果以及未确认申请的原单号重试。0 元订单不发起资金退款，超过支付成功后 365 天的订单不允许新申请。

微信申请接口返回成功只表示受理，不能直接标记退款成功。当前状态：`created`（结果待确认）、`processing`（微信已受理）、`success`、`closed`、`abnormal`。同一订单的未确认、处理中、成功或异常退款会阻止创建另一笔退款；网络失败继续使用原退款单号、金额和原因。确认关闭的退款可以在原订单重新发起新申请。异常退款需按微信要求在商户平台处理，再在后台刷新结果。

仅通过公钥验签和金额、订单号、交易号、商户号、退款单号核对的退款成功结果，才会在同一数据库事务中更新退款记录、订单状态及会员权益。回调和主动查询共用处理逻辑，成功状态不会被较早的受理响应覆盖，重复通知不会再次扣减权益。`payment_events` 同时记录退款回调及申请/查询处理结果。

会员权益规则：仅回收退款订单尚未使用的时长；后续订单的权益区间前移，其购买时长完整保留。已经使用完的订单退款不影响后续购买权益。仅当剩余会员时长耗尽才将该订阅设为过期。新的付款和 0 元开通均记录独立权益起止区间，订单金额和天数仍取下单快照。

迁移 `000043` 新增退款表和订单权益区间字段。对历史数据，仅当所有已支付订单的天数可以准确解释订阅起止时间时，才按原支付顺序回填；无法还原、被人工修改或存在区间冲突的订单，退款成功后标记 `entitlement_status=manual`（权益待人工核对），不自动修改其他会员权益。后台确认弹窗和退款详情会提示这类情况。`reclaimed` 表示已回收未使用时长，`none` 表示本单已无剩余时长。

部署前执行：

```bash
go run . migrate up
```

退款通知由主后端接收，新增公开路由 `/api/v1/payment/wechat/refund/notify`。默认使用支付通知 URL 的同域名和同路径前缀，将 `/api/v1/payment/wechat/notify` 替换为退款路径；非标准路径需要显式设置：

```bash
export WECHAT_PAY_REFUND_NOTIFY_URL='https://你的域名/api/v1/payment/wechat/refund/notify'
```

也可在 `wechat_pay.refund_notify_url` 中配置。申请退款时会把该地址传给微信，因此无需新增密钥或平台证书。回调经过 Nginx 时需要保留原始报文和 `Wechatpay-*` 头，转发至主后端，不能套用户登录校验。订单及退款记录查看、退款状态查询允许已登录管理员访问；发起退款和重试退款单独注册到 `RequireSuperAdmin` 路由组，普通管理员调用返回 HTTP 403。资金操作的服务层再次校验服务端上下文中的超级管理员身份及操作人 ID；后台根据 `is_super_admin` 隐藏普通管理员的退款按钮和确认弹窗。申请人的管理员 ID 从服务端登录上下文读取，客户端不能自行指定权限。

验证使用临时 RSA 密钥、签名通知、模拟微信 HTTP 应答及隔离 SQLMock；不发起真实退款。覆盖受理不回收权益、重复回调、回调与查询乱序、网络超时原号重试、金额/身份不匹配、免费/过期订单拦截及连续购买后的权益回收。

官方文档：[退款申请](https://pay.wechatpay.cn/doc/v3/merchant/4012791903)、[查询单笔退款](https://pay.wechatpay.cn/doc/v3/merchant/4012791904)、[退款结果通知](https://pay.wechatpay.cn/doc/v3/merchant/4012791906)。
