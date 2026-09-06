#!/usr/bin/env python3
"""
从 worksheetContext.json 生成确定性的填值模板 (fillTemplates.json)。

用法:
  python3 generate_fill_templates.py <worksheetContext.json路径> [输出路径]

输出:
  fillTemplates.json — 每张工作表的可填字段清单，包含：
    - fieldKey: 直接用于 batch_create_records 的 fields[].id（已从 alias/id 中确定性提取）
    - name: 字段中文名（供 LLM 理解语义）
    - type: 字段类型
    - validOptions: 选项字段的合法选项值列表（仅限选项类型字段）
    - dataSource: 关联字段的目标工作表 ID（仅限 Relation 类型）
    - isTitle: 标题字段标记
    - isSelfRelation: 自关联标记
"""

import json
import sys
from pathlib import Path

# 不可填值的字段类型 — 这些字段在 batch_create_records 中不传
SKIP_TYPES = {"Divider", "Formula", "AutoNumber", "AutoID", "Lookup"}

# 系统字段的固定 ID — 所有表都有，不应由用户写入
SYSTEM_FIELD_IDS = {"rowid", "ownerid", "caid", "ctime", "utime", "uaid"}


def generate_fill_templates(worksheet_context):
    """从 worksheetContext 生成填值模板列表。"""

    # 全局 worksheetId → worksheetName 映射（用于关联依赖的人类可读翻译）
    id_to_name = {ws["worksheetId"]: ws["worksheetName"] for ws in worksheet_context}

    templates = []

    for ws in worksheet_context:
        worksheet_id = ws["worksheetId"]
        worksheet_name = ws["worksheetName"]
        fillable_fields = []
        relation_deps = []
        title_field_found = False

        for field in ws.get("fields", []):
            field_id = field.get("id", "")
            field_alias = field.get("alias", "")
            field_name = field.get("name", "")
            field_type = field.get("type", "")
            data_source = field.get("dataSource", "")
            source_field = field.get("sourceField", "")
            options = field.get("options", [])

            # ── 跳过系统字段 ──
            if field_id in SYSTEM_FIELD_IDS:
                continue
            if field_alias.startswith("_"):
                continue

            # ── 跳过不可填值的类型 ──
            if field_type in SKIP_TYPES:
                continue

            # ── 跳过反向关联字段（alias 为空的 Relation = 系统自动创建的反向字段，只读） ──
            # 注意：明道云的正向和反向关联都有 sourceField，不能用 sourceField 判断
            if field_type == "Relation" and not field_alias:
                continue

            # ── 确定 fieldKey（alias 优先，为空则用 id） ──
            field_key = field_alias if field_alias else field_id

            # ── 构建字段模板 ──
            field_template = {
                "fieldKey": field_key,
                "name": field_name,
                "type": field_type,
            }

            # ── 标记标题字段（启发式：第一个 Text 类型字段） ──
            if not title_field_found and field_type == "Text":
                field_template["isTitle"] = True
                title_field_found = True

            # ── 提取有效选项（选项类型字段） ──
            if options and field_type in ("SingleSelect", "MultipleSelect", "Dropdown"):
                valid_options = [
                    opt["value"]
                    for opt in options
                    if not opt.get("isDelete", False)
                ]
                field_template["validOptions"] = valid_options

            # ── 提取关联信息 ──
            if field_type == "Relation" and data_source:
                field_template["dataSource"] = data_source
                if data_source == worksheet_id:
                    field_template["isSelfRelation"] = True
                else:
                    relation_deps.append(data_source)

            fillable_fields.append(field_template)

        # ── 将关联依赖的 ID 翻译为表名 ──
        relation_dep_names = sorted(set(
            id_to_name.get(dep_id, dep_id) for dep_id in relation_deps
        ))

        template = {
            "worksheetId": worksheet_id,
            "worksheetName": worksheet_name,
            "fieldCount": len(fillable_fields),
            "fillableFields": fillable_fields,
            "relationDeps": relation_dep_names,
        }

        templates.append(template)

    return templates


def main():
    if len(sys.argv) < 2:
        print(
            "用法: python3 generate_fill_templates.py <worksheetContext.json路径> [输出路径]",
            file=sys.stderr,
        )
        sys.exit(1)

    input_path = Path(sys.argv[1])
    if not input_path.exists():
        print(f"❌ 文件不存在: {input_path}", file=sys.stderr)
        sys.exit(1)

    output_path = (
        Path(sys.argv[2]) if len(sys.argv) > 2 else input_path.parent / "fillTemplates.json"
    )

    with open(input_path, "r", encoding="utf-8") as f:
        worksheet_context = json.load(f)

    templates = generate_fill_templates(worksheet_context)

    with open(output_path, "w", encoding="utf-8") as f:
        json.dump(templates, f, ensure_ascii=False, indent=2)

    # 输出摘要
    print(f"\n📋 填值模板生成完成（共 {len(templates)} 张表）：", file=sys.stderr)
    for t in templates:
        deps = f"，依赖: {', '.join(t['relationDeps'])}" if t["relationDeps"] else ""
        print(f"  • {t['worksheetName']}: {t['fieldCount']} 个可填字段{deps}", file=sys.stderr)
    print(f"\n✅ 输出: {output_path}", file=sys.stderr)


if __name__ == "__main__":
    main()
