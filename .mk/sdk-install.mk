# Publish the locally-built Java SDK to ~/.m2 so the java api-test examples can
# resolve it. Lives here, not inline in the generated Makefile, to survive
# ci-mgmt regen: install_java_sdk is generated empty, so this only adds a
# prerequisite (merged, no recipe override) plus a recipe on the .make sentinel.
install_java_sdk: .make/install_java_sdk
.make/install_java_sdk: .make/build_java
	cd sdk/java/ && gradle --console=plain publishToMavenLocal
	@touch $@
