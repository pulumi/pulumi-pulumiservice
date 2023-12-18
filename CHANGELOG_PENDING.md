### Improvements

- Support `import` for the `Team` resource. [#207](https://github.com/pulumi/pulumi-pulumiservice/pull/207)

### Bug Fixes

- Fix `Read` for TeamStackPermission so resources are not deleted from state on refresh. Note: TeamStackPermission resources created before v0.17.0 will now return an error if attempting a refresh, but those (re)created with the new version will support `refresh`. [#205](https://github.com/pulumi/pulumi-pulumiservice/pull/205)
- Fix `Update` for StackTags so tag names can be updated. [#206](https://github.com/pulumi/pulumi-pulumiservice/pull/206)

