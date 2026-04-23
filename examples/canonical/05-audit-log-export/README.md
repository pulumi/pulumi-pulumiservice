# 05 — Audit log export

Continuously stream Pulumi Cloud audit events to S3 for compliance.

## Resources

- `pulumiservice:orgs/audit:LogExport` — export configuration

## Why this pattern

Compliance frameworks (SOC 2, ISO 27001, HIPAA) require tamper-evident
audit trails retained for 1–7 years. Pulumi Cloud keeps events for
90 days by default; export is how you stretch that to compliance-grade
retention and pipe events into your SIEM.

Do not rely on the UI "Download" button for this — it's for ad-hoc
investigation, not continuous compliance.

## Prerequisites

- An S3 bucket with versioning + object-lock enabled (the `acme-corp-pulumi-audit-logs`).
- An IAM role trusted by Pulumi's audit-log exporter with `s3:PutObject`.
  See [Pulumi Cloud audit log docs](https://www.pulumi.com/docs/pulumi-cloud/audit-logs/).

## Run

```
pulumi stack init dev
pulumi config set organizationName acme-corp
pulumi config set exportBucket acme-corp-pulumi-audit-logs
pulumi config set exportRoleArn arn:aws:iam::...:role/PulumiAuditLogExporter
pulumi up
```

Then verify end-to-end:

```
aws s3 ls s3://acme-corp-pulumi-audit-logs/pulumi-audit/ --recursive
```

## Other languages

Use `pulumi convert` — see the
[convert docs](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
