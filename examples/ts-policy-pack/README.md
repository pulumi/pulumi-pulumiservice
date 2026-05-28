# Policy Pack Example (TypeScript)

This example publishes a Policy Pack to Pulumi Cloud using the `PolicyPack`
resource, and binds it to a `PolicyGroup`.

The pack's source lives in [`./policy-pack`](./policy-pack). On Create the
provider tarballs that directory and uploads it; the policy metadata is
extracted by running the policy analyzer plugin against the source (matching
`pulumi policy publish` behavior).

## Running

```bash
pulumi config set organizationName <your-org>
pulumi config set digits $(date +%s)        # unique resource names per run
pulumi config set versionTag 1.0.0          # bump on re-publish, see below
pulumi up
```

Pulumi Cloud tombstones policy-pack version tags on delete, so re-publishing
the same `name`+`versionTag` returns 409. Either bump `versionTag` between runs
or use a fresh `digits` (which makes the pack name unique).

## Converting to Other Languages

You can convert this example to other programming languages using the Pulumi
conversion tool:

```bash
pulumi convert --language python --out ../py-policy-pack-converted
pulumi convert --language go --out ../go-policy-pack-converted
```

For more information, see the [pulumi convert documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
