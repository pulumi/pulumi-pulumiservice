# 04 — Full deployment-settings pipeline

A GitOps pipeline for one Pulumi stack: git-driven previews, drift
detection, and TTL on ephemeral review stacks.

## Resources

- `pulumiservice:stacks/deployments:Settings` — git wiring + executor config
- `pulumiservice:stacks/deployments:DriftSchedule` — periodic drift check
- `pulumiservice:stacks/deployments:TtlSchedule` — auto-destroy timer

## Why this pattern

Pulumi Deployments only shines when all three pieces are in place at once:

- Without **settings**, nothing runs automatically; you're just running
  `pulumi up` from a laptop.
- Without **drift detection**, console-clickers can silently diverge
  state from code.
- Without **TTL**, preview stacks accumulate AWS bills.

Ship all three together or you'll end up manually setting up the missing
ones later (and forgetting some).

## Run

```
pulumi stack init dev
pulumi config set organizationName acme-corp
pulumi up
```

After this applies, pushing to `main` on the underlying GitHub repo
triggers `pulumi up` automatically.

## Other languages

Use `pulumi convert` — see the
[convert docs](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
