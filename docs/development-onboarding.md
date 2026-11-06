# 源仓跨电脑开发指南

本文用于在新电脑上恢复源仓开发环境。项目由三个独立 Git 仓库组成：

```text
indu/
├── indu-scope-backend/             # GoFrame 后端 API
├── indu-scope-frontend/            # 面向用户的前端仓库
│   ├── miniapp/                    # Taro 4 微信小程序（当前项目）
│   └── web/                        # 未来 Web 端预留目录
└── indu-scope-manager/             # React + Umi 管理后台
```

管理后台不是 `indu-scope-frontend` 的子目录，而是同级独立仓库。

## 1. 获取代码

在新电脑上先安装 Git，然后选择一个统一的工作目录：

```bash
mkdir -p ~/code/indu
cd ~/code/indu

git clone git@github.com:suxiangdong/indu-scope-backend.git
git clone git@github.com:suxiangdong/indu-scope-frontend.git
git clone git@github.com:suxiangdong/indu-scope-manager.git
```

如果新电脑没有配置 GitHub SSH key，可将上面的地址替换成 HTTPS 地址。

## 2. 必需软件和服务

| 软件/服务 | 建议版本 | 用途 |
| --- | --- | --- |
| Git | 最新稳定版 | 拉取和提交代码 |
| Go | 1.23+ | 运行后端和数据库迁移 |
| Node.js | 22+ | 构建小程序和管理后台 |
| npm | 10+ | 安装 JavaScript 依赖；本项目不要求 pnpm |
| MySQL | 8.x 或兼容版本 | 业务数据库 |
| Redis | 6+ | 登录 session 和缓存 |
| 微信开发者工具 | 最新稳定版 | 小程序编译、预览和真机调试 |

先确认基础版本：

```bash
git --version
go version
node --version
npm --version
```

MySQL 和 Redis 可以运行在本机，也可以使用局域网/云服务器。新电脑必须能访问配置中的地址和端口。

## 3. 后端配置和启动

进入后端目录，安装 Go 依赖并创建本机配置：

```bash
cd ~/code/indu/indu-scope-backend
go mod download
cp manifest/config/config.example.yaml manifest/config/config.yaml
```

只有需要执行 `make dao`、`make service` 或 `make ctrl` 时才需要安装 GoFrame CLI；日常启动和迁移不依赖它：

```bash
go install github.com/gogf/gf/v2/cmd/gf@v2.10.3
```

编辑 `manifest/config/config.yaml`，至少填写以下内容：

```yaml
server:
  address: ":8000"

database:
  default:
    type: "mysql"
    link: "mysql:用户名:密码@tcp(主机:3306)/数据库?charset=utf8mb4&parseTime=true&loc=Local"

jwt:
  secret: "至少32个字符的随机字符串"

wechat:
  app_id: "微信小程序 AppID"
  app_secret: "微信小程序 AppSecret"
  api_base_url: "https://api.weixin.qq.com"

redis:
  default:
    address: "redis主机:6379"
    user: ""
    pass: ""
    db: 0
```

OSS 需要上传营业执照和身份证时，再配置 `oss` 节点或对应环境变量：

```yaml
oss:
  endpoint: "https://oss-cn-beijing.aliyuncs.com"
  bucket: "env-develop"
  access_key_id: ""
  access_key_secret: ""
  key_prefix: "indu/verification"
```

AccessKey 只放在新电脑的本地配置或部署环境变量中，不要写入 Git。开发桶是 `env-develop`，生产桶是 `env-production`，地域为华北 2（北京）。

执行数据库迁移：

```bash
go run . migrate up
```

迁移命令从 `database.default.link` 读取连接串，当前最新迁移包括用户角色待完善状态、实名认证、OSS 对象位置、店铺平台和审核字段。已经执行过的迁移不要修改，需要新增结构时新增一组 `.up.sql`/`.down.sql` 文件。

启动后端：

```bash
go run .
```

默认地址：`http://127.0.0.1:8000`。

首次初始化开发管理员（数据库迁移完成后执行一次）：

```bash
mysql -h 主机 -P 3306 -u 用户名 -p 数据库 < docs/sql/dev_admin_user.sql
```

开发管理员账号是 `cffyadmin`，临时密码是 `123456`，首次登录后应立即修改。生产环境不要使用该 SQL。

后端检查：

```bash
GOCACHE=/tmp/indu-scope-go-cache go test ./...
```

如果数据库结构变化后需要重新生成 GoFrame 文件：

```bash
export INDU_SCOPE_DATABASE_DSN='mysql:用户名:密码@tcp(主机:3306)/数据库?charset=utf8mb4&parseTime=true&loc=Local'
make dao
make service
make ctrl
```

## 4. 小程序前端（Taro 4 + Vite）

小程序目录：`indu-scope-frontend/miniapp`。当前 AppID 为 `wx550ed33a6840ccec`，配置已经写入 `project.config.json` 和各环境文件。

```bash
cd ~/code/indu/indu-scope-frontend/miniapp
npm install
npm run typecheck
```

本地 H5 开发：

```bash
npm run dev:h5
```

浏览器打开 `http://localhost:10086/`。H5 的 API 地址来自 `.env.development`：

```dotenv
TARO_APP_API_BASE_URL=http://127.0.0.1:8000
```

微信小程序开发构建：

```bash
npm run dev:weapp
```

构建产物在 `miniapp/dist/`。微信开发者工具导入 `miniapp` 项目目录，项目配置会自动将小程序根目录指向 `dist/`；也可以直接导入 `miniapp/dist`。生产构建：

```bash
npm run build:weapp:prod
npm run build:h5
```

### 微信开发者工具和真机预览的区别

- 在本机微信开发者工具中，打开“详情 → 本地设置”，可勾选“不校验合法域名、web-view（业务域名）、TLS 版本以及 HTTPS 证书”，这样本机 `http://127.0.0.1:8000` 可以调试。
- 手机预览/真机调试不能访问电脑的 `127.0.0.1`。必须把后端部署到手机可访问的 HTTPS 域名，并在微信公众平台配置 request 合法域名，然后把 `.env.production` 中的 `TARO_APP_API_BASE_URL` 改为该域名重新构建。
- `request:fail url not in domain list` 是微信域名校验问题；`ERR_CONNECTION_REFUSED` 是后端未启动、地址错误或手机无法访问该地址，两者不是登录回调问题。
- 小程序手机号登录在用户点击授权按钮时才调用 `wx.login`，后端在同一个接口中完成 code 换 OpenID、手机号校验和登录；启动小程序不会静默调用 `wx.login`。

### 当前业务流程

新用户登录后 `role=0`，需要进入“我的 → 去认证/完善资料”选择一次采购方或供货方。供货方只能提交企业认证；采购方可以提交个人或企业认证。认证审核中的资料不能再次修改，审核被驳回后才可重新提交。

## 5. 管理后台

```bash
cd ~/code/indu/indu-scope-manager
npm install
npm run tsc
PORT=8001 npm run dev
```

管理后台默认访问 `http://localhost:8001`，开发代理会把 `/api/` 请求转发到 `http://127.0.0.1:8000`。因此启动顺序应是先启动后端，再启动管理后台。

常用命令：

```bash
PORT=8001 npm run dev # 开发环境，无 mock，固定使用 8001，避免和后端 8000 冲突
npm run tsc       # TypeScript 检查
npm run build     # 生产构建
npm run preview   # 预览构建结果
```

当前仓库的 `preview` 脚本使用 8000 端口，和后端默认端口相同；联调时优先使用开发命令，做预览时先停止后端或调整预览端口。

管理后台主要页面：

- `/user/login`：管理员登录
- `/operations/users`：前台用户管理
- `/operations/verifications`：个人/企业认证审核
- `/system/store-platforms`：店铺平台维护
- `/catalog/categories`：类目管理
- `/catalog/attributes`：属性和属性选项管理
- `/catalog/attribute-groups`：属性分组及类目绑定

如未配置独立的 `API_BASE_URL`，开发环境使用 Umi 代理；配置模板见 `indu-scope-manager/.env.example`。

## 6. 推荐启动顺序

每次换电脑或重新开始工作时按以下顺序：

1. 启动 MySQL。
2. 启动 Redis。
3. 在后端执行 `go run . migrate up`。
4. 启动后端 `go run .`。
5. 需要管理后台时启动 `PORT=8001 npm run dev`。
6. 需要小程序时在 `miniapp` 启动 H5 或微信构建。

## 7. Git 和本地文件注意事项

换电脑前，必须把需要继续的代码和文档提交并推送到远端；只存在当前电脑工作区的未提交修改不会随着 `git clone` 迁移：

```bash
git status
git branch --show-current
git add README.md docs/ internal/ api/ manifest/   # 按实际修改范围选择，不要盲目提交密钥
git commit -m "docs: organize cross-computer development guide"
git push origin 当前分支
```

如果同一分支还有其他未完成工作，先逐项检查 `git diff`，不要为了换电脑执行破坏性回退。新电脑首次开始工作时再执行：

```bash
git pull --rebase
```

开始工作前建议分别检查三个仓库：

```bash
git status
```

以下文件/目录不应提交：

- `manifest/config/config.yaml`（后端数据库、JWT、微信和 OSS 密钥）
- `.env*` 中的真实生产密钥
- `node_modules/`、`dist/`、`.pnpm-store/`
- `project.private.config.json`（微信开发者工具的个人设置）

迁移文件、API 定义和生成的 GoFrame 文件属于项目代码，不能因为换电脑而删除。换电脑后如果发现代码版本不一致，先分别在三个仓库执行 `git status` 和 `git log -1`，确认分支后再继续开发。

## 8. 常见问题速查

| 现象 | 处理 |
| --- | --- |
| `pnpm: command not found` | 直接使用文档中的 `npm install` 和 `npm run ...`，不依赖 pnpm。 |
| `未配置 TARO_APP_API_BASE_URL` | 检查 `miniapp/.env.development` 或生产环境变量，并重新构建。 |
| `request:fail url not in domain list` | 仅本机开发者工具可关闭域名校验；手机预览必须使用 HTTPS 合法域名。 |
| `ERR_CONNECTION_REFUSED` | 确认后端正在监听 8000，且真机不能使用电脑的 `127.0.0.1`。 |
| 管理后台接口 401/未授权 | 先登录管理员账号，确认浏览器保存了管理员 JWT，并检查后端 session/Redis。 |
| 数据库表不存在或字段缺失 | 在后端配置正确后重新执行 `go run . migrate up`。 |
| OSS `SignatureDoesNotMatch` | 检查 Endpoint、Bucket、AccessKey 是否属于同一地域和账号，密钥不要带多余空格。 |

更细的属性、属性分组和类目绑定说明见 [`attribute-management.md`](attribute-management.md)；认证接口和数据结构说明见本仓库 `README.md`。
