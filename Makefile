# Ensure GOBIN is not set during build so that tools are installed to the correct path
unexport GOBIN

# GO related variables
GO           ?= go
GOFMT        ?= $(GO)fmt
FIRST_GOPATH := $(firstword $(subst :, ,$(shell $(GO) env GOPATH)))
GOOPTS       ?=
GOHOSTOS     ?= $(shell $(GO) env GOHOSTOS)
GOHOSTARCH   ?= $(shell $(GO) env GOHOSTARCH)
PKGS         := ./...

GOTEST       := $(GO) test
GOTEST_DIR   := test

ifeq ($(GOHOSTARCH),amd64 arm64)
	ifeq ($(GOHOSTOS),$(filter $(GOHOSTOS),linux freebsd darwin windows))
		# Only supported on amd64 and arm64
		test-flags := -race
	endif
endif

GOLANGCI_LINT :=
GOLANGCI_LINT_OPTS ?=
GOLANGCI_LINT_VERSION ?= v1.64.8
# golangci-lint only supports linux, darwin and windows platforms on i386/amd64/arm64.
# windows isn't included here because of the path separator being different.
ifeq ($(GOHOSTOS),$(filter $(GOHOSTOS),linux darwin))
	ifeq ($(GOHOSTARCH),$(filter $(GOHOSTARCH),amd64 i386 arm64))
		GOLANGCI_LINT := $(FIRST_GOPATH)/bin/golangci-lint
	endif
endif

## test: running tests
.PHONY: test
test: lint
	@echo ">> running tests"
	$(GOTEST) $(test-flags) $(GOOPTS) $(PKGS)

## lint: running code inspection
.PHONY: lint
lint: $(GOLANGCI_LINT)
ifdef GOLANGCI_LINT
	@echo ">> running golangci-lint"
ifdef GO111MODULE
# 'go list' needs to be executed before staticcheck to prepopulate the modules cache.
# Otherwise staticcheck might fail randomly for some reason not yet explained.
	GO111MODULE=$(GO111MODULE) $(GO) list -e -compiled -test=true -export=false -deps=true -find=false -tags= -- ./... > /dev/null
	GO111MODULE=$(GO111MODULE) $(GOLANGCI_LINT) run $(GOLANGCI_LINT_OPTS) $(PKGS)
else
	$(GOLANGCI_LINT) run $(GOLANGCI_LINT_OPTS) $(PKGS)
endif
endif

# install golangci-lint if not exist
ifdef GOLANGCI_LINT
$(GOLANGCI_LINT):
	mkdir -p $(FIRST_GOPATH)/bin
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_LINT_VERSION)/install.sh \
		| sed -e '/install -d/d' \
		| sh -s -- -b $(FIRST_GOPATH)/bin $(GOLANGCI_LINT_VERSION)
endif
