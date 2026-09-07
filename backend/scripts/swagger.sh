#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
DIRS="cmd/server,pkg/response,internal/auth,internal/form,internal/user,internal/rbac,internal/workflow,internal/notification,internal/member,internal/interview,internal/meeting,internal/task,internal/internship,internal/audit,internal/stats,internal/file,internal/scheduler,internal/search,internal/cache,internal/export,internal/dict,internal/configstore,internal/finance,internal/discipline,internal/contract"
exec swag init -d "$DIRS" -g main.go -o docs --parseInternal
