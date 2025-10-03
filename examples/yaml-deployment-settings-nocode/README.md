# No-Code Deployment Settings Example

This example demonstrates how to configure deployment settings for no-code deployments using template sources.

## Overview

No-code deployments allow you to deploy Pulumi programs from organization templates without managing a version control system (VCS) repository. This enables rapid infrastructure provisioning using pre-configured templates.

## Key Features

- **Template Source**: Uses a template URL instead of a Git repository
- **No VCS Required**: Deploy directly from templates without GitHub/GitLab integration
- **Quick Provisioning**: Instant deployment without managing underlying IaC code

## Configuration

The example configures deployment settings with:

- `sourceContext.template.sourceUrl`: URL of the template to deploy
- `executorContext`: Custom executor image for deployments
- `operationContext`: Environment variables and execution options
- `cacheOptions`: Dependency caching configuration

## Running the Example

```bash
pulumi up
```

## Converting to Other Languages

This YAML example can be converted to other programming languages using [`pulumi convert`](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/):

```bash
# Convert to TypeScript
pulumi convert --language typescript --out ../ts-deployment-settings-nocode

# Convert to Python
pulumi convert --language python --out ../py-deployment-settings-nocode

# Convert to Go
pulumi convert --language go --out ../go-deployment-settings-nocode
```

## Learn More

- [Pulumi Deployments Documentation](https://www.pulumi.com/docs/pulumi-cloud/deployments/)
- [No-Code Deployments](https://www.pulumi.com/blog/announcing-pulumi-idp/#no-code)
- [Organization Templates](https://www.pulumi.com/docs/pulumi-cloud/developer-platforms/templates/)
