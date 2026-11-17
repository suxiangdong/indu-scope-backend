# 工业视界后端

基于 GoFrame 2.10 的工业产品 API 服务，为 Taro 4 小程序和未来的 Flutter App 提供统一后端接口。

## 技术栈

- Go 1.23+
- GoFrame v2.10.3
- RESTful API
- OpenAPI/Swagger
- GoFrame gdb（业务数据访问）
- GoFrame ghttp session + Redis（登录态管理）

前端只负责页面展示、交互状态和 API 调用；权限、参数业务规则、数据状态流转和设备业务逻辑全部由后端处理。

业务代码通过 GoFrame `gdb` 使用 ORM 链式查询和模型映射。服务启动时从
`manifest/config/config.yaml` 读取数据库连接池和 Redis 配置，认证逻辑在签发或校验 Token 时通过 `g.Cfg()` 临时读取 JWT 密钥；该文件已加入 Git 忽略规则，真实凭据不会提交到仓库。
数据库迁移仍由 go-migrate 执行，但连接串同样从该配置文件读取。

## 本地运行

```bash
go run .
```

首次运行前，复制 `manifest/config/config.example.yaml` 为
`manifest/config/config.yaml`，再配置 `database.default.link`、`redis.default` 和 `jwt.secret`。
JWT 密钥至少需要 32 个字符；真实配置文件不会提交到 Git。

默认服务地址为 `http://127.0.0.1:8000`。

生产环境可在 `server.swagger_enabled` 设置为 `false`，同时关闭 Swagger UI 和 `/api.json` OpenAPI 文档；开发环境保持为 `true` 即可。

换电脑时请先阅读 [`docs/development-onboarding.md`](docs/development-onboarding.md)，其中包含三个仓库的克隆方式、MySQL/Redis 配置、迁移、Taro 小程序、微信开发者工具和管理后台的完整启动顺序。

## 项目结构

```text
api/                       API 请求和响应定义（由 gf gen ctrl 维护接口文件）
internal/controller/       HTTP 控制器（由 gf gen ctrl 生成）
internal/logic/            业务逻辑实现
internal/service/          服务接口（由 gf gen service 生成）
internal/dao/              数据访问对象（由 gf gen dao 生成）
internal/model/do/         DAO 数据对象（由 gf gen dao 生成）
internal/model/entity/     数据实体（由 gf gen dao 生成）
internal/table/            表字段定义（由 gf gen dao -gt 生成）
manifest/config/           服务配置
internal/migration/        go-migrate 执行包和内嵌迁移脚本
internal/cmd/              GoFrame gcmd 服务和数据库迁移命令
docs/postman/              Postman API 集合
```

## 文档导航

- [`docs/development-onboarding.md`](docs/development-onboarding.md)：换电脑后的完整环境恢复、启动顺序、微信开发者工具和常见问题
- [`docs/attribute-management.md`](docs/attribute-management.md)：属性、属性分组、类目绑定的数据模型和管理接口
- `docs/postman/`：前台和管理后台 API 调试集合
- `docs/sql/dev_admin_user.sql`：仅用于开发环境的管理员初始化 SQL

接口参数绑定和基础校验使用 GoFrame 请求对象与校验标签；跨字段业务规则、权限和状态流转放在 `internal/logic` 处理。

前台请求通过 `internal/middleware.Ctx`，管理后台请求通过
`internal/middleware.AdminCtx`，调用生成的 `service.Bizctx()` 注入业务上下文。
实现位于 `internal/logic/bizctx`，上下文结构位于 `internal/model/bizctx`，其中包含
GoFrame 的 `ghttp.Session` 和当前用户。`internal/logic/session` 只负责通过
`SetUser`、`GetUser`、`RemoveUser` 管理当前请求的用户 session；JWT 本身不设置过期时间，
JWT 字符串作为 Redis session id，但每次受保护请求都必须能在 Redis session 中找到对应用户。
新增前台业务路由时，在 `middleware.Ctx` 后绑定 `middleware.RequireUser`；新增管理后台路由时，
在 `middleware.AdminCtx` 后绑定 `middleware.RequireAdmin`，即可在进入控制器前拒绝没有有效 session 的请求。

## GoFrame 代码生成

项目使用与 GoFrame v2.10.3 对齐的 `gf` CLI。`gf gen dao` 需要通过 `-l` 接收数据库连接串，
当前 `make dao` 目标仍使用临时环境变量把连接串传给代码生成器；服务运行和迁移不会读取该环境变量：

```bash
export INDU_SCOPE_DATABASE_DSN='mysql:用户名:密码@tcp(主机:3306)/数据库?charset=utf8mb4&parseTime=true&loc=Local'
make dao       # 生成 dao、do、entity、table；自动排除 schema_migrations
make service   # 根据 internal/logic 生成 internal/service 接口
make ctrl      # 根据 api 定义生成 controller 和 API 接口文件
```

`dao`、`do`、`entity`、`table` 和生成的服务接口文件由 GoFrame CLI 维护，数据库结构变化后重新执行 `make dao`。业务实现放在 `internal/logic`，控制器接口实现放在 `internal/controller` 的生成文件中。

## 数据库迁移

迁移使用 `golang-migrate/migrate`，迁移文件内嵌到后端二进制中，不依赖运行目录。连接串从
`manifest/config/config.yaml` 的 `database.default.link` 读取：

```bash
go run . migrate up
```

支持的命令：

```bash
go run . migrate up    # 执行所有待执行迁移
go run . migrate down  # 回滚最近一次迁移
go run . migrate force --version=28  # 仅在确认 dirty 迁移未完成后修复版本元数据
```

如果迁移执行失败并提示 `dirty`，先确认失败迁移的 DDL 是否已经生效，再将版本强制回退到失败迁移的前一个版本，然后重新执行 `migrate up`。`force` 只修改迁移元数据，不会执行或回滚业务 SQL。

迁移版本按编号递增。已经执行过的迁移不要修改，后续结构变化请新增一组 `.up.sql` 和 `.down.sql` 文件。当前迁移包含用户认证、公司资料、采购方销售渠道、多级类目、通用产品属性、产品库存和管理后台管理员账号。迁移 `000018_allow_pending_user_role` 允许新用户暂时使用 `role=0`，完成资料后再选择业务角色。

类目表使用 `parent_id = 0` 表示一级类目；迁移 `000006_categories_use_zero_parent` 会将历史 `NULL` 值转换为 `0`，并将字段改为非空且默认值为 `0`。迁移 `000007_create_category_attribute_groups` 先保存了类目级分组，迁移 `000008_normalize_attribute_groups` 再将其转换为可复用的全局分组、属性关系和类目分组绑定；迁移 `000009_add_ids_to_all_tables` 为此前使用用户主键或联合主键的业务表补充统一的自增 `id` 主键，并保留原有业务唯一约束；迁移 `000010_add_attribute_group_display_name` 为分组增加面向前端展示的显示名称。

迁移 `000045_serialize_category_tree_writes` 新增 `backend_write_locks` 并预置分类树锁记录。分类新增、移动和删除在同一事务内取得该记录的行锁后，再检查父节点、祖先链和同级名称，防止多实例并发操作产生循环或孤立节点。部署包含此修复的版本前，先执行 `go run . migrate up`。

## 认证接口

接口前缀为 `/api/v1/auth`：

```text
POST /api/v1/auth/register  手机号注册
POST /api/v1/auth/login     手机号密码登录
POST /api/v1/auth/wechat/phone-login   微信手机号组件登录
GET  /api/v1/auth/verification        获取当前认证状态
GET  /api/v1/auth/verification/store-platforms  获取启用的店铺平台
POST /api/v1/auth/verification/business-license  上传营业执照到 OSS
POST /api/v1/auth/verification/business-license/ocr  识别已上传的营业执照
POST /api/v1/auth/verification/id-card  上传身份证正面或反面到 OSS（`side=front|back`）
POST /api/v1/auth/verification/personal  提交个人认证
POST /api/v1/auth/verification/company   提交企业认证
GET  /api/v1/auth/me        获取当前用户
POST /api/v1/auth/logout    用户退出登录
```

用户点击登录并完成手机号授权后，前端调用 `wx.login` 获取临时 `login_code`，同时取得
`getPhoneNumber` 组件返回的一次性 `phone_code`，将两个 code 一起提交到
`wechat/phone-login`。后端在同一个方法内通过 `jscode2session` 获取 OpenID、调用微信手机号
接口验证手机号，自动创建或绑定用户，再返回正式的 Bearer access token。微信应用级
`access_token` 由后端启动时预热并在进程内缓存，每 30 分钟刷新一次，不会为每次手机号授权
重复获取。

认证接口要求携带用户 `Authorization: Bearer <access_token>`。新用户微信登录后 `role=0`，需要在“完善资料”页面选择一次采购方或供货方；已有 `role=1/2` 的用户不显示角色选择，也不能通过认证接口修改角色。供货方只能提交企业资料，不能提交个人认证或身份证；采购方可以选择个人或企业资料。个人认证先选择店铺平台、填写店铺地址，分别上传身份证正面和反面，
再提交姓名、身份证号及两个文件的 Endpoint、bucket、对象路径；身份证号按业务要求以明文保存，接口仍只返回脱敏后的号码；企业认证提交企业名称、统一社会信用代码、注册地址、企业介绍、
办公地址、工厂地址和至少一位联系人，分别保存到 `company_verifications`、
`company_profiles` 和 `company_contacts`。认证状态返回 `unsubmitted`、`pending`、
`rejected` 或 `verified`；被驳回的资料可以修改后重新提交，审核中的资料不允许修改。

营业执照通过 `POST /api/v1/auth/verification/business-license` 上传。后端使用
OSS 私有桶保存对象，并在 `user_files` 中保存 Endpoint、bucket、对象键、文件名、类型、大小和 ETag，
企业认证表保存 `business_license_endpoint`、`business_license_bucket` 和
`business_license_path`，并校验这三个值属于当前用户上传的对象。个人认证同样使用
`id_front_endpoint`/`id_front_bucket`/`id_front_path` 和
`id_back_endpoint`/`id_back_bucket`/`id_back_path` 保存身份证文件位置。三者合起来
构成完整的 OSS 对象定位信息，例如 `oss-cn-beijing.aliyuncs.com`、`env-develop`、
`verification/business_license/...`。OSS 配置建议通过
`OSS_ENDPOINT`、`OSS_BUCKET`、`OSS_ACCESS_KEY_ID`、`OSS_ACCESS_KEY_SECRET` 和
`OSS_KEY_PREFIX` 环境变量注入；不要将 AccessKey 提交到代码仓库。
迁移 `000013_add_verification_object_paths` 新增这些对象位置字段；迁移
`000016_remove_verification_file_ids` 已移除认证表中的旧 `*_file_id` 字段，文件元数据统一由
`user_files` 管理。
迁移 `000014_add_verification_object_endpoints` 为 `user_files` 及认证表补充 Endpoint，
并为历史对象回填当前北京地域的 Endpoint。由于 OSS 使用私有 ACL，拼接出的地址仅用于
定位对象；实际访问仍应由后端生成签名 URL。
迁移 `000017_add_personal_store_and_plain_id` 新增 `store_platforms` 店铺平台表，以及个人认证的
`store_platform_id`、`store_url` 和明文 `id_number` 字段。店铺平台由管理后台的“店铺平台”页面维护，
禁用的平台不会出现在小程序认证表单中；已被认证资料使用的平台不能删除。
当前 OSS 地域为华北 2（北京），Endpoint 为 `https://oss-cn-beijing.aliyuncs.com`；开发环境
使用 `env-develop`，生产环境使用 `env-production`。部署生产环境时设置
`OSS_BUCKET=env-production` 覆盖开发配置。

营业执照上传成功后，调用 `POST /api/v1/auth/verification/business-license/ocr` 可使用阿里云
`RecognizeBusinessLicense` 识别企业名称、统一社会信用代码和营业地址。接口复用已有的
`OSS_ACCESS_KEY_ID`/`OSS_ACCESS_KEY_SECRET`（或 `oss.access_key_id`/`oss.access_key_secret`）配置，
后端会为私有 OSS 对象生成临时签名 URL；小程序不会接触 AccessKey。OCR 默认使用
`ocr-api.cn-hangzhou.aliyuncs.com`，如需切换地域可设置 `OCR_ENDPOINT` 或 `ocr.endpoint`。

管理后台认证审核接口前缀为 `/api/v1/admin/verifications`，需要管理员 Bearer Token：

```text
GET  /api/v1/admin/verifications         查询个人/企业认证列表，可按 type、review_status 筛选
GET  /api/v1/admin/verifications/detail  查询认证详情
POST /api/v1/admin/verifications/review  通过或驳回认证，驳回时必须填写 reason
```

迁移 `000015_add_verification_review_fields` 为个人和企业认证增加审核状态、驳回原因、
审核人和审核时间，并将历史已有 `verified_at` 的记录标记为已通过。

店铺平台管理接口前缀为 `/api/v1/admin/store-platforms`，管理后台可新增、修改、启用/禁用和删除平台：

```text
GET  /api/v1/admin/store-platforms         查询平台
POST /api/v1/admin/store-platforms/create  新增平台
POST /api/v1/admin/store-platforms/update  修改平台
POST /api/v1/admin/store-platforms/delete  删除平台（已使用的平台不可删除）
```

前端用户管理接口前缀为 `/api/v1/admin/customer-users`，用于查询注册用户和调整账号状态：

```text
GET  /api/v1/admin/customer-users         查询前端用户，可按手机号、角色、用户类型和状态筛选
POST /api/v1/admin/customer-users/status  修改用户状态（0 未认证、1 已认证、2 已禁用）
```

微信配置位于 `manifest/config/config.yaml`：

```yaml
wechat:
  app_id: "微信小程序 AppID"
  app_secret: "微信小程序 AppSecret"
  api_base_url: "https://api.weixin.qq.com"
```

注册请求示例：

```json
{
  "phone": "13800138001",
  "password": "Password123!",
  "user_type": 1
}
```

`role` 为 `0` 待完善、`1` 采购方、`2` 供货方；`user_type` 为 `1` 个人用户、`2` 公司用户。注册后用户状态为 `0`（未认证），认证通过后由后台更新为 `1`。密码使用 bcrypt 存储，接口返回的访问令牌为不设置过期时间的 JWT；Redis session 被删除后令牌立即失效。

Postman 文件位于 `docs/postman/indu-scope-auth.postman_collection.json`，导入后设置 `base_url`、`phone` 和 `password` 即可调用。

## 管理后台认证

管理后台使用独立的 `admin_users` 表，登录接口为：

```text
POST /api/v1/admin/auth/login  管理员账号或手机号登录
GET  /api/v1/admin/me          获取当前管理员
POST /api/v1/admin/auth/logout 管理员退出登录
```

开发环境的超级管理员初始化 SQL 位于 `docs/sql/dev_admin_user.sql`，执行后可使用账号 `cffyadmin` 和临时密码 `123456` 登录。Postman 文件位于 `docs/postman/indu-scope-admin-auth.postman_collection.json`。

管理员登录后可通过以下接口维护产品类目：

```text
GET  /api/v1/admin/categories
POST /api/v1/admin/categories/create
POST /api/v1/admin/categories/update
POST /api/v1/admin/categories/delete
```

属性管理接口：

```text
GET  /api/v1/admin/attributes
POST /api/v1/admin/attributes/create
POST /api/v1/admin/attributes/update
POST /api/v1/admin/attributes/delete
GET  /api/v1/admin/attributes/options?attribute_id=...
POST /api/v1/admin/attributes/options/create
POST /api/v1/admin/attributes/options/update
POST /api/v1/admin/attributes/options/delete
GET  /api/v1/admin/attribute-groups
POST /api/v1/admin/attribute-groups/create
POST /api/v1/admin/attribute-groups/update
POST /api/v1/admin/attribute-groups/delete
GET  /api/v1/admin/attribute-groups/attributes?group_id=...
POST /api/v1/admin/attribute-groups/attributes/save
POST /api/v1/admin/attribute-groups/attributes/delete
GET  /api/v1/admin/category-groups?category_id=...
POST /api/v1/admin/category-groups/bind
POST /api/v1/admin/category-groups/unbind
GET  /api/v1/admin/category-attributes?category_id=...&attribute_id=...
POST /api/v1/admin/category-attributes/save    # 类目级覆盖配置
POST /api/v1/admin/category-attributes/delete  # 删除类目级覆盖
```

属性 `value_type` 约定为 `1` 文本、`2` 整数、`3` 小数、`4` 布尔、`5` 日期、`6` JSON；`input_type` 决定管理端和产品端使用的输入控件。属性、属性选项和属性分组的编码由系统自动生成，创建后不可由管理端修改；属性编码全局唯一，属性选项编码在同一属性下唯一。

属性分组由 `attribute_groups` 独立管理，`attribute_group_attributes` 保存分组与属性的多对多关系，`category_attribute_group_bindings` 保存分组与叶子类目的绑定。`category_attributes` 保留为类目级覆盖配置，属性不再通过管理端直接绑定到类目。

属性分组的完整数据模型、接口调用顺序、删除约束和跨电脑开发步骤见
[`docs/attribute-management.md`](docs/attribute-management.md)。
