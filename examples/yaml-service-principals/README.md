# yaml-service-principals

Demonstrates service principals, the named machine identities of a Pulumi
Cloud organization, and the identity a stack's deployments run as:

- **`pulumiservice:ServicePrincipal`**: creates a service principal with the
  built-in Reader role, looked up with `pulumiservice:getOrganizationRole`.
- **`pulumiservice:OrganizationMemberEntityAccess`**: grants the service
  principal open access to an ESC environment and write access to a stack,
  using its `login` as the username.
- **`pulumiservice:DeploymentSettings`** with
  `operationContext.servicePrincipal`: makes the stack's deployments run as the
  service principal. Pulumi Cloud requires the service principal to have write
  access to the stack, so the settings depend on that grant.
- **`pulumiservice:DeploymentSettings`** with `operationContext.roleId`: makes
  another stack's deployments run with the built-in Maintainer role.

## Prerequisites

The organization must have service principals enabled.

## Convert to another language

Use `pulumi convert` to translate this program to TypeScript, Python, Go,
C#, or Java:

```bash
pulumi convert --from yaml --language typescript --out ../ts-service-principals
```

See the [`pulumi convert` documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/)
for more options.
