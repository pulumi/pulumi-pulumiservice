# Disjoint Path Params Spike — Result

**Date:** 2026-05-06
**Outcome:** PASS — proceed with Option 1 from `tasks/2026-05-06-disjoint-path-params.md`.

## Question

Does Pulumi's engine resolve downstream `${parent.pathParamField}` references when `pathParamField` is declared only in `inputProperties` (not echoed into the resource's `properties` output schema)?

## Method

1. Applied a one-off patch to `provider/pkg/rest/schema.go` removing the `mergePathParamsAsInputs(outputs, ...)` block — path params declared only as inputs, not as outputs.
2. Ran the existing `examples/v2/stack-tags/yaml/` example via the in-process provider test harness:
   ```
   cd examples && go test -tags v2 -v -run "TestV2/stack-tags/yaml" -timeout 10m
   ```
   Backend: dev review-stacks instance.
3. The example exercises the spike pattern: a `Stack` with three path params (`orgName`, `projectName`, `stackName`); two `Tag` resources downstream that reference `${parentStack.projectName}` and `${parentStack.stackName}`.

## Result

```
=== RUN   TestV2/stack-tags/yaml
    examples_v2_test.go:339: starting providers
    examples_v2_test.go:339: creating stack test-7de84a79
    examples_v2_test.go:363: deploying
    examples_v2_test.go:370: previewing update
    examples_v2_test.go:372: refreshing
    examples_v2_test.go:374: destroying
--- PASS: TestV2/stack-tags/yaml (18.94s)
PASS
```

Full lifecycle succeeded: `up` (Create), preview-no-changes, `refresh` (Read), `destroy` (Delete). All `${parentStack.*}` references resolved against the upstream's input-only path-param values.

## Conclusion

Pulumi's engine resolves cross-resource references against the resource's full state (inputs + outputs), not against the schema's `properties` map alone. Dropping path params from `properties` does not break downstream references in IaC programs.

**Action:** Proceed with Phase B → C+D → E of the disjoint-path-params plan. Spike patch reverted; `git diff provider/pkg/rest/schema.go` is empty.

## Permanent regression coverage

The existing `TestV2/stack-tags/yaml` already exercises the disjoint-reference pattern end-to-end. No additional regression test is needed; Phase E1 of the plan was simplified to "verify TestV2/stack-tags/yaml still passes after the disjointness change lands" rather than promoting a new spike harness.

## Post-change retest

After C+D landed, re-ran `TestV2/stack-tags/yaml` against the same dev backend with the final disjointness changes applied (no spike patch — this is the production code path).

```
=== RUN   TestV2/stack-tags/yaml
    examples_v2_test.go:339: starting providers
    examples_v2_test.go:339: creating stack test-6428f986
    examples_v2_test.go:363: deploying
    examples_v2_test.go:370: previewing update
    examples_v2_test.go:372: refreshing
    examples_v2_test.go:374: destroying
--- PASS: TestV2/stack-tags/yaml (48.80s)
```

Full E2E succeeded: Create, preview-no-changes (no spurious diffs from path-param removal), Refresh (Read recovered path params from `req.Inputs`), Destroy (Delete sourced path params from `req.OldInputs`). All `${parentStack.*}` references resolved correctly. Disjointness is validated end-to-end.
