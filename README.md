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
```

迁移版本按编号递增。已经执行过的迁移不要修改，后续结构变化请新增一组 `.up.sql` 和 `.down.sql` 文件。当前迁移包含用户认证、公司资料、采购方销售渠道、多级类目、通用产品属性、产品库存和管理后台管理员账号。

类目表使用 `parent_id = 0` 表示一级类目；迁移 `000006_categories_use_zero_parent` 会将历史 `NULL` 值转换为 `0`，并将字段改为非空且默认值为 `0`。迁移 `000007_create_category_attribute_groups` 先保存了类目级分组，迁移 `000008_normalize_attribute_groups` 再将其转换为可复用的全局分组、属性关系和类目分组绑定。

## 认证接口

接口前缀为 `/api/v1/auth`：

```text
POST /api/v1/auth/register  手机号注册
POST /api/v1/auth/login     手机号密码登录
GET  /api/v1/auth/me        获取当前用户
POST /api/v1/auth/logout    用户退出登录
```

注册请求示例：

```json
{
  "phone": "13800138001",
  "password": "Password123!",
  "role": 1,
  "user_type": 1
}
```

`role` 为 `1` 采购方、`2` 供货方；`user_type` 为 `1` 个人用户、`2` 公司用户。注册后用户状态为 `0`（未认证），认证通过后由后台更新为 `1`。密码使用 bcrypt 存储，接口返回的访问令牌为不设置过期时间的 JWT；Redis session 被删除后令牌立即失效。

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

属性 `value_type` 约定为 `1` 文本、`2` 整数、`3` 小数、`4` 布尔、`5` 日期、`6` JSON；`input_type` 决定管理端和产品端使用的输入控件。属性编码全局唯一，属性选项编码在同一属性下唯一。

属性分组由 `attribute_groups` 独立管理，`attribute_group_attributes` 保存分组与属性的多对多关系，`category_attribute_group_bindings` 保存分组与叶子类目的绑定。`category_attributes` 保留为类目级覆盖配置，属性不再通过管理端直接绑定到类目。

属性分组的完整数据模型、接口调用顺序、删除约束和跨电脑开发步骤见
[`docs/attribute-management.md`](docs/attribute-management.md)。
