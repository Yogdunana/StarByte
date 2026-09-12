# 特性开关 / 灰度一期（#98）

招新前可用的最小切片：布尔、用户名单、角色/部门、百分比 + Redis 热更新。  
定时开关、多变体 AB 分析、客户端 WASM、多环境矩阵 UI **明确不做**。

登录、CAS、入会申请、工作台 `/knowledge` 编辑 **不会**被开关拦截。

迁移：`000066_feature_flags`（`000064_knowledge` 已由 #180 占用，`000065` 留给 #176 backup phase-2）。错误码 `34000-34999`（`33000-33999` 是知识库）。

## 如何打开一个开关

系统管理 → 特性开关（需 `feature:read`，写操作另需 `feature:create|update|manage`）。  
社长 / 超管 / 副社长默认拥有上述权限。

### 给 10% 用户

1. 找到或新建开关，类型选 **百分比**
2. `percent = 10`，可选填写 `salt`（改盐值会重洗分桶）
3. 打开「启用」或调用 `POST /api/v1/system/features/:id/toggle`
4. 哈希为 `fnv64a(flag_key + salt + user_id) % 100`，同一用户稳定命中
5. 匿名访客对百分比 / 名单 **fail closed**

### 给指定用户名单

1. 类型选 **用户名单**
2. `user_ids` 填 UUID（每行一个）
3. 启用开关

### 给角色或部门

1. 类型选 **角色/部门**
2. 填写 `role_codes`（如 `minister`）和/或 `department_ids`
3. 任一命中即开启

布尔类型：启用后对所有人（含未登录）为真。

## 热更新

写操作会：

1. 落库并写审计
2. 删除 Redis `feature:snapshot`
3. 向 `feature:invalidate` pub/sub 广播
4. 本进程与其它实例重载内存快照

无需重启 `starbyte-server`。

## 当前接好的门闸

| Key | 默认 | 作用 |
|-----|------|------|
| `cms.public` | 关 | `/about-us`、`/docs`、`/docs/:slug`、公开 `/:slug`，以及 `GET /api/v1/knowledge/public/*` |
| `announcement.feed` | 开 | 工作台「最新公告」+ 公告列表/详情（公告写操作不拦；有发布权限的同事可绕过） |
| `membership.portal` | 关 | `/member/portal`（新门户，**不**替代入会申请） |

前端：`useFeature('cms.public')` / `<FeatureEnabled flag="cms.public">`。  
后端：`feature.Evaluate(flag, subject)` / `RequireFlag`（匿名为零 subject）。

SDK：`GET /api/v1/features/me?keys=cms.public,announcement.feed`（公开，带 JWT 则按用户评估）。
