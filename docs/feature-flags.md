# 特性开关 / 灰度（#98）

完整验收：布尔、用户名单、角色/部门、百分比、**多变体 AB**；用户 / 角色 / 百分比 / **环境** 定向；**定时上/下线**；Redis 热更新；操作审计与回滚；曝光分析。

登录、CAS、入会申请、工作台 `/knowledge` 编辑 **不会**被开关拦截。

| 项 | 值 |
|----|----|
| 迁移 | `000066_feature_flags` + `000071_feature_flags_phase2` |
| 错误码 | `34000-34999`（`34008` = 没有可回滚快照） |
| 热更新 | Redis `feature:snapshot` + pub/sub `feature:invalidate` |

## 管理入口

系统管理 → 特性开关（`feature:read`；写操作另需 `feature:create|update|manage`）。  
社长 / 超管 / 副社长默认拥有上述权限。

## 开关类型

| 类型 | 说明 |
|------|------|
| `boolean` | 启用后对所有命中环境/窗口的调用方为真（含未登录） |
| `user_allowlist` | `user_ids`（UUID）。匿名 fail closed |
| `role_dept` | `role_codes` 或 `department_ids`，任一命中即开 |
| `percentage` | `fnv64a(flag_key + salt + user_id) % 100 < percent`。匿名 fail closed |
| `ab_test` | 多个变体按权重稳定分桶；`control` / `off` 默认不当作门闸开启。匿名 fail closed（不取 `variants[0]`） |

### 给 10% 用户

1. 类型选 **百分比**，`percent = 10`，可选 `salt`
2. 打开「启用」
3. 匿名访客 fail closed

### 多变体 AB

1. 类型选 **AB 测试**
2. 至少两个变体，填写 `key` + `weight`（权重按比例归一）
3. 变体可单独标记是否作为门闸开启；未填时 `control`/`off` 为关，其余为开
4. 同一用户同一盐值稳定命中同一变体
5. `GET /api/v1/features/me` 与评估接口会写入曝光（登录用户），管理页「曝光分析」按 `variant + enabled` 统计独立用户（boolean / 百分比空变体也会分开计开/关）

前端：`useFeature('exp.hero').variant`。后端：`feature.Evaluate` 的 `Result.Variant`。

### 环境定向

`APP_ENV` 已是 `dev` / `test` / `prod`。规则里 `environments` 留空 = 全部环境；填写后仅所列环境生效。未设置 `APP_ENV` 时，带环境名单的开关 **fail closed**。

### 定时上/下线（不重启）

`starts_at` / `ends_at`（RFC3339）写在规则里：

- **评估时立即生效**：`enabled=true` 且未到 `starts_at` → `schedule_pending`；过了 `ends_at` → `schedule_expired`
- **后台约 30s 落库**：窗口开始且仍关闭 → 打开并记 `schedule_on`；窗口结束且仍开启 → 关闭并记 `schedule_off`
- **人工覆盖优先**：`UpdatedAt` 晚于 `starts_at` / `ends_at` 时，ticker 不再改回人工开关；评估层仍按窗口门闸。落库用 `id + updated_at` 条件更新，只写 `enabled`/`updated_at`；读后被别人改过则跳过且不记审计
- 列表里的「当前生效」= `enabled && 环境命中 && 窗口内`（不含用户定向）

推荐：把开关设为启用，再填未来的 `starts_at` / `ends_at`。评估不会等 ticker。

## 热更新

写操作会：

1. 落库并写审计
2. 从数据库重建内存，并 `SET` Redis `feature:snapshot`（避免 `DEL` 失败后旧快照被灌回）
3. `SET` 失败则删掉旧快照；删不掉就不广播，避免对端 `Get` 到过期数据。进程启动时 Redis 写失败仍会从数据库填内存并启动定时器/订阅
4. 向 `feature:invalidate` 广播，其它实例重载快照

无需重启 `starbyte-server`。

## 回滚

`POST /api/v1/system/features/:id/rollback`（`feature:update`）恢复最近一次 update / toggle / schedule / rollback 的 `before_json`。没有快照时返回 `34008`。

## 当前接好的门闸

| Key | 默认 | 作用 |
|-----|------|------|
| `cms.public` | 关 | `/about-us`、`/docs`、`/docs/:slug`、公开 `/:slug`，以及 `GET /api/v1/knowledge/public/*` |
| `announcement.feed` | 开 | 工作台整块「最新公告」（关则不渲染、不请求）+ 公告列表/详情（写操作不拦；有发布权限可绕过） |
| `membership.portal` | 关 | `/member/portal`（关则不请求门户接口；**不**替代入会申请） |

前端：`useFeature('cms.public')` / `<FeatureEnabled flag="cms.public">`。  
后端：`feature.Evaluate(flag, subject)` / `RequireFlag`（匿名为零 subject；`Subject.Environment` 来自 `APP_ENV`）。

SDK：`GET /api/v1/features/me?keys=cms.public,announcement.feed`（公开，带 JWT 则按用户评估）。`keys` 去重、非法格式丢弃，最多 32 个。

## 明确不做

- **客户端 WASM SDK**：过重，本期不做。浏览器继续走 HTTP SDK（`/features/me` + `useFeature`）。若以后要离线评估，再单独加 WASM 包。
