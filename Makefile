.PHONY: gifs

all: gifs

VERSION=v0.2.8

TAPES=$(shell ls doc/vhs/*tape)
gifs: $(TAPES)
	for i in $(TAPES); do vhs < $$i; done

docker-lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.50.1 golangci-lint run -v

lint:
	golangci-lint run -v --enable=exhaustive

test:
	go test ./...

build:
	go generate ./...
	go build

goreleaser:
	goreleaser release --snapshot --rm-dist

tag-release:
	git tag ${VERSION}

release:
	git push origin ${VERSION}
	GOPROXY=proxy.golang.org go list -m github.com/go-go-golems/glazed@${VERSION}

exhaustive:
	golangci-lint run -v --enable=exhaustive

GLAZE_BINARY=$(shell which glaze)

install:
	go build -o ./dist/glaze ./cmd/glaze && \
		cp ./dist/glaze $(GLAZE_BINARY)
