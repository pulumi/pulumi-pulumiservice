# YAML RBAC pass-through example

Demonstrates the pass-through grammar for `OrganizationRole.permissions` with three role shapes:

1. **`kind: PermissionDescriptorCompose`** — composing two base roles into a third (the customer's UI-import case).
2. **`on: { team: <id> }`** — the new `team` entity-type sugar on a structured `kind: allow` role.
3. **`kind: PermissionDescriptorCondition` with `And(Equal, Equal)`** — a non-collapsible boolean expressed via the pass-through grammar.

To run:

```bash
pulumi config set digits 12345
pulumi up
```

To convert this example to another language, see the [`pulumi convert`](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/) docs.
