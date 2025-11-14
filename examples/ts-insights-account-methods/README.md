# Insights Account Resource Methods Example (TypeScript)

This example demonstrates how to use the **resource methods** available on the
`InsightsAccount` resource:

- `triggerScan()` - Trigger an on-demand scan of cloud resources
- `getStatus()` - Retrieve the current scan status and metadata

## What are Resource Methods?

Resource methods are functions that can be called on a resource instance to
perform operations or retrieve information. Unlike resource properties (which
are declarative), methods are imperative operations that execute at runtime.

For more information about resource methods, see the
[Pulumi blog post on resource methods](https://www.pulumi.com/blog/resource-methods-for-pulumi-packages/).

## Prerequisites

- Pulumi CLI installed
- Pulumi Service access token set in `PULUMI_ACCESS_TOKEN` environment variable
- An organization in Pulumi Cloud (defaults to `service-provider-test-org`)
- AWS OIDC role configured for Pulumi Insights (see
  [Pulumi Cloud OIDC](https://www.pulumi.com/docs/pulumi-cloud/oidc/))
- Node.js and npm/yarn installed

> **Note:** This example uses placeholder AWS credentials for demonstration
> purposes. To actually deploy and scan AWS resources, you must update the ESC
> environment with valid AWS credentials (OIDC or access keys).

## What This Example Does

1. **Creates an ESC Environment** with AWS credentials for scanning
2. **Creates an InsightsAccount** resource connected to the environment
3. **Calls `triggerScan()`** to initiate an on-demand scan of AWS resources
4. **Calls `getStatus()`** to retrieve the current scan status and metadata
5. **Exports the results** including scan ID, status, timestamps, and resource
   count

## Running the Example

```bash
# Install dependencies
yarn install

# Set a unique identifier for this run
pulumi config set digits $(date +%s)

# Optional: Set a custom organization name
pulumi config set organizationName your-org-name

# Deploy the resources and call the methods
pulumi up
```

## Resource Methods

### triggerScan()

Triggers an on-demand scan of the insights account's cloud resources.

**Use cases:**
- Immediately scan after infrastructure changes
- Test scan configuration before enabling scheduled scans
- Get an up-to-date resource inventory on demand

**Returns:**
- `scanId` - Unique identifier for the triggered scan
- `status` - Current status ("running", "failed", "succeeded")
- `timestamp` - When the scan was triggered (ISO 8601 format)

**Example:**
```typescript
const scanResult = insightsAccount.triggerScan();
export const scanId = scanResult.apply(result => result.scanId);
export const scanStatus = scanResult.apply(result => result.status);
```

### getStatus()

Retrieves the current status of the insights account including information
about the last scan.

**Returns:**
- `accountId` - The insights account identifier
- `accountName` - Name of the account
- `status` - Current scan status ("running", "failed", "succeeded")
- `lastScanId` - ID of the most recent scan
- `lastScanTime` - When the last scan completed (ISO 8601 format)
- `nextScanTime` - When the next scheduled scan will run (if scheduled
  scanning is enabled)
- `resourceCount` - Number of resources discovered in the last scan

**Example:**
```typescript
const status = insightsAccount.getStatus();
export const lastScanTime = status.apply(result => result.lastScanTime);
export const resourceCount = status.apply(result => result.resourceCount);
```

## Viewing Outputs

After running `pulumi up`, you can view the exported values:

```bash
pulumi stack output scanId
pulumi stack output currentStatus
pulumi stack output resourceCount
```

## How It Works

The example demonstrates the key pattern for using resource methods:

1. **Create the resource** as usual:
   ```typescript
   const account = new service.InsightsAccount("insights-account", {
       organizationName: "my-org",
       accountName: "my-account",
       provider: "aws",
       environment: "my-project/aws-creds",
   });
   ```

2. **Call methods on the resource instance**:
   ```typescript
   const scanResult = account.triggerScan();
   const status = account.getStatus();
   ```

3. **Use the results** in outputs or other resources:
   ```typescript
   export const scanId = scanResult.apply(result => result.scanId);
   export const resourceCount = status.apply(result => result.resourceCount);
   ```

The methods execute during the `pulumi up` operation and their results are
captured as stack outputs.

## API Endpoints Used

This example uses the following Pulumi Cloud API endpoints:

- `POST /api/preview/insights/{orgName}/accounts/{accountName}/scan` -
  Trigger scan
- `GET /api/preview/insights/{orgName}/accounts/{accountName}/scan` -
  Get scan status

## Cleaning Up

```bash
pulumi destroy
```

This will delete both the InsightsAccount and the ESC Environment.

## Language Support

Resource methods are available in all Pulumi SDK languages:
- TypeScript/JavaScript (this example)
- Python
- Go
- .NET (C#/F#)
- Java

> **Note:** Resource methods are **not available in YAML** because YAML is
> declarative and doesn't support imperative method calls. To use resource
> methods, you must use one of the SDK languages listed above.
