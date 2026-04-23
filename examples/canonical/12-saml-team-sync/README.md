# 12 — SAML/SSO with team sync

Bind Pulumi teams to IdP groups so membership follows from your
existing Okta/Azure AD/Google Workspace source of truth.

## Resources

- `pulumiservice:orgs/teams:Team` × 3 — bound to external IdP group IDs

## Why this pattern

Per-tool user administration scales poorly: every new tool adds another
place to add/remove users when someone joins or leaves. Mirroring
IdP groups into Pulumi teams means:

- **Onboarding is automatic.** New hire added to the "Platform" Okta
  group? Next login they're in the Pulumi `platform` team with its
  permissions.
- **Offboarding is automatic.** User leaves the IdP? All Pulumi access
  disappears on next session refresh.
- **Audits are trivial.** "Who has write access to production stacks?"
  becomes "who is in the IdP group mapped to production-deployers?"

## Prerequisites

SAML SSO must already be configured on the Pulumi Cloud org. The
`groups` claim (or equivalent) must be present in the SAML assertion.
See [Pulumi Cloud SAML docs](https://www.pulumi.com/docs/pulumi-cloud/admin/saml-sso/).

## Finding group IDs

- **Okta** — admin console → Directory → Groups → click group →
  Profile tab → "ID" field.
- **Azure AD** — Entra ID → Groups → select group → "Object ID".
- **Google Workspace** — admin console → Directory → Groups → group →
  URL contains the group ID.

## Run

```
pulumi stack init prod
pulumi config set organizationName acme-corp
pulumi config set platformGroupId okta-group-id:00g...
pulumi up
```

Then have a user in the IdP group log into Pulumi Cloud; they should
appear in the corresponding Pulumi team within a session.

## Other languages

Use `pulumi convert` — see the
[convert docs](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
