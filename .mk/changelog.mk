# Add a changelog entry: prompts for the kind, the description, and an optional
# issue or PR number, then writes a fragment to .changes/unreleased/.
#
# Lives here, not in the Makefile, because ci-mgmt regenerates the Makefile and
# drops any target it doesn't know about. Files in .mk/ survive regen.
changelog:
	mise exec -- changie new
.PHONY: changelog
