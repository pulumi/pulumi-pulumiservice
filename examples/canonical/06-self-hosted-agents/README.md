# 06 — Self-hosted agent pool

Provision a Pulumi agent pool and wire a stack to deploy via it.

## Resources

- `pulumiservice:orgs/agents:AgentPool` — the pool registration
- `pulumiservice:stacks/deployments:Settings` — stack → pool binding

## Why this pattern

Pulumi's SaaS executors live in Pulumi's infrastructure. That's fine for
most stacks. It stops working when your target cloud account has no
public ingress (e.g., VPC-only EKS clusters, PrivateLink-only databases,
regulated government tenancies).

Self-hosted agents solve this: the executor runs on your hardware,
inside your network, and only the control plane (updates, state, RBAC)
talks to `api.pulumi.com`. No inbound ports from Pulumi to you.

## Run

```
pulumi stack init dev
pulumi config set organizationName acme-corp
pulumi up
pulumi stack output --show-secrets registrationToken
```

Then deploy the agent binary on each worker VM using the printed token.

## Other languages

Use `pulumi convert` — see the
[convert docs](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
