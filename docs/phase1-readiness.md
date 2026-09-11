# 一期可用性检查（更新：2026-09-07）

对照 `TEAM_DEV_GUIDE.md` §9.4 的 33 张门禁表、架构文档 §10，以及本仓库实现。

## 结论

**除学校 HAP / 企微 / 外网域名跳转外，一期业务与支持项已齐。** 校园网 CAS 已接 `authserver.smbu.edu.cn`（回调 `starbyte.smbu.edu.cn`）。财务、纪律处分、合同已有 API 与产品页；壳层支持 i18n 与亮/暗/跟随系统主题；README、部署、入门、API、架构文档已补全。

前端仍有 4 个占位页：流程实例、我的待办、角色管理、权限管理（后端 API 已有），不挡一期上线试用。

## 门禁表（33 张）对照 GitHub

| 优先级 | 已完成 | 缺口 |
|--------|--------|------|
| P0 7 项 | #1 #2 #12 #13 #14 #15 #16 | 无 |
| P1 核心 8 项 | #3 #4 #5 #17 #18 #19 #20 #21 | 无 |
| P2 业务 | #6 #7 #8 #9 #10 #11 #25 #26 #28；#22 #23 #24；#27 已关 | 无（OAuth 不在此列） |
| P3 支持 | #29 #30 #32 #33；#31 壳层+新模块+占位页 i18n 与主题 | 存量业务页文案迁 i18n（不挡一期试用） |

## 本批新增

- `#22` `GET/POST/PUT/DELETE /api/v1/finance/records`，分类、汇总；`GET /finance/export` 501 预留
- `#23` 处分 CRUD + 审批/撤销/申诉 + `discipline_notice`；有已发布流程 `discipline_approve` 时挂接实例
- `#24` 合同 CRUD + 模板 + 附件 `file_id`；调度 handler `contract_expiry`；`GET /contracts/expiring`
- `#31` react-i18next（zh-CN / en-US）+ 主题持久化。壳层（登录、顶栏、菜单、设置、403/404、占位页）与财务/处分/合同页使用 `t()`；存量业务页仍以中文为默认文案
- `#33` `docs/deployment.md`、`docs/getting-started.md`、`docs/api.md`，并更新 README / 架构 §10–11
