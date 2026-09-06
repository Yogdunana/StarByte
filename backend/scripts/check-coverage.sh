#!/usr/bin/env bash
# 生成覆盖率并检查核心包门禁（>=80%）。
# 有 PostgreSQL（TEST_DATABASE_REQUIRED=1）时额外门禁需要库的包。
set -euo pipefail
cd "$(dirname "$0")/.."

COVER_FILE="${COVER_FILE:-coverage.out}"
echo "=== running tests with coverage ==="
go test -count=1 -coverprofile="$COVER_FILE" \
  ./internal/auth/... ./internal/user/... ./internal/rbac/... ./internal/workflow/... \
  ./pkg/testutil/...

echo
echo "=== go tool cover (total of profiled packages) ==="
go tool cover -func="$COVER_FILE" | tail -1

python3 - "$COVER_FILE" <<'PY'
import collections
import os
import sys

path = sys.argv[1]
pkgs = collections.defaultdict(lambda: [0, 0])  # covered, total
with open(path) as f:
    next(f, None)  # mode line
    for line in f:
        line = line.strip()
        if not line:
            continue
        parts = line.rsplit(" ", 2)
        if len(parts) != 3:
            continue
        filepart, num_stmt, count = parts
        file_path = filepart.rsplit(":", 1)[0]
        if file_path.rsplit("/", 1)[-1].startswith("mock_"):
            continue
        pkg = file_path.rsplit("/", 1)[0]
        n = int(num_stmt)
        pkgs[pkg][1] += n
        if int(count) > 0:
            pkgs[pkg][0] += n

print()
print("=== per-package statement coverage ===")
for pkg in sorted(pkgs):
    cov, tot = pkgs[pkg]
    pct = 100.0 * cov / tot if tot else 0.0
    print(f"{pkg}: {pct:.1f}% ({cov}/{tot})")

gates = [
    "github.com/Yogdunana/StarByte/backend/internal/auth/repo",
    "github.com/Yogdunana/StarByte/backend/internal/rbac",
    "github.com/Yogdunana/StarByte/backend/internal/rbac/model",
]
if os.environ.get("TEST_DATABASE_REQUIRED") == "1":
    gates.append("github.com/Yogdunana/StarByte/backend/internal/auth/service")

print()
print("=== coverage gates (>=80%) ===")
fail = 0
for pkg in gates:
    if pkg not in pkgs:
        print(f"FAIL {pkg}: not in coverage profile")
        fail = 1
        continue
    cov, tot = pkgs[pkg]
    pct = 100.0 * cov / tot if tot else 0.0
    if pct + 1e-9 < 80:
        print(f"FAIL {pkg} {pct:.1f}% < 80%")
        fail = 1
    else:
        print(f"OK   {pkg} {pct:.1f}%")
sys.exit(fail)
PY
