# Pulumi Service Provider

[![Slack](http://www.pulumi.com/images/docs/badges/slack.svg)](https://slack.pulumi.com)
[![NPM version](https://badge.fury.io/js/%40pulumi%2Fpulumiservice.svg)](https://www.npmjs.com/package/@pulumi/pulumiservice)
[![Python version](https://badge.fury.io/py/pulumi-pulumiservice.svg)](https://pypi.org/project/pulumi-pulumiservice)
[![NuGet version](https://badge.fury.io/nu/pulumi.pulumiservice.svg)](https://badge.fury.io/nu/pulumi.pulumiservice)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice)](https://pkg.go.dev/github.com/pulumi/pulumi-pulumiservice/sdk/go)
[![License](https://img.shields.io/npm/l/%40pulumi%2Fpulumiservice.svg)](https://github.com/pulumi/pulumi-pulumiservice/blob/main/LICENSE)

A Pulumi Resource Provider for The Pulumi Service.

The Pulumi Service Provider is built on top of the [Pulumi Service REST API](https://pulumi.com/docs/reference/service-rest-api) which is another feature available to our customers to programmatically configuring the Pulumi Service. The Pulumi Service REST API includes functionality to interact with and manipulate any kind of metadata managed by Pulumi. That includes Projects and Stacks, Previews and Updates, Organizations and Audit Logs. We have already seen Cloud Engineering teams using the Pulumi REST API to build all sorts of custom functionality. These new capabilities are especially powerful when used in combination with the [Automation API](https://pulumi.com/automation).

#### Supported Resources

- [Webhooks](https://pulumi.com/docs/intro/pulumi-service/webhooks)
- [Teams](https://pulumi.com/docs/intro/pulumi-service/teams)
  - You can grant a team access to stacks via the `TeamStackPermission` resource
- [StackTags](https://pulumi.com/docs/reference/cli/pulumi_stack_tag)
- [AccessTokens](https://pulumi.com/docs/intro/pulumi-service/accounts)

For reference documentation, visit the [Pulumi Service API Documentation](https://www.pulumi.com/registry/packages/pulumiservice/api-docs/) on the [Pulumi Registry](https://www.pulumi.com/registry/packages/pulumiservice/)

## Installing

This package is available in many languages in the standard packaging formats.

### Node.js (Java/TypeScript)

To use from JavaScript or TypeScript in Node.js, install using either `npm`:

    $ npm install @pulumi/pulumiservice

or `yarn`:

    $ yarn add @pulumi/pulumiservice

### Python

To use from Python, install using `pip`:

    $ pip install pulumi_pulumiservice

### Go

To use from Go, use `go get` to grab the latest version of the library

    $ go get github.com/pulumi/pulumi-pulumiservice/sdk/go

### .NET

To use from .NET, install using `dotnet add package`:

    $ dotnet add package Pulumi.PulumiService

## Setup

Ensure that you have ran `pulumi login`. Run `pulumi whoami` to verify that you are logged in.

### Configuration Options

Use `pulumi config set pulumiservice:<option>` or pass options to the [constructor of `new pulumiservice.Provider`][1].

| Option        | Environment Variable Name | Required/Optional | Description                                                                           |
|---------------|---------------------------|-------------------|---------------------------------------------------------------------------------------|
| `accessToken` | `PULUMI_ACCESS_TOKEN`     | Optional          | Overrides [Pulumi Service Access Tokens][2]                                           |
| `apiUrl`      | `PULUMI_BACKEND_URL`      | Optional          | Allows overriding default [Pulumi Service API URL][3] for [self hosted customers][4]. |
|               |                           |                   |                                                                                       |

## Examples

```typescript
import * as aws from "@pulumi/awsx"
import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const team = new service.Team("team", {
    name: "pulumi-service-team",
    displayName: "Pulumi Service",
    description: "The Pulumi Service Team",
    organizationName: "pulumi",
    teamType: "pulumi",
    members: [
        "piers",
        "bryce",
        "casey"
        "evan",
        "devon",
        "meagan"
        "myles",
        "steve"
    ],
});

export const members = team.members;
```

Check out the [examples/](examples/) directory for more examples.

[1]: https://pulumi.com/registry/packages/pulumiservice/api-docs/provider
[2]: https://www.pulumi.com/docs/intro/pulumi-service/accounts/#access-tokens
[3]: https://www.pulumi.com/docs/reference/service-rest-api
[4]: https://www.pulumi.com/docs/guides/self-hosted/
