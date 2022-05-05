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

* [Webhooks](https://pulumi.com/docs/intro/pulumi-service/webhooks)
* [Teams](https://pulumi.com/docs/intro/pulumi-service/teams)
* [StackTags](https://pulumi.com/docs/reference/cli/pulumi_stack_tag)
* [AccessTokens](https://pulumi.com/docs/intro/pulumi-service/accounts)

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

To provision resources with the Pulumi Service provider, you need to have Pulumi Service credentials. Pulumi Service maintains documentation on how to create access tokens [here](https://www.pulumi.com/docs/intro/pulumi-service/accounts/#access-tokens). 

While you can use this provider to provision access tokens, you'll still need to have an access token available to generate an access token with the provider.

### Set environment variables

Once you have an access token, its easy to set the environment variables. The Pulumi Service Provider uses the same environment variables as Pulumi does.

```bash
$ export PULUMI_ACCESS_TOKEN=<PULUMI_ACCESS_TOKEN>

$ export PULUMI_BACKEND_URL=<PULUMI_BACKEND_URL> # For self hosted customers. defaults to https://api.pulumi.com
```

### Configuration Options

Use `pulumi config set pulumiservice:<option>` or pass options to the [constructor of `new pulumiservice.Provider`](https://pulumi.com/registry/packages/pulumiservice/api-docs/provider).

| Option | Required/Optional | Description |
|-----|------|----|
| `accessToken`| Required | [Pulumi Service Access Tokens](https://www.pulumi.com/docs/intro/pulumi-service/accounts/#access-tokens) |
| `apiUrl`| Optional | Allows overriding default [Pulumi Service API URL](https://www.pulumi.com/docs/reference/service-rest-api) for [self hosted customers](https://www.pulumi.com/docs/guides/self-hosted/).


## Examples
```typescript
import * as aws from "@pulumi/awsx"
import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const team = new service.Team("team", {
    name: "pulumi-up-team",
    displayName: "PulumiUP Team",
    description: "This was created with Pulumi",
    organizationName: "pulumi-up",
    teamType: "pulumi",
    members: [
        "piers",
        "myles",
        "meagan"
    ],
});

export const members = team.members;
```

Check out the [examples/](examples/) directory for more examples.

