# yaml-service-principals

Demonstrates service principals, the named machine identities of a Pulumi
Cloud organization:

- **`pulumiservice:ServicePrincipal`** – create a service principal with the
  built-in Reader role, looked up with `pulumiservice:getOrganizationRole`.
- **`pulumiservice:ServicePrincipalAccessToken`** – create an expiring access
  token that authenticates as the service principal. Its `value` is a secret
  output.
- **`pulumiservice:OrganizationMemberEntityAccess`** – grant the service
  principal access to one ESC environment, using its `login` as the username.

## Prerequisites

- The service principal API must be available to the organization.
- The organization's plan must include team access tokens.

## Convert to another language

Use `pulumi convert` to translate this program to TypeScript, Python, Go,
C#, or Java:

```bash
pulumi convert --from yaml --language typescript --out ../ts-service-principals
```

See the [`pulumi convert` documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/)
for more options.
