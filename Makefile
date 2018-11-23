PATH := ${PWD}/bin:${PATH}
export PATH

.DEFAULT_GOAL := all

REVISION ?= $(shell git describe --always)
BUILD_DATE ?= $(shell date +'%Y-%m-%dT%H:%M:%SZ')

GO_BUILD_FLAGS := -v
GO_LDFLAGS := -ldflags "-X main.revision=$(REVISION) -X main.buildDate=$(BUILD_DATE)"
GO_TEST_FLAGS := -v -timeout 3m
GO_COVER_FLAGS := -coverprofile coverage.txt -covermode atomic
SRC_FILES := $(shell go list -f '{{range .GoFiles}}{{printf "%s/%s\n" $$.Dir .}}{{end}}' ./...)

XC_ARCH := 386 amd64
XC_OS := darwin linux windows


#  Apps
#----------------------------------------------------------------
BIN_DIR := ./bin
OUT_DIR := ./dist
BINS :=

define cmd-tmpl

$(eval NAME := $(notdir $(1)))
$(eval OUT := $(addprefix $(BIN_DIR)/,$(NAME)))

$(OUT): $(SRC_FILES)
	go build $(GO_BUILD_FLAGS) $(GO_LDFLAGS) -o $(OUT) $(1)

.PHONY: $(NAME)
$(NAME): $(OUT)

.PHONY: $(NAME)-install
$(NAME)-install:
	go install $(GO_BUILD_FLAGS) $(GO_LDFLAGS) $(1)

.PHONY: $(NAME)-package
$(NAME)-package:
	gex gox \
		$(GO_LDFLAGS) \
		-os="$(XC_OS)" \
		-arch="$(XC_ARCH)" \
		-output="$(OUT_DIR)/$(NAME)_{{.OS}}_{{.Arch}}" \
		$(1)

$(eval BINS += $(NAME))

endef

$(foreach src,$(wildcard ./cmd/*),$(eval $(call cmd-tmpl,$(src))))

.PHONY: all
all: $(addprefix $(BIN_DIR)/,$(BINS))

.PHONY: packages
packages: $(addsuffix -package,$(BINS))

.PHONY: install
install: $(addsuffix -install,$(BINS))


#  commands
#----------------------------------------------------------------
.PHONY: setup
setup:
ifdef CI
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
endif
	dep ensure -v -vendor-only
	@go get github.com/izumin5210/gex/cmd/gex
	gex --build --verbose

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)/* $(OUT_DIR)/*

.PHONY: lint
lint:
ifdef CI
	gex reviewdog -reporter=github-pr-review
else
	gex reviewdog -diff="git diff master"
endif

.PHONY: test
test:
	go test $(GO_TEST_FLAGS) ./...

.PHONY: cover
cover:
	go test $(GO_TEST_FLAGS) $(GO_COVER_FLAGS) ./...

.PHONY: test-e2e
test-e2e: $(BIN_DIR)/ery
	go test $(GO_TEST_FLAGS) ./_tests/e2e
