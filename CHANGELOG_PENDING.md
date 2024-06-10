### Improvements

- Added Update logic to Deployment Settings resource [#299](https://github.com/pulumi/pulumi-pulumiservice/issues/299)

### Bug Fixes

- Fixed environment tests breaking due to name collision [#296](https://github.com/pulumi/pulumi-pulumiservice/issues/296)
- Fixed import for Schedules [#270](https://github.com/pulumi/pulumi-pulumiservice/issues/270)
- Fixed noisy refresh for Team resource [#314](https://github.com/pulumi/pulumi-pulumiservice/pull/314)

### Miscellaneous

- Migrated all Diff methods to use GetOldInputs instead of GetOlds to avoid manually removing properties [#297](https://github.com/pulumi/pulumi-pulumiservice/issues/297)
