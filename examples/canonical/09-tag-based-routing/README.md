# 09 — Stack tagging + per-tag webhooks

Tag stacks by environment + criticality, then route events to
differently-urgent channels based on tags.

## Resources

- `pulumiservice:stacks/tags:Tags` — environment/criticality tags
- `pulumiservice:stacks/hooks:Webhook` × 2 — PagerDuty (prod) + Slack (dev)

## Why this pattern

The naïve setup — one webhook that fires on every event to one channel —
produces alert fatigue and burns on-call trust in about six weeks.
Routing by tag scales: the same code path can create tier-1 stacks
with PagerDuty integration and tier-3 stacks with a quiet log sink,
without the hooks resource growing any per-stack conditionals.

## Run

```
pulumi stack init dev
pulumi config set organizationName acme-corp
pulumi config set --secret pagerdutyUrl https://events.pagerduty.com/integration/...
pulumi config set --secret slackDevUrl https://hooks.slack.com/services/...
pulumi up
```

## Other languages

Use `pulumi convert` — see the
[convert docs](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
