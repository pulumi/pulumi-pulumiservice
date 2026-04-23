# 03 — ESC with rotation + hooks

An ESC environment with a rotated secret, a daily rotation schedule, and a
Slack webhook that fires on rotation events.

## Resources

- `pulumiservice:esc:Environment` — the environment + its definition
- `pulumiservice:esc/schedules:Rotation` — cron trigger
- `pulumiservice:esc/hooks:Hook` — Slack webhook on rotation events

## Why this pattern

Rotating secrets only matters if someone notices when it fails.
Wiring a webhook at the same time as the rotation turns "silent rotation
failure" into a visible event the on-call can respond to before consumers
start hitting expired creds.

## Run

```
pulumi stack init dev
pulumi config set organizationName acme-corp
pulumi config set --secret slackWebhookUrl https://hooks.slack.com/services/...
pulumi up
```

Wait for the first scheduled rotation (or trigger one manually via
`pulumi env rotate`) and confirm the Slack message lands.

## Other languages

Use `pulumi convert` — see the
[convert docs](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
