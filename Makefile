.DEFAULT_GOAL := all

ORG := srvc
PROJ := ery
PKG := github.com/$(ORG)/$(PROJ)

VERSION_MAJOR ?= 0
VERSION_MINOR ?= 2
VERSION_BUILD ?= 1

VERSION ?= v$(VERSION_MAJOR).$(VERSION_MINOR).$(VERSION_BUILD)
REVISION ?= $(shell git describe --always)
BUILD_DATE ?= $(shell date +'%Y-%m-%dT%H:%M:%SZ')
RELEASE_TYPE ?= $(if $(shell git tag --contains $(REVISION) | grep $(VERSION)),stable,canary)

BIN_DIR := ./bin

#  build
#----------------------------------------------------------------
LDFLAGS := "-X main.version=$(VERSION) -X main.revision=$(REVISION) -X main.buildDate=$(BUILD_DATE) -X main.releaseType=$(RELEASE_TYPE)"
GO_BUILD_FLAGS := -v -ldflags $(LDFLAGS)
CMDS := $(notdir $(abspath $(wildcard cmd/*)))

.PHONY: $(CMDS)
$(CMDS):
	@go build $(GO_BUILD_FLAGS) -o ./$(BIN_DIR)/$@ $(PKG)/cmd/$@

.PHONY: all
all: $(CMDS)

#  commands
#----------------------------------------------------------------
.PHONY: setup
setup: dep

.PHONY: dep
dep:
	@dep ensure -v -vendor-only

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)

.PHONY: clobber
clobber: clean
	rm -rf vendor

#  lint
#----------------------------------------------------------------
.PHONY: lint
lint: gofmt golint

.PHONY: gofmt
gofmt:
	@find . -name "*.go" | grep -v vendor/ | xargs gofmt -l -s -e

.PHONY: golint
golint:
	@go list ./... | xargs golint -set_exit_status

#  test
#----------------------------------------------------------------
GO_TEST_FLAGS := -v
GO_COVER_FLAGS := -coverpkg ./... -coverprofile coverage.txt -covermode atomic

.PHONY: test
test:
	@go test $(GO_TEST_FLAGS) ./...

.PHONY: cover
cover:
	@go test $(GO_TEST_FLAGS) $(GO_COVER_FLAGS) ./...
