#!/usr/bin/env python3
"""devctl plugin for the Glazed docs browser (with SSR sidecar).

Starts two supervised services:
  1. glazed.ssr-sidecar — Node.js Express SSR server (port 8089)
  2. glazed.docs-server — Go docs server with --ssr-url proxy (port 8088)

Protocol: NDJSON stdio v2.
"""

import json
import os
import shutil
import sys

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def emit(obj):
    sys.stdout.write(json.dumps(obj) + "\n")
    sys.stdout.flush()


def log(msg):
    sys.stderr.write(f"[glazed-plugin] {msg}\n")
    sys.stderr.flush()


def e_unsupported(rid, op):
    return {
        "type": "response",
        "request_id": rid,
        "ok": False,
        "error": {"code": "E_UNSUPPORTED", "message": f"unsupported op: {op}"},
    }


# ---------------------------------------------------------------------------
# Config
# ---------------------------------------------------------------------------

GO_PORT = 8088
SSR_PORT = 8089
DB_DIR = "/tmp/help-dbs"

# ---------------------------------------------------------------------------
# Handshake
# ---------------------------------------------------------------------------

emit({
    "type": "handshake",
    "protocol_version": "v2",
    "plugin_name": "glazed",
    "capabilities": {
        "ops": ["config.mutate", "validate.run", "launch.plan"],
    },
})

# ---------------------------------------------------------------------------
# Request loop
# ---------------------------------------------------------------------------

for line in sys.stdin:
    line = line.strip()
    if not line:
        continue

    req = json.loads(line)
    rid = req.get("request_id", "")
    op = req.get("op", "")
    ctx = req.get("ctx", {}) or {}
    inp = req.get("input", {}) or {}

    dry_run = bool(ctx.get("dry_run", False))
    repo_root = ctx.get("repo_root", "")

    try:
        # -------------------------------------------------------------------
        # config.mutate
        # -------------------------------------------------------------------
        if op == "config.mutate":
            emit({
                "type": "response",
                "request_id": rid,
                "ok": True,
                "output": {
                    "config_patch": {
                        "set": {
                            "services.docs-server.port": GO_PORT,
                            "services.docs-server.url": f"http://127.0.0.1:{GO_PORT}",
                            "services.ssr-sidecar.port": SSR_PORT,
                            "services.ssr-sidecar.url": f"http://127.0.0.1:{SSR_PORT}",
                            "env.SSR_PORT": str(SSR_PORT),
                            "env.API_BASE": f"http://127.0.0.1:{GO_PORT}/api",
                        },
                        "unset": [],
                    },
                },
            })

        # -------------------------------------------------------------------
        # validate.run
        # -------------------------------------------------------------------
        elif op == "validate.run":
            errors = []
            warnings = []

            # Check Go
            if not shutil.which("go"):
                errors.append({
                    "code": "E_MISSING_TOOL",
                    "message": "go is not installed or not on PATH",
                })

            # Check Node
            if not shutil.which("node"):
                errors.append({
                    "code": "E_MISSING_TOOL",
                    "message": "node is not installed or not on PATH",
                })

            # Check pnpm
            if not shutil.which("pnpm"):
                errors.append({
                    "code": "E_MISSING_TOOL",
                    "message": "pnpm is not installed or not on PATH",
                })

            # Check node_modules
            nm = os.path.join(repo_root, "web", "node_modules") if repo_root else ""
            if nm and not os.path.isdir(nm):
                warnings.append({
                    "code": "W_MISSING_DIR",
                    "message": "web/node_modules missing — run 'cd web && pnpm install' before 'devctl up'",
                })

            # Check test DB dir
            if not os.path.isdir(DB_DIR):
                warnings.append({
                    "code": "W_MISSING_DIR",
                    "message": f"{DB_DIR} missing — the docs server will have no data. "
                               f"Create test DBs or point to an existing directory.",
                })

            emit({
                "type": "response",
                "request_id": rid,
                "ok": True,
                "output": {
                    "valid": len(errors) == 0,
                    "errors": errors,
                    "warnings": warnings,
                },
            })

        # -------------------------------------------------------------------
        # launch.plan
        # -------------------------------------------------------------------
        elif op == "launch.plan":
            if dry_run:
                log("dry-run: computing plan without side effects")

            go_bin = os.path.join(repo_root, ".bin", "glaze") if repo_root else "glaze"

            emit({
                "type": "response",
                "request_id": rid,
                "ok": True,
                "output": {
                    "services": [
                        {
                            "name": "glazed.ssr-sidecar",
                            "cwd": "web",
                            "command": [
                                "bash", "--noprofile", "--norc", "-lc",
                                "mkdir -p dist && "
                                f"if [ ! -f dist/ssr/entry-server.js ]; then "
                                f"  pnpm build:all; "
                                f"fi && "
                                f"exec node server.mjs",
                            ],
                            "env": {
                                "SSR_PORT": str(SSR_PORT),
                                "API_BASE": f"http://127.0.0.1:{GO_PORT}/api",
                            },
                            "health": {
                                "type": "http",
                                "url": f"http://127.0.0.1:{SSR_PORT}/health",
                                "timeout_ms": 30000,
                            },
                        },
                        {
                            "name": "glazed.docs-server",
                            "command": [
                                "bash", "--noprofile", "--norc", "-lc",
                                f"go build -o {go_bin} ./cmd/glaze && "
                                f"exec {go_bin} serve "
                                f"  --from-sqlite-dir {DB_DIR} "
                                f"  --address :{GO_PORT} "
                                f"  --ssr-url http://127.0.0.1:{SSR_PORT}",
                            ],
                            "env": {},
                            "health": {
                                "type": "http",
                                "url": f"http://127.0.0.1:{GO_PORT}/api/packages",
                                "timeout_ms": 30000,
                            },
                        },
                    ],
                },
            })

        # -------------------------------------------------------------------
        # Unknown op
        # -------------------------------------------------------------------
        else:
            emit(e_unsupported(rid, op))

    except Exception as exc:
        log(f"ERROR handling {op}: {exc}")
        emit({
            "type": "response",
            "request_id": rid,
            "ok": False,
            "error": {"code": "E_INTERNAL", "message": str(exc)},
        })
