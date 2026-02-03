#!/usr/bin/env python3
import re
import pathlib

ROOT = pathlib.Path(".")
TARGET_GLOBS = ["**/*.md"]
EXCLUDE_DIRS = {"ttmp", ".git", "node_modules", "vendor"}

pattern = re.compile(r"\bInitializeStruct(?!From)")


def should_skip(path: pathlib.Path) -> bool:
    return bool(set(path.parts) & EXCLUDE_DIRS)


def main() -> None:
    for pattern_glob in TARGET_GLOBS:
        for path in ROOT.glob(pattern_glob):
            if path.is_dir() or should_skip(path):
                continue
            try:
                data = path.read_text()
            except Exception:
                continue
            new_data = pattern.sub("DecodeSectionInto", data)
            if new_data != data:
                path.write_text(new_data)


if __name__ == "__main__":
    main()
