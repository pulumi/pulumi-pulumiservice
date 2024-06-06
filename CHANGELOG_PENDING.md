### Improvements

- Added Update logic to Deplyoment Settings resourse [#299](https://github.com/pulumi/pulumi-pulumiservice/issues/299)

### Bug Fixes

- Fixed environment tests breaking due to name collision [#296](https://github.com/pulumi/pulumi-pulumiservice/issues/296)

### Miscellaneous

- Migrated all Diff methods to use GetOldInputs instead of GetOlds to avoid manually removing properties [#297](https://github.com/pulumi/pulumi-pulumiservice/issues/297)