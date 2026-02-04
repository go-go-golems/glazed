#!/usr/bin/env python3
"""Import git diffs into sqlite for querying and basic API-change analysis.

Usage:
  python3 import_git_diff_to_sqlite.py \
    --repo /abs/path/to/glazed \
    --base origin/main \
    --db /abs/path/to/db.sqlite \
    --summary-json /abs/path/to/summary.json
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import os
import re
import sqlite3
import subprocess
import sys
from pathlib import Path


def run(cmd: list[str], cwd: str | None = None) -> str:
    result = subprocess.run(cmd, cwd=cwd, text=True, capture_output=True, check=True)
    return result.stdout


def git(repo: str, *args: str) -> str:
    return run(["git", "-C", repo, *args])


def parse_name_status(output: str) -> dict[str, str]:
    status_by_path: dict[str, str] = {}
    for line in output.strip().splitlines():
        if not line.strip():
            continue
        parts = line.split("\t", 1)
        if len(parts) != 2:
            continue
        status, path = parts
        status_by_path[path] = status
    return status_by_path


def parse_numstat(output: str) -> dict[str, tuple[int, int]]:
    stats: dict[str, tuple[int, int]] = {}
    for line in output.strip().splitlines():
        if not line.strip():
            continue
        parts = line.split("\t")
        if len(parts) != 3:
            continue
        add_s, del_s, path = parts
        try:
            additions = int(add_s) if add_s != "-" else 0
            deletions = int(del_s) if del_s != "-" else 0
        except ValueError:
            additions, deletions = 0, 0
        stats[path] = (additions, deletions)
    return stats


def strip_comments(src: str) -> str:
    out: list[str] = []
    i = 0
    n = len(src)
    state = "code"
    while i < n:
        ch = src[i]
        if state == "code":
            if ch == '"':
                state = "dquote"
                out.append(ch)
                i += 1
                continue
            if ch == "'":
                state = "squote"
                out.append(ch)
                i += 1
                continue
            if ch == "`":
                state = "raw"
                out.append(ch)
                i += 1
                continue
            if ch == "/" and i + 1 < n:
                nxt = src[i + 1]
                if nxt == "/":
                    state = "line_comment"
                    out.append(" ")
                    i += 2
                    continue
                if nxt == "*":
                    state = "block_comment"
                    out.append(" ")
                    i += 2
                    continue
            out.append(ch)
            i += 1
            continue
        if state == "line_comment":
            if ch == "\n":
                out.append("\n")
                state = "code"
            i += 1
            continue
        if state == "block_comment":
            if ch == "*" and i + 1 < n and src[i + 1] == "/":
                out.append(" ")
                i += 2
                state = "code"
                continue
            if ch == "\n":
                out.append("\n")
            i += 1
            continue
        if state == "dquote":
            if ch == "\\" and i + 1 < n:
                out.append(ch)
                out.append(src[i + 1])
                i += 2
                continue
            out.append(ch)
            if ch == '"':
                state = "code"
            i += 1
            continue
        if state == "squote":
            if ch == "\\" and i + 1 < n:
                out.append(ch)
                out.append(src[i + 1])
                i += 2
                continue
            out.append(ch)
            if ch == "'":
                state = "code"
            i += 1
            continue
        if state == "raw":
            out.append(ch)
            if ch == "`":
                state = "code"
            i += 1
            continue
    return "".join(out)


def parse_names_list(segment: str) -> list[str]:
    before_eq = segment.split("=", 1)[0]
    names_part = before_eq.strip()
    names = []
    for name in names_part.split(","):
        name = name.strip()
        if not name:
            continue
        if re.match(r"^[A-Z][A-Za-z0-9_]*$", name):
            names.append(name)
    return names


def extract_go_symbols(src: str) -> tuple[str, list[dict]]:
    stripped = strip_comments(src)
    package = ""
    pkg_match = re.search(r"^\s*package\s+([A-Za-z0-9_]+)", stripped, re.M)
    if pkg_match:
        package = pkg_match.group(1)

    symbols: list[dict] = []
    const_var_block: str | None = None
    type_block = False

    lines = stripped.splitlines()
    for idx, line in enumerate(lines, 1):
        if type_block:
            if re.match(r"^\s*\)", line):
                type_block = False
                continue
            m = re.match(r"^\s*([A-Z][A-Za-z0-9_]*)\b", line)
            if m:
                name = m.group(1)
                symbols.append({
                    "name": name,
                    "kind": "type",
                    "receiver": "",
                    "line": idx,
                    "signature": f"type {name}",
                })
            continue

        if const_var_block:
            if re.match(r"^\s*\)", line):
                const_var_block = None
                continue
            names = parse_names_list(line)
            for name in names:
                symbols.append({
                    "name": name,
                    "kind": const_var_block,
                    "receiver": "",
                    "line": idx,
                    "signature": f"{const_var_block} {name}",
                })
            continue

        block_match = re.match(r"^\s*(const|var|type)\s*\(\s*$", line)
        if block_match:
            kind = block_match.group(1)
            if kind == "type":
                type_block = True
            else:
                const_var_block = kind
            continue

        m = re.match(r"^\s*type\s+([A-Z][A-Za-z0-9_]*)\b", line)
        if m:
            name = m.group(1)
            symbols.append({
                "name": name,
                "kind": "type",
                "receiver": "",
                "line": idx,
                "signature": f"type {name}",
            })
            continue

        m = re.match(r"^\s*(const|var)\s+(.+)$", line)
        if m:
            kind = m.group(1)
            names = parse_names_list(m.group(2))
            for name in names:
                symbols.append({
                    "name": name,
                    "kind": kind,
                    "receiver": "",
                    "line": idx,
                    "signature": f"{kind} {name}",
                })
            continue

        m = re.match(r"^\s*func\s+\(([^)]*)\)\s+([A-Z][A-Za-z0-9_]*)\s*\(", line)
        if m:
            receiver_raw = m.group(1).strip()
            receiver_type = receiver_raw.split()[-1] if receiver_raw else ""
            name = m.group(2)
            symbols.append({
                "name": name,
                "kind": "method",
                "receiver": receiver_type,
                "line": idx,
                "signature": f"func ({receiver_raw}) {name}",
            })
            continue

        m = re.match(r"^\s*func\s+([A-Z][A-Za-z0-9_]*)\s*\(", line)
        if m:
            name = m.group(1)
            symbols.append({
                "name": name,
                "kind": "func",
                "receiver": "",
                "line": idx,
                "signature": f"func {name}",
            })
            continue

    return package, symbols


def create_schema(conn: sqlite3.Connection) -> None:
    conn.executescript(
        """
        CREATE TABLE IF NOT EXISTS diff_meta (
            key TEXT PRIMARY KEY,
            value TEXT
        );

        CREATE TABLE IF NOT EXISTS diff_files (
            path TEXT PRIMARY KEY,
            status TEXT,
            additions INTEGER,
            deletions INTEGER,
            diff_text TEXT
        );

        CREATE TABLE IF NOT EXISTS diff_hunks (
            path TEXT,
            hunk_index INTEGER,
            old_start INTEGER,
            old_lines INTEGER,
            new_start INTEGER,
            new_lines INTEGER,
            header TEXT,
            hunk_text TEXT,
            PRIMARY KEY (path, hunk_index)
        );

        CREATE TABLE IF NOT EXISTS symbols (
            version TEXT,
            path TEXT,
            package TEXT,
            name TEXT,
            kind TEXT,
            receiver TEXT,
            line INTEGER,
            signature TEXT
        );

        CREATE TABLE IF NOT EXISTS symbol_changes (
            change_type TEXT,
            path TEXT,
            package TEXT,
            kind TEXT,
            receiver TEXT,
            name TEXT
        );
        """
    )


def parse_hunks(diff_text: str) -> list[dict]:
    hunks = []
    hunk_index = -1
    current_lines: list[str] = []
    current_header = ""
    old_start = old_lines = new_start = new_lines = 0

    for line in diff_text.splitlines():
        if line.startswith("@@"):
            if hunk_index >= 0:
                hunks.append({
                    "hunk_index": hunk_index,
                    "old_start": old_start,
                    "old_lines": old_lines,
                    "new_start": new_start,
                    "new_lines": new_lines,
                    "header": current_header,
                    "hunk_text": "\n".join(current_lines),
                })
            hunk_index += 1
            current_lines = [line]
            current_header = line
            m = re.match(r"@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@", line)
            if m:
                old_start = int(m.group(1))
                old_lines = int(m.group(2) or "1")
                new_start = int(m.group(3))
                new_lines = int(m.group(4) or "1")
            else:
                old_start = old_lines = new_start = new_lines = 0
            continue
        if hunk_index >= 0:
            current_lines.append(line)
    if hunk_index >= 0:
        hunks.append({
            "hunk_index": hunk_index,
            "old_start": old_start,
            "old_lines": old_lines,
            "new_start": new_start,
            "new_lines": new_lines,
            "header": current_header,
            "hunk_text": "\n".join(current_lines),
        })
    return hunks


def main() -> int:
    parser = argparse.ArgumentParser(description="Import git diffs into sqlite")
    parser.add_argument("--repo", required=True, help="Path to git repo root")
    parser.add_argument("--base", default="origin/main", help="Base ref to diff against")
    parser.add_argument("--db", required=True, help="Output sqlite db path")
    parser.add_argument("--summary-json", default="", help="Optional path to write summary JSON")
    args = parser.parse_args()

    repo = os.path.abspath(args.repo)
    db_path = os.path.abspath(args.db)
    os.makedirs(os.path.dirname(db_path), exist_ok=True)

    base_ref = args.base
    head_ref = git(repo, "rev-parse", "HEAD").strip()
    base_commit = git(repo, "rev-parse", base_ref).strip()

    name_status = parse_name_status(git(repo, "diff", "--name-status", f"{base_ref}...HEAD"))
    numstat = parse_numstat(git(repo, "diff", "--numstat", f"{base_ref}...HEAD"))

    conn = sqlite3.connect(db_path)
    create_schema(conn)

    conn.execute("DELETE FROM diff_meta")
    conn.execute("DELETE FROM diff_files")
    conn.execute("DELETE FROM diff_hunks")
    conn.execute("DELETE FROM symbols")
    conn.execute("DELETE FROM symbol_changes")

    meta = {
        "base_ref": base_ref,
        "base_commit": base_commit,
        "head_ref": "HEAD",
        "head_commit": head_ref,
        "generated_at": dt.datetime.utcnow().isoformat() + "Z",
        "repo": repo,
    }
    conn.executemany("INSERT INTO diff_meta (key, value) VALUES (?, ?)", meta.items())

    for path, status in sorted(name_status.items()):
        diff_text = git(repo, "diff", f"{base_ref}...HEAD", "--", path)
        additions, deletions = numstat.get(path, (0, 0))
        conn.execute(
            "INSERT INTO diff_files (path, status, additions, deletions, diff_text) VALUES (?, ?, ?, ?, ?)",
            (path, status, additions, deletions, diff_text),
        )
        for hunk in parse_hunks(diff_text):
            conn.execute(
                """
                INSERT INTO diff_hunks
                (path, hunk_index, old_start, old_lines, new_start, new_lines, header, hunk_text)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                """,
                (
                    path,
                    hunk["hunk_index"],
                    hunk["old_start"],
                    hunk["old_lines"],
                    hunk["new_start"],
                    hunk["new_lines"],
                    hunk["header"],
                    hunk["hunk_text"],
                ),
            )

        if not path.endswith(".go"):
            continue

        if status != "D":
            head_content = git(repo, "show", f"HEAD:{path}")
            pkg, symbols = extract_go_symbols(head_content)
            for sym in symbols:
                conn.execute(
                    """
                    INSERT INTO symbols
                    (version, path, package, name, kind, receiver, line, signature)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                    """,
                    ("head", path, pkg, sym["name"], sym["kind"], sym["receiver"], sym["line"], sym["signature"]),
                )

        if status != "A":
            try:
                base_content = git(repo, "show", f"{base_ref}:{path}")
            except subprocess.CalledProcessError:
                base_content = ""
            if base_content:
                pkg, symbols = extract_go_symbols(base_content)
                for sym in symbols:
                    conn.execute(
                        """
                        INSERT INTO symbols
                        (version, path, package, name, kind, receiver, line, signature)
                        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                        """,
                        ("base", path, pkg, sym["name"], sym["kind"], sym["receiver"], sym["line"], sym["signature"]),
                    )

    conn.commit()

    base_rows = conn.execute(
        "SELECT path, package, name, kind, receiver FROM symbols WHERE version='base'"
    ).fetchall()
    head_rows = conn.execute(
        "SELECT path, package, name, kind, receiver FROM symbols WHERE version='head'"
    ).fetchall()

    base_set = set(base_rows)
    head_set = set(head_rows)

    added = sorted(head_set - base_set)
    removed = sorted(base_set - head_set)

    conn.executemany(
        "INSERT INTO symbol_changes (change_type, path, package, kind, receiver, name) VALUES ('added', ?, ?, ?, ?, ?)",
        added,
    )
    conn.executemany(
        "INSERT INTO symbol_changes (change_type, path, package, kind, receiver, name) VALUES ('removed', ?, ?, ?, ?, ?)",
        removed,
    )
    conn.commit()

    if args.summary_json:
        summary = {
            "meta": meta,
            "counts": {
                "files": len(name_status),
                "symbols_added": len(added),
                "symbols_removed": len(removed),
            },
            "symbols_added": [
                {
                    "path": path,
                    "package": pkg,
                    "kind": kind,
                    "receiver": recv,
                    "name": name,
                }
                for (path, pkg, name, kind, recv) in added
            ],
            "symbols_removed": [
                {
                    "path": path,
                    "package": pkg,
                    "kind": kind,
                    "receiver": recv,
                    "name": name,
                }
                for (path, pkg, name, kind, recv) in removed
            ],
        }
        with open(args.summary_json, "w", encoding="utf-8") as f:
            json.dump(summary, f, indent=2)

    conn.close()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
