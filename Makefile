# Makefile

# Quiet by default; run `make V=1 bench-all` to see commands.
Q :=
ifneq ($(V),1)
Q := @
MAKEFLAGS += --no-print-directory
endif

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# Ensure local tools in ./bin are found by exec.LookPath (used in benches)
export PATH := $(LOCALBIN):$(PATH)

## Tool Binaries
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint
RENVSUBST     = $(LOCALBIN)/renvsubst

## Tool Versions
# renovate: datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION ?= v2.10.1
# renovate: datasource=github-releases depName=containeroo/renvsubst
RENVSUBST_VERSION ?= v0.10.0

# Default: no prefix (override via: make patch VERSION_PREFIX=v)
VERSION_PREFIX ?= v

##@ Tagging

LATEST_TAG = $(shell \
	if [ -n "$(VERSION_PREFIX)" ]; then \
		git tag --list "$(VERSION_PREFIX)*" --sort=-v:refname | head -n 1 ; \
	else \
		git tag --list --sort=-v:refname | head -n 1 ; \
	fi)
VERSION    = $(shell [ -n "$(LATEST_TAG)" ] && echo $(LATEST_TAG) | sed "s/^$(VERSION_PREFIX)//" || echo "0.0.0")

patch: ## Create a new patch release (x.y.Z+1)
	@NEW_VERSION=$$(echo "$(VERSION)" | awk -F. '{printf "%d.%d.%d", $$1, $$2, $$3+1}') && \
	git tag "$(VERSION_PREFIX)$${NEW_VERSION}" && \
	echo "Tagged $(VERSION_PREFIX)$${NEW_VERSION}"

minor: ## Create a new minor release (x.Y+1.0)
	@NEW_VERSION=$$(echo "$(VERSION)" | awk -F. '{printf "%d.%d.0", $$1, $$2+1}') && \
	git tag "$(VERSION_PREFIX)$${NEW_VERSION}" && \
	echo "Tagged $(VERSION_PREFIX)$${NEW_VERSION}"

major: ## Create a new major release (X+1.0.0)
	@NEW_VERSION=$$(echo "$(VERSION)" | awk -F. '{printf "%d.0.0", $$1+1}') && \
	git tag "$(VERSION_PREFIX)$${NEW_VERSION}" && \
	echo "Tagged $(VERSION_PREFIX)$${NEW_VERSION}"

tag:  ## Show latest tag
	@echo "Latest version: $(if $(LATEST_TAG),$(LATEST_TAG),<none>)"

push: ## Push tags to remote
	git push --tags

##@ Development

.PHONY: download
download: ## Download go packages
	go mod download

.PHONY: run
run: ## Run the app
	go run ./cmd/vex/main.go

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: fmt vet ## Run unit tests.
	go test -covermode=atomic -count=1 -parallel=4 -timeout=5m ./...

.PHONY: cover
cover: ## Display test coverage
	go test -coverprofile=coverage.out -covermode=atomic -count=1 -parallel=4 -timeout=5m ./...
	go tool cover -html=coverage.out

.PHONY: clean
clean: ## Clean up generated files
	rm -f coverage.out coverage.html cpu.out mem.out

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter.
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint with --fix.
	$(GOLANGCI_LINT) run --fix

##@ Build (vex)

COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS ?= -s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT)

VEX_GO_BIN := $(LOCALBIN)/vex-go

.PHONY: build-vex-go
build-vex-go: $(LOCALBIN) ## Build vex with the standard Go toolchain
	@echo ">> Building vex (Go) -> $(VEX_GO_BIN)"
	@CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -o '$(VEX_GO_BIN)' ./cmd/vex/main.go

##@ Benchmarks

PKG_VEX      ?= ./benchmarks/
PKG_ENVSUBST ?= ./benchmarks/

BENCH_TIME    ?= 2s
BENCH_COUNT   ?= 1
BENCH_CPU_OUT ?= cpu.out
BENCH_MEM_OUT ?= mem.out

.PHONY: bench-vex
bench-vex: build-vex-go
	@echo ">> Benchmarking vex (Go build)"
	$(Q)VEX_BIN=$(VEX_GO_BIN) go test -run=^$$ -bench '^BenchmarkVex' -benchmem -benchtime=$(BENCH_TIME) -count=$(BENCH_COUNT) $(PKG_VEX)

.PHONY: bench-envsubst
bench-envsubst:
	$(Q)go test -run=^$$ -bench '^BenchmarkEnvsubst' -benchmem -benchtime=$(BENCH_TIME) -count=$(BENCH_COUNT) $(PKG_ENVSUBST)

.PHONY: bench-renvsubst
bench-renvsubst: renvsubst
	$(Q)RENV_BIN=$(RENVSUBST) go test -run=^$$ -bench '^BenchmarkRenvsubst' -benchmem -benchtime=$(BENCH_TIME) -count=$(BENCH_COUNT) $(PKG_ENVSUBST)

.PHONY: bench-all
bench-all:
	$(Q)$(MAKE) bench-vex
	$(Q)$(MAKE) renvsubst -s
	$(Q)RENV_BIN=$(RENVSUBST) $(MAKE) bench-renvsubst -s
	$(Q)$(MAKE) bench-envsubst

.PHONY: bench-profile
bench-profile: ## Run Vex bench with cpu/mem profiles
	go test -run=^$$ -bench '^BenchmarkVex' -benchmem -benchtime=$(BENCH_TIME) -cpuprofile=$(BENCH_CPU_OUT) -memprofile=$(BENCH_MEM_OUT) $(PKG_VEX)

##@ Perf Mode Helpers (best-effort; Linux effective, macOS informative)

.PHONY: perf-on
perf-on: ## Linux: set performance governor; macOS: print guidance
	@uname_s=$$(uname -s); \
	if [ "$$uname_s" = "Linux" ]; then \
	  if command -v cpupower >/dev/null 2>&1; then \
	    echo ">> sudo cpupower frequency-set -g performance"; \
	    sudo cpupower frequency-set -g performance; \
	  elif command -v cpufreq-set >/dev/null 2>&1; then \
	    echo ">> sudo cpufreq-set -r -g performance"; \
	    sudo cpufreq-set -r -g performance; \
	  else \
	    echo "!! cpupower/cpufreq-set not found; install linux-tools/cpufrequtils"; \
	  fi; \
	elif [ "$$uname_s" = "Darwin" ]; then \
	  echo ">> macOS doesn't expose a standard way to pin perf or disable turbo."; \
	  echo "   Tip: close background apps; plug in power; use the same thermal state."; \
	else \
	  echo ">> Unsupported OS $$uname_s; skipping."; \
	fi

.PHONY: perf-off
perf-off: ## Linux: set powersave governor; macOS: print guidance
	@uname_s=$$(uname -s); \
	if [ "$$uname_s" = "Linux" ]; then \
	  if command -v cpupower >/dev/null 2>&1; then \
	    echo ">> sudo cpupower frequency-set -g powersave"; \
	    sudo cpupower frequency-set -g powersave; \
	  elif command -v cpufreq-set >/dev/null 2>&1; then \
	    echo ">> sudo cpufreq-set -r -g powersave"; \
	    sudo cpufreq-set -r -g powersave; \
	  else \
	    echo "!! cpupower/cpufreq-set not found; install linux-tools/cpufrequtils"; \
	  fi; \
	elif [ "$$uname_s" = "Darwin" ]; then \
	  echo ">> macOS: no standard way to toggle; returning to normal usage is enough."; \
	else \
	  echo ">> Unsupported OS $$uname_s; skipping."; \
	fi

##@ Dependencies

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): $(LOCALBIN)
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

# ---------------- renvsubst (Rust) via GitHub release asset ----------------

UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_M),x86_64)
  RUST_ARCH := x86_64
else ifeq ($(UNAME_M),amd64)
  RUST_ARCH := x86_64
else ifeq ($(UNAME_M),arm64)
  RUST_ARCH := aarch64
else ifeq ($(UNAME_M),aarch64)
  RUST_ARCH := aarch64
else
  RUST_ARCH := $(UNAME_M)
endif

ifeq ($(UNAME_S),Darwin)
  RUST_OS := apple-darwin
  RUST_LINUX_FLAVOR :=
else ifeq ($(UNAME_S),Linux)
  RUST_OS := unknown-linux-gnu
  RUST_LINUX_FLAVOR := musl
else
  $(warning Unsupported host OS '$(UNAME_S)'; assuming Linux gnu)
  RUST_OS := unknown-linux-gnu
  RUST_LINUX_FLAVOR := musl
endif

RENVSUBST_VERSIONED := $(RENVSUBST)-$(RENVSUBST_VERSION)
RENVSUBST_ASSET_GNU  := renvsubst-$(RENVSUBST_VERSION)-$(RUST_ARCH)-$(RUST_OS).tar.gz
RENVSUBST_ASSET_MUSL := renvsubst-$(RENVSUBST_VERSION)-$(RUST_ARCH)-unknown-linux-musl.tar.gz
RENVSUBST_URL_GNU    := https://github.com/containeroo/renvsubst/releases/download/$(RENVSUBST_VERSION)/$(RENVSUBST_ASSET_GNU)
RENVSUBST_URL_MUSL   := https://github.com/containeroo/renvsubst/releases/download/$(RENVSUBST_VERSION)/$(RENVSUBST_ASSET_MUSL)

.PHONY: renvsubst
renvsubst: $(RENVSUBST) ## Ensure ./bin/renvsubst symlink exists at requested version.

$(RENVSUBST): $(RENVSUBST_VERSIONED)
	@ln -sf "$(RENVSUBST_VERSIONED)" "$(RENVSUBST)"
	@echo ">> Linked $(RENVSUBST) -> $(RENVSUBST_VERSIONED)"

$(RENVSUBST_VERSIONED): $(LOCALBIN)
	@set -e; \
	echo ">> Installing renvsubst $(RENVSUBST_VERSION) into $(LOCALBIN)"; \
	TMPD="$$(mktemp -d)"; \
	URL="$(RENVSUBST_URL_GNU)"; \
	echo ">> Probing $$URL"; \
	if ! curl -fsI "$$URL" >/dev/null 2>&1; then \
	  if [ -n "$(RUST_LINUX_FLAVOR)" ]; then \
	    URL="$(RENVSUBST_URL_MUSL)"; \
	    echo ">> Falling back to $$URL"; \
	  fi; \
	fi; \
	curl -fsSL "$$URL" -o "$$TMPD/renvsubst.tgz"; \
	tar -xzf "$$TMPD/renvsubst.tgz" -C "$$TMPD"; \
	BIN="$$(find "$$TMPD" -type f -name renvsubst -perm +111 -o -type f -name renvsubst)"; \
	if [ -z "$$BIN" ]; then echo "renvsubst binary not found in archive"; rm -rf "$$TMPD"; exit 1; fi; \
	install -m 0755 "$$BIN" "$(RENVSUBST_VERSIONED)"; \
	rm -rf "$$TMPD"; \
	echo ">> Installed $(RENVSUBST_VERSIONED)"

define go-install-tool
@[ -f "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) || true ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $(1)-$(3) $(1)
endef

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

