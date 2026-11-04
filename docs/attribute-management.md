# 属性与属性分组管理

本文记录当前版本的属性管理模型，方便后续开发和在另一台电脑继续工作。

## 1. 数据模型

属性和类目不再直接绑定，关系分成三层：

```text
attributes
    │  多对多
    ▼
attribute_group_attributes
    ▲
    │
attribute_groups
    │  多对多
    ▼
category_attribute_group_bindings
    │
    ▼
叶子类目 categories
```

表职责：

| 表 | 职责 |
| --- | --- |
| `attributes` | 全局属性定义，例如“平板重量” |
| `attribute_options` | 单选、多选属性的选项 |
| `attribute_groups` | 可复用的属性分组，例如“基础参数” |
| `attribute_group_attributes` | 分组包含哪些属性及分组内排序 |
| `category_attribute_group_bindings` | 哪些分组绑定到哪些叶子类目 |
| `category_attributes` | 类目级覆盖配置，不再作为主要绑定关系 |

`category_attributes` 中的覆盖配置包括 `is_required`、`is_filterable`、`is_searchable` 和 `sort_order`。当前管理端暂时隐藏筛选和搜索配置，等搜索引擎建立后再开放。

一级类目使用 `categories.parent_id = 0`，不能绑定属性分组；只有没有子类目的叶子类目可以绑定分组。

## 2. 推荐操作顺序

1. 在“属性管理”创建全局属性。属性可以暂时不选择分组。
2. 在“属性分组”创建分组。
3. 在分组的“管理属性”中加入属性并调整排序。
4. 在分组的“绑定类目”中选择叶子类目并绑定。
5. 在绑定记录上点击“配置属性”，设置该类目下属性是否必填和排序。

同一分组可以复用到多个叶子类目。同一属性也可以加入多个分组。一个属性如果通过多个分组出现在同一类目下，系统按分组排序和属性排序去重；类目级覆盖优先于默认值。

## 3. 管理端接口

### 属性

```text
GET  /api/v1/admin/attributes
POST /api/v1/admin/attributes/create
POST /api/v1/admin/attributes/update
POST /api/v1/admin/attributes/delete
GET  /api/v1/admin/attributes/options?attribute_id=...
POST /api/v1/admin/attributes/options/create
POST /api/v1/admin/attributes/options/update
POST /api/v1/admin/attributes/options/delete
```

创建或更新属性时可以传递：

```json
{
  "code": "tablet_weight",
  "name": "平板重量",
  "value_type": 2,
  "input_type": "single_line",
  "group_ids": [1, 3]
}
```

只有 `select` 和 `multi_select` 输入方式显示和维护选项。

### 分组与类目

```text
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
```

### 类目级覆盖

```text
GET  /api/v1/admin/category-attributes?category_id=...
POST /api/v1/admin/category-attributes/save
POST /api/v1/admin/category-attributes/delete
```

这里的 `category-attributes` 返回的是分组绑定后的有效属性，并叠加类目级覆盖；它不负责创建独立的类目属性绑定。保存覆盖时必须传递已绑定到该类目的 `group_id` 和该分组中的 `attribute_id`。

## 4. 删除和变更约束

- 有分组成员关系、类目绑定或覆盖配置的分组不能删除。
- 已绑定类目的分组不能移除其中的属性。
- 编辑属性时，如果尝试从已绑定类目的分组中移除属性，接口会拒绝操作。
- 已被类目覆盖配置或商品属性值使用的属性不能删除。
- 被商品属性值使用的选项不能删除。
- 非叶子类目不能绑定分组，也不能保存类目属性覆盖。

## 5. 迁移

当前迁移版本为 `000008_normalize_attribute_groups`：

- `000006_categories_use_zero_parent`：将一级类目的 `parent_id` 统一为 `0`。
- `000007_create_category_attribute_groups`：建立旧的类目级分组结构。
- `000008_normalize_attribute_groups`：将旧结构转换为全局分组、分组属性关系和类目分组绑定。

执行迁移：

```bash
go run . migrate up
```

迁移文件已经内嵌到后端程序中，不能修改已执行的迁移；后续数据库变更必须新增编号递增的 `.up.sql` 和 `.down.sql` 文件。

## 6. 换电脑继续开发

```bash
git clone git@github.com:suxiangdong/indu-scope-backend.git
git clone git@github.com:suxiangdong/indu-scope-manager.git

cd indu-scope-backend
cp manifest/config/config.example.yaml manifest/config/config.yaml
# 填写本机数据库、Redis 和 JWT 配置
go mod download
go run . migrate up
GOCACHE=/tmp/indu-scope-go-cache go test ./...

cd ../indu-scope-manager
npm install
npm run tsc
npm run build
```

后端默认监听 `http://127.0.0.1:8000`。为避免端口冲突，管理端建议使用 `PORT=8001 npm run dev`，访问 `http://localhost:8001`；前端 `/api/` 请求会代理到后端地址。不要提交 `manifest/config/config.yaml`、`.env` 或任何真实凭据。
