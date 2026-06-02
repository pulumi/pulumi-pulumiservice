# Declare the provider binary's real inputs so incremental builds are correct.
# Under useProviderBinarySchemaGen, ci-mgmt generates `bin/$(PROVIDER)` with no
# prerequisites, so make never rebuilds it on source changes. This prerequisite-
# only rule (no recipe) is merged into the generated rule without overriding it.
# Lives here, not inline in the Makefile, to survive ci-mgmt regen (ci-mgmt#2285).
#
# Keep in sync with the provider's //go:embed directives. Do NOT list the
# generated provider/cmd/$(PROVIDER)/schema.json: it is built FROM the binary,
# so it would create a circular dependency.
PROVIDER_SOURCES := $(shell find provider/cmd provider/pkg -name '*.go') \
                    provider/pkg/cloud/spec.json \
                    provider/pkg/cloud/metadata.json \
                    provider/pkg/provider/manual-schema.json \
                    provider/pkg/provider/README.md \
                    go.mod go.sum

bin/$(PROVIDER): $(PROVIDER_SOURCES)
