.PHONY: gifs

VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse --short HEAD)
DIRTY ?= $(shell git diff --quiet || echo "dirty")

LDFLAGS=-ldflags "-X main.version=$(VERSION)-$(COMMIT)-$(DIRTY)"

all: gifs

TAPES=$(shell ls doc/vhs/*tape)
gifs: $(TAPES)
	for i in $(TAPES); do vhs < $$i; done

docker-lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.50.1 golangci-lint run -v

lint:
	golangci-lint run -v --enable=exhaustive

lintmax:
	golangci-lint run -v --enable=exhaustive --max-same-issues=100

test:
	go test ./...

build:
	go generate ./...
	go build $(LDFLAGS) ./...

bench:
	go test -bench=./... -benchmem

goreleaser:
	goreleaser release --skip=sign --snapshot --clean

tag-major:
	git tag $(shell svu major)

tag-minor:
	git tag $(shell svu minor)

tag-patch:
	git tag $(shell svu patch)

release:
	git push --tags
	GOPROXY=proxy.golang.org go list -m github.com/go-go-golems/glazed@$(shell svu current)

exhaustive:
	golangci-lint run -v --enable=exhaustive

GLAZE_BINARY=$(shell which glaze)

install:
	go build $(LDFLAGS) -o ./dist/glaze ./cmd/glaze && \
		cp ./dist/glaze $(GLAZE_BINARY)
