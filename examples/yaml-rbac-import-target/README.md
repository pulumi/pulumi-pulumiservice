# yaml-rbac-import-target

Empty fixture used by `TestYamlRbacComposeImport` as the destination of a
`pulumi import` invocation. The test verifies that a custom role authored
with `__type: PermissionDescriptorCompose` imports cleanly without the
prior "unknown `__type`" error from earlier translators.

This fixture has no resources of its own — the test populates it via
`pulumi import` at runtime. See `examples/examples_yaml_test.go`'s
`TestYamlRbacComposeImport` for the full flow.

## Converting to other languages

Pulumi can convert this YAML program (post-import) to other supported
languages with `pulumi convert` — see the
[`pulumi convert` docs](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/)
for the available targets and flags.
