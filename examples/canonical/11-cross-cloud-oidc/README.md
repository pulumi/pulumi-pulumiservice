# 11 — Cross-cloud OIDC federation

Federate Pulumi Cloud identity into AWS, Azure, and GCP so deployments
use short-lived OIDC-minted credentials — no static cloud access keys
anywhere.

## Resources

- `pulumiservice:orgs/oidc:Issuer` × 3 — one trust relationship per cloud

## Why this pattern

Static cloud access keys in CI are the number-one credential-leak
vector. OIDC federation collapses the credential surface:

- No long-lived AWS access keys, Azure client secrets, or GCP service
  account JSON keys.
- A compromised deployment pipeline is time-boxed — stolen tokens
  expire in an hour, not next year.
- Key rotation stops being a quarterly operational pain.

The setup has a fixed cost (federation config on each cloud side) and
a permanent reward.

## Per-cloud setup (outside Pulumi)

- **AWS** — create an IAM OIDC provider pointing at `https://oidc.pulumi.com`,
  then a role with a trust policy federated to it.
  See [AWS federated users docs](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc.html).
- **Azure** — create a federated identity credential on an Azure AD app,
  issuer `https://oidc.pulumi.com`, subject `pulumi.com/<org-name>`.
- **GCP** — create a Workload Identity Pool + OIDC provider; bind
  Pulumi identities to a Google service account.

## Run

```
pulumi stack init dev
pulumi config set organizationName acme-corp
pulumi config set awsRoleArn arn:aws:iam::...:role/PulumiOidcDeployer
pulumi up
```

## Other languages

Use `pulumi convert` — see the
[convert docs](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
