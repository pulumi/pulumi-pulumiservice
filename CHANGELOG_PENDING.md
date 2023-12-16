### Improvements

### Bug Fixes

- Fix `Read` for TeamStackPermission so resources are not deleted from state on refresh. Note: TeamStackPermission resources created before v0.17.0 will maintain the existing behavior, but those (re)created with the new version will support `refresh`. [#205](https://github.com/pulumi/pulumi-pulumiservice/pull/205)

