# 02 — GitHub + OIDC CI

Configure Pulumi Cloud to accept deployments from GitHub Actions via OIDC.
No long-lived Pulumi access tokens in CI.

## Resources

- `pulumiservice:orgs/oidc:Issuer` — trusts GitHub Actions' OIDC provider

## Why this pattern

Long-lived access tokens are the highest-risk credential in any Pulumi
deployment setup. OIDC replaces them with workflow-scoped tokens that
expire after a few minutes and carry claims (repo, branch, environment)
Pulumi Cloud can gate on. A leaked token from a specific workflow run
is useless outside that context.

## On the GitHub side

Add this to your workflow:

```yaml
permissions:
  id-token: write   # required to mint the OIDC token
  contents: read
steps:
  - uses: pulumi/auth-actions@v1
    with:
      organization: acme-corp
      requested-token-type: urn:pulumi:token-type:access_token:team
      scope: user:readers team:deployers
  - uses: pulumi/actions@v5
    with:
      command: up
      stack-name: acme-corp/infra/prod
```

## Run

```
pulumi stack init infra
pulumi config set organizationName acme-corp
pulumi up
```

## Other languages

Use `pulumi convert` — see the
[convert docs](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
