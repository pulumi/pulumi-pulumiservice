# Contributing to Pulumi Service Provider

First, thanks for contributing to Pulumi Service Provider and helping make it better. We appreciate the help!
This repository is one of many across the Pulumi ecosystem and we welcome contributions to them all.

## Code of Conduct

Please make sure to read and observe our [Contributor Code of Conduct](./CODE-OF-CONDUCT.md).

## Communications

You are welcome to join the [Pulumi Community Slack](https://slack.pulumi.com/) for questions and a community of like-minded folks.
We discuss features and file bugs on GitHub via [Issues](https://github.com/pulumi/pulumi-pulumiservice/issues) as well as [Discussions](https://github.com/pulumi/pulumi-pulumiservice/discussions).

### Issues

Feel free to pick up any existing issue that looks interesting to you or fix a bug you stumble across while using Pulumi Service Provider. No matter the size, we welcome all improvements.

### Feature Work

For larger features, we'd appreciate it if you open a [new issue](https://github.com/pulumi/pulumi-pulumiservice/issues/new) before investing a lot of time so we can discuss the feature together.
Please also be sure to browse [current issues](https://github.com/pulumi/pulumi-pulumiservice/issues) to make sure your issue is unique, to lighten the triage burden on our maintainers.
Finally, please limit your pull requests to contain only one feature at a time. Separating feature work into individual pull requests helps speed up code review and reduces the barrier to merge.

## Developing

Here's a quick list of helpful make commands:

1. `make ensure`, which restores/installs any build dependencies
1. `make build`, which generates models from provider's `schema.json`, builds the provider and builds all SDKs into the `sdk` folder
1. `make install`, which installs Pulumi Service Provider

## Testing

Ideally, every change should include unit tests, and every new resource a matching example in the `examples` folder.

You should also test changes manually using a Pulumi program that uses the updated SDKs. Here are some language-specific hints:

### .NET

- To import generated Nuget package, use this command inside your Pulumi program
  - `dotnet add package Pulumi.PulumiService -s {your path}/pulumi-pulumiservice/sdk/dotnet/bin/Debug/ -v X.XX.XX`

## Submitting a Pull Request

For contributors we use the [standard fork based workflow](https://gist.github.com/Chaser324/ce0505fbed06b947d962): Fork this repository, create a topic branch, and when ready, open a pull request from your fork.

We require a changelog entry for pretty much all PRs. Add a line in `CHANGELOG.md` describing your change and link to an issue. In rare cases where your PR is a minor change, like formatting or a typo fix, apply `impact/no-changelog-required` label to your PR instead.

### Pulumi employees

Pulumi employees have write access to Pulumi repositories and should push directly to branches rather than forking the repository. Tests can run directly without approval for PRs based on branches rather than forks.

Please ensure that you nest your branches under a unique identifier such as your name (e.g. `pulumipus/cool_feature`).

## Creating a Release

This section is for Pulumi employees only.

To release a new version of the provider, follow steps below:

- Trigger a release in #release-ops
- Github Actions will automatically build, test and then publish the new release to all the various package managers
- Once that is done, you will see your version in [Releases](https://github.com/pulumi/pulumi-pulumiservice/releases)

## Releasing the API Surface

This section is for Pulumi employees only.

Resources under the `pulumiservice:api:*` namespace ship via an auto-tag workflow rather than `#release-ops`. A release has two phases: refresh the spec to open a release PR, then merge that PR to tag and publish.

### Refresh the spec and open the release PR

1. Run the [`refresh-api-spec`](./.github/workflows/refresh-v1-spec.yml) workflow, passing the `version` to release.
2. It opens a PR from `auto/refresh-api-spec` with the `auto-release` label and `Release-Version:` marker already set. CI runs automatically; work through the reviewer checklist in the PR body.

### Tag and publish

1. Merge the PR to `main`.
2. [`tag-api-release`](./.github/workflows/tag-v1-release.yml) reads the marker and pushes the release tag.
3. The tag triggers `release.yml` to build and publish; a `finalize` job then marks the release `--latest` and triggers a docs rebuild in `pulumi/registry`.

## Getting Help

We're sure there are rough edges and we appreciate you helping out. If you want to talk with other folks in the Pulumi community (including members of the Pulumi team) come hang out in the `#contribute` channel on the [Pulumi Community Slack](https://slack.pulumi.com/).
