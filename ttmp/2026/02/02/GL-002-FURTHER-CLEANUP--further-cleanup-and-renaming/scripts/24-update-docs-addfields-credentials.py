#!/usr/bin/env python3
from __future__ import annotations

from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[6]

ADD_FIELDS_TARGETS = [
    REPO_ROOT / "pkg/doc/topics/sections-guide.md",
    REPO_ROOT / "pkg/doc/tutorials/migrating-to-facade-packages.md",
]

CREDENTIALS_TARGET = REPO_ROOT / "pkg/doc/topics/16-adding-field-types.md"


def replace_in_file(path: Path, replacements: list[tuple[str, str]]) -> bool:
    text = path.read_text(encoding="utf-8")
    updated = text
    for old, new in replacements:
        updated = updated.replace(old, new)
    if updated != text:
        path.write_text(updated, encoding="utf-8")
        return True
    return False


def main() -> int:
    changed = False
    for target in ADD_FIELDS_TARGETS:
        changed = replace_in_file(target, [("AddFlags", "AddFields")]) or changed

    changed = replace_in_file(
        CREDENTIALS_TARGET,
        [
            ("credentials-param", "credentials-field"),
            ("CredentialsParam", "CredentialsField"),
            ("paramDef", "fieldDef"),
        ],
    ) or changed

    return 0 if changed else 0


if __name__ == "__main__":
    raise SystemExit(main())
