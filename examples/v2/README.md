# v2 example PCL programs

PCL (Pulumi Configuration Language) representations of the existing v1 examples
under `examples/ts-*`, `examples/py-*`, etc., rewritten to use the new
`pulumiservice:v2:*` resource types.

PCL is Pulumi's intermediate representation for code conversion. Each example
here is the canonical version; convert to your target language with
`pulumi convert --from pcl --language <lang>`. One source, all SDKs.

## Layout

```
examples/v2/
  README.md                  ← this file
  webhooks/main.pp           ← Org-scoped webhooks (was: ts-webhooks)
  oidc-issuer/main.pp        ← OIDC issuers (was: ts-oidc-issuer)
  …
```

## Generating a runnable program

From inside an example directory:

```bash
# TypeScript
pulumi convert --from pcl --language typescript --out ../../v2-ts-webhooks

# Python
pulumi convert --from pcl --language python --out ../../v2-py-webhooks

# Pulumi YAML
pulumi convert --from pcl --language yaml --out ../../v2-yaml-webhooks

# Go / .NET / Java similar
pulumi convert --from pcl --language go     --out ../../v2-go-webhooks
pulumi convert --from pcl --language csharp --out ../../v2-cs-webhooks
pulumi convert --from pcl --language java   --out ../../v2-java-webhooks
```

Then in the generated directory, link the local SDK and run:

```bash
cd ../v2-ts-webhooks
yarn link "@pulumi/pulumiservice"          # or pip install -e for Python, etc.
pulumi stack init dev
pulumi config set v2-webhooks:secretValue super-secret --secret
pulumi up
```

## Coverage map: which v1 examples have a v2 PCL counterpart

| v1 example | v2 PCL | Notes |
|---|---|---|
| ts-webhooks | webhooks/ | v1 single Webhook → v2 OrganizationWebhook (StackWebhook + Webhook_esc_environments not yet ported) |
| ts-oidc-issuer | oidc-issuer/ | v2 doesn't expose `policies`; that's a separate resource pending RBAC polymorphism support |
| ts-teams | — | v2 Team has no inputs (Update-as-create with empty request body schema). Needs metadata.fields work first. |
| ts-deployment-settings | (todo) | upsert via PatchDeploymentSettings; should port cleanly |
| ts-environments | (todo) | v2 splits into `Environment_esc_environments` / `Environment_preview_environments` |
| ts-rbac | (partial) | Role available; OrganizationMember + TeamRoleAssignment not in v2 yet |
| ts-approval-rules | (todo) | PolicyGroup available; ApprovalRule not in v2 |
| ts-schedules | (todo) | ScheduledDeployment + EnvironmentSchedule available; DriftSchedule + TtlSchedule not |
| ts-stack-tags | — | StackTag has no clean Create+Get pair in spec |
| ts-access-tokens | — | Token endpoints don't form a CRUD set |
| ts-template-source | — | Not in v2 metadata |
| ts-team-stack-permissions | — | Not in v2 metadata |
| ts-insights-account-invokes | — | InsightsAccount not in v2 |

Items marked **(todo)** can be added by porting the v1 example to v2 PCL —
the v2 surface is wired in `provider/pkg/cloud/metadata.json` and the SDK
already exists under `sdk/<lang>/v2/`. Items marked **—** require either
expanding `metadata.json` to expose the missing v2 surface or accepting that
the resource has no v2 equivalent yet.

## Caveat: v2 transport not yet wired

`pulumi up` against v2 resources will currently fail at the first CRUD call
with `rest: no transport resolver registered`. The v2 dispatcher needs the
authenticated HTTP client to be plumbed through Configure (TODO). PCL
conversion + schema validation work today; live deploys do not.
