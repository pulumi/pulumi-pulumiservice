# 08 — Approval workflow with change gates

Gate production deployments behind human approval from a
segregation-of-duties team.

## Resources

- `pulumiservice:orgs/teams:Team` — the approver team
- `pulumiservice:changegates:ApprovalRule` — the gate itself

## Why this pattern

Regulated workloads (SOX, PCI, HIPAA) require that production changes
be approved by someone other than the author. Slack approvals and
`CODEOWNERS` files don't cut it for compliance audits — auditors want
a tamper-evident record of (change, approver, timestamp, justification)
held in a system-of-record.

Change gates live in Pulumi Cloud's audit log (combine with example 05
for compliance-grade retention) and block the deployment itself, not
just the PR.

## Run

```
pulumi stack init prod
pulumi config set organizationName acme-corp
pulumi up
```

Then try `pulumi up` on a stack tagged `environment=production` — the
deployment will pend until an approver clicks approve in the Console.

## Other languages

Use `pulumi convert` — see the
[convert docs](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
