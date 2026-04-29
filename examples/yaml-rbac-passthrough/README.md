# YAML RBAC pass-through example

Demonstrates the pass-through grammar for `OrganizationRole.permissions` by
composing two simple roles into a third using `kind: PermissionDescriptorCompose`.

To run:

```bash
pulumi config set digits 12345
pulumi up
```

To convert this example to another language, see the [`pulumi convert`](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/) docs.
