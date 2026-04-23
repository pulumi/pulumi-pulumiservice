# Upgrading from `@pulumi/pulumiservice` v1 to v2

v2 is a full rewrite of the Pulumi Service Provider. It is **generated
from the Pulumi Cloud OpenAPI specification** rather than hand-coded,
which means v2 covers the full Pulumi Cloud API surface — around 10×
more resources than v1 — and stays in sync with new Pulumi Cloud
features automatically.

The API surface you care about almost certainly exists in both versions,
but **resource tokens and a small number of property names changed** so
we could organize the much larger v2 surface sensibly. This guide walks
through the migration.

**Breaking change classification:** v2 is a major version bump. Programs
written against v1 will not compile against v2 without the token and
property renames below. Your Pulumi state (the resources themselves in
Pulumi Cloud) is **not affected** — no resource is recreated, modified,
or deleted as part of the upgrade.

## Quick summary

1. Bump the package in your project's manifest (`package.json`,
   `requirements.txt`, etc.) to `@pulumi/pulumiservice@^2.0.0` (or the
   language equivalent).
2. Update every `pulumiservice:index:*` token to its new sub-module token
   (see [Resource rename table](#resource-rename-table)).
3. Rename any properties that moved (mostly none).
4. Run `pulumi state rename` for each resource so Pulumi's state mirror
   matches the new tokens (see [State rename recipes](#state-rename-recipes)).
5. Run `pulumi preview`. The diff should show token renames only and no
   property changes.
6. Apply with `pulumi up`.

Nothing in Pulumi Cloud changes — this is a client-side rename.

## What changed and why

**From flat to sub-modules.** v1 put every resource in the flat
`pulumiservice:index:*` namespace — 20 resources, all siblings. v2 has
on the order of 100 resources, organized by domain:

| v1 | v2 |
|---|---|
| `pulumiservice:index:AgentPool` | `pulumiservice:orgs/agents:AgentPool` |
| `pulumiservice:index:Stack` | `pulumiservice:stacks:Stack` |
| `pulumiservice:index:Environment` | `pulumiservice:esc:Environment` |
| `pulumiservice:index:Webhook` | `pulumiservice:stacks/hooks:Webhook` |
| `pulumiservice:index:Team` | `pulumiservice:orgs/teams:Team` |

The sub-modules follow the Pulumi Cloud URL structure (`/api/orgs/...`,
`/api/esc/...`, `/api/stacks/...`) and match the feel of every other
native Pulumi provider (`azure-native:storage:StorageAccount`,
`google-native:container/v1:Cluster`).

**Stutter-free type names.** In v1, type names carried their own
context (`InsightsAccount`, `OidcIssuer`, `StackTag`) because the flat
namespace offered none. In v2, the sub-module provides context, so
type names drop the redundant prefix:

| v1 token | v2 token |
|---|---|
| `pulumiservice:index:OidcIssuer` | `pulumiservice:orgs/oidc:Issuer` |
| `pulumiservice:index:InsightsAccount` | `pulumiservice:orgs/insights:Account` |
| `pulumiservice:index:StackTag` | `pulumiservice:stacks/tags:Tag` |
| `pulumiservice:index:DeploymentSchedule` | `pulumiservice:stacks/deployments:Schedule` |

The full list is in the [resource rename table](#resource-rename-table).
A handful of v1 names survive unchanged in v2 because the un-prefixed
name would be too generic or clash with `pulumi.Provider` (e.g.,
`AgentPool`, `PolicyGroup`, `IdentityProvider`).

**Schema derived from the spec.** In v1 a new Pulumi Cloud endpoint
required a human to hand-code a resource in Go. In v2 an endpoint
shows up as a new entry in the coverage report, gets wired into
`resource-map.yaml`, and the generator does the rest. The payoff:
every new Pulumi Cloud feature is a PR away from being in the SDK.

## Resource rename table

v1 tokens on the left, v2 tokens on the right. Any resource not listed
here had no v1 equivalent — it's new in v2.

| v1 token | v2 token |
|---|---|
| `pulumiservice:index:AccessToken` | `pulumiservice:orgs/tokens:AccessToken` |
| `pulumiservice:index:AgentPool` | `pulumiservice:orgs/agents:AgentPool` |
| `pulumiservice:index:ApprovalRule` | `pulumiservice:changegates:ApprovalRule` |
| `pulumiservice:index:DeploymentSchedule` | `pulumiservice:stacks/deployments:Schedule` |
| `pulumiservice:index:DeploymentSettings` | `pulumiservice:stacks/deployments:Settings` |
| `pulumiservice:index:DriftSchedule` | `pulumiservice:stacks/deployments:DriftSchedule` |
| `pulumiservice:index:Environment` | `pulumiservice:esc:Environment` |
| `pulumiservice:index:EnvironmentRotationSchedule` | `pulumiservice:esc/schedules:Rotation` |
| `pulumiservice:index:EnvironmentVersionTag` | `pulumiservice:esc/versions:Tag` |
| `pulumiservice:index:InsightsAccount` | `pulumiservice:orgs/insights:Account` |
| `pulumiservice:index:OidcIssuer` | `pulumiservice:orgs/oidc:Issuer` |
| `pulumiservice:index:OrgAccessToken` | `pulumiservice:orgs/tokens:OrgAccessToken` |
| `pulumiservice:index:PolicyGroup` | `pulumiservice:orgs/policies:PolicyGroup` |
| `pulumiservice:index:Stack` | `pulumiservice:stacks:Stack` |
| `pulumiservice:index:StackTag` | `pulumiservice:stacks/tags:Tag` |
| `pulumiservice:index:StackTags` | `pulumiservice:stacks/tags:Tags` |
| `pulumiservice:index:Team` | `pulumiservice:orgs/teams:Team` |
| `pulumiservice:index:TeamAccessToken` | `pulumiservice:orgs/tokens:TeamAccessToken` |
| `pulumiservice:index:TeamStackPermission` | `pulumiservice:stacks/permissions:TeamStackPermission` |
| `pulumiservice:index:TemplateSource` | `pulumiservice:orgs/templates:Source` |
| `pulumiservice:index:TtlSchedule` | `pulumiservice:stacks/deployments:TtlSchedule` |
| `pulumiservice:index:Webhook` | `pulumiservice:stacks/hooks:Webhook` |

**Removed in v2.0** (no equivalent yet; awaiting public API endpoints):

- `pulumiservice:index:TeamEnvironmentPermission` — returns in a later
  2.x when the public spec exposes its endpoints.

In each language SDK, the import path changes accordingly. A concrete
TypeScript example:

```ts
// v1
import * as ps from "@pulumi/pulumiservice";
const pool = new ps.AgentPool("pool", { ... });

// v2
import * as ps from "@pulumi/pulumiservice";
const pool = new ps.orgs.agents.AgentPool("pool", { ... });
```

Python:

```python
# v1
import pulumi_pulumiservice as ps
pool = ps.AgentPool("pool", ...)

# v2
from pulumi_pulumiservice.orgs import agents
pool = agents.AgentPool("pool", ...)
```

## Property renames

v2 preserves v1 property names across the board. No resource in the
v1→v2 mapping above renames a property. If your v1 program sets
`webhook.payloadUrl`, `team.teamType`, or `agentPool.forceDestroy`,
v2 accepts those names verbatim.

**New property conventions.** Across v2:

- Output-only server-assigned IDs are exposed consistently as
  `<resource>Id` (`agentPoolId`, `tokenId`, `issuerId`, `gateId`, …)
  even when the underlying API response field is just `"id"`.
- Secret properties (`tokenValue`, `webhook.secret`, …) keep their v1
  names and are marked `secret: true` in the schema (same as v1).
- Optional write-only properties (e.g. one-time setup tokens) are
  unchanged from v1.

## State rename recipes

Pulumi's state stores each resource keyed by URN, which includes the
type token. When you change a token, run `pulumi state rename` so
Pulumi's records mirror the new code without re-creating the resource.

Do this **after** you've updated your code to v2 but **before** you
run `pulumi up`.

```bash
# One-line form, per resource:
pulumi state rename <name> --type-token pulumiservice:orgs/agents:AgentPool

# Or for bulk renames, run `pulumi state` in an interactive session:
pulumi state edit
```

Template this across your stack — the [resource rename table](#resource-rename-table)
is exhaustive for v1. For large programs, the pattern is:

```bash
for urn in $(pulumi stack --show-urns | grep "pulumiservice:index:AgentPool"); do
  pulumi state rename --urn "$urn" --type-token pulumiservice:orgs/agents:AgentPool
done
```

## What does NOT require a re-creation

All of the following are preserved **by design** across the v1 → v2
upgrade:

- Your Pulumi Cloud resources themselves (agent pools, stacks,
  environments, teams, webhooks, tokens, …) are untouched.
- Resource IDs keep their format and value. A v1 AgentPool with ID
  `org/pool-name/pool-abc-123` is a v2 AgentPool with the same ID.
- Access tokens' values don't rotate; webhooks don't re-register.
- Stack outputs exported by other stacks referencing these resources
  remain valid.

## What's new in v2

Beyond the rename, v2 exposes Pulumi Cloud surfaces that v1 couldn't:

- **`pulumiservice:orgs/audit:LogExport`** — continuous audit-log
  export to S3 / Splunk / Datadog.
- **`pulumiservice:changegates:ApprovalRule`** — block production
  deployments behind a human approver team (SOX / HIPAA).
- **`pulumiservice:integrations:*`** — GitHub, GitLab, Bitbucket,
  Azure DevOps integration resources.
- **Most of `pulumiservice:esc/*`** — schedules, hooks, version tags,
  drafts, cloud-setup wizards.
- **`pulumiservice:orgs/teams:Team` externalMembershipBinding** —
  bind teams to IdP groups for SAML/SSO auto-sync.

See `examples/canonical/` for end-to-end programs demonstrating these.

## Troubleshooting

**`preview` shows both `delete` and `create` for every resource.** You
forgot to run `pulumi state rename` (or `pulumi state edit`). The engine
sees the old-token URN in state and the new-token URN in your code as
two different resources. Run the rename, re-preview, and the diff
should collapse to an empty change.

**A resource's property is undefined / missing.** A small number of v1
properties were only ever populated in certain states. If you relied
on one, see [Property renames](#property-renames) and the v2 SDK docs
for the resource — the underlying Pulumi Cloud field is still present,
but may be named more consistently.

**My CI still uses v1.** Pin your CI's `@pulumi/pulumiservice` to
`^1.x` while you migrate, then flip to `^2.0.0` once your main branch
is green on v2.

## Getting help

- [File a GitHub issue](https://github.com/pulumi/pulumi-pulumiservice/issues) —
  include the v1→v2 rename step you're stuck on.
- [Pulumi Community Slack](https://slack.pulumi.com/) — `#general`.
