.DEFAULT_GOAL := all

# Makefile variables.
PROJECT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

# Name of output binary.
BIN_NAME := $(or $(BIN_NAME),x-windows-list)

# Artifacts output directory.
ARTIFACTS_DIR := $(abspath $(or $(ARTIFACTS_DIR), $(addprefix $(PROJECT_DIR),/./BUILD)))

# Version for go mod tidy -compat flag.
GO_MOD_COMPAT_VERSION := 1.20

# Configure golang module proxy URI.
GOPROXY ?= proxy.golang.org

# Version of linter.
#  - https://github.com/golangci/golangci-lint/releases/tag/v1.52.2
#  - https://github.com/mgechev/revive/tree/v1.3.1
GOLANGCI_LINT_VERSION := $(or $(GOLANGCI_LINT_VERSION),v1.52.2)
GOLANGCI_LINT_VERSION_FULL := $(subst v,golangci-lint has version ,$(GOLANGCI_LINT_VERSION))

# Set common utilities environs.
ENV_BIN := $(or $(ENV_BIN),env)
GIT_BIN := $(or $(GIT_BIN),git)
GO_BIN := $(or $(GO_BIN),go)
GREP_BIN := $(or $(GREP_BIN),grep)
ID_BIN := $(or $(ID_BIN),id)
MKDIR_BIN := $(or $(MKDIR_BIN),mkdir)
PWD_BIN := $(or $(PWD_BIN),pwd)
SH_BIN := $(or $(SH_BIN),sh)
TEST_BIN := $(or $(TEST_BIN),test)
WGET_BIN := $(or $(WGET_BIN),wget)

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set).
ifeq (,$(shell $(GO_BIN) env GOBIN))
	GOBIN = $(shell $(GO_BIN) env GOPATH)/bin
else
	GOBIN = $(shell $(GO_BIN) env GOBIN)
endif

# Build flags.
GO_BUILD_ENV = CGO_ENABLED=1 GOOS=linux GOARCH=amd64 GOPROXY=$(GOPROXY)
CGO_BUILD_ENV = CGO_CFLAGS="-Wno-deprecated -Wno-deprecated-declarations"
GO_BUILD_FLAGS ?=
LDFLAGS = -s

# Set golangci-lint binary path.
GOLANGCI_LINT_BIN=$(GOBIN)/golangci-lint

all: check build

# Run all checks and generators.
.PHONY: check
check: tidy verify lint test check-git-clean

# Download golangci-lint if needed.
.PHONY: golangci-lint
golangci-lint:
ifeq ("$(wildcard $(GOLANGCI_LINT_BIN))","")
	$(WGET_BIN) -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		$(SH_BIN) -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION)
else
	$(GOLANGCI_LINT_BIN) --version | \
		$(GREP_BIN) -qE '^$(GOLANGCI_LINT_VERSION_FULL)' || \
			$(WGET_BIN) -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
				$(SH_BIN) -s -- -b $(GOBIN) $(GOLANGCI_LINT_VERSION)
endif

# Print environment configuration.
.PHONY: env-info
env-info:
	@echo
	@echo \# Current user information:
	@$(ID_BIN)
	@echo
	@echo \# Current working directory:
	@$(PWD_BIN)
	@echo
	@echo \# Golang related environment variables:
	@$(GO_BIN) env
	@echo
	@echo \# Compiler version:
	@$(GO_BIN) version
	@echo
	@echo \# Git status:
	@$(GIT_BIN) status -s
	@echo

# Run linter.
.PHONY: lint
lint: golangci-lint
	$(GOLANGCI_LINT_BIN) -v run --sort-results ./...

# Run linter with fix flag.
.PHONY: lint-n-fix
lint-n-fix: golangci-lint
	$(GOLANGCI_LINT_BIN) -v run --fix --sort-results ./...

# Run tests.
.PHONY: test
test:
	$(ENV_BIN) $(GO_BUILD_ENV) $(GO_BIN) test -v -failfast ./...

# Update golang dependencies.
.PHONY: tidy
tidy:
	$(TEST_BIN) -d vendor || $(ENV_BIN) $(GO_BUILD_ENV) $(GO_BIN) mod tidy -v -compat=$(GO_MOD_COMPAT_VERSION)

# Update vendor directory.
.PHONY: vendor
vendor:
	$(GO_BIN) mod vendor -v && $(GIT_BIN) status -s

# Verify golang dependencies.
.PHONY: verify
verify:
	$(TEST_BIN) -d vendor || $(GO_BIN) mod verify

# Create artifacts directory.
.PHONY: artifacts-dir
artifacts-dir:
ifeq ($(ARTIFACTS_DIR),)
	@echo "ARTIFACTS_DIR variable must be set!"
	exit 1
endif
	$(MKDIR_BIN) -vp "$(ARTIFACTS_DIR)"

# Build application.
.PHONY: build
build: artifacts-dir env-info
	$(ENV_BIN) $(CGO_BUILD_ENV) $(GO_BUILD_ENV) $(GO_BIN) build -v $(GO_BUILD_FLAGS) -ldflags='$(LDFLAGS)' -a -o '$(ARTIFACTS_DIR)/$(BIN_NAME)'
	@echo

# Fail when directory tree is dirty.
.PHONY: check-git-clean
check-git-clean:
	@status=$$($(GIT_BIN) status --porcelain=v1); \
	if [ ! -z "$${status}" ]; then \
		echo "Error: working directory tree is dirty."; \
		$(GIT_BIN) diff; \
		exit 1; \
	fi

# Show git diff helper (with excludes).
.PHONY: diff
diff:
	$(GIT_BIN) --no-pager diff --diff-algorithm=minimal --ignore-all-space -- ":(exclude)vendor/*" ":(exclude)go.sum"
