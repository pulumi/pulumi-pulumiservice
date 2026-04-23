# 10 — Internal template catalog

A git-hosted catalog of Pulumi templates for `pulumi new`.

## Resources

- `pulumiservice:orgs/templates:Source` — catalog registration

## Why this pattern

Platform engineering's job is to make the paved road irresistible.
If the paved road is a wiki page titled "how to set up a new service"
with 14 manual steps, developers will — correctly — not use it.

A template catalog makes the paved road one command. `pulumi new <template>`
produces a project that already knows your tagging convention, your
monitoring defaults, your cost-allocation setup, and your deployment
pipeline wiring.

## Run

```
pulumi stack init dev
pulumi config set organizationName acme-corp
pulumi up
```

Then any org member can:

```
pulumi new acme-corp/acme-golden-path/typescript-microservice
```

## Other languages

Use `pulumi convert` — see the
[convert docs](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
