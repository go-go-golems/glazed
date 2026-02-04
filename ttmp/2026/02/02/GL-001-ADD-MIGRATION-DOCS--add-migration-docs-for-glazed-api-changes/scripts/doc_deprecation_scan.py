#!/usr/bin/env python3
"""Scan documentation files for legacy Glazed API names.

Outputs JSON with per-file matches and counts.
"""
from __future__ import annotations

import argparse
import json
import os
import re
from pathlib import Path

PATTERNS = [
    ("pkg/cmds/layers import", r"github\.com/go-go-golems/glazed/pkg/cmds/layers"),
    ("pkg/cmds/parameters import", r"github\.com/go-go-golems/glazed/pkg/cmds/parameters"),
    ("pkg/cmds/middlewares import", r"github\.com/go-go-golems/glazed/pkg/cmds/middlewares"),
    ("layers.ParameterLayer", r"layers\.ParameterLayer"),
    ("layers.ParameterLayers", r"layers\.ParameterLayers"),
    ("ParameterLayer", r"\bParameterLayer\b"),
    ("ParameterLayers", r"\bParameterLayers\b"),
    ("ParsedLayer", r"\bParsedLayer\b"),
    ("ParsedLayers", r"\bParsedLayers\b"),
    ("parameters.ParameterDefinition", r"parameters\.ParameterDefinition"),
    ("parameters.ParameterDefinitions", r"parameters\.ParameterDefinitions"),
    ("ParameterDefinition", r"\bParameterDefinition\b"),
    ("ParameterDefinitions", r"\bParameterDefinitions\b"),
    ("ParameterType", r"\bParameterType\w*\b"),
    ("AddFlags", r"\bAddFlags\b"),
    ("AddFlag", r"\bAddFlag\b"),
    ("CobraParameterLayer", r"\bCobraParameterLayer\b"),
    ("ExecuteMiddlewares", r"\bExecuteMiddlewares\b"),
    ("ParseFromCobraCommand", r"\bParseFromCobraCommand\b"),
    ("GatherArguments", r"\bGatherArguments\b"),
    ("UpdateFromEnv", r"\bUpdateFromEnv\b"),
    ("SetFromDefaults", r"\bSetFromDefaults\b"),
    ("LoadParametersFromFile", r"\bLoadParametersFromFile\b"),
    ("LoadParametersFromFiles", r"\bLoadParametersFromFiles\b"),
    ("WithParseStepSource", r"\bWithParseStepSource\b"),
    ("CommandDefinition", r"\bCommandDefinition\b"),
]

SCOPE_DIRS = [
    "glazed/pkg/doc",
    "prompto/glazed",
    "glazed/README.md",
]


def collect_files(root: Path) -> list[Path]:
    files: list[Path] = []
    for item in SCOPE_DIRS:
        p = root / item
        if p.is_file():
            files.append(p)
            continue
        if p.is_dir():
            files.extend(p.rglob("*.md"))
    return sorted(set(files))


def scan_file(path: Path) -> dict:
    rel = str(path)
    matches = []
    counts = {}
    try:
        text = path.read_text(encoding="utf-8", errors="replace")
    except Exception:
        return {
            "path": rel,
            "error": "read-failed",
            "matches": [],
            "counts": {},
        }

    for label, pattern in PATTERNS:
        regex = re.compile(pattern)
        counts[label] = 0
        for i, line in enumerate(text.splitlines(), 1):
            if regex.search(line):
                counts[label] += 1
                matches.append({
                    "label": label,
                    "pattern": pattern,
                    "line": i,
                    "text": line.strip(),
                })
    counts = {k: v for k, v in counts.items() if v > 0}
    return {
        "path": rel,
        "matches": matches,
        "counts": counts,
    }


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--root", required=True, help="Repo root")
    parser.add_argument("--out", required=True, help="Output JSON path")
    args = parser.parse_args()

    root = Path(os.path.abspath(args.root))
    out_path = Path(os.path.abspath(args.out))
    out_path.parent.mkdir(parents=True, exist_ok=True)

    files = collect_files(root)
    results = [scan_file(p) for p in files]

    summary = {
        "root": str(root),
        "file_count": len(files),
        "patterns": [label for label, _ in PATTERNS],
        "results": results,
    }
    out_path.write_text(json.dumps(summary, indent=2), encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
