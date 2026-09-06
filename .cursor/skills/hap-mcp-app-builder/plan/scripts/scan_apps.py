#!/usr/bin/env python3
"""
scan_apps.py — 扫描 apps/ 目录，输出每个应用的名称和搭建状态。
               同时检查 GitHub 远程版本是否有更新（2 秒超时，失败静默跳过）。

用法：python scan_apps.py <projectRoot>
示例：python scan_apps.py /Users/user/应用搭建测试
输出：JSON 对象 { apps: [...], update?: { available, local, remote, notes } }
"""

import json
import os
import re
import sys
import urllib.request

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))

# 优先从命令行参数获取项目根目录，否则 fallback 到脚本目录向上 4 级
project_root = (
    os.path.abspath(sys.argv[1])
    if len(sys.argv) > 1
    else os.path.abspath(os.path.join(SCRIPT_DIR, "..", "..", "..", ".."))
)
apps_dir = os.path.join(project_root, "apps")


def find_version_file(start_dir):
    """向上逐级搜索版本文件（兼容完整仓库和仅 skills/ 目录两种安装方式）
    优先查找 plugin.json，其次 version.json"""
    d = start_dir
    for _ in range(6):
        for name in ("plugin.json", "version.json"):
            candidate = os.path.join(d, name)
            if os.path.isfile(candidate):
                return candidate
        parent = os.path.dirname(d)
        if parent == d:
            break
        d = parent
    return None


def scan_apps():
    """扫描 apps/ 目录下的应用"""
    results = []
    if not os.path.isdir(apps_dir):
        return results

    for entry in sorted(os.listdir(apps_dir)):
        entry_path = os.path.join(apps_dir, entry)
        if not os.path.isdir(entry_path):
            continue

        plan_path = os.path.join(entry_path, "hap-plan.json")
        if not os.path.isfile(plan_path):
            continue

        # 读取应用名
        app_name = entry
        try:
            with open(plan_path, "r", encoding="utf-8") as f:
                plan = json.load(f)
            if plan.get("appName"):
                app_name = plan["appName"]
        except Exception:
            pass

        # 读取搭建进度
        ctx_path = os.path.join(entry_path, "hap-context.json")
        progress = None
        status = "planned"

        if os.path.isfile(ctx_path):
            try:
                with open(ctx_path, "r", encoding="utf-8") as f:
                    ctx = json.load(f)
                progress = ctx.get("progress")
                status = "completed" if progress == "completed" else "in_progress"
            except Exception:
                pass

        results.append({"dir": entry, "appName": app_name, "progress": progress, "status": status})

    return results


def find_git_root(start_dir):
    """向上查找 .git 目录，返回仓库根目录路径，找不到返回 None"""
    d = start_dir
    for _ in range(10):
        if os.path.isdir(os.path.join(d, ".git")):
            return d
        parent = os.path.dirname(d)
        if parent == d:
            break
        d = parent
    return None


def check_update():
    """检查 GitHub 远程版本（2 秒超时，失败静默跳过）"""
    plugin_path = find_version_file(SCRIPT_DIR)
    if not plugin_path:
        return None

    try:
        with open(plugin_path, "r", encoding="utf-8") as f:
            pkg = json.load(f)
        local_version = pkg.get("version")
        repo_url = pkg.get("repository")
        branch = pkg.get("branch", "main")
        skill_path = pkg.get("skillPath", "")
    except Exception:
        return None

    if not local_version or not repo_url:
        return None

    match = re.search(r"github\.com/([^/]+/[^/]+)", repo_url)
    if not match:
        return None

    # 计算版本文件在仓库中的相对路径
    git_root = find_git_root(os.path.dirname(plugin_path))
    if git_root:
        remote_file = os.path.relpath(plugin_path, git_root)
    else:
        # 非 git 安装方式，用 skillPath 拼接路径
        if skill_path:
            remote_file = f"{skill_path}/{os.path.basename(plugin_path)}"
        else:
            remote_file = os.path.basename(plugin_path)

    raw_url = f"https://raw.githubusercontent.com/{match.group(1)}/{branch}/{remote_file}"

    try:
        with urllib.request.urlopen(raw_url, timeout=2) as resp:
            remote = json.loads(resp.read())
        if remote.get("version") and remote["version"] != local_version:
            return {
                "available": True,
                "local": local_version,
                "remote": remote["version"],
                "notes": remote.get("releaseNotes"),
                "repository": repo_url,
                "branch": branch,
                "skillPath": skill_path,
            }
    except Exception:
        pass

    return None


if __name__ == "__main__":
    output = {"apps": scan_apps()}
    update = check_update()
    if update:
        output["update"] = update
    print(json.dumps(output, ensure_ascii=False, indent=2))
