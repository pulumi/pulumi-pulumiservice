# 01 — Organization bootstrap

Day 0 provisioning for a new Pulumi organization: three teams, a CI token,
and a default policy group.

## Resources

- `pulumiservice:orgs/teams:Team` × 3 — admins, deployers, readers
- `pulumiservice:orgs/tokens:TeamAccessToken` — CI token scoped to deployers
- `pulumiservice:orgs/policies:PolicyGroup` — baseline guardrails

## Why this pattern

- **Team tokens over org tokens.** Rotating an org token means updating
  every CI secret. A team token is rotated by removing the team member
  or deleting the token, and can be reissued without cascading changes.
- **Three teams, not one.** "Everyone" is the wrong default; start with
  least-privilege readers, then opt users into deployers per-stack.
- **A policy group at bootstrap.** Applying guardrails retroactively is
  painful. Create the group empty if you must, and wire packs in later.

## Run

```
pulumi stack init dev
pulumi up
pulumi stack output --show-secrets ciTokenValue  # store this somewhere safe
```

## Other languages

Use `pulumi convert --language typescript` (or python/go/csharp/java). See
the [convert docs](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
