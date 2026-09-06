#!/usr/bin/env python3
"""
Step 3 全脚本方案：直接调用明道云 REST API 获取工作表结构并生成 worksheetContext.json。

用法:
  python3 refresh_fields.py --token "md_pss_id xxx" ./apps/图书借阅/hap-context.json

输入:
  --token     明道云认证 token（Authorization header 值）
  context     hap-context.json 的路径（从中读取 appId 和 worksheetIdByName）

输出:
  与 hap-context.json 同目录下的 worksheetContext.json
"""

import argparse
import json
import sys
from pathlib import Path
from urllib.request import Request, urlopen
from urllib.error import HTTPError, URLError

API_BASE = "https://api2.mingdao.com"


def fetch_worksheet_structure(worksheet_id, token, app_id):
    """调用 REST API 获取单张工作表结构。"""
    url = f"{API_BASE}/v3/app/worksheets/{worksheet_id}"
    headers = {
        "Authorization": token,
        "HAP-Appid": app_id,
    }
    req = Request(url, headers=headers, method="GET")
    try:
        with urlopen(req, timeout=30) as resp:
            return json.loads(resp.read().decode("utf-8"))
    except HTTPError as e:
        body = e.read().decode("utf-8", errors="replace")
        print(f"  ❌ HTTP {e.code}: {body[:200]}", file=sys.stderr)
        return None
    except URLError as e:
        print(f"  ❌ 网络错误: {e.reason}", file=sys.stderr)
        return None


def extract_fields(raw_data):
    """从 API 响应中提取字段列表。"""
    if not isinstance(raw_data, dict):
        return []
    data = raw_data.get("data", raw_data)
    if isinstance(data, dict):
        return data.get("fields", data.get("controls", []))
    return []


def normalize_field(field):
    """标准化单个字段为 worksheetContext 格式。"""
    result = {
        "id": field.get("id", field.get("controlId", "")),
        "alias": field.get("alias", ""),
        "name": field.get("name", field.get("controlName", "")),
        "type": field.get("type", ""),
    }

    # 选项字段
    options = field.get("options", [])
    if options:
        result["options"] = [
            {"key": opt.get("key", opt.get("value", "")), "value": opt.get("value", "")}
            for opt in options
            if not opt.get("isDelete", False)
        ]

    # 关联字段
    data_source = field.get("dataSource", "")
    if data_source:
        result["dataSource"] = data_source

    # 来源字段（反向关联）
    source_field = field.get("sourceField", "")
    if source_field:
        result["sourceField"] = source_field

    return result


def main():
    parser = argparse.ArgumentParser(description="刷新工作表字段结构")
    parser.add_argument("--token", required=True, help="明道云认证 token（如 md_pss_id xxx）")
    parser.add_argument("context", help="hap-context.json 文件路径")
    args = parser.parse_args()

    context_path = Path(args.context)
    if not context_path.exists():
        print(f"❌ 文件不存在: {context_path}", file=sys.stderr)
        sys.exit(1)

    with open(context_path, "r", encoding="utf-8") as f:
        context = json.load(f)

    app_id = context.get("appId", "")
    worksheet_id_by_name = context.get("worksheetIdByName", {})

    if not app_id:
        print("❌ hap-context.json 中缺少 appId", file=sys.stderr)
        sys.exit(1)
    if not worksheet_id_by_name:
        print("❌ hap-context.json 中缺少 worksheetIdByName", file=sys.stderr)
        sys.exit(1)

    print(f"📋 开始刷新字段结构（共 {len(worksheet_id_by_name)} 张表）", file=sys.stderr)

    worksheet_context = []
    errors = []

    for ws_name, ws_id in worksheet_id_by_name.items():
        print(f"  ⏳ {ws_name} ({ws_id})...", file=sys.stderr, end="")
        raw = fetch_worksheet_structure(ws_id, args.token, app_id)
        if raw is None:
            errors.append(ws_name)
            continue

        fields = extract_fields(raw)
        normalized = [normalize_field(f) for f in fields]

        worksheet_context.append({
            "worksheetId": ws_id,
            "worksheetName": ws_name,
            "fields": normalized,
        })
        print(f" ✅ {len(normalized)} 个字段", file=sys.stderr)

    # 写入输出
    output_path = context_path.parent / "worksheetContext.json"
    with open(output_path, "w", encoding="utf-8") as f:
        json.dump(worksheet_context, f, ensure_ascii=False, indent=2)

    # 摘要
    print(f"\n{'=' * 40}", file=sys.stderr)
    print(f"✅ worksheetContext 构建完成（{len(worksheet_context)}/{len(worksheet_id_by_name)} 张表）", file=sys.stderr)
    for ws in worksheet_context:
        print(f"  • {ws['worksheetName']}: {len(ws['fields'])} 个字段", file=sys.stderr)

    if errors:
        print(f"\n⚠️ 失败 {len(errors)} 张: {', '.join(errors)}", file=sys.stderr)
        sys.exit(1)

    print(f"\n📁 输出: {output_path}", file=sys.stderr)


if __name__ == "__main__":
    main()
