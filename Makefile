# Makefile for Pulumi Service Provider (custom native provider)
# This provider does not use tfgen - it uses schema.json directly with pulumi package gen-sdk

SHELL := /bin/bash

PACK             := pulumiservice
PACKDIR          := sdk
PROJECT          := github.com/pulumi/pulumi-pulumiservice
NODE_MODULE_NAME := @pulumi/pulumiservice
NUGET_PKG_NAME   := Pulumi.PulumiService

PROVIDER        := pulumi-resource-${PACK}
CODEGEN         := pulumi-gen-${PACK}
# Override during CI using `make [TARGET] PROVIDER_VERSION=""` or by setting a PROVIDER_VERSION environment variable
# Local & branch builds will just used this fixed default version unless specified
PROVIDER_VERSION ?= 0.0.0-alpha.0+dev
# Use this normalised version everywhere rather than the raw input to ensure consistency.
VERSION_GENERIC = $(shell pulumictl convert-version --language generic --version "$(PROVIDER_VERSION)")
LDFLAGS         := "-X github.com/pulumi/pulumi-pulumiservice/provider/pkg/version.Version=$(VERSION_GENERIC)"
BUILD_PATH      := $(PROJECT)/provider/cmd/$(PROVIDER)

SCHEMA_FILE     := provider/cmd/pulumi-resource-pulumiservice/schema.json
GOPATH          := $(shell go env GOPATH)

WORKING_DIR     := $(shell pwd)
TESTPARALLELISM := 10

# The pulumi binary to use during generation
PULUMI := .pulumi/bin/pulumi

# Create a `.make` directory for tracking targets
_ := $(shell mkdir -p .make bin .pulumi/bin)

.PHONY: ensure
ensure::
	go mod tidy
	cd sdk && go mod tidy

.PHONY: build
build:: install_plugins provider build_sdks

.PHONY: build_sdks
build_sdks: build_dotnet build_go build_nodejs build_python build_java

.PHONY: generate
generate: generate_sdks schema

.PHONY: generate_sdks
generate_sdks: generate_dotnet generate_go generate_nodejs generate_python generate_java

.PHONY: schema
schema::
	(cd provider && VERSION=$(VERSION_GENERIC) go generate cmd/$(PROVIDER)/main.go)

# Alias for CI compatibility - native providers don't have a separate codegen binary
.PHONY: codegen
codegen: schema

# Alias for CI compatibility
.PHONY: generate_schema
generate_schema: schema

.PHONY: provider
provider:: $(SCHEMA_FILE)
	(cd provider && go build -o $(WORKING_DIR)/bin/$(PROVIDER) -ldflags $(LDFLAGS) $(BUILD_PATH))
	@# Create a symlink for the codegen binary to satisfy CI workflows
	@# Native providers don't have a separate codegen binary
	@ln -sf $(PROVIDER) $(WORKING_DIR)/bin/$(CODEGEN)

.PHONY: provider_debug
provider_debug::
	(cd provider && go build -o $(WORKING_DIR)/bin/$(PROVIDER) -gcflags="all=-N -l" -ldflags $(LDFLAGS) $(BUILD_PATH))

.PHONY: test_provider
test_provider::
	cd provider && go test -v -short -coverprofile="coverage.txt" -coverpkg="./..." -parallel $(TESTPARALLELISM) -timeout 2h ./...

.PHONY: test
test: export PATH := $(WORKING_DIR)/bin:$(PATH)
test:
	cd examples && go test -v -tags=all -parallel $(TESTPARALLELISM) -timeout 2h

generate_dotnet: .make/generate_dotnet
build_dotnet: .make/build_dotnet
.make/generate_dotnet: $(PULUMI) $(SCHEMA_FILE)
	rm -rf sdk/dotnet
	$(PULUMI) package gen-sdk $(SCHEMA_FILE) --language dotnet --out sdk
	cd sdk/dotnet/ && \
		printf "module fake_dotnet_module // Exclude this directory from Go tools\n\ngo 1.17\n" > go.mod && \
		echo "$(VERSION_GENERIC)" >version.txt
	@touch $@
.make/build_dotnet: .make/generate_dotnet
	mkdir -p $(WORKING_DIR)/nuget
	cd sdk/dotnet/ && dotnet build
	@touch $@
.PHONY: generate_dotnet build_dotnet

generate_go: .make/generate_go
build_go: .make/build_go
.make/generate_go: $(PULUMI) $(SCHEMA_FILE)
	rm -rf sdk/go
	$(PULUMI) package gen-sdk $(SCHEMA_FILE) --language go --out sdk
	@touch $@
.make/build_go: .make/generate_go
	cd sdk && go list "$$(grep -e "^module" go.mod | cut -d ' ' -f 2)/go/..." | xargs -I {} bash -c 'go build {} && go clean -i {}'
	@touch $@
.PHONY: generate_go build_go

generate_java: .make/generate_java
build_java: .make/build_java
.make/generate_java: $(PULUMI) $(SCHEMA_FILE)
	rm -rf sdk/java
	$(PULUMI) package gen-sdk $(SCHEMA_FILE) --language java --out sdk
	printf "module fake_java_module // Exclude this directory from Go tools\n\ngo 1.17\n" > sdk/java/go.mod
	@touch $@
.make/build_java: .make/generate_java
	cd sdk/java/ && \
		gradle --console=plain build && \
		gradle --console=plain javadoc
	@touch $@
.PHONY: generate_java build_java

generate_nodejs: .make/generate_nodejs
build_nodejs: .make/build_nodejs
.make/generate_nodejs: $(PULUMI) $(SCHEMA_FILE)
	rm -rf sdk/nodejs
	$(PULUMI) package gen-sdk $(SCHEMA_FILE) --language nodejs --out sdk
	printf "module fake_nodejs_module // Exclude this directory from Go tools\n\ngo 1.17\n" > sdk/nodejs/go.mod
	@touch $@
.make/build_nodejs: .make/generate_nodejs
	cd sdk/nodejs/ && \
		yarn install && \
		yarn run tsc && \
		cp ../../README.md ../../LICENSE package.json yarn.lock ./bin/
	@touch $@
.PHONY: generate_nodejs build_nodejs

generate_python: .make/generate_python
build_python: .make/build_python
.make/generate_python: $(PULUMI) $(SCHEMA_FILE)
	rm -rf sdk/python
	$(PULUMI) package gen-sdk $(SCHEMA_FILE) --language python --out sdk
	printf "module fake_python_module // Exclude this directory from Go tools\n\ngo 1.17\n" > sdk/python/go.mod
	cp README.md sdk/python/
	@touch $@
.make/build_python: .make/generate_python
	cd sdk/python/ && \
		python3 -m venv venv && \
		./venv/bin/python -m pip install build==1.2.1 && \
		./venv/bin/python -m build .
	@touch $@
.PHONY: generate_python build_python

.PHONY: clean
clean:
	rm -rf sdk/{dotnet,nodejs,go,python,java}
	rm -rf bin/*
	rm -rf .make/*

.PHONY: install_sdks
install_sdks: install_dotnet_sdk install_go_sdk install_java_sdk install_nodejs_sdk install_python_sdk

install_dotnet_sdk: .make/install_dotnet_sdk
.make/install_dotnet_sdk: .make/build_dotnet
	mkdir -p nuget
	find sdk/dotnet/bin -name '*.nupkg' -print -exec cp -p "{}" ${WORKING_DIR}/nuget \;
	if ! dotnet nuget list source | grep "${WORKING_DIR}/nuget"; then \
		dotnet nuget add source "${WORKING_DIR}/nuget" --name "${WORKING_DIR}/nuget" \
	; fi
	@touch $@
.PHONY: install_dotnet_sdk

.PHONY: install_go_sdk
install_go_sdk:

.PHONY: install_java_sdk
install_java_sdk:

install_nodejs_sdk: .make/install_nodejs_sdk
.make/install_nodejs_sdk: .make/build_nodejs
	yarn link --cwd $(WORKING_DIR)/sdk/nodejs/bin
	@touch $@
.PHONY: install_nodejs_sdk

.PHONY: install_python_sdk
install_python_sdk:

.PHONY: lint_provider
lint_provider: upstream
	cd provider && golangci-lint run --path-prefix provider -c ../.golangci.yml --timeout 10m

.PHONY: lint_provider.fix
lint_provider.fix: upstream
	cd provider && golangci-lint run --path-prefix provider -c ../.golangci.yml --fix --timeout 10m

.PHONY: lint
lint: lint_provider
	cd sdk && golangci-lint run -c ../.golangci.yml --timeout 10m
	cd examples && golangci-lint run -c ../.golangci.yml --build-tags all --timeout 10m

.PHONY: fmt
fmt:
	cd provider && golangci-lint fmt -c ../.golangci.yml
	cd sdk && golangci-lint fmt -c ../.golangci.yml
	cd examples && golangci-lint fmt -c ../.golangci.yml

# Install plugins
.PHONY: install_plugins
install_plugins: .make/install_plugins
.make/install_plugins: $(PULUMI)
	@touch $@

# Install pulumi CLI
$(PULUMI):
	@if [ ! -f $(PULUMI) ]; then \
		curl -fsSL https://get.pulumi.com | sh -s -- --version $(shell cat .pulumi.version) --install-root $(WORKING_DIR)/.pulumi; \
	fi

# Apply patches to the upstream submodule, if it exists
.PHONY: upstream
upstream: .make/upstream
.make/upstream: $(wildcard patches/*) $(shell ./scripts/upstream.sh file_target 2>/dev/null || echo)
	@if [ -f ./scripts/upstream.sh ]; then ./scripts/upstream.sh init; fi
	@touch $@

# Regenerate CI configuration
.PHONY: ci-mgmt
ci-mgmt: .ci-mgmt.yaml
	go run github.com/pulumi/ci-mgmt/provider-ci@master generate

.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Main Targets"
	@echo "  build (default)     Build the provider and all SDKs"
	@echo "  generate            Generate all SDKs and schema"
	@echo "  schema              Generate schema.json using go generate"
	@echo "  codegen             Generate schema (alias for CI compatibility)"
	@echo "  generate_schema     Generate schema (alias for CI compatibility)"
	@echo "  provider            Build the provider binary"
	@echo "  lint_provider       Run the linter on the provider"
	@echo "  lint                Run the linter on provider, sdk, and examples"
	@echo "  fmt                 Format code using golangci-lint fmt"
	@echo "  test_provider       Run the provider tests"
	@echo "  test                Run the example tests"
	@echo "  clean               Clean up generated files"
	@echo ""

include scripts/plugins.mk
include scripts/crossbuild.mk

# Permit providers to extend the Makefile with provider-specific Make includes.
include $(wildcard .mk/*.mk)
