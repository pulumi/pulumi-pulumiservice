# Insights Account Example (YAML)

This example demonstrates:
- Creating an ESC `Environment` resource with AWS credentials
- Creating an `InsightsAccount` resource for cloud resource scanning
- Configuring scheduled scans with cron expressions
- Setting provider-specific configuration (AWS regions)

## Prerequisites

- Pulumi CLI installed
- Pulumi Service access token set in `PULUMI_ACCESS_TOKEN` environment variable
- An organization in Pulumi Cloud (defaults to `service-provider-test-org`)
- AWS OIDC role configured for Pulumi Insights (see [Pulumi Cloud OIDC](https://www.pulumi.com/docs/pulumi-cloud/oidc/))

## What is Pulumi Insights?

Pulumi Insights discovers and analyzes cloud resources across your infrastructure, providing:
- Visibility into all cloud resources (including those not managed by Pulumi)
- Policy compliance checking
- Cost analysis and optimization recommendations
- Resource relationships and dependency mapping

An Insights Account connects to a cloud provider (AWS, Azure, GCP) and scans for resources on a schedule.

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
- **environment**: ESC environment containing cloud credentials
- **cron**: Schedule expression for automated scanning (daily at 2 AM UTC)
- **providerConfig**: Provider-specific settings (e.g., AWS regions to scan)

## Cleaning Up

```bash
pulumi destroy
```

## Converting to Other Languages

You can convert this YAML example to other programming languages using the Pulumi conversion tool:

```bash
pulumi convert --language typescript --out ../ts-insights-account-converted
pulumi convert --language python --out ../py-insights-account-converted
pulumi convert --language go --out ../go-insights-account-converted
```

For more information, see the [pulumi convert documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
