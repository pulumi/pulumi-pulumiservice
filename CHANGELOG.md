CHANGELOG
=========

## 0.26.0

### Improvements

- Added support for ESC Webhooks and Webhook Groups [#401](https://github.com/pulumi/pulumi-pulumiservice/issues/409)

## 0.25.0

### Improvements

- Added support for ESC Projects [#401](https://github.com/pulumi/pulumi-pulumiservice/issues/401)

## 0.24.1

### Bug Fixes

- Fixes type error panic with environments [#400](https://github.com/pulumi/pulumi-pulumiservice/issues/400)

## 0.24.0

### Improvements

- Added TemplateSource resource [#387](https://github.com/pulumi/pulumi-pulumiservice/issues/387)

### Bug Fixes

- Fixes a bug where the `Environment` resource would print its contents on error. [#390](https://github.com/pulumi/pulumi-pulumiservice/pull/390)
- Fixes a bug where the `Environment` resource would error if a FileAsset was used. [#391](https://github.com/pulumi/pulumi-pulumiservice/pull/391)
- Fixes a bug where the `AgentPool` resource lost some outputs on `refresh`. [#395](https://github.com/pulumi/pulumi-pulumiservice/pull/395)

## 0.23.2

### Bug Fixes
- Improving error messages and input validation [#374](https://github.com/pulumi/pulumi-pulumiservice/issues/374)
- Fixing secrets leak [#376](https://github.com/pulumi/pulumi-pulumiservice/issues/376)[#377](https://github.com/pulumi/pulumi-pulumiservice/issues/377)

## 0.23.1

### Bug Fixes
- Fixing webhook exposing secret values bug [#371](https://github.com/pulumi/pulumi-pulumiservice/issues/371)

## 0.23.0

### Improvements

- Add a Stack Resource [#357](https://github.com/pulumi/pulumi-pulumiservice/pull/357)

## 0.22.3

### Bug Fixes

- Fixed import by refactoring Read method of OrgAccessToken resource [311](https://github.com/pulumi/pulumi-pulumiservice/issues/311)
- Fixed import by refactoring Read method of TeamAccessToken resource [311](https://github.com/pulumi/pulumi-pulumiservice/issues/311)
- Fixed import by refactoring Read method of StackTag resource [311](https://github.com/pulumi/pulumi-pulumiservice/issues/311)
- Fixed token value clearing on refresh [359](https://github.com/pulumi/pulumi-pulumiservice/issues/359)

## 0.22.2

### Bug Fixes

- Fixed import by refactoring Read method of EnvironmentVersionTag resource [311](https://github.com/pulumi/pulumi-pulumiservice/issues/311)
- Fix a panic in DeploymentSettings [#356](https://github.com/pulumi/pulumi-pulumiservice/pull/356)

## 0.22.1

### Bug Fixes
- Fixed import by refactoring Read method of AccessToken resource + minor refactor [#311](https://github.com/pulumi/pulumi-pulumiservice/issues/311)
- Fixed import by refactoring Read method of AgentPool resource + minor refactor [#311](https://github.com/pulumi/pulumi-pulumiservice/issues/311)
- Fixing noisy diff in DS OIDC object [#330](https://github.com/pulumi/pulumi-pulumiservice/issues/330)
- Removed accessToken provider parameter defaults from schema to prevent leaks [#350](https://github.com/pulumi/pulumi-pulumiservice/issues/350)

### Miscellaneous
- Added CHANGELOG_PENDING file to ignore-list of the `main` workflow [[#340](https://github.com/pulumi/pulumi-pulumiservice/issues/340)]

## 0.22.0

### Bug Fixes
- DeploymentSettings resource will now successfully store secrets, but outputs for secret values became ciphertext. [#123](https://github.com/pulumi/pulumi-pulumiservice/issues/123)
- DeploymentSettings will no longer have noisy diff on update and refresh [#123](https://github.com/pulumi/pulumi-pulumiservice/issues/123)
- DeploymentSettings can now be successfully imported [#123](https://github.com/pulumi/pulumi-pulumiservice/issues/123)

## 0.21.4

### Bug Fixes

- Fixed Environment Get function by fixing the resource's Read method [#319](https://github.com/pulumi/pulumi-pulumiservice/issues/319)

### Miscellaneous

- Fixed integ tests [#328](https://github.com/pulumi/pulumi-pulumiservice/issues/328)

## 0.21.3

### Improvements

- Added Update logic to Deployment Settings resource [#299](https://github.com/pulumi/pulumi-pulumiservice/issues/299)

### Bug Fixes

- Fixed Read failure on 404 from Pulumi Service [#312](https://github.com/pulumi/pulumi-pulumiservice/issues/312)
- Fixed environment tests breaking due to name collision [#296](https://github.com/pulumi/pulumi-pulumiservice/issues/296)
- Fixed import for Schedules [#270](https://github.com/pulumi/pulumi-pulumiservice/issues/270)
- Fixed noisy refresh for Team resource [#314](https://github.com/pulumi/pulumi-pulumiservice/pull/314)

### Miscellaneous

- Migrated all Diff methods to use GetOldInputs instead of GetOlds to avoid manually removing properties [#297](https://github.com/pulumi/pulumi-pulumiservice/issues/297)

## 0.21.2

### Bug Fixes

- Version bump of `github.com/pulumi/esc` package to v0.9.1 as old one no longer works, which broke EnvironmentVersionTag resource

## 0.21.1

### Improvements

- Added revision field to Environment resource [#290](https://github.com/pulumi/pulumi-pulumiservice/issues/290)
- Added EnvironmentVersionTag resource [#291](https://github.com/pulumi/pulumi-pulumiservice/issues/291)

## 0.21

### Improvements

- Add TeamEnvironmentPermission resource [#179](https://github.com/pulumi/pulumi-pulumiservice/issues/179)
- Add support for Environment resource [#255](https://github.com/pulumi/pulumi-pulumiservice/issues/255)

## 0.20.2

### Bug Fixes

- Prevent noisy drift if agentPoolId is empty [#268](https://github.com/pulumi/pulumi-pulumiservice/issues/268)
- Add missing DeploymentSettings output properties [#267](https://github.com/pulumi/pulumi-pulumiservice/issues/267)

## 0.20.1

### Bug Fixes

- Fixed refresh breaking schedules bug [#257](https://github.com/pulumi/pulumi-pulumiservice/issues/257)
- Adjusting provider to work with updated model for DockerImage [#262](https://github.com/pulumi/pulumi-pulumiservice/issues/262)

## 0.20.0

### Improvements

- Adding support for DeploymentSchedule resource. [#248](https://github.com/pulumi/pulumi-pulumiservice/issues/248)
- Adding support for DriftSchedule and TtlSchedule resource. [#250](https://github.com/pulumi/pulumi-pulumiservice/issues/250)
- Adding support for new WebhookFilters. [#254](https://github.com/pulumi/pulumi-pulumiservice/pull/254)

## 0.19.0

### Improvements

- Support `deleteAfterDestroy` option for the `DeploymentSettings` resource. [#244](https://github.com/pulumi/pulumi-pulumiservice/pull/244)

## 0.18.0

### Improvements

- Support `AgentPool` resource. [#228](https://github.com/pulumi/pulumi-pulumiservice/pull/228)
- Support `agentPoolId` for the `Deployment Settings` resource. [#228](https://github.com/pulumi/pulumi-pulumiservice/pull/228)

### Miscellaneous

- Updated pulumi/pulumi and other dependencies to latest versions. [#233](https://github.com/pulumi/pulumi-pulumiservice/pull/233)

## 0.17.0

### Improvements

- Support `import` for the `Team` resource. [#207](https://github.com/pulumi/pulumi-pulumiservice/pull/207)

### Bug Fixes

- Fix `Read` for TeamStackPermission so resources are not deleted from state on refresh. Note: TeamStackPermission resources created before v0.17.0 will now return an error if attempting a refresh, but those (re)created with the new version will support `refresh`. [#205](https://github.com/pulumi/pulumi-pulumiservice/pull/205)
- Fix `Update` for StackTags so tag names can be updated. [#206](https://github.com/pulumi/pulumi-pulumiservice/pull/206)
