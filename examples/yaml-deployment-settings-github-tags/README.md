# Deployment Settings with Git Tag Deployments Example (YAML)

This example configures deployment settings that run `pulumi up` when a git tag matching `tagFilters` is pushed to a GitHub repository. Patterns prefixed with `!` exclude matching tags, so `!*-rc*` skips release candidates.

`deployTags` can't be combined with `deployCommits`, which defaults to `true`, so the example sets `deployCommits: false`.

The organization must have the Pulumi GitHub app installed on the repository.

To convert this example to other programming languages, see [pulumi convert](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
