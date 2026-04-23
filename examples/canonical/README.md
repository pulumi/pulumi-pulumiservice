# Canonical examples

These are realistic end-to-end scenarios — each a user story built from
multiple resources working together — not per-resource smoke tests.
Every scenario ships a YAML program as the primary reference. Three
scenarios additionally ship hand-written TypeScript / Python / Go /
C# / Java variants to exercise each language SDK's idioms and to
showcase the most common provider patterns.

| # | Scenario | YAML | Typed variants |
|---|---|---|---|
| 1 | Organization bootstrap | [`01-organization-bootstrap`](./01-organization-bootstrap/) | ts · py · go · cs · java |
| 2 | GitHub + OIDC CI | [`02-github-oidc-ci`](./02-github-oidc-ci/) | — |
| 3 | ESC rotation + hooks | [`03-esc-rotation-hooks`](./03-esc-rotation-hooks/) | — |
| 4 | Deployment pipeline (upsert, nested objects) | [`04-deployment-pipeline`](./04-deployment-pipeline/) | ts · py · go · cs · java |
| 5 | Audit log export | [`05-audit-log-export`](./05-audit-log-export/) | — |
| 6 | Self-hosted agent pool | [`06-self-hosted-agents`](./06-self-hosted-agents/) | — |
| 7 | Tiered team access (multiple resources of one kind) | [`07-tiered-team-access`](./07-tiered-team-access/) | ts · py · go · cs · java |
| 8 | Approval workflow with change gates | [`08-approval-change-gates`](./08-approval-change-gates/) | — |
| 9 | Stack tagging + per-tag webhooks | [`09-tag-based-routing`](./09-tag-based-routing/) | — |
| 10 | Template catalog | [`10-template-catalog`](./10-template-catalog/) | — |
| 11 | Cross-account OIDC federation | [`11-cross-cloud-oidc`](./11-cross-cloud-oidc/) | — |
| 12 | SAML/SSO with team sync | [`12-saml-team-sync`](./12-saml-team-sync/) | — |

## Why only three scenarios in every language?

The three showcase scenarios (01, 04, 07) are chosen to cover the
distinct SDK patterns end users actually encounter:

- **01** — basic CRUD with a single-field secret output.
- **04** — PATCH-based upsert with nested object configuration
  (`github`, `sourceContext`, `executorContext`, `operationContext`).
- **07** — multiple resources of one kind programmatically parameterised
  by a loop.

The other nine scenarios compose the same SDK primitives in different
arrangements; hand-writing them in five languages would duplicate
semantics without adding test coverage.

## Getting a typed variant for any scenario

Run `pulumi convert` against the YAML:

```bash
cd examples/canonical/02-github-oidc-ci
pulumi convert --from yaml --language typescript --out ./my-converted
# or --language python / go / csharp / java
```

See [the `pulumi convert` docs](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/)
for the full option surface.

## All examples use v2 tokens

Sub-module tokens throughout: `pulumiservice:orgs/teams:Team`,
`pulumiservice:orgs/agents:AgentPool`, `pulumiservice:stacks/deployments:Settings`,
etc. See [`docs/UPGRADE-v1-to-v2.md`](../../docs/UPGRADE-v1-to-v2.md)
for the rename table.
