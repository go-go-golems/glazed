#!/usr/bin/env python3
from __future__ import annotations

import subprocess
from pathlib import Path
from typing import Iterable


SCRIPT_PATH = Path(__file__).resolve()
TICKET_ROOT = SCRIPT_PATH.parents[1]
REPO_ROOT = TICKET_ROOT.parents[4]
POSTMORTEM = TICKET_ROOT / "analysis/04-postmortem-gl-002-refactor-and-tooling.md"
RENAME_MAP = TICKET_ROOT / "scripts/12-rename-symbols.yaml"
SCRIPTS_DIR = TICKET_ROOT / "scripts"


def run(cmd: list[str], cwd: Path) -> str:
    result = subprocess.run(cmd, cwd=cwd, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    return result.stdout.strip()


def get_upstream(cwd: Path) -> str:
    try:
        return run(["git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"], cwd)
    except subprocess.CalledProcessError:
        return "origin/main"


def format_list(lines: Iterable[str]) -> str:
    return "\n".join(f"- {line}" for line in lines if line)


def parse_rename_map(path: Path) -> list[str]:
    if not path.exists():
        return []
    mappings: list[str] = []
    current_pkg = None
    current_old = None
    current_new = None
    for raw in path.read_text(encoding="utf-8").splitlines():
        stripped = raw.strip()
        if not stripped or stripped.startswith("#"):
            continue
        if stripped.startswith("- "):
            # reset for new mapping
            current_pkg = None
            current_old = None
            current_new = None
            stripped = stripped[2:]
        if stripped.startswith("pkg:"):
            current_pkg = stripped.split(":", 1)[1].strip()
        elif stripped.startswith("old:"):
            current_old = stripped.split(":", 1)[1].strip()
        elif stripped.startswith("new:"):
            current_new = stripped.split(":", 1)[1].strip()
        if current_pkg and current_old and current_new:
            mappings.append(f"{current_pkg}: {current_old} -> {current_new}")
            current_pkg = None
            current_old = None
            current_new = None
    return mappings


def list_scripts(path: Path) -> list[str]:
    if not path.exists():
        return []
    items = sorted(p.name for p in path.iterdir() if p.is_file())
    return items


def main() -> int:
    upstream = get_upstream(REPO_ROOT)
    commit_lines = run(["git", "log", "--reverse", f"{upstream}..HEAD", "--format=%h %s"], REPO_ROOT).splitlines()

    diff_lines = run(["git", "diff", "--name-status", "--find-renames", f"{upstream}..HEAD"], REPO_ROOT).splitlines()
    rename_lines = [line for line in diff_lines if line.startswith("R")]

    rename_map = parse_rename_map(RENAME_MAP)
    script_list = list_scripts(SCRIPTS_DIR)

    appendix = []
    appendix.append("## Appendices")
    appendix.append("")

    appendix.append("### Appendix A: Commit list (this branch vs upstream)")
    appendix.append("")
    appendix.append(format_list(commit_lines) or "- (no commits found)")
    appendix.append("")

    appendix.append("### Appendix B: File renames (git diff --name-status --find-renames)")
    appendix.append("")
    if rename_lines:
        appendix.append("- Format: R<score> <old> <new>")
        appendix.append(format_list(rename_lines))
    else:
        appendix.append("- (no renames detected)")
    appendix.append("")

    appendix.append("### Appendix C: Symbol rename map (AST tool YAML)")
    appendix.append("")
    if rename_map:
        appendix.append(format_list(rename_map))
    else:
        appendix.append("- (rename map not found)")
    appendix.append("")

    appendix.append("### Appendix D: Ticket scripts inventory")
    appendix.append("")
    if script_list:
        appendix.append(format_list(script_list))
    else:
        appendix.append("- (no scripts found)")
    appendix.append("")

    appendix.append("### Appendix E: Validation commands (as executed)")
    appendix.append("")
    appendix.append("- `rg -n -i \"layer|parameter\" glazed -g '!**/ttmp/**'`")
    appendix.append("- `go test ./...`")
    appendix.append("- `golangci-lint run -v --max-same-issues=100`")
    appendix.append("- `gosec -exclude=G101,G304,G301,G306,G204 -exclude-dir=ttmp -exclude-dir=.history ./...`")
    appendix.append("- `govulncheck ./...`")
    appendix.append("")

    appendix_text = "\n".join(appendix).rstrip() + "\n"

    content = POSTMORTEM.read_text(encoding="utf-8")
    marker = "\n## Appendices\n"
    if marker in content:
        pre = content.split(marker, 1)[0].rstrip() + "\n\n"
        new_content = pre + appendix_text
    else:
        new_content = content.rstrip() + "\n\n" + appendix_text

    POSTMORTEM.write_text(new_content, encoding="utf-8")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
