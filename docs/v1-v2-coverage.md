# v1 / v2 Resource Coverage

| v1 (`pulumiservice:index:`) | v2 (`pulumiservice:v2/`) |
|---|---|
| `AccessToken` | `tokens:PersonalToken` |
| `AgentPool` | `agents:Pool` |
| `ApprovalRule` | `pulumiservice:v2:Gate` |
| `DeploymentSchedule` | `deployments:ScheduledDeployment` |
| `DeploymentSettings` | `deployments:Settings` |
| `DriftSchedule` | `deployments:ScheduledDeployment` (partial) |
| `Environment` | `esc:Environment` |
| `EnvironmentRotationSchedule` | `esc:EnvironmentSchedule` |
| `EnvironmentVersionTag` | `esc:RevisionTag` |
| `InsightsAccount` | `insights:Account` |
| `OidcIssuer` | `auth:OidcIssuer` |
| `OrgAccessToken` | `tokens:OrgToken` |
| `OrganizationMember` | `pulumiservice:v2:OrganizationMember` |
| `OrganizationRole` | `pulumiservice:v2:Role` |
| `PolicyGroup` | `pulumiservice:v2:PolicyGroup` |
| `Stack` | `stacks:Stack` |
| `StackTag` | `stacks:Tag` |
| `StackTags` | `stacks:Tag` (singular only) |
| `Team` | `teams:Team` |
| `TeamAccessToken` | `tokens:TeamToken` |
| `TeamEnvironmentPermission` | `auth:Policy` |
| `TeamRoleAssignment` | `teams:Role` |
| `TeamStackPermission` | `auth:Policy` |
| `TemplateSource` | `pulumiservice:v2:OrgTemplateCollection` |
| `TtlSchedule` | `deployments:ScheduledDeployment` (partial) |
| `Webhook` | `pulumiservice:v2:OrganizationWebhook`, `stacks:Webhook`, `esc:Webhook` |
| — | `agents:Task` |
| — | `auth:Policy` |
| — | `auth:SAML` |
| — | `esc:EnvironmentDraft` |
| — | `esc:EnvironmentSettings` |
| — | `esc:EnvironmentTag` |
| — | `esc:OpenEnvironmentRequest` |
| — | `esc:Webhook` |
| — | `escPreview:Environment` |
| — | `escPreview:EnvironmentTag` |
| — | `escPreview:RevisionTag` |
| — | `escPreview:Webhook` |
| — | `insights:ScheduledScanSettings` |
| — | `integrations:AzureDevOpsIntegration` |
| — | `integrations:BitBucketIntegration` |
| — | `integrations:CustomVCSIntegration` |
| — | `integrations:CustomVCSRepository` |
| — | `integrations:GitHubEnterpriseIntegration` |
| — | `integrations:GitHubIntegration` |
| — | `integrations:GitLabIntegration` |
| — | `services:Service` |
| — | `services:Item` |
| — | `stacks:Config` |
| — | `stacks:Webhook` |
| — | `pulumiservice:v2:AuditLogExportConfiguration` |
| — | `pulumiservice:v2:DefaultOrganization` |
| — | `pulumiservice:v2:PolicyIssue` |
| — | `pulumiservice:v2:PolicyPack` |
