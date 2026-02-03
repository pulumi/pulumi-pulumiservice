SHELL := bash

PROJECT_NAME := Pulumi Service Resource Provider

PACK             := pulumiservice
PACKDIR          := sdk
PROJECT          := github.com/pulumi/pulumi-pulumiservice
PROVIDER_PATH    := provider
NODE_MODULE_NAME := @pulumi/pulumiservice
NUGET_PKG_NAME   := Pulumi.PulumiService

PROVIDER        := pulumi-resource-${PACK}
# Override during CI using `make [TARGET] PROVIDER_VERSION=""` or by setting a PROVIDER_VERSION environment variable
# Local & branch builds will just used this fixed default version unless specified
PROVIDER_VERSION ?= 1.0.0-alpha.0+dev
# Use this normalised version everywhere rather than the raw input to ensure consistency.
VERSION_GENERIC = $(shell pulumictl convert-version --language generic --version "$(PROVIDER_VERSION)")
LDFLAGS         := -X main.Version=$(VERSION_GENERIC)
BUILD_PATH      := $(PROJECT)/provider/cmd/$(PROVIDER)

SCHEMA_FILE     := provider/cmd/pulumi-resource-pulumiservice/schema.json
GOPATH			:= $(shell go env GOPATH)

WORKING_DIR     := $(shell pwd)
TESTPARALLELISM := 4

# Ensure all directories exist before evaluating targets to avoid issues with `touch` creating directories.
_ := $(shell mkdir -p .make bin .pulumi/bin)

# Ensure helpmakego is installed
_ := $(shell go build -o bin/helpmakego github.com/iwahbe/helpmakego)

# The pulumi binary to use during generation
PULUMI := pulumi

ensure:: | mise_install
	go mod tidy
	cd sdk && go mod tidy

# Installs all necessary tools with mise and records completion in a sentinel
# file so dependent targets can participate in make's caching behaviour. The
# environment is refreshed via an order-only prerequisite so it still runs on
# every invocation without invalidating the sentinel.
mise_install: .make/mise_install | mise_env

.PHONY: mise_env
mise_env:
	@mise env -q  > /dev/null

.make/mise_install:
	@mise install -q
	@touch $@

# Prepare the workspace for building the provider and SDKs
# Importantly this is run by CI ahead of restoring the bin directory and resuming SDK builds
prepare_local_workspace: .make/mise_install
prepare_local_workspace: | mise_env

build_sdks: provider dotnet_sdk go_sdk nodejs_sdk python_sdk java_sdk

bin/pulumi-resource-pulumiservice: $(shell bin/helpmakego provider/cmd/pulumi-resource-pulumiservice) | mise_install
	go build -C provider -o ../$@ -ldflags "$(LDFLAGS)" $(BUILD_PATH)

build_provider_cmd = GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 go build $(PULUMI_PROVIDER_BUILD_PARALLELISM) -o "$(3)" -ldflags "$(LDFLAGS)" $(PROJECT)/$(PROVIDER_PATH)/cmd/$(PROVIDER)

.PHONY: provider
provider: bin/pulumi-resource-pulumiservice

provider_debug:: | mise_install
	(cd provider && go build -o $(WORKING_DIR)/bin/${PROVIDER} -gcflags="all=-N -l" -ldflags "$(LDFLAGS)" $(BUILD_PATH))

test_provider::
	cd provider/pkg && go test -short -v -count=1 -coverprofile="coverage.txt" -coverpkg=./... -timeout 2h -parallel ${TESTPARALLELISM} ./...

dotnet_sdk: bin/pulumi-resource-pulumiservice
	rm -rf sdk/dotnet
	$(PULUMI) package gen-sdk ./$< --language dotnet
	cd sdk/dotnet/ && \
		printf "module fake_dotnet_module // Exclude this directory from Go tools\n\ngo 1.17\n" > go.mod && \
		echo "$(VERSION_GENERIC)" >version.txt && \
		dotnet build

go_sdk: bin/pulumi-resource-pulumiservice
	rm -rf sdk/go
	$(PULUMI) package gen-sdk ./$< --language go

nodejs_sdk: bin/pulumi-resource-pulumiservice
	rm -rf sdk/nodejs
	$(PULUMI) package gen-sdk ./$< --language nodejs
	cd sdk/nodejs && \
		yarn install --no-progress && \
		yarn run build && \
		cp package.json yarn.lock ./bin/

python_sdk: bin/pulumi-resource-pulumiservice
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

java_sdk: bin/pulumi-resource-pulumiservice
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

lint:: | mise_install
	if [ -d provider ]; then \
		pushd provider && golangci-lint run --timeout 10m --config ../.golangci.yml && popd ; \
	fi
	if [ -d examples ]; then \
		pushd examples && golangci-lint run --timeout 10m --build-tags all --config ../.golangci.yml && popd ; \
	fi


install:: install_nodejs_sdk install_dotnet_sdk
	cp $(WORKING_DIR)/bin/${PROVIDER} ${GOPATH}/bin

GO_TEST := go test -v -count=1 -cover -timeout 2h -parallel ${TESTPARALLELISM}

install_dotnet_sdk::
	mkdir -p nuget
	find sdk/dotnet/bin -name '*.nupkg' -print -exec cp -p "{}" ${WORKING_DIR}/nuget \;
	if ! dotnet nuget list source | grep "${WORKING_DIR}/nuget"; then \
		dotnet nuget add source "${WORKING_DIR}/nuget" --name "${WORKING_DIR}/nuget" \
	; fi

install_python_sdk::
	#target intentionally blank

install_go_sdk::
	#target intentionally blank

install_nodejs_sdk:: | mise_install
	-yarn unlink --cwd $(WORKING_DIR)/sdk/nodejs/bin
	yarn link --cwd $(WORKING_DIR)/sdk/nodejs/bin

install_java_sdk:: | mise_install
	cd sdk/java && gradle publishToMavenLocal


$(SCHEMA_FILE): bin/pulumi-resource-pulumiservice | mise_install
	$(PULUMI) package get-schema ./$<  | \
		jq 'del(.version)' > $@

######################
# ci-mgmt onboarding #
######################

# TODO(https://github.com/pulumi/ci-mgmt/issues/1131): Use default target implementations.

.PHONY: test
test: export PATH := $(WORKING_DIR)/bin:$(PATH)
test:
	cd examples && go test -v -tags=all -parallel $(TESTPARALLELISM) -timeout 2h $(value GOTESTARGS)

install_plugins: export PULUMI_HOME := $(WORKING_DIR)/.pulumi
install_plugins: export PATH := $(WORKING_DIR)/.pulumi/bin:$(PATH)
install_plugins: .pulumi/bin/pulumi

install_sdks: install_nodejs_sdk install_dotnet_sdk install_go_sdk install_python_sdk install_java_sdk

build_nodejs: nodejs_sdk
build_python: python_sdk
build_java: java_sdk
build_dotnet: dotnet_sdk
build_go: go_sdk

schema: provider/cmd/pulumi-resource-pulumiservice/schema.json

include scripts/crossbuild.mk
