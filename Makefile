.PHONY: gifs

all: gifs

TAPES=$(shell ls doc/vhs/*tape)
gifs: $(TAPES)
	for i in $(TAPES); do vhs < $$i; done

lint:
	docker run --rm -v $(shell pwd):/app -w /app golangci/golangci-lint:v1.50.1 golangci-lint run -v

test:
	go test ./...

build:
	go generate ./...
	go build

release:
	goreleaser release --snapshot --rm-dist