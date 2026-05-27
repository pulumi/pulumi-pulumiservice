# Policy Pack Example (YAML)

This example publishes a Policy Pack to Pulumi Cloud using the `PolicyPack`
resource and binds it to a `PolicyGroup`. Policies are supplied inline so the
provider does not need to spawn a language-analyzer plugin to introspect the
pack — useful when the publishing machine doesn't have the pack's runtime
toolchain installed.

The pack's source lives in [`./policy-pack`](./policy-pack); on Create the
provider tarballs the directory and uploads it.

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

You can convert this YAML example to other programming languages using the
Pulumi conversion tool:

```bash
pulumi convert --language typescript --out ../ts-policy-pack-converted
pulumi convert --language python --out ../py-policy-pack-converted
pulumi convert --language go --out ../go-policy-pack-converted
```

For more information, see the [pulumi convert documentation](https://www.pulumi.com/docs/iac/cli/commands/pulumi_convert/).
