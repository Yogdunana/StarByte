# Git 工作流规范

> **版本**: v1.0
> **日期**: 2026-08-24
> **适用范围**: StarByte 全项目

---

## 1. 工作流模型

采用 **GitHub Flow** 工作流：

```
main ──●───────────●───────────●─────  主分支，生产环境代码
       │           │           │
       └── feature/xxx ──┘  feature/yyy ──┘
```

### 核心原则
1. **main 分支**：始终保持可部署状态，代码必须稳定
2. **feature 分支**：所有新功能、修复都在 feature 分支上开发
3. **Pull Request**：代码通过 PR 合入 main，必须经过 Code Review
4. **Squash Merge**：合并时压缩为一个 commit，保持 main 历史干净

---

## 2. 分支命名规范

### 2.1 分支类型前缀

| 前缀 | 说明 | 示例 |
|------|------|------|
| `feature/` | 新功能开发 | `feature/user-module` |
| `fix/` | Bug 修复 | `fix/login-error` |
| `hotfix/` | 紧急修复（直接从 main 切出） | `hotfix/prod-crash` |
| `refactor/` | 代码重构 | `refactor/service-layer` |
| `docs/` | 文档更新 | `docs/api-doc` |
| `chore/` | 构建/工具/依赖更新 | `chore/update-deps` |

### 2.2 命名规则
- 前缀 + 简短描述，使用 kebab-case（小写 + 连字符）
- 描述要清晰，能看出这个分支在做什么
- 可以带模块名前缀，便于识别

```
# 好的命名
feature/workflow-engine
fix/member-approval-bug
feature/user-profile-page
chore/upgrade-gin-v1-10

# 不好的命名
feature/xxx          # 不清晰
fix-bug              # 太笼统
new_feature          # 没有前缀，驼峰
```

---

## 3. 开发流程

### 3.1 开始新功能

```bash
# 1. 切换到 main 分支
git checkout main

# 2. 拉取最新代码
git pull origin main

# 3. 创建 feature 分支
git checkout -b feature/your-feature-name
```

### 3.2 开发过程中

```bash
# 定期同步 main 分支的最新代码（建议每天至少一次）
git checkout main
git pull origin main
git checkout feature/your-feature-name
git rebase main
# 或 git merge main
```

### 3.3 提交代码

```bash
# 查看修改
git status
git diff

# 添加文件
git add path/to/file

# 提交（遵循 commit 规范）
git commit -m "feat: add user login api"

# 推送到远程
git push origin feature/your-feature-name
```

### 3.4 提 PR
1. 在 GitHub 上创建 Pull Request
2. 填写 PR 模板（标题、描述、测试截图等）
3. 指定 Reviewer
4. 等待 Code Review 和 CI 检查通过

### 3.5 合并代码
- **使用 Squash Merge**：一个 PR 对应一个 commit
- 合并后自动删除源分支
- 合并完成后，本地可以删除 feature 分支

```bash
# 本地清理
git checkout main
git pull origin main
git branch -d feature/your-feature-name
```

---

## 4. Commit 规范

采用 **Conventional Commits** 规范。

### 4.1 Commit 格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

### 4.2 Type 类型

| type | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | 修复 bug |
| `docs` | 文档更新 |
| `style` | 代码格式调整（不影响代码运行） |
| `refactor` | 重构（不是新功能，也不是修 bug） |
| `perf` | 性能优化 |
| `test` | 增加测试 |
| `chore` | 构建过程或辅助工具的变动 |
| `ci` | CI/CD 相关变更 |
| `revert` | 回滚之前的 commit |

### 4.3 Scope（可选）
- 影响的模块或范围
- 如 `user`、`workflow`、`frontend`、`backend`

### 4.4 Subject
- 简短描述，不超过 50 个字符
- 使用动词开头（add、fix、update、remove 等）
- 不用句号结尾
- 中文或英文都可以，但项目内保持统一（推荐中文）

### 4.5 Body（可选）
- 详细描述变更内容
- 可以分点说明
- 说明为什么这么改

### 4.6 Footer（可选）
- 关联 Issue：`Closes #123`
- 破坏性变更：`BREAKING CHANGE: xxx`

### 4.7 示例

```
feat(workflow): 添加条件分支节点支持

- 实现排他网关节点类型
- 集成 govaluate 表达式引擎
- 支持在流程设计器中配置分支条件

Closes #45
```

```
fix(user): 修复用户列表分页参数错误

- 修复 page 参数默认值为 0 的问题
- 添加参数校验，确保 page >= 1
```

```
docs: 更新后端开发规范

- 补充错误码规范
- 添加数据库迁移说明
```

---

## 5. Pull Request 规范

### 5.1 PR 标题
- 格式同 Commit：`type(scope): subject`
- 清晰描述 PR 做了什么

### 5.2 PR 描述
使用 PR 模板，包含以下内容：

1. **变更描述**：这个 PR 做了什么？
2. **关联 Issue**：关联哪个 Issue/Ticket
3. **测试方式**：如何验证功能？
4. **截图/GIF**：前端变更必须附截图
5. **注意事项**：有没有需要特别注意的地方？
6. **检查清单**：
   - [ ] 代码遵循项目开发规范
   - [ ] 已进行自测
   - [ ] 已添加/更新必要的文档
   - [ ] CI 检查通过

### 5.3 PR 大小
- 单个 PR 控制在 **500 行以内**（不含自动生成文件）
- 超过的话考虑拆成多个 PR
- 大功能可以分多个 PR 逐步合入，使用 feature flag 控制

### 5.4 Code Review
- 每个 PR 至少需要 **1 个人 Review 通过**
- 核心模块（流程引擎、权限系统）需要 **2 个人 Review**
- Reviewer 要认真看，不能随便 approve
- 有疑问及时在 PR 里评论讨论

### 5.5 合并条件
- ✅ 至少 1 个 Approve（核心模块 2 个）
- ✅ CI 检查全部通过
- ✅ 没有未解决的评论
- ✅ 没有冲突（有冲突先解决）
- ✅ Squash Merge

---

## 6. 常用 Git 操作

### 6.1 暂存修改
```bash
# 暂存当前修改
git stash

# 查看暂存列表
git stash list

# 恢复暂存
git stash pop
```

### 6.2 修改最后一次提交
```bash
# 修改最后一次提交的 message
git commit --amend

# 把新的修改加到最后一次提交
git add forgotten-file
git commit --amend --no-edit
```

### 6.3 回退提交
```bash
# 软回退（保留修改，只撤销 commit）
git reset --soft HEAD~1

# 硬回退（彻底丢弃修改，慎用！）
git reset --hard HEAD~1

# 回退到指定 commit
git reset --hard <commit-hash>
```

### 6.4 撤销已经 push 的提交
```bash
# 使用 revert（推荐，会生成新的反向 commit）
git revert <commit-hash>
git push origin main

# 不推荐：force push（会改历史，只有自己的分支才能这么干）
git reset --hard HEAD~1
git push origin feature/xxx --force
```

---

## 7. 注意事项

1. **永远不要 force push main 分支**
2. **不要把大文件提交到 Git**（图片、视频、依赖包等），用 .gitignore 排除
3. **不要提交敏感信息**（密码、密钥、Token 等），用环境变量
4. **提交前先自测**，确保代码能正常运行
5. **保持 commit 原子性**：一个 commit 做一件事
6. **及时同步 main 分支**：避免最后合并时冲突太多
7. **PR 标题和描述要认真写**：Reviewer 靠这个理解你做了什么
