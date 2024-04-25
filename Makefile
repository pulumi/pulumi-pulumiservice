PROJECT_NAME := Pulumi Service Resource Provider

PACK             := pulumiservice
PACKDIR          := sdk
PROJECT          := github.com/pulumi/pulumi-pulumiservice
NODE_MODULE_NAME := @pulumi/pulumi-service
NUGET_PKG_NAME   := Pulumi.PulumiService

PROVIDER        := pulumi-resource-${PACK}
VERSION         ?= $(shell pulumictl get version)
PROVIDER_PATH   := provider
VERSION_PATH     := ${PROVIDER_PATH}/pkg/version.Version

SCHEMA_FILE     := provider/cmd/pulumi-resource-pulumiservice/schema.json
export GOPATH	:= $(shell go env GOPATH)

WORKING_DIR     := $(shell pwd)
TESTPARALLELISM := 4

# The pulumi binary to use during generation
PULUMI := .pulumi/bin/pulumi

export PULUMI_IGNORE_AMBIENT_PLUGINS = true

ensure::
	cd provider && go mod tidy
	cd sdk && go mod tidy
	cd examples && go mod tidy

gen::

build_sdks: dotnet_sdk go_sdk nodejs_sdk python_sdk java_sdk

gen_sdk_prerequisites: $(PULUMI)

schema: $(SCHEMA_FILE) # schema is human remember-able alias for $(SCHEMA_FILE)

.PHONY: $(SCHEMA_FILE)
$(SCHEMA_FILE): provider
	$(PULUMI) package get-schema $(WORKING_DIR)/bin/${PROVIDER} | \
		jq 'del(.version)' > $(SCHEMA_FILE)

.PHONY: provider
provider:
	(cd provider && VERSION=${VERSION} go generate cmd/${PROVIDER}/main.go)
	(cd provider && go build -o $(WORKING_DIR)/bin/${PROVIDER} -ldflags "-X ${PROJECT}/${VERSION_PATH}=${VERSION}" $(PROJECT)/${PROVIDER_PATH}/cmd/$(PROVIDER))

provider_debug::
	(cd provider && go build -o $(WORKING_DIR)/bin/${PROVIDER} -gcflags="all=-N -l" -ldflags "-X ${PROJECT}/${VERSION_PATH}=${VERSION}" $(PROJECT)/${PROVIDER_PATH}/cmd/$(PROVIDER))

test_provider::
	cd provider/pkg && go test -short -v -count=1 -cover -timeout 2h -parallel ${TESTPARALLELISM} ./...

dotnet_sdk: DOTNET_VERSION := $(shell pulumictl get version --language dotnet)
dotnet_sdk: gen_sdk_prerequisites
	rm -rf sdk/dotnet
	$(PULUMI) package gen-sdk $(SCHEMA_FILE) --language dotnet
	cd ${PACKDIR}/dotnet/&& \
		echo "${DOTNET_VERSION}" >version.txt && \
		dotnet build /p:Version=${DOTNET_VERSION}

go_sdk: gen_sdk_prerequisites
	rm -rf sdk/go
	$(PULUMI) package gen-sdk $(SCHEMA_FILE) --language go

nodejs_sdk: VERSION := $(shell pulumictl get version --language javascript)
nodejs_sdk: gen_sdk_prerequisites
	rm -rf sdk/nodejs
	$(PULUMI) package gen-sdk $(SCHEMA_FILE) --language nodejs
	cd ${PACKDIR}/nodejs/ && \
		yarn install && \
		yarn run tsc && \
		cp ../../README.md ../../LICENSE package.json yarn.lock ./bin/ && \
		sed -i.bak -e 's/\$${VERSION}/$(VERSION)/g' ./bin/package.json

python_sdk: PYPI_VERSION := $(shell pulumictl get version --language python)
python_sdk: gen_sdk_prerequisites
	rm -rf sdk/python
	$(PULUMI) package gen-sdk $(SCHEMA_FILE) --language python
	cp README.md ${PACKDIR}/python/
	cd ${PACKDIR}/python/ && \
		python3 setup.py clean --all 2>/dev/null && \
		rm -rf ./bin/ ../python.bin/ && cp -R . ../python.bin && mv ../python.bin ./bin && \
		sed -i.bak -e 's/^VERSION = .*/VERSION = "$(PYPI_VERSION)"/g' -e 's/^PLUGIN_VERSION = .*/PLUGIN_VERSION = "$(VERSION)"/g' ./bin/setup.py && \
		rm ./bin/setup.py.bak && \
		cd ./bin && python3 setup.py build sdist

GRADLE_DIR := $(WORKING_DIR)/.gradle
GRADLE := $(GRADLE_DIR)/gradlew
java_sdk: RESOURCE_FOLDER := src/main/resources/com/pulumi/pulumiservice
java_sdk: gen_sdk_prerequisites
	rm -rf sdk/java/{.gradle,build,src}
	$(PULUMI) package gen-sdk $(SCHEMA_FILE) --language java
	cp $(GRADLE_DIR)/settings.gradle sdk/java/settings.gradle
	cp $(GRADLE_DIR)/build.gradle sdk/java/build.gradle
	cd sdk/java && \
	mkdir -p $(RESOURCE_FOLDER) && \
	  echo "$(VERSION)" > $(RESOURCE_FOLDER)/version.txt && \
	  echo '{"resource": true,"name": "pulumiservice","version": "$(VERSION)"}' > $(RESOURCE_FOLDER)/plugin.json && \
	  PULUMI_JAVA_SDK_VERSION=0.10.0 $(GRADLE) --console=plain build && \
	  PULUMI_JAVA_SDK_VERSION=0.10.0 $(GRADLE) --console=plain publishToMavenLocal

.PHONY: build
build:: gen provider dotnet_sdk go_sdk nodejs_sdk python_sdk java_sdk

# Required for the codegen action that runs in pulumi/pulumi
only_build:: build

lint::
	for DIR in "provider" "sdk" "examples" ; do \
		pushd $$DIR && golangci-lint run -c ../.golangci.yml --timeout 10m && popd ; \
	done


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
	cd sdk/java && $(GRADLE) publishToMavenLocal


# Keep the version of the pulumi binary used for code generation in sync with the version
# of the dependency used by github.com/pulumi/pulumi-pulumiservice/provider

$(PULUMI): HOME := $(WORKING_DIR)
$(PULUMI): provider/go.mod
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
