#!/usr/bin/env python3
"""Inventory files mentioning layer/parameter and extract Go identifiers."""
from __future__ import annotations

import argparse
import re
from pathlib import Path

WORD_RE = re.compile(r"\b(?:parameter|parameters|layer|layers)\b", re.IGNORECASE)
IDENT_RE = re.compile(r"\b[A-Za-z_][A-Za-z0-9_]*\b")

EXCLUDE_DIRS = {".git", "node_modules", "vendor", "dist", "build", "ttmp", ".ttmp", "coverage", "bin", "tmp"}


def iter_files(root: Path) -> list[Path]:
    files = []
    for path in root.rglob("*"):
        if not path.is_file():
            continue
        if any(part in EXCLUDE_DIRS for part in path.parts):
            continue
        files.append(path)
    return files


def classify_file(path: Path) -> str:
    suffix = path.suffix.lower()
    if suffix in {".go"}:
        return "go"
    if suffix in {".md", ".markdown", ".txt", ".rst"}:
        return "doc"
    if suffix in {".yaml", ".yml", ".json", ".toml"}:
        return "data"
    return "other"


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", default=".")
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    root = Path(args.root).resolve()
    matches: dict[Path, dict[str, set[str] | list[str]]] = {}

    for path in iter_files(root):
        try:
            content = path.read_text(encoding="utf-8", errors="ignore")
        except Exception:
            continue
        if not WORD_RE.search(content):
            continue

        file_type = classify_file(path)
        entry: dict[str, set[str] | list[str]] = {
            "type": {file_type},
            "idents": set(),
            "snippets": [],
        }
        if file_type == "go":
            for ident in IDENT_RE.findall(content):
                if "Parameter" in ident or "Layer" in ident or "parameter" in ident or "layer" in ident:
                    entry["idents"].add(ident)
        else:
            # Capture a few matching lines for context
            snippets = []
            for line in content.splitlines():
                if WORD_RE.search(line):
                    snippets.append(line.strip())
                if len(snippets) >= 5:
                    break
            entry["snippets"] = snippets

        matches[path] = entry

    lines: list[str] = []
    lines.append("# Layer/Parameter Inventory")
    lines.append("")
    lines.append(f"Root: `{root}`")
    lines.append("")

    by_type: dict[str, list[Path]] = {"go": [], "doc": [], "data": [], "other": []}
    for path, meta in matches.items():
        file_type = next(iter(meta["type"]))
        by_type[file_type].append(path)

    total_files = sum(len(v) for v in by_type.values())
    lines.append(f"Total files with matches: **{total_files}**")
    for key in ["go", "doc", "data", "other"]:
        lines.append(f"- {key}: {len(by_type[key])}")

    for key in ["go", "doc", "data", "other"]:
        lines.append("")
        lines.append(f"## {key.upper()} files")
        for path in sorted(by_type[key]):
            rel = path.relative_to(root)
            lines.append("")
            lines.append(f"### `{rel}`")
            meta = matches[path]
            if key == "go":
                idents = sorted(meta["idents"])  # type: ignore[arg-type]
                if idents:
                    lines.append("Identifiers:")
                    lines.append("")
                    for ident in idents:
                        lines.append(f"- {ident}")
                else:
                    lines.append("Identifiers: (none found)")
            else:
                snippets = meta["snippets"]  # type: ignore[assignment]
                if snippets:
                    lines.append("Snippets:")
                    lines.append("")
                    for snippet in snippets:
                        lines.append(f"- {snippet}")
                else:
                    lines.append("Snippets: (none captured)")

    Path(args.output).write_text("\n".join(lines) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
