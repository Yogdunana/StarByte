# 000070 入会审批模板部署核对

修复后的 000070 只升级与 000061 种子完全相同的 JSONB，保留任何自定义发布图。迁移号不变；已执行到 70/71 的数据库不会重新执行此文件，因此本修复不能自动恢复已经被旧版迁移停用的自定义图。

部署方先执行以下只读查询，核对迁移状态及所有入会流程版本：

```sql
SELECT version, dirty FROM schema_migrations;
SELECT v.id, v.version, v.status, v.published_at, v.created_at, v.bpmn_data
FROM flow_definition_versions v
JOIN flow_definitions d ON d.id = v.definition_id
WHERE d.key = 'member_application'
ORDER BY v.version DESC;
```

如果旧版 000070 已执行，而且迁移前有自定义发布流程：对比紧邻迁移生成版本之前的图与预期审批规则。确认被替换后，通过流程设计器将确认过的旧图另存为新版本并发布；不要仅凭“四个节点”自动选择旧版本，也不要删除版本或改写已有实例引用。新申请采用修正后的发布版本；迁移后已经发起的实例需逐项核对，重新发布不会自动改变其流程快照。

未曾自定义且实际发布图符合干事→部长→社长预期时，无需恢复。不要通过回滚 000071/000070 来恢复流程，以免影响已写入的数据。

回归检查已接入 Database Migration Check。也可在完成 migrations up 的隔离测试库运行：

```sh
psql "$TEST_DATABASE_URL" -X -v ON_ERROR_STOP=1 -f backend/migrations/tests/member_application_v2.sql
```

测试使用临时表并回滚，覆盖默认种子升级与重复执行、自定义角色/会签保留、仅更改连线的图保留。
