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

## Resource surfaces

Resources are organized into two surfaces under one package:

- **v0 (package root)** — Mature, hand-maintained resources accessible directly off the package import (e.g. `pulumiservice.Stack`). **In maintenance mode**: bug fixes and security updates only; no new resources or features. Existing programs continue to work without any code changes.
- **api (`pulumiservice.api`)** — **Preview.** Actively developed, generated at runtime from the public Pulumi Cloud OpenAPI specification. Resource shape and module layout may change before GA; not yet recommended for production. Coverage expands as new operations are mapped from the spec.

Resources from both surfaces can be used in the same program. There is no forced migration: existing users stay on v0 indefinitely, or migrate individual resources to the api surface by updating their IaC code (resource type + input shape) and adding Pulumi `aliases` on the new api declaration so state rebinds in place. v0 has known coverage gaps relative to the full Cloud API; the api surface closes those gaps over time.

## Installing

This package is available in many languages in the standard packaging formats.

### Node.js (Javascript/TypeScript)

To use from JavaScript or TypeScript in Node.js, install using either `npm`:

```sh
npm install @pulumi/pulumiservice
```

or `yarn`:

```sh
yarn add @pulumi/pulumiservice
```

### Python

To use from Python, install using `pip`:

```sh
pip install pulumi_pulumiservice
```

### Go

To use from Go, use `go get` to grab the latest version of the library

```sh
go get github.com/pulumi/pulumi-pulumiservice/sdk/go
```

### .NET

To use from .NET, install using `dotnet add package`:

```sh
dotnet add package Pulumi.PulumiService
```

### Java

To use from Java, add an entry to your `build.gradle` file:

```groovy
implementation 'com.pulumi:pulumiservice:%Fill in latest version from the badge up top%'
```

Or to your `pom.xml` file:

```xml
<dependency>
    <groupId>com.pulumi</groupId>
    <artifactId>pulumiservice</artifactId>
    <version>%Fill in latest version from the badge up top%</version>
</dependency>
```

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
        "meagan",
        "myles",
        "steve"
    ],
});

export const members = team.members;
```

Check out the [examples/](examples/) directory for more examples.

[1]: https://www.pulumi.com/registry/packages/pulumiservice/api-docs/provider/
[2]: https://www.pulumi.com/docs/pulumi-cloud/access-management/access-tokens/
[3]: https://www.pulumi.com/docs/pulumi-cloud/reference/cloud-rest-api/
[4]: https://www.pulumi.com/docs/pulumi-cloud/self-hosted/
