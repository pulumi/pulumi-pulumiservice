# Insights Account Example (YAML)

This example demonstrates:

- Creating an ESC `Environment` resource with AWS credentials
- Creating an `InsightsAccount` resource for cloud resource scanning
- Configuring scheduled scans with cron expressions
- Setting provider-specific configuration (AWS regions)
- **Using resource methods**: `triggerScan()` and `getStatus()` for on-demand operations

## Prerequisites

- Pulumi CLI installed
- Pulumi Service access token set in `PULUMI_ACCESS_TOKEN` environment variable
- An organization in Pulumi Cloud (defaults to `service-provider-test-org`)
- AWS OIDC role configured for Pulumi Insights (see
  [Pulumi Cloud OIDC](https://www.pulumi.com/docs/pulumi-cloud/oidc/))

> **Note:** This example uses placeholder AWS credentials for
> demonstration purposes. To actually deploy and scan AWS resources, you
> must update the ESC environment with valid AWS credentials (OIDC or
> access keys).

## What is Pulumi Insights?

Pulumi Insights discovers and analyzes cloud resources across your
infrastructure, providing:

- Visibility into all cloud resources (not just Pulumi-managed ones)
- Policy compliance checking
- Cost analysis and optimization recommendations
- Resource relationships and dependency mapping

An Insights Account connects to a cloud provider (AWS, Azure, GCP) and
scans for resources on a schedule.

## Running the Example

```bash
# Set a unique identifier for this run
pulumi config set digits $(date +%s)

# Deploy the resources
pulumi up
```

## Configuration

The example configures:

- **provider**: Cloud provider type (`aws`, `azure`, or `gcp`)
- **environment**: ESC environment containing cloud credentials.
  Format: `project/environment` with optional `@version` suffix
  (e.g., `my-project/prod-env` or `my-project/prod-env@v1.0`)
- **cron**: Schedule expression for automated scanning
  (daily at 2 AM UTC)
- **providerConfig**: Provider-specific settings
  (e.g., AWS regions to scan)

## Resource Methods

The `InsightsAccount` resource provides two methods for runtime operations:

### triggerScan()

Triggers an on-demand scan of the cloud resources. This is useful for:
- Immediately scanning after infrastructure changes
- Testing scan configuration
- Getting up-to-date resource inventory

Returns:
- `scanId`: Unique identifier for the scan
- `status`: Current status of the scan (e.g., "queued", "running")
- `message`: Optional informational message
- `timestamp`: When the scan was triggered

Example in TypeScript:
```typescript
const account = new pulumiservice.InsightsAccount("insights-account", {
    /* ... config ... */
});

const scanResult = account.triggerScan();
export const scanId = scanResult.scanId;
```

### getStatus()

Retrieves the current status and metadata of the insights account. Returns:
- `accountId`: The insights account identifier
- `accountName`: Name of the account
- `status`: Current account status (e.g., "active", "scanning", "idle")
- `lastScanId`: ID of the most recent scan
- `lastScanTime`: When the last scan completed
- `nextScanTime`: When the next scheduled scan will run
- `resourceCount`: Number of resources discovered in the last scan

Example in TypeScript:
```typescript
const status = account.getStatus();
export const lastScanTime = status.lastScanTime;
export const resourceCount = status.resourceCount;
```

> **Note**: Resource methods in YAML require SDK generation. To use these methods,
> convert this example to TypeScript, Python, Go, .NET, or Java using
> `pulumi convert`.

## Cleaning Up

```bash
pulumi destroy
```

## Converting to Other Languages

You can convert this YAML example to other programming languages using
the Pulumi conversion tool:

```bash
pulumi convert --language typescript --out ../ts-insights-account-converted
pulumi convert --language python --out ../py-insights-account-converted
pulumi convert --language go --out ../go-insights-account-converted
```

For more information, see the
[pulumi convert documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
