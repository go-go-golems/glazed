#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path

TICKET_ROOT = Path(__file__).resolve().parents[1]
OUTPUT_ROOT = Path("/tmp/remarkable-gl-002")

INPUTS = [
    TICKET_ROOT / "README.md",
    TICKET_ROOT / "design-doc/01-further-cleanup-and-renaming-plan.md",
    TICKET_ROOT / "analysis/01-exhaustive-parameter-layer-audit.md",
    TICKET_ROOT / "analysis/02-parameter-layer-symbol-inventory.md",
    TICKET_ROOT / "analysis/03-layer-parameter-inventory.md",
    TICKET_ROOT / "sources/01-glazed-cleanup-notes.md",
    TICKET_ROOT / "reference/01-diary.md",
]

REPLACEMENTS = [
    ("\\n", "\\\\n"),
    ("\\t", "\\\\t"),
    ("“", "\""),
    ("”", "\""),
    ("‘", "'"),
    ("’", "'"),
]


def sanitize(text: str) -> str:
    updated = text
    for old, new in REPLACEMENTS:
        updated = updated.replace(old, new)
    return updated


def main() -> int:
    OUTPUT_ROOT.mkdir(parents=True, exist_ok=True)
    for path in INPUTS:
        rel = path.relative_to(TICKET_ROOT)
        out_path = OUTPUT_ROOT / rel
        out_path.parent.mkdir(parents=True, exist_ok=True)
        out_path.write_text(sanitize(path.read_text(encoding="utf-8")), encoding="utf-8")
        print(out_path)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
