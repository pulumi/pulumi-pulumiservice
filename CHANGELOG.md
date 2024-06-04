CHANGELOG
=========

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
