# Changelog

## 2026-05-02

- Initial workspace created


## 2026-05-02

Step 2: Removed Logstash from glazed/pkg/cmds/logging/ (commit 08b8905)

### Related Files

- /home/manuel/workspaces/2026-05-02/remove-logstash-glazed/glazed/pkg/cmds/logging/section.go — Removed Logstash fields and flags


## 2026-05-02

Step 2 (cont): Related additional changed files

### Related Files

- /home/manuel/workspaces/2026-05-02/remove-logstash-glazed/glazed/pkg/cmds/logging/init-early.go — Removed Logstash early arg parsing


## 2026-05-02

Step 2 (cont): Related additional changed files

### Related Files

- /home/manuel/workspaces/2026-05-02/remove-logstash-glazed/glazed/pkg/cmds/logging/logstash_writer.go — Deleted entire LogstashWriter implementation


## 2026-05-02

Step 2 (cont): Related additional changed files

### Related Files

- /home/manuel/workspaces/2026-05-02/remove-logstash-glazed/glazed/pkg/cmds/logging/init.go — Removed Logstash initialization and cobra flag reads


## 2026-05-02

Step 3: Removed Logstash references from glazed docs (commit dc9baab)

### Related Files

- /home/manuel/workspaces/2026-05-02/remove-logstash-glazed/glazed/pkg/doc/topics/logging-section.md — Removed Logstash flags


## 2026-05-02

Step 4: Removed clay Logstash example and references (commit 0dee348)

### Related Files

- /home/manuel/workspaces/2026-05-02/remove-logstash-glazed/clay/examples/simple/logging_layer_example.go — Removed Logstash references from description


## 2026-05-02

Step 5: Committed docmgr ticket docs (commit 18eef8b)

### Related Files

- /home/manuel/workspaces/2026-05-02/remove-logstash-glazed/glazed/ttmp/2026/05/02/GL-003--remove-logstash-logging-support-from-glazed/reference/01-diary.md — Final diary with all steps recorded


## 2026-05-02

Step 6: Fixed golangci-lint install in glazed and clay Makefiles to download prebuilt binary from .golangci-lint-version (commits 2bcbfca, c048bba)

### Related Files

- /home/manuel/workspaces/2026-05-02/remove-logstash-glazed/clay/.gitignore — Added .bin/ to ignore downloaded linter binary
- /home/manuel/workspaces/2026-05-02/remove-logstash-glazed/clay/Makefile — Added golangci-lint-install target
- /home/manuel/workspaces/2026-05-02/remove-logstash-glazed/glazed/Makefile — Changed golangci-lint-install from go install to official install script

