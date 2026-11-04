# Regenerate the Example Usage markdown that infer resources and functions
# embed in their descriptions, from the Pulumi YAML sources in each package's
# examples/ directory. See provider/tools/gen-examples.
#
# Lives here, not in the Makefile, because ci-mgmt regenerates the Makefile and
# drops any target it doesn't know about. Files in .mk/ survive regen.
#
# `pulumi convert` loads the provider plugin, so build it first and put it on
# PATH. That binary embeds the old markdown, so remove it afterward: `make
# provider` only builds a missing binary.
#
# PHONY because the repo root has an examples/ directory, which would otherwise
# make this target always up to date.
examples: export PATH := $(WORKING_DIR)/bin:$(PATH)
examples: bin/$(PROVIDER)
	cd provider && go generate ./pkg/functions ./pkg/resources
	rm -f bin/$(PROVIDER)
.PHONY: examples
