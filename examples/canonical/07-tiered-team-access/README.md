# 07 — Tiered team access

Three teams with three permission tiers (admin / write / read) applied
across two stacks. The RBAC baseline for any non-trivial organization.

## Resources

- `pulumiservice:orgs/teams:Team` × 3
- `pulumiservice:stacks/permissions:TeamStackPermission` × 5

## Why this pattern

One "everyone-can-do-anything" team is a policy violation waiting to
happen. The minimal sustainable shape is:

- **admins** — break-glass, small, rotated regularly
- **service-owners** — can deploy their own thing, not others
- **developers** — read everywhere, deploy nothing without an explicit grant

Pulumi's `read` / `write` / `admin` levels map cleanly: read gives
state/config visibility, write adds `pulumi up`, admin adds permission
changes and stack deletion.

## Run

```
pulumi stack init prod
pulumi config set organizationName acme-corp
pulumi up
```

## Other languages

Use `pulumi convert` — see the
[convert docs](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
