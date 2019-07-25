.DEFAULT_GOAL := build

PATH := ${PWD}/bin:${PATH}
export PATH
export GO111MODULE=on

.PHONY: tools
tools:
	go generate -tags tools

.PHONY: gen
gen: tools
	go generate ./...

.PHONY: build
build:
	go build -o=./bin/ery ./cmd/ery

.PHONY: install
install:
	go install ./cmd/ery

.PHONY: lint
lint: tools
ifdef CI
	reviewdog -reporter=github-pr-review
else
	reviewdog -diff="git diff master"
endif

.PHONY: test
test:
	go test -race -v ./...

.PHONY: cover
cover:
	go test -race -v -coverprofile coverage.txt -covermode atomic ./...

.PHONY: test-e2e
test-e2e: build
	go test -v ./_tests/e2e
