# CHANGELOG

## Unreleased

### Improvements

- Added support for no-code deployments with template source in DeploymentSettings [#521](https://github.com/pulumi/pulumi-pulumiservice/issues/521)

## 0.31.0

### Improvements

- Added ApprovalRule resource [#539](https://github.com/pulumi/pulumi-pulumiservice/pull/539)
- Minor fixes for the updated Approvals API [#543](https://github.com/pulumi/pulumi-pulumiservice/pull/543)

## 0.30.0

### Bug Fixes

- Fixed OIDC Issuer policies order to prevent accidental drifts [#542](https://github.com/pulumi/pulumi-pulumiservice/pull/542)

### Improvements

- Added EnvironmentRotationSchedule resource [#536](https://github.com/pulumi/pulumi-pulumiservice/pull/536)

## 0.29.3

### Improvements

- Added support for per-team maximum open durations for ESC environments [#525](https://github.com/pulumi/pulumi-pulumiservice/pull/525)

## 0.29.2

### Bug Fixes

- Fix a bug in TTLSchedules preventing correct handling of deleteBeforeDestroy [#517](https://github.com/pulumi/pulumi-pulumiservice/pull/517)

## 0.29.1

### Bug Fixes

- Fixed Schedules' Read method erroring on non-existent schedule [#510](https://github.com/pulumi/pulumi-pulumiservice/pull/510)
- Added ability to pass in backend url in provider constructor [#515](https://github.com/pulumi/pulumi-pulumiservice/pull/515)

## 0.28.0

### Improvements

- Added OIDC Issuer resource [#349](https://github.com/pulumi/pulumi-pulumiservice/issues/349)

## 0.27.4

### Improvements

- Added secret support for all fields in DeploymentSettings [#467](https://github.com/pulumi/pulumi-pulumiservice/pull/467)
- Added cascading secrets to Deployment settings [#475](https://github.com/pulumi/pulumi-pulumiservice/pull/475)

## 0.27.3

### Bug Fixes

- Ensure the `commit` property on DeploymentSettings is propagated through the provider. [#470](https://github.com/pulumi/pulumi-pulumiservice/pull/470)

## 0.27.2

### Bug Fixes

- Updating the timestamp of a one-time schedule will now cause a replacement, allowing schedules that have already run to be rescheduled correctly: [#455](https://github.com/pulumi/pulumi-pulumiservice/pull/455)

## 0.27.1

### Bug Fixes

- Fixed eternal drift in Webhook resource when `secret` field is supplied [#369](https://github.com/pulumi/pulumi-pulumiservice/issues/369)
- Fixed Environment resource secrets regression [#442](https://github.com/pulumi/pulumi-pulumiservice/issues/442)

## 0.27.0

### Improvements

- Allow force deleting agent pools. [#435](https://github.com/pulumi/pulumi-pulumiservice/pull/435)
- Add support for CacheOptions in DeploymentSettings to enable Dependency Caching [#436](https://github.com/pulumi/pulumi-pulumiservice/pull/436)

### Bug Fixes

- Generate TypedDict types for the Python SDK [#437](https://github.com/pulumi/pulumi-pulumiservice/pull/437)

## 0.26.5

### Improvements

- Make SourceContext optional in DeploymentSettings [#427](https://github.com/pulumi/pulumi-pulumiservice/pulls/427)

  In some advanced use cases, for example if your source code is baked into a custom image, or you are obtaining
  the source code from a different source, you may not want to specify a `SourceContext` in your `DeploymentSettings`.

### Bug Fixes

- Fixing TeamEnvironmentPermission, project field was not working [#429](https://github.com/pulumi/pulumi-pulumiservice/issues/429)

## 0.26.3

### Bug Fixes

- Fix panic when using computed values in environment properties [#423](https://github.com/pulumi/pulumi-pulumiservice/issues/423)

## 0.26.2

### Bug Fixes

- Fix panic in environment definition [#418](https://github.com/pulumi/pulumi-pulumiservice/issues/418)

## 0.26.1

### Bug Fixes

- Fixes failing pulumi refresh on org tokens with a slash in its name [#388](https://github.com/pulumi/pulumi-pulumiservice/issues/388)
- Fix panic when using computed values in environment definition [#411](https://github.com/pulumi/pulumi-pulumiservice/issues/411)

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
