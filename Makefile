SHELL := bash

PROJECT_NAME := Pulumi Service Resource Provider

PACK             := pulumiservice
PACKDIR          := sdk
PROJECT          := github.com/pulumi/pulumi-pulumiservice
NODE_MODULE_NAME := @pulumi/pulumiservice
NUGET_PKG_NAME   := Pulumi.PulumiService

PROVIDER        := pulumi-resource-${PACK}
# Override during CI using `make [TARGET] PROVIDER_VERSION=""` or by setting a PROVIDER_VERSION environment variable
# Local & branch builds will just used this fixed default version unless specified
PROVIDER_VERSION ?= 1.0.0-alpha.0+dev
# Use this normalised version everywhere rather than the raw input to ensure consistency.
VERSION_GENERIC = $(shell pulumictl convert-version --language generic --version "$(PROVIDER_VERSION)")
LDFLAGS         := "-X main.Version=$(VERSION_GENERIC)"
BUILD_PATH      := $(PROJECT)/provider/cmd/$(PROVIDER)

SCHEMA_FILE     := provider/cmd/pulumi-resource-pulumiservice/schema.json
GOPATH			:= $(shell go env GOPATH)

WORKING_DIR     := $(shell pwd)
TESTPARALLELISM := 4

# Ensure helpmakego is installed
_ := $(shell go build -o bin/helpmakego github.com/iwahbe/helpmakego)

# The pulumi binary to use during generation
PULUMI := .pulumi/bin/pulumi

ensure::
	go mod tidy
	cd sdk && go mod tidy

build_sdks: provider dotnet_sdk go_sdk nodejs_sdk python_sdk java_sdk

bin/pulumi-resource-pulumiservice: $(shell bin/helpmakego provider/cmd/pulumi-resource-pulumiservice)
	go build -C provider -o ../$@ -ldflags $(LDFLAGS) $(BUILD_PATH)

.PHONY: provider
provider: bin/pulumi-resource-pulumiservice

provider_debug::
	(cd provider && go build -o $(WORKING_DIR)/bin/${PROVIDER} -gcflags="all=-N -l" -ldflags $(LDFLAGS) $(BUILD_PATH))

test_provider::
	cd provider/pkg && go test -short -v -count=1 -coverprofile="coverage.txt" -coverpkg=./... -timeout 2h -parallel ${TESTPARALLELISM} ./...

dotnet_sdk: bin/pulumi-resource-pulumiservice $(PULUMI)
	rm -rf sdk/dotnet
	$(PULUMI) package gen-sdk ./$< --language dotnet
	cd sdk/dotnet/ && \
		printf "module fake_dotnet_module // Exclude this directory from Go tools\n\ngo 1.17\n" > go.mod && \
		echo "${VERSION_GENERIC}" >version.txt && \
		dotnet build

go_sdk: bin/pulumi-resource-pulumiservice $(PULUMI)
	rm -rf sdk/go
	$(PULUMI) package gen-sdk ./$< --language go

nodejs_sdk: bin/pulumi-resource-pulumiservice $(PULUMI)
	rm -rf sdk/nodejs
	$(PULUMI) package gen-sdk ./$< --language nodejs
	cd sdk/nodejs && \
		yarn install --no-progress && \
		yarn run build && \
		cp package.json yarn.lock ./bin/

python_sdk: bin/pulumi-resource-pulumiservice $(PULUMI)
	rm -rf sdk/python
	$(PULUMI) package gen-sdk ./$< --language python
	cd sdk/python/ && \
		printf "module fake_python_module // Exclude this directory from Go tools\n\ngo 1.17\n" > go.mod && \
		cp ../../README.md . && \
		rm -rf ./bin/ ../python.bin/ && cp -R . ../python.bin && mv ../python.bin ./bin && \
		python3 -m venv venv && \
		./venv/bin/python -m pip install build && \
		cd ./bin && \
		../venv/bin/python -m build .

java_sdk: bin/pulumi-resource-pulumiservice $(PULUMI)
	rm -rf sdk/java
	$(PULUMI) package gen-sdk ./$< --language java
	cd sdk/java && \
		printf "module fake_java_module // Exclude this directory from Go tools\n\ngo 1.17\n" > go.mod && \
		cp ../../README.md . && \
		gradle --console=plain build

.PHONY: build
build:: provider dotnet_sdk go_sdk nodejs_sdk python_sdk java_sdk

# Required for the codegen action that runs in pulumi/pulumi
only_build:: build

lint::
	if [ -d provider ]; then \
		pushd provider && golangci-lint run --timeout 10m && popd ; \
	fi
	if [ -d examples ]; then \
		pushd examples && golangci-lint run --timeout 10m --build-tags all && popd ; \
	fi


install:: install_nodejs_sdk install_dotnet_sdk
	cp $(WORKING_DIR)/bin/${PROVIDER} ${GOPATH}/bin

GO_TEST := go test -v -count=1 -cover -timeout 2h -parallel ${TESTPARALLELISM}

install_dotnet_sdk::
	rm -rf $(WORKING_DIR)/nuget/$(NUGET_PKG_NAME).*.nupkg
	mkdir -p $(WORKING_DIR)/nuget
	find . -name '*.nupkg' -print -exec cp -p {} ${WORKING_DIR}/nuget \;

install_python_sdk::
	#target intentionally blank

install_go_sdk::
	#target intentionally blank

install_nodejs_sdk::
	-yarn unlink --cwd $(WORKING_DIR)/sdk/nodejs/bin
	yarn link --cwd $(WORKING_DIR)/sdk/nodejs/bin

install_java_sdk::
	cd sdk/java && gradle publishToMavenLocal


# Keep the version of the pulumi binary used for code generation in sync with the version
# of the dependency used by github.com/pulumi/pulumi-pulumiservice/provider

$(PULUMI): HOME := $(WORKING_DIR)
$(PULUMI): go.mod
	@ PULUMI_VERSION="$$(cd provider && go list -m github.com/pulumi/pulumi/pkg/v3 | awk '{print $$2}')"; \
	if [ -x $(PULUMI) ]; then \
		CURRENT_VERSION="$$($(PULUMI) version)"; \
		if [ "$${CURRENT_VERSION}" != "$${PULUMI_VERSION}" ]; then \
			echo "Upgrading $(PULUMI) from $${CURRENT_VERSION} to $${PULUMI_VERSION}"; \
			rm $(PULUMI); \
		fi; \
	fi; \
	if ! [ -x $(PULUMI) ]; then \
		curl -fsSL https://get.pulumi.com | sh -s -- --version "$${PULUMI_VERSION#v}"; \
	fi

$(SCHEMA_FILE): bin/pulumi-resource-pulumiservice $(PULUMI)
	$(PULUMI) package get-schema ./$<  | \
		jq 'del(.version)' > $@

######################
# ci-mgmt onboarding #
######################

# TODO(https://github.com/pulumi/ci-mgmt/issues/1131): Use default target implementations.

shard:
	@(cd examples && go run github.com/blampe/shard@latest --total $(TOTAL) --index $(INDEX) --output env) >> "$(GITHUB_ENV)"

test_shard:
	cd examples && \
		go test -tags=all -v -count=1 -coverprofile="coverage.txt" -coverpkg=./... -timeout 3h -parallel ${TESTPARALLELISM} -run "$(SHARD_TESTS)" $(SHARD_PATHS)

install_plugins: export PULUMI_HOME := $(WORKING_DIR)/.pulumi
install_plugins: export PATH := $(WORKING_DIR)/.pulumi/bin:$(PATH)
install_plugins: .pulumi/bin/pulumi

bin/linux-amd64/$(PROVIDER): TARGET := linux-amd64
bin/linux-arm64/$(PROVIDER): TARGET := linux-arm64
bin/darwin-amd64/$(PROVIDER): TARGET := darwin-amd64
bin/darwin-arm64/$(PROVIDER): TARGET := darwin-arm64
bin/windows-amd64/$(PROVIDER).exe: TARGET := windows-amd64
bin/%/$(PROVIDER) bin/%/$(PROVIDER).exe:
	@# check the TARGET is set
	test $(TARGET)
	cd provider && \
		export GOOS=$$(echo "$(TARGET)" | cut -d "-" -f 1) && \
		export GOARCH=$$(echo "$(TARGET)" | cut -d "-" -f 2) && \
		export CGO_ENABLED=0 && \
		go build -o "${WORKING_DIR}/$@" $(PULUMI_PROVIDER_BUILD_PARALLELISM) -ldflags $(LDFLAGS) $(BUILD_PATH)

bin/$(PROVIDER)-v$(VERSION_GENERIC)-linux-amd64.tar.gz: bin/linux-amd64/$(PROVIDER)
bin/$(PROVIDER)-v$(VERSION_GENERIC)-linux-arm64.tar.gz: bin/linux-arm64/$(PROVIDER)
bin/$(PROVIDER)-v$(VERSION_GENERIC)-darwin-amd64.tar.gz: bin/darwin-amd64/$(PROVIDER)
bin/$(PROVIDER)-v$(VERSION_GENERIC)-darwin-arm64.tar.gz: bin/darwin-arm64/$(PROVIDER)
bin/$(PROVIDER)-v$(VERSION_GENERIC)-windows-amd64.tar.gz: bin/windows-amd64/$(PROVIDER).exe
bin/$(PROVIDER)-v$(VERSION_GENERIC)-%.tar.gz:
	@mkdir -p dist
	@# $< is the last dependency (the binary path from above) e.g. bin/linux-amd64/pulumi-resource-xyz
	@# $@ is the current target e.g. bin/pulumi-resource-xyz-v1.2.3-linux-amd64.tar.gz
	tar --gzip -cf $@ README.md LICENSE -C $$(dirname $<) .

provider_dist-linux-amd64: bin/$(PROVIDER)-v$(VERSION_GENERIC)-linux-amd64.tar.gz
provider_dist-linux-arm64: bin/$(PROVIDER)-v$(VERSION_GENERIC)-linux-arm64.tar.gz
provider_dist-darwin-amd64: bin/$(PROVIDER)-v$(VERSION_GENERIC)-darwin-amd64.tar.gz
provider_dist-darwin-arm64: bin/$(PROVIDER)-v$(VERSION_GENERIC)-darwin-arm64.tar.gz
provider_dist-windows-amd64: bin/$(PROVIDER)-v$(VERSION_GENERIC)-windows-amd64.tar.gz
provider_dist: provider_dist-linux-amd64 provider_dist-linux-arm64 provider_dist-darwin-amd64 provider_dist-darwin-arm64 provider_dist-windows-amd64

install_sdks: install_nodejs_sdk install_dotnet_sdk install_go_sdk install_python_sdk install_java_sdk

build_nodejs: nodejs_sdk
build_python: python_sdk
build_java: java_sdk
build_dotnet: dotnet_sdk
build_go: go_sdk

schema: provider/cmd/pulumi-resource-pulumiservice/schema.json
