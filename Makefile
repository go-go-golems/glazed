.PHONY: all test build lint lintmax docker-lint golangci-lint-install gosec govulncheck goreleaser tag-major tag-minor tag-patch release bump-glazed install version

all: test build

VERSION ?= $(shell svu current)
LDFLAGS ?= -X main.version=$(VERSION)
GORELEASER_ARGS ?= --skip=sign --snapshot --clean
GORELEASER_TARGET ?= --single-target
GOLANGCI_LINT_VERSION ?= $(shell cat .golangci-lint-version)
GOLANGCI_LINT_BIN ?= $(CURDIR)/.bin/golangci-lint

docker-lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run -v

golangci-lint-install:
	mkdir -p $(dir $(GOLANGCI_LINT_BIN))
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(dir $(GOLANGCI_LINT_BIN)) $(GOLANGCI_LINT_VERSION)

lint: golangci-lint-install
	$(GOLANGCI_LINT_BIN) run -v

lintmax: golangci-lint-install
	$(GOLANGCI_LINT_BIN) run -v --max-same-issues=100

gosec:
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec -exclude=G101,G304,G301,G306,G204 -exclude-dir=ttmp -exclude-dir=.history -exclude-dir=cmd/examples ./...

govulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

test:
	go test $(shell go list ./... | grep -v 'ttmp')

build:
	go generate ./...
	go build -tags "sqlite_fts5 embed" -ldflags "$(LDFLAGS)" ./cmd/glaze

goreleaser:
	GOWORK=off goreleaser release $(GORELEASER_ARGS) $(GORELEASER_TARGET)

tag-major:
	git tag $(shell svu major)

tag-minor:
	git tag $(shell svu minor)

tag-patch:
	git tag $(shell svu patch)

release:
	git push origin --tags
	GOPROXY=proxy.golang.org go list -m github.com/go-go-golems/glazed@$(shell svu current)

version:
	@echo $(VERSION)

bump-glazed:
	go get github.com/go-go-golems/glazed@latest
	go get github.com/go-go-golems/clay@latest
	go mod tidy

install:
	go generate ./... && \
		go build -tags "sqlite_fts5 embed" -ldflags "$(LDFLAGS)" -o ./dist/glaze ./cmd/glaze && \
		cp ./dist/glaze $(shell which glaze)
