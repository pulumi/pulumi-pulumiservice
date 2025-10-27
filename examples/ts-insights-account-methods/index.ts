import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const digits = config.require("digits");
const organizationName = config.get("organizationName") || "service-provider-test-org";

// Create an ESC environment with AWS credentials for scanning
const credentialsEnv = new service.Environment("credentials-env", {
    organization: organizationName,
    project: `insights-project-${digits}`,
    name: `insights-credentials-${digits}`,
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

// Create an Insights Account
const insightsAccount = new service.InsightsAccount("insights-account", {
    organizationName: organizationName,
    accountName: `test-insights-account-${digits}`,
    provider: "aws",
    environment: pulumi.interpolate`${credentialsEnv.project}/${credentialsEnv.name}`,
});

// Export basic resource properties
export const insightsAccountId = insightsAccount.insightsAccountId;
export const accountName = insightsAccount.accountName;
export const scheduledScanEnabled = insightsAccount.scheduledScanEnabled;

// Demonstrate resource methods
// Resource methods are functions that can be called on a resource instance
// to perform operations or retrieve information.

// Example: Trigger an on-demand scan
// This initiates a scan of the cloud resources associated with this insights account.
// If a scan is already running, it returns the existing scan ID instead of triggering a new one.
// This makes the method idempotent - safe to call multiple times.
const scanResult = insightsAccount.triggerScan();

// Note: scanId and timestamp are optional outputs from triggerScan()
// When a scan is first queued, it may not have been assigned an ID yet.
// The scan ID is only created once the scan actually starts running.
// Similarly, the timestamp is only available once the scan begins.
// We export these with fallback values to avoid "undefined value" warnings.
export const scanId = pulumi.output(scanResult.scanId).apply(id => id ?? "not-yet-assigned");
export const scanStatus = scanResult.status;
export const scanTimestamp = pulumi.output(scanResult.timestamp).apply(ts => ts ?? "not-yet-started");
