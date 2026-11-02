# Saved Custom Agent

This example creates a saved Custom Agent using the organization's existing `Member` role and exports its durable `agentId`. Pass that ID to `delivery.Agent.agent`. Renaming or updating the definition keeps the ID.

Set `organizationName` and a unique `agentName`, then run `pulumi up`. The provider's caller needs permission to create and manage Custom Agents in that organization. `approvalMode: auto` allows unattended runs; `manual` pauses each run for a person's approval.

`enabledIntegrations: []` enables none. Omit it to inherit all organization integrations, or supply objects with catalog IDs such as `{id: honeycomb}`.

To convert this example to another supported language, use [`pulumi convert`](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
