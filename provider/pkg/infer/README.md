# Infer Resources

This directory contains resources that have been migrated to use the `pulumi-go-provider` infer framework.

## What is Infer?

The infer framework automatically generates Pulumi schemas from Go types, reducing boilerplate and providing type safety.

## Migration Status

Resources are being gradually migrated from `provider/pkg/resources/` to this directory.

### Phase 1 POC (In Progress)
- [ ] StackTag
- [ ] OrgAccessToken
- [ ] AgentPool

See `Convert-to-infer.md` and `docs/INFER_MIGRATION.md` for the complete migration plan.

## Adding a New Infer Resource

1. Create the resource file: `<resource_name>.go`
2. Define the resource struct, args, and state
3. Implement CRUD methods with infer signatures
4. Add `Annotate()` for documentation
5. Register in `provider/pkg/provider/hybrid.go`
6. Test and validate

See `docs/INFER_MIGRATION.md` for detailed instructions.
