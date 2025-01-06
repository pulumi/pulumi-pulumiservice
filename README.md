# Pulumi Service Provider

[![Slack](http://www.pulumi.com/images/docs/badges/slack.svg)](https://slack.pulumi.com)
[![NPM version](https://badge.fury.io/js/%40pulumi%2Fpulumiservice.svg)](https://www.npmjs.com/package/@pulumi/pulumiservice)
[![Python version](https://badge.fury.io/py/pulumi-pulumiservice.svg)](https://pypi.org/project/pulumi-pulumiservice)
[![NuGet version](https://badge.fury.io/nu/pulumi.pulumiservice.svg)](https://badge.fury.io/nu/pulumi.pulumiservice)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/pulumi/pulumi-pulumiservice/sdk/go/pulumiservice)](https://pkg.go.dev/github.com/pulumi/pulumi-pulumiservice/sdk/go)
[![License](https://img.shields.io/npm/l/%40pulumi%2Fpulumiservice.svg)](https://github.com/pulumi/pulumi-pulumiservice/blob/main/LICENSE)

Pulumi Service Provider for creating Pulumi Cloud resources.

The Pulumi Service Provider (PSP) is built on top of the [Pulumi Cloud REST API](https://www.pulumi.com/docs/pulumi-cloud/reference/cloud-rest-api/), allowing Pulumi customers to create Pulumi Cloud resources using Pulumi programs. That includes Stacks, Environments, Teams, Tokens, Webhooks, Tags, Deployment Settings, Deployment Schedules and much more! Pulumi Service Provider is especially powerful when used in combination with the [Automation API](https://pulumi.com/automation).

For a full list of supported resources, visit the [Pulumi Registry](https://www.pulumi.com/registry/packages/pulumiservice/). For the REST API reference documentation, visit [Pulumi Cloud API Documentation](https://www.pulumi.com/docs/pulumi-cloud/reference/cloud-rest-api/).

For now, PSP is in public preview, but we're working on releasing v1.0 in the near future.

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

### Java

To use from Java, add an entry to your `build.gradle` file:

    $ implementation 'com.pulumi:pulumiservice:%Fill in latest version from the badge up top%'
    
Or to your `pom.xml` file:

    $   <dependency>
    $       <groupId>com.pulumi</groupId>
    $       <artifactId>pulumiservice</artifactId>
    $       <version>%Fill in latest version from the badge up top%</version>
    $   </dependency>

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
