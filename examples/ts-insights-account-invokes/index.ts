import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const digits = config.require("digits");
const orgName =
  config.get("organizationName") ||
  process.env.PULUMI_TEST_OWNER ||
  "service-provider-test-org";

// Create an ESC environment with AWS credentials for the insights account
const credentialsEnv = new service.Environment("credentials-env", {
  organization: orgName,
  project: `insights-invoke-project-${digits}`,
  name: `insights-invoke-credentials-${digits}`,
  yaml: new pulumi.asset.StringAsset(`values:
  aws:
    login:
      fn::open::aws-login:
        oidc:
          roleArn: arn:aws:iam::123456789012:role/PulumiInsightsRole
          sessionName: pulumi-insights-session
  environmentVariables:
    AWS_REGION: us-west-2
`),
});

// Create an InsightsAccount resource
const insightsAccount = new service.InsightsAccount("insights-account", {
  organizationName: orgName,
  accountName: `test-invoke-account-${digits}`,
  provider: "aws",
  environment: pulumi.interpolate`${credentialsEnv.project}/${credentialsEnv.name}`,
  scanSchedule: "none",
  tags: {
    "test-type": "invoke-integration",
    "created-by": "pulumi-test",
  },
});

// Use getInsightsAccountOutput to fetch the account we just created
// We apply on insightsAccountId to ensure the invoke is not called before the resource is created
const fetchedAccount = insightsAccount.insightsAccountId.apply((_) => {
  return service.getInsightsAccountOutput({
    organizationName: orgName,
    accountName: insightsAccount.accountName,
  });
});

// Use getInsightsAccountsOutput to list all accounts in the organization
// We apply on insightsAccountId to ensure the invoke is not called before the resource is created
const allAccounts = insightsAccount.insightsAccountId.apply((_) => {
  return service.getInsightsAccountsOutput({
    organizationName: orgName,
  });
});

// Export values to verify the invoke functions work correctly
export const resourceAccountName = insightsAccount.accountName;
export const resourceInsightsAccountId = insightsAccount.insightsAccountId;

// Export values from the getInsightsAccount invoke
export const fetchedAccountName = fetchedAccount.accountName;
export const fetchedInsightsAccountId = fetchedAccount.insightsAccountId;
export const fetchedProvider = fetchedAccount.provider;
export const fetchedEnvironment = fetchedAccount.environment;
export const fetchedScanSchedule = fetchedAccount.scanSchedule;
export const fetchedScheduledScanEnabled = fetchedAccount.scheduledScanEnabled;

// Export accounts list from getInsightsAccounts invoke
export const accountsCount = fetchedAccount
  .apply(() => allAccounts.accounts)
  .apply((accounts) => accounts.length);

// Verify that our created account is in the list
export const createdAccountInList = pulumi
  .all([allAccounts.accounts, insightsAccount.accountName])
  .apply(([accounts, name]) => accounts.some((a) => a.accountName === name));
