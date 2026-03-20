.PHONY: build test tui

GOCACHE ?= $(CURDIR)/.cache/go-build

build:
	GOCACHE=$(GOCACHE) go build ./...

test:
	GOCACHE=$(GOCACHE) go test ./...

tui:
	GOCACHE=$(GOCACHE) go run ./cmd --tui
