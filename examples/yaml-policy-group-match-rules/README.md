# yaml-policy-group-match-rules

Builds policy group match rules with **`pulumiservice:buildStackMatchRule`**
and **`pulumiservice:buildAccountMatchRule`**. A match rule adds stacks or
Insights accounts to a policy group by their tags: an entity matches if it
matches any of the `rules`, and a rule matches when all of its `conditions`
hold. Each condition compares one tag with `equals` (the default),
`notEquals`, `exists`, or `notExists`.

The functions return the expression Pulumi Cloud takes in a policy group's
`stackMatchRule` and `accountMatchRule` fields, so you don't write its
`__type` format by hand. The program passes both to a
`pulumiservice:api:PolicyGroup`. Match rules are a preview feature of Pulumi Cloud,
and account rules are only supported on `audit` policy groups.

## Convert to another language

Use `pulumi convert` to translate this program to TypeScript, Python, Go,
C#, or Java:

```bash
pulumi convert --from yaml --language typescript --out ../ts-policy-group-match-rules
```

See the [`pulumi convert` documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/)
for more options.
