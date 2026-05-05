# Service Catalog v2 demo

Provisions a Pulumi Cloud **service catalog** entirely from one Pulumi
program. v1 pulumiservice has no equivalent — there is no v1 resource
that models org-wide service catalog entries, owners, items, or
properties. The closest legacy approach is shelling out to
`/api/orgs/{org}/services` from a `command:local.Command`.

## What gets created

One `pulumi up` against any org provisions:

- 1 × `pulumiservice:v2/teams:Team` — `catalog-owner-<suffix>` (the owner)
- 3 × `pulumiservice:Stack` — placeholder stacks the catalog points at
- 3 × `pulumiservice:v2/services:Service` — `checkout-api`, `search-frontend`,
  `notification-worker`, each with owner / items / tier / oncall /
  slack / runbook metadata

## Run

```bash
# Yaml (no install step):
cd yaml
pulumi stack init demo
pulumi up

# TypeScript (uses the local SDK build):
cd ../typescript
yarn install
yarn link "@pulumi/pulumiservice"   # one-time, after `make install_nodejs_sdk`
pulumi stack init demo
pulumi up
```

Override `serviceOrg` / `nameSuffix` via `pulumi config set` to retarget
the org or namespace the demo's resource names.

## Demo runbook (≈5 min)

1. **Open Pulumi Cloud** → your org → **Services** tab. Show the empty
   (or sparsely populated) catalog so the audience has the "before"
   image.
2. **`pulumi up`** in the `yaml/` directory. Walk through the preview
   tree: 7 resources, ordered by dependency (team → stacks → services).
3. **Reload the Services tab.** Three cards appear:
   `checkout-api-demo`, `search-frontend-demo`, `notification-worker-demo`.
   Click into one to show: owner team, linked stack item, ordered
   property rows (tier → oncall → slack → runbook).
4. **Drift moment.** Edit a property in `yaml/Main.yaml` (e.g., change
   `value: tier-2` → `value: tier-1` on `searchService`) and run
   `pulumi up`. Point out the diff Pulumi computes against the live
   catalog before applying.
5. **`pulumi destroy`.** Catalog returns to its prior state. Show the
   Services tab is empty again.

## v1 contrast

Show this snippet from a v1 program — the only way to seed the catalog
without v2:

```typescript
new command.local.Command("seedCatalog", {
    create: `curl -X POST $PULUMI_BACKEND_URL/api/orgs/${org}/services \
        -H "Authorization: token $PULUMI_ACCESS_TOKEN" \
        -d @service-payload.json`,
    delete: `curl -X DELETE ...`,
});
```

No diff, no drift detection, no typed inputs, no preview of the change
shape. Switching to v2 turns the catalog into proper Pulumi state.

## Cleanup

```bash
pulumi destroy --yes
pulumi stack rm demo --yes
```

The `forceDestroy: true` on each `Stack` lets `pulumi destroy` tear the
placeholder stacks down even if their state isn't empty.
