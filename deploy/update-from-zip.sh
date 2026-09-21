#!/usr/bin/env bash
#
# 用 zip 覆盖更新 StarByte 代码（供应链加固版）。
#
# 安全目标：堵住「微信群 zip → 宿主机 root」的供应链 RCE。
#   - 白名单准入：一组受保护的基础设施文件/目录，永不覆盖。
#   - SHA256 校验清单：可选强校验，不匹配即中止，绝不写入。
#   - zip-slip 防护：拒绝含绝对路径或 `..` 段条目的 zip。
#   - 非普通文件防护：拒绝含符号链接等特殊文件条目的 zip。
#     这条不是锦上添花 —— 受保护清单是靠 `find -type f` 枚举出来的，而符号链接
#     不是普通文件，既进不了「可覆盖」清单也进不了「受保护」清单，最后会被
#     tar 原样写进仓库，直接把 backend/entrypoint.sh、deploy/cli/starbyte、
#     .github/** 这些「永不覆盖」的目标换成攻击者内容，绕过整套白名单。
#
# 设计要点：
#   - 始终保留现有 deploy/.env（不覆盖密钥）。
#   - 绝不执行 docker volume / compose down -v。
#   - 所有校验（zip-slip → 非普通文件 → SHA256 → 受保护清单）在覆盖【之前】完成；
#     任一失败都「不写入任何文件」。
set -euo pipefail

#######################################
# 配置 & 工具函数
#######################################

PROG="$(basename "${BASH_SOURCE[0]}")"

usage() {
  cat <<'EOF'
用法: bash deploy/update-from-zip.sh [选项] <StarByte.zip>

用 zip 覆盖更新到仓库根目录，并做供应链加固。

安全校验（全部在覆盖之前完成；任一失败都不写入任何文件）：
  • zip-slip 防护           拒绝含绝对路径或 `..` 路径段的条目
  • 非普通文件防护          拒绝含符号链接/设备文件等条目的 zip
                            （符号链接会绕过下面的受保护清单，故直接拒绝，
                              合法发行包不会包含它们）
  • 受保护基础设施清单      一组文件/目录永不覆盖，且会被报告/告警
  • SHA256 校验清单（可选） 强校验，不匹配即中止

选项：
  --sha256 <hex>     提供 zip 文件的期望 SHA256（64 位十六进制）。
                     与自动探测的清单二选一，显式参数优先级更高。
  --strict           偏执模式：只要 zip 里出现任何受保护的基础设施
                     文件/目录，立即中止且不写入任何文件。
                     默认【不】开启——否则正常更新也会被卡死。
  -h, --help         显示本帮助。

受保护清单（相对仓库根，大小写不敏感、子目录同样匹配）：
  • 所有 Dockerfile / Dockerfile.*（任意层级）
  • deploy/docker-compose.yml、deploy/docker-compose.dev.yml
  • deploy/cli/**（整个 CLI 目录，含 starbyte）
  • deploy/update-from-zip.sh 自身 及 deploy/*.sh
      （脚本自身受保护，防止被 zip 替换导致「自我替换」；
       其余 deploy 下的脚本同样受保护）
  • backend/entrypoint.sh、backend/configs/**
  • frontend/nginx.conf、frontend/security-headers.conf
  • .github/**（CI 工作流）
  • 根 Makefile、.gitignore、.dockerignore
  • backend/.dockerignore、frontend/.dockerignore

默认（非 --strict）：zip 里出现的受保护文件会被排除，结束时醒目列出
「跳过了哪些」「zip 里有 N 个受保护文件已被忽略」。

SHA256 清单（强烈推荐）：
  发布方生成：  sha256sum StarByte.zip > StarByte.zip.sha256
  本脚本会自动探测同目录下的 <zip>.sha256
  （兼容 sha256sum 输出 `<hex>  <文件名>` 与纯 hex 行）。
  提供清单 → 强校验，不匹配即中止；
  未提供   → 打印醒目警告后继续（救火场景不硬失败，但风险自担）。

zip 可以是 GitHub 的 Source code（顶层带 StarByte-main/ 目录），
也可以是仓库根文件直接打包。

更新后请重建应用（不会强拉 MinIO）：
  starbyte rebuild all
  # 或
  docker compose -f deploy/docker-compose.yml up -d --build --no-deps --force-recreate backend frontend
EOF
}

# 计算文件的 SHA256（优先 sha256sum，回退 shasum / python3）
compute_sha256() {
  local f="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$f" | awk '{print $1}'
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$f" | awk '{print $1}'
  elif command -v python3 >/dev/null 2>&1; then
    python3 - "$f" <<'PY'
import sys, hashlib
h = hashlib.sha256()
with open(sys.argv[1], 'rb') as fp:
    for chunk in iter(lambda: fp.read(65536), b''):
        h.update(chunk)
print(h.hexdigest())
PY
  else
    echo "缺少 sha256sum / shasum / python3，无法计算哈希。" >&2
    return 1
  fi
}

# 判断相对路径（仓库根相对，使用 / 分隔）是否受保护。
# 返回 0 = 受保护（永不覆盖）；返回 1 = 可覆盖。
is_protected() {
  local p="$1"
  p="${p//\\//}"                 # 统一分隔符
  local lower="${p,,}"           # 大小写不敏感
  local base
  base="$(basename "$p")"
  base="${base,,}"

  # 1) 任意层级的 Dockerfile / Dockerfile.*
  if [[ "$base" == "dockerfile" || "$base" == dockerfile.* ]]; then
    return 0
  fi

  # 2) 受保护的目录前缀（整目录永不被覆盖）
  local d
  for d in ".github" "deploy/cli" "backend/configs"; do
    if [[ "$lower" == "$d" || "$lower" == "$d/"* ]]; then
      return 0
    fi
  done

  # 3) 显式文件（大小写不敏感精确匹配）
  local f
  for f in \
      "deploy/docker-compose.yml" \
      "deploy/docker-compose.dev.yml" \
      "deploy/update-from-zip.sh" \
      "backend/entrypoint.sh" \
      "frontend/nginx.conf" \
      "frontend/security-headers.conf" \
      "makefile" \
      ".gitignore" \
      ".dockerignore" \
      "backend/.dockerignore" \
      "frontend/.dockerignore" ; do
    if [[ "$lower" == "$f" ]]; then
      return 0
    fi
  done

  # 4) deploy 下的其它脚本（除本脚本本身已在上面显式保护）
  if [[ "$lower" == deploy/*.sh ]]; then
    if [[ "$lower" != "deploy/update-from-zip.sh" ]]; then
      return 0
    fi
  fi

  return 1
}

#######################################
# 参数解析
#######################################

STRICT=0
EXPECTED_HASH=""
ZIP=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help) usage; exit 0 ;;
    --strict) STRICT=1; shift ;;
    --sha256)
      if [[ $# -lt 2 ]]; then echo "错误: --sha256 需要一个值" >&2; exit 2; fi
      EXPECTED_HASH="$2"; shift 2 ;;
    --sha256=*) EXPECTED_HASH="${1#--sha256=}"; shift ;;
    -*) echo "未知选项: $1" >&2; usage >&2; exit 2 ;;
    *)
      if [[ -n "$ZIP" ]]; then echo "错误: 多余的参数: $1" >&2; exit 2; fi
      ZIP="$1"; shift ;;
  esac
done

if [[ -z "$ZIP" ]]; then
  usage >&2
  exit 2
fi
if [[ ! -f "$ZIP" ]]; then
  echo "找不到 zip: $ZIP" >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
COMPOSE_FILE="$REPO_ROOT/deploy/docker-compose.yml"
ENV_FILE="$REPO_ROOT/deploy/.env"

if [[ ! -f "$COMPOSE_FILE" ]]; then
  echo "未找到 $COMPOSE_FILE，请在 StarByte 仓库内执行本脚本。" >&2
  exit 1
fi

# mktemp：若设置了 $TMPDIR（受限/沙箱环境常见）则在其下创建，否则回退默认行为
if [[ -n "${TMPDIR:-}" ]]; then
  WORKDIR="$(mktemp -d "$TMPDIR/XXXXXXXXX")"
else
  WORKDIR="$(mktemp -d)"
fi
cleanup() { rm -rf "$WORKDIR"; }
trap cleanup EXIT

if [[ -f "$ENV_FILE" ]]; then
  cp -a "$ENV_FILE" "$WORKDIR/preserved.env"
  echo "已暂存 deploy/.env"
fi

#######################################
# 步骤 1/4: zip-slip + 非普通文件防护（解压前先列条目）
#######################################
echo "=== 步骤 1/4: 校验 zip 条目（zip-slip + 非普通文件防护）==="

list_zip_entries() {
  if command -v unzip >/dev/null 2>&1; then
    unzip -Z1 "$ZIP"
  elif command -v python3 >/dev/null 2>&1; then
    python3 - "$ZIP" <<'PY'
import sys, zipfile
with zipfile.ZipFile(sys.argv[1]) as z:
    for n in z.namelist():
        print(n)
PY
  else
    echo "需要 unzip 或 python3 才能读取 zip 条目。" >&2
    exit 1
  fi
}

# 列出 zip 中类型不是「普通文件 / 目录」的条目（符号链接、FIFO、设备文件等）。
#
# 为什么要在解压【之前】查：unzip 在 Unix 上会忠实还原符号链接（unzip(1)：
# "the only file types restored by unzip are regular files, directories and
# symbolic (soft) links"）。而符号链接落在解压树里之后：
#   1) 它不是 `find -type f` 的命中对象 —— 于是既进不了「可覆盖」清单，
#      也进不了「受保护」清单，白名单对它完全失效；
#   2) 最后的 `tar -C "$SRC" -cf - .` 默认把符号链接当符号链接存（不加 -h
#      就不会解引用），解到仓库里就把目标换成了攻击者指定的链接；
#   3) 于是 backend/entrypoint.sh（镜像入口）、deploy/cli/starbyte（运维用
#      sudo 执行）、.github/**（CI 工作流）都能被不可信 zip 替换 —— 供应链 RCE。
#
# 探测工具不可用 / 探测失败时输出为空，此时靠解压之后的兜底复查拦截。
list_zip_nonregular() {
  if command -v unzip >/dev/null 2>&1; then
    # zipinfo 长格式（unzip -Z -l）每行首字符即文件类型：
    #   `-` 普通文件、`d` 目录、`?` 未知（FAT/DOS 来源，按普通文件看待），
    #   其余（l 符号链接 / b 块设备 / c 字符设备 / p FIFO / s socket）一律可疑。
    # 表头/表尾行以 A、Z、数字等开头，不会命中 /^[lbcps]/，故不会误报。
    # 报错信息只需要条目名，故按时间戳字段把它从长格式行里切出来；
    # 切不出来就原样打印整行（判定只看首字符，不依赖这里的解析结果）。
    unzip -Z -l "$ZIP" 2>/dev/null | awk '
      /^[-d?]/ { next }
      /^[lbcps]/ {
        if (match($0, /[0-9]+-[A-Za-z]+-[0-9]+ [0-9]+:[0-9]+ /)) {
          print substr($0, RSTART + RLENGTH)
        } else {
          print $0
        }
      }'
  elif command -v python3 >/dev/null 2>&1; then
    python3 - "$ZIP" <<'PY'
import stat, sys, zipfile

with zipfile.ZipFile(sys.argv[1]) as z:
    for info in z.infolist():
        # 只看外部属性里的 Unix 类型位，不能拿 stat.S_ISREG 直接判：
        # 很多打包器（包括 Python 自己的 zipfile.writestr）只写权限位、
        # 不写类型位（如 0o600），此时 S_ISREG(0o600) 为假，会把普通文件误判成
        # 特殊文件，导致正常更新被拒。
        kind = stat.S_IFMT(info.external_attr >> 16)
        # kind == 0：没有类型位（上述情况，或 FAT/DOS 创建的包），按普通文件处理。
        if kind == 0 or kind == stat.S_IFREG or kind == stat.S_IFDIR:
            continue
        print(info.filename)
PY
  fi
}

ZIP_ENTRIES_FILE="$WORKDIR/entries.txt"
list_zip_entries > "$ZIP_ENTRIES_FILE"

SLIP_FOUND=0
while IFS= read -r entry; do
  [[ -z "$entry" ]] && continue
  norm="${entry//\\//}"
  if [[ "$norm" == /* ]]; then
    echo "  ✗ 拒绝：条目含绝对路径: $entry" >&2
    SLIP_FOUND=1
  fi
  if [[ "$norm" == *"/../"* || "$norm" == "../"* || "$norm" == *"/.." ]]; then
    echo "  ✗ 拒绝：条目含 '..' 路径段: $entry" >&2
    SLIP_FOUND=1
  fi
done < "$ZIP_ENTRIES_FILE"

if [[ $SLIP_FOUND -ne 0 ]]; then
  echo "错误: zip 含潜在路径穿越（zip-slip）条目，已中止，未写入任何文件。" >&2
  exit 1
fi
echo "  ✓ 未发现 zip-slip 条目"

# 1b) 非普通文件（符号链接 / 特殊文件）防护。
# 探测失败（工具不可用或输出格式不符）不在这里硬失败：解压后的兜底复查会兜住，
# 而此处误判会把正常更新卡死。
ZIP_NONREG_FILE="$WORKDIR/nonregular.txt"
list_zip_nonregular > "$ZIP_NONREG_FILE" || true

if [[ -s "$ZIP_NONREG_FILE" ]]; then
  echo "  ✗ 拒绝：zip 含符号链接/特殊文件条目（受保护清单无法拦截它们）：" >&2
  while IFS= read -r entry; do
    [[ -n "$entry" ]] && echo "      $entry" >&2
  done < "$ZIP_NONREG_FILE"
  echo "错误: zip 含非普通文件条目，已中止，未写入任何文件。" >&2
  exit 1
fi
echo "  ✓ 未发现符号链接/特殊文件条目"

#######################################
# 步骤 2/4: SHA256 完整性校验（覆盖之前）
#######################################
echo "=== 步骤 2/4: SHA256 完整性校验 ==="

MANIFEST="${ZIP}.sha256"
RESOLVED_HASH=""
HASH_SOURCE=""
FALLBACK_HASH=""
FALLBACK_NAME=""

if [[ -n "$EXPECTED_HASH" ]]; then
  RESOLVED_HASH="$EXPECTED_HASH"
  HASH_SOURCE="--sha256 参数"
elif [[ -f "$MANIFEST" ]]; then
  ZIP_BASENAME="$(basename "$ZIP")"
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    line="${line%$'\r'}"
    if [[ "$line" =~ ^[0-9a-fA-F]{64}$ ]]; then
      RESOLVED_HASH="$line"
      HASH_SOURCE="清单 $(basename "$MANIFEST")（纯 hex）"
      break
    fi
    if [[ "$line" =~ ^([0-9a-fA-F]{64})[[:space:]]+[*]?(.*)$ ]]; then
      hex="${BASH_REMATCH[1]}"
      name="${BASH_REMATCH[2]}"
      name="${name%$'\r'}"
      if [[ "$(basename "$name")" == "$ZIP_BASENAME" ]]; then
        RESOLVED_HASH="$hex"
        HASH_SOURCE="清单 $(basename "$MANIFEST")（匹配 $ZIP_BASENAME）"
        break
      fi
      FALLBACK_HASH="$hex"
      FALLBACK_NAME="$name"
    fi
  done < "$MANIFEST"
  if [[ -z "$RESOLVED_HASH" && -n "$FALLBACK_HASH" ]]; then
    RESOLVED_HASH="$FALLBACK_HASH"
    HASH_SOURCE="清单 $(basename "$MANIFEST")（回退：$FALLBACK_NAME）"
  fi
fi

if [[ -z "$RESOLVED_HASH" ]]; then
  echo "  ⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠" >&2
  echo "  ⚠  警告：未提供校验清单（--sha256 或 <zip>.sha256），" >&2
  echo "  ⚠  本次更新【未做任何完整性校验】。" >&2
  echo "  ⚠  来源不可信时存在被篡改/投毒风险，请确认 zip 来源可信！" >&2
  echo "  ⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠" >&2
else
  ACTUAL_HASH="$(compute_sha256 "$ZIP")" || true
  if [[ -z "$ACTUAL_HASH" ]]; then
    echo "错误：无法计算 zip 的 SHA256，已中止。" >&2
    exit 1
  fi
  if [[ "${ACTUAL_HASH,,}" != "${RESOLVED_HASH,,}" ]]; then
    echo "错误：SHA256 校验失败（来源: $HASH_SOURCE）。" >&2
    echo "  期望: $RESOLVED_HASH" >&2
    echo "  实际: $ACTUAL_HASH" >&2
    echo "  已中止，未写入任何文件。" >&2
    exit 1
  fi
  echo "  ✓ SHA256 校验通过（来源: $HASH_SOURCE）"
fi

#######################################
# 步骤 3/4: 解压到临时目录
#######################################
echo "=== 步骤 3/4: 解压到临时目录 ==="

EXTRACT="$WORKDIR/extract"
mkdir -p "$EXTRACT"

if command -v unzip >/dev/null 2>&1; then
  unzip -q "$ZIP" -d "$EXTRACT"
elif command -v python3 >/dev/null 2>&1; then
  python3 -m zipfile -e "$ZIP" "$EXTRACT"
else
  echo "需要 unzip 或 python3 才能解压。" >&2
  exit 1
fi

# 顶层若只有一个目录，视为 GitHub 源码包
shopt -s nullglob
entries=("$EXTRACT"/*)
shopt -u nullglob
SRC="$EXTRACT"
if [[ ${#entries[@]} -eq 1 && -d "${entries[0]}" ]]; then
  SRC="${entries[0]}"
fi

if [[ ! -f "$SRC/deploy/docker-compose.yml" && ! -f "$SRC/README.md" ]]; then
  echo "zip 内容不像 StarByte 仓库根（缺少 deploy/docker-compose.yml 或 README.md）。" >&2
  exit 1
fi

# 兜底复查（解压之后、枚举/写入之前）：解压器是否还原符号链接因平台而异
# （Linux 的 Info-ZIP unzip 会还原；python3 -m zipfile 不会，它把链接目标
# 写成普通文件的内容，那种情况下会被下面的 -type f + 受保护清单正常拦下）。
# 所以这里不依赖步骤 1 的探测结果，直接在解压树上找「既不是普通文件也不是目录」
# 的任何东西。这是权威判定：命中即中止，且此刻尚未向仓库写入任何文件。
SPECIAL_FOUND="$(find "$SRC" ! -type f ! -type d -print 2>/dev/null || true)"
if [[ -n "$SPECIAL_FOUND" ]]; then
  echo "错误: 解压结果含非普通文件（符号链接/特殊文件），已中止，未写入任何文件：" >&2
  while IFS= read -r p; do
    [[ -n "$p" ]] && echo "  - $p" >&2
  done <<< "$SPECIAL_FOUND"
  exit 1
fi

# 枚举 SRC 下所有文件，区分受保护与可写
TOTAL_FILES=0
PROTECTED_LIST=()
WRITE_LIST=()

while IFS= read -r rel; do
  [[ -z "$rel" ]] && continue
  # find 在 $SRC 内执行，输出已是仓库根相对路径（去掉前导 ./）
  # 跳过 .git 与 deploy/.env（原有行为保留）
  if [[ "$rel" == ".git" || "$rel" == .git/* ]]; then continue; fi
  if [[ "$rel" == "deploy/.env" || "$rel" == "deploy/.env/"* ]]; then continue; fi
  TOTAL_FILES=$((TOTAL_FILES+1))
  if is_protected "$rel"; then
    PROTECTED_LIST+=("$rel")
  else
    WRITE_LIST+=("$rel")
  fi
done < <(cd "$SRC" && find . -type f | sed 's|^\./||')

#######################################
# 步骤 4/4: 覆盖预览 + 写入
#######################################
echo "=== 步骤 4/4: 覆盖预览 ==="
echo "将覆盖文件数（不含 .git / .env）：    ${#WRITE_LIST[@]}"
echo "跳过（受保护）文件数：               ${#PROTECTED_LIST[@]}"

# --strict：出现任何受保护文件即中止，不写入
if [[ $STRICT -eq 1 ]]; then
  if [[ ${#PROTECTED_LIST[@]} -gt 0 ]]; then
    echo "错误: --strict 模式发现 zip 内含 ${#PROTECTED_LIST[@]} 个受保护的基础设施文件，已中止，未写入任何文件：" >&2
    for p in "${PROTECTED_LIST[@]}"; do echo "  - $p" >&2; done
    exit 1
  fi
fi

# 给运维一次中断机会（仅在交互终端）
if [[ -t 1 ]]; then
  echo "确认覆盖？按 Enter 继续，或 Ctrl-C 中止。"
  read -r _ || true
fi

# 从 SRC 中剔除受保护文件/目录（确定性做法，避免 tar --exclude
# 对前导 ./ 的匹配歧义；被剔除项绝不会进入写出的 tar 流）
for p in "${PROTECTED_LIST[@]}"; do
  rm -rf "$SRC/$p"
done

echo "正在覆盖: $REPO_ROOT  （来源: $SRC）"
tar -C "$SRC" --exclude='.git' --exclude='deploy/.env' -cf - . \
  | tar -C "$REPO_ROOT" -xf -

if [[ -f "$WORKDIR/preserved.env" ]]; then
  mkdir -p "$(dirname "$ENV_FILE")"
  cp -a "$WORKDIR/preserved.env" "$ENV_FILE"
  echo "已还原 deploy/.env"
else
  echo "目录中原先没有 deploy/.env（未新建密钥文件）"
fi

# 受保护文件被忽略的醒目告警
if [[ ${#PROTECTED_LIST[@]} -gt 0 ]]; then
  echo "⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠" >&2
  echo "⚠  警告：zip 里有 ${#PROTECTED_LIST[@]} 个受保护的基础设施文件已被忽略（未覆盖）：" >&2
  for p in "${PROTECTED_LIST[@]}"; do echo "  ⚠   - $p" >&2; done
  echo "⚠  这通常说明 zip 可能被伪造或被改动过；请确认发布来源可信。" >&2
  echo "⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠⚠" >&2
fi

if ! grep -q 'extra_hosts' "$COMPOSE_FILE" \
  || ! grep -q 'authserver.smbu.edu.cn' "$COMPOSE_FILE"; then
  echo "警告: 更新后的 deploy/docker-compose.yml 没有 authserver extra_hosts。" >&2
  echo "  CAS callback 可能再解析到 IPv6 并 502。请换含本仓库 compose 的 zip，或手工加回 extra_hosts。" >&2
fi

cat <<EOF

代码已更新。Docker named volumes 未被本脚本触碰。
本次跳过的受保护基础设施文件数：${#PROTECTED_LIST[@]}（详见上方警告）。
下一步（按需）：
  starbyte rebuild all
  # 或 docker compose -f deploy/docker-compose.yml up -d --build --no-deps --force-recreate backend frontend

切勿: docker compose down -v   （会删 postgres/redis/minio 数据）
EOF
