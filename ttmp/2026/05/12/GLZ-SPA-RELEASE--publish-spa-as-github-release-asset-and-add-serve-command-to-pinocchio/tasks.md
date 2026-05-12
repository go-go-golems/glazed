# Tasks

## TODO

- [ ] Add tasks here

- [x] Phase 1: Modify glazed .goreleaser.yaml — add tar czf to before hooks, add release.extra_files
- [ ] Phase 1: Tag and release glazed, verify glazed-spa.tar.gz appears on GitHub Release
- [x] Phase 2: Create pinocchio pkg/spa/ package (embed.go, embed_none.go, spa.go)
- [x] Phase 2: Add make fetch-spa to pinocchio Makefile and .gitignore
- [x] Phase 2: Create cmd/pinocchio/cmds/serve.go and help_loader.go
- [x] Phase 2: Wire serve command into pinocchio main.go
- [x] Phase 2: Update pinocchio .goreleaser.yaml with fetch-spa and -tags embed
- [x] Phase 2: Test end-to-end: make fetch-spa && go build -tags embed && ./pinocchio serve
