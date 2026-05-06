# Path-Param Disjointness Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make path parameters disjoint between `inputProperties` and `properties` in V2 resources, so identity-bearing fields are owned by the program (inputs) and the resource ID, never echoed redundantly into outputs.

**Architecture:** Lean on `IDFormat` to carry identity. Path parameters live only in `inputProperties` (program-owned, declared replace-on-change) and inside the synthesized resource ID. `Read` and `Delete` recover them from `req.OldInputs` (always present at delete time) and from the parsed ID (for the import path). Drop the `mergePathParamsAsInputs(outputs, …)` schema injection and the `populatePathParams(state, …)` runtime injection. Keep the read-after-create path working by sourcing path params from `req.Properties` (inputs) instead of from state during Create.

**Tech Stack:** Go, Pulumi schema spec (`github.com/pulumi/pulumi/pkg/v3/codegen/schema`), `pulumi-go-provider` middleware/dispatch, V2 metadata-driven resource framework at `provider/pkg/rest/`.

**Risk gate:** Phase A is a spike — if Pulumi's runtime cannot resolve `${parent.pathParam}` references against an input-only field, abandon strict disjointness and switch to Option 2 (semantic-only ownership tagging). All later phases assume the spike succeeds.

---

## Background (for engineers new to V2)

V2 is a metadata-driven resource layer in `provider/pkg/rest/`. One Go file (`resource.go`) implements all CRUD verbs for every `pulumiservice:v2:*` resource, parameterized by a JSON metadata document at `provider/pkg/cloud/metadata.json` and the embedded OpenAPI spec at `provider/pkg/cloud/spec.json`. The `Spec` parser and `Metadata` parser sit in `spec.go` and `metadata.go`. The schema generator (`schema.go`) builds a Pulumi `PackageSpec` from the same two documents.

Today, path parameters appear in **two places** in the resource shape:

1. **`inputProperties`** — `operationInputs(spec, create, rm)` walks the create op's path/query parameters (schema.go:189-206) and adds them as inputs with `WillReplaceOnChanges: true`.
2. **`properties` (outputs)** — `mergePathParamsAsInputs(outputs, &requiredOutputs, op, rm)` (schema.go:132-139) explicitly forces every create + read path-parameter into the outputs map.

At runtime, `populatePathParams(state, inputs)` (resource.go:652-677) is called after Create's read-after-create (resource.go:372, 376) and after Update (resource.go:619). It walks the create + read ops' path parameters and copies their values from `inputs` into `state`. The reason: most Pulumi Cloud responses are sparse (e.g. `POST /api/stacks/{org}/{name}` returns `{}` on success) and would otherwise leave path parameters out of state — breaking downstream `${parent.organizationName}` references.

The appendix principle says: *"create method's outputs are disjoint from its inputs, so it's clear if the program or cloud owns a given field"*. Path parameters are program-owned (the user types them), so they belong in inputs only. The ID already encodes them; recovering them from the ID at delete/read time is structurally cleaner than echoing them into outputs.

**All 45 v2 resources today already declare `idFormat`** (verified via `python3 -c "import json; …"` audit). The backfill phase is empty.

---

## File Structure

```
provider/pkg/rest/
├── schema.go        — BuildSchema, buildResource, operationInputs, operationOutputs, mergePathParamsAsInputs
├── resource.go      — Check, Create, Read, Update, Delete, populatePathParams, parseIDIntoInputs
├── schema_test.go   — TestBuildSchemaSucceeds (full-spec smoke test)
├── resource_test.go — TestCreate*, TestRead, end-to-end CRUD against synthetic specs
└── check_test.go    — TestCheck*, parseIDIntoInputs unit tests
```

**Files this plan touches:**
- `provider/pkg/rest/schema.go` — add IDFormat validation; remove path-param echo into outputs.
- `provider/pkg/rest/resource.go` — adjust Create's read-after-create source; remove `populatePathParams` calls; widen `parseIDIntoInputs` to fill missing keys.
- `provider/pkg/rest/schema_test.go` — new validation test; assertion that outputs don't carry path params.
- `provider/pkg/rest/resource_test.go` — invert path-param assertions in `TestCreateReadsAfterCreate`; new test for read-after-create source recovery.
- `provider/pkg/rest/check_test.go` — extend `parseIDIntoInputs` test coverage.
- `examples/yaml-stack/` (or similar) — integration spike for `${parent.pathParam}` resolution.

**Files this plan does NOT touch:**
- `provider/pkg/cloud/metadata.json` — no per-resource changes needed (all 45 already have idFormat).
- `provider/pkg/cloud/spec.json` — embedded OpenAPI spec, untouched.
- V1 resources (`provider/pkg/resources/*.go`, `provider/pkg/provider/manual-schema.json`) — out of scope.

---

## Phase A: Verification spike (BLOCKING)

The whole plan rests on one assumption: **`${webhook.organizationName}` continues to resolve when `organizationName` is declared only in `inputProperties` and not in `properties`**. If Pulumi's engine and SDK codegen don't preserve input values as observable on a resource handle, dropping path params from outputs breaks downstream references — and we must abandon Option 1 in favor of Option 2 (semantic-only tagging).

Reference docs / signal we want to confirm:
- Pulumi schema docgen treats `inputProperties` as the constructor surface and `properties` as the resource's readable state.
- The engine stores both inputs and outputs in state; SDK getters typically expose union.
- AWS provider declares overlapping fields explicitly — possibly because input-only doesn't expose getters.

We're testing whether V2 can drop the duplicate declaration **and** retain runtime DX.

### Task A1: Build a synthetic v2 resource with input-only path param

**Files:**
- Create: `provider/pkg/rest/spike_disjoint_test.go` (throwaway, deleted at end of phase)
- No production-code edits

- [ ] **Step 1: Write a Go test that constructs a `schema.ResourceSpec` with the `name` field in `inputProperties` only**

```go
// Copyright 2016-2026, Pulumi Corporation.
package rest

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

// TestSpike_InputOnlyPropertyShape confirms that a property declared only in
// InputProperties (not in ObjectTypeSpec.Properties) round-trips through
// schema.PackageSpec marshaling without being silently mirrored. This is a
// build-time sanity check; the live runtime check is in Task A2.
func TestSpike_InputOnlyPropertyShape(t *testing.T) {
	rs := schema.ResourceSpec{
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Type: "object",
			Properties: map[string]schema.PropertySpec{
				"id": {TypeSpec: schema.TypeSpec{Type: "string"}},
			},
			Required: []string{"id"},
		},
		InputProperties: map[string]schema.PropertySpec{
			"organizationName": {
				TypeSpec:             schema.TypeSpec{Type: "string"},
				WillReplaceOnChanges: true,
			},
			"name": {
				TypeSpec: schema.TypeSpec{Type: "string"},
			},
		},
		RequiredInputs: []string{"organizationName", "name"},
	}
	pkg := schema.PackageSpec{
		Name:      "spikepkg",
		Resources: map[string]schema.ResourceSpec{"spikepkg:index:Foo": rs},
	}
	out, err := json.Marshal(pkg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), `"organizationName"`) {
		t.Fatalf("organizationName missing from emitted schema:\n%s", out)
	}
	if strings.Contains(string(out), `"properties":{"id":{"type":"string"},"organizationName"`) {
		t.Errorf("organizationName leaked into outputs.properties (expected input-only):\n%s", out)
	}
}
```

- [ ] **Step 2: Run the test**

Run: `cd provider/pkg && go test -run TestSpike_InputOnlyPropertyShape -v ./rest/`
Expected: PASS — confirms schema marshaling treats inputs and outputs as separate maps.

### Task A2: End-to-end `${parent.pathParam}` resolution check

This is the load-bearing check. We need a real Pulumi stack with a v2-shaped resource where a path-param input is referenced by a downstream resource.

**Files:**
- Create: `examples/yaml-spike-disjoint/Pulumi.yaml`
- Create: `examples/yaml-spike-disjoint/README.md`
- Create: `examples/yaml-spike-disjoint/main.go` (or use a TS variant — pick whatever the test harness already supports)

- [ ] **Step 1: Pick the simplest existing v2 resource to spike against**

Use `pulumiservice:v2:Tag` (the stack-tag resource at `metadata.json` line ~256, idFormat `{orgName}/{project}/{name}`). It has 3 path params and a body with `value`. It's small.

- [ ] **Step 2: Write a Pulumi YAML program that creates a Tag and references its `orgName` from a downstream Tag**

```yaml
# examples/yaml-spike-disjoint/Pulumi.yaml
name: yaml-spike-disjoint
runtime: yaml
description: Spike — validate ${parent.pathParam} resolves with input-only path params

config:
  organizationName:
    type: string
    default: service-provider-test-org
  projectName:
    type: string
    default: spike-project
  stackName:
    type: string
    default: spike-stack-${pulumi.stack}

resources:
  upstreamTag:
    type: pulumiservice:v2:Tag
    properties:
      orgName: ${organizationName}
      project: ${projectName}
      stack: ${stackName}
      name: spike-upstream
      value: hello

  downstreamTag:
    type: pulumiservice:v2:Tag
    properties:
      orgName: ${upstreamTag.orgName}      # this must resolve
      project: ${upstreamTag.project}      # this must resolve
      stack: ${upstreamTag.stack}          # this must resolve
      name: spike-downstream
      value: world

outputs:
  resolvedOrg: ${upstreamTag.orgName}
  resolvedProject: ${upstreamTag.project}
```

- [ ] **Step 3: Provisional baseline — run the spike against current `main` (path params still in outputs)**

```bash
cd examples/yaml-spike-disjoint
pulumi stack init spike --non-interactive
PULUMI_TEST_OWNER=${PULUMI_TEST_OWNER:-service-provider-test-org} pulumi up --yes --non-interactive
pulumi stack output resolvedOrg
pulumi destroy --yes --non-interactive
pulumi stack rm spike --yes
```

Expected: PASS, outputs printed. This proves the spike topology is valid before we tear out the underlying mechanism.

- [ ] **Step 4: Apply a one-off local patch that removes path params from outputs only for `Tag`, then re-run**

Edit `provider/pkg/rest/schema.go` and bracket the `mergePathParamsAsInputs(outputs, …)` block with a temporary `if rm.Operations.Create != "CreateTag" {` so it skips path-param echoing for Tag specifically. Rebuild the provider (`make provider`), then:

```bash
cd examples/yaml-spike-disjoint
pulumi up --yes --non-interactive
```

Expected: One of two outcomes —
- **PASS** with `${upstreamTag.orgName}` resolving to `service-provider-test-org` → proceed with this plan.
- **FAIL** with "missing required input `orgName`" or "cannot read property of undefined" → ABORT this plan and switch to Option 2.

- [ ] **Step 5: Document spike result in `tasks/2026-05-06-disjoint-path-params-spike.md`**

Write a one-paragraph summary: outcome, command run, error (if any), decision (proceed / abort). Commit alongside the spike artifacts.

- [ ] **Step 6: Revert the local patch**

```bash
git checkout provider/pkg/rest/schema.go
```

Leave the `examples/yaml-spike-disjoint/` directory in place for use as a regression test in Phase E.

- [ ] **Step 7: Commit spike artifacts**

```bash
git add examples/yaml-spike-disjoint tasks/2026-05-06-disjoint-path-params-spike.md
git commit -m "test: add disjoint-path-params spike harness and result"
```

---

## Phase B: Build-time IDFormat validation (independent, mergeable)

This phase is non-breaking and useful on its own. It prevents future regressions where someone adds a v2 resource with path parameters but forgets `idFormat`. Land it independently of Phase C/D so we get the safety net regardless of whether disjointness lands.

### Task B1: Add `hasPathParams` helper in schema.go

**Files:**
- Modify: `provider/pkg/rest/schema.go` (helper, used by validation)

Note: an unexported `hasPathParams(op *Operation) bool` already exists in `resource.go:528-535`. Promote it to a shared package-level helper (still unexported, just shared between schema.go and resource.go), or duplicate as `opHasPathParams`. Pick duplication if the import direction would otherwise cycle.

- [ ] **Step 1: Verify the existing helper is reachable**

Run: `grep -n "func hasPathParams" provider/pkg/rest/*.go`
Expected: one hit in `resource.go`. Same package, so reachable from schema.go without import changes.

- [ ] **Step 2: No code change — proceed to Task B2**

### Task B2: Add `validateIDFormat` and call it from `buildResource`

**Files:**
- Modify: `provider/pkg/rest/schema.go:62` (`buildResource` entry)

- [ ] **Step 1: Write the failing test first**

Add to `provider/pkg/rest/schema_test.go`:

```go
// TestBuildResourceRejectsPathParamsWithoutIDFormat confirms that any resource
// declaring path parameters but no idFormat fails build with a clear error.
// idFormat is the canonical identity carrier; without it, import is broken
// and (after Phase C/D) Read/Delete cannot recover path params from state.
func TestBuildResourceRejectsPathParamsWithoutIDFormat(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "Body": {"type": "object", "properties": {"name": {"type": "string"}}}
	  }},
	  "paths": {
	    "/things/{org}": {
	      "post": {
	        "operationId": "CreateThing",
	        "parameters": [{"name": "org", "in": "path", "required": true, "schema": {"type": "string"}}],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Body"}}}},
	        "responses": {"200": {"content": {"application/json": {"schema": {"type": "object"}}}}}
	      }
	    }
	  }
	}`
	spec, err := ParseSpec([]byte(specJSON))
	if err != nil {
		t.Fatalf("spec: %v", err)
	}
	rm := ResourceMeta{Operations: Operations{Create: "CreateThing"}}
	_, err = buildResource(spec, nil, "test:index:Thing", rm)
	if err == nil || !strings.Contains(err.Error(), "idFormat") {
		t.Fatalf("expected idFormat error, got: %v", err)
	}
}

func TestBuildResourceAcceptsResourceWithoutPathParams(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "Body": {"type": "object", "properties": {"name": {"type": "string"}}}
	  }},
	  "paths": {
	    "/things": {
	      "post": {
	        "operationId": "CreateThing",
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Body"}}}},
	        "responses": {"200": {"content": {"application/json": {"schema": {"type": "object", "properties": {"id": {"type": "string"}}}}}}}
	      }
	    }
	  }
	}`
	spec, _ := ParseSpec([]byte(specJSON))
	rm := ResourceMeta{Operations: Operations{Create: "CreateThing"}}
	if _, err := buildResource(spec, nil, "test:index:Thing", rm); err != nil {
		t.Errorf("path-param-free resource should build without idFormat: %v", err)
	}
}
```

Add `"strings"` to the import list of `schema_test.go` if not already present.

- [ ] **Step 2: Run test to verify it fails**

Run: `cd provider/pkg && go test -run TestBuildResourceRejectsPathParamsWithoutIDFormat -v ./rest/`
Expected: FAIL — current `buildResource` doesn't check idFormat.

- [ ] **Step 3: Implement the validation**

In `provider/pkg/rest/schema.go`, modify `buildResource` (line 62) to add a precondition check after the operation-resolution block (around line 89, before `operationInputs` is called):

```go
// Validate idFormat presence: any resource with path parameters across its
// CRUD ops must declare an idFormat. The ID is the canonical identity
// carrier — without it, Read on import cannot recover path parameters and
// Delete (after path params are removed from output state) has no fallback.
opsWithPaths := []*Operation{create, read}
if uop, _ := spec.Op(rm.Operations.Update); uop != nil {
	opsWithPaths = append(opsWithPaths, uop)
}
if dop, _ := spec.Op(rm.Operations.Delete); dop != nil {
	opsWithPaths = append(opsWithPaths, dop)
}
anyHasPath := false
for _, op := range opsWithPaths {
	if op != nil && hasPathParams(op) {
		anyHasPath = true
		break
	}
}
if anyHasPath && rm.IDFormat == "" {
	return nil, fmt.Errorf("idFormat is required for resources with path parameters")
}
```

- [ ] **Step 4: Run tests**

Run: `cd provider/pkg && go test -v -count=1 ./rest/`
Expected: All pass, including the two new tests.

- [ ] **Step 5: Run the full-spec build smoke test**

Run: `cd provider/pkg && go test -run TestBuildSchemaSucceeds -v ./rest/`
Expected: PASS — all 45 production resources already have idFormat (verified earlier), so this is a no-op for current metadata.

- [ ] **Step 6: Commit**

```bash
git add provider/pkg/rest/schema.go provider/pkg/rest/schema_test.go
git commit -m "rest: require idFormat for resources with path parameters"
```

---

## Phase C+D: Strict disjointness — atomic schema + runtime change

C and D are coupled: dropping path-param outputs without adjusting Create breaks read-after-create; adjusting Create without dropping outputs leaves the schema declaring a field that no longer reaches state. Land them as one PR with sequential commits per logical step.

### Task C1: Widen `parseIDIntoInputs` to fill missing keys

**Files:**
- Modify: `provider/pkg/rest/resource.go:484-504`
- Modify: `provider/pkg/rest/check_test.go`

Today `parseIDIntoInputs` returns inputs unchanged whenever `inputs.Len() > 0` — it's an import-only helper. After our change, Read on refresh may receive `req.Properties` without path params (because state no longer carries them). Inputs themselves still carry path params on refresh, but Delete (and post-Phase-D Read) needs a robust merge regardless.

The new contract: parse the ID, **fill missing keys**, never overwrite existing ones.

- [ ] **Step 1: Write a failing test**

Add to `provider/pkg/rest/check_test.go`:

```go
// TestParseIDIntoInputs_FillsMissingKeys confirms that parseIDIntoInputs
// merges ID-derived path params into a non-empty inputs map, only filling
// keys that aren't already present. This is the Delete/refresh case where
// inputs may be partially populated.
func TestParseIDIntoInputs_FillsMissingKeys(t *testing.T) {
	spec := syntheticSpec(t)
	r := fooResource(spec, nil, "{org}/{name}", false)
	original := property.NewMap(map[string]property.Value{
		"name": property.New("explicit-name"),
	})
	got := r.parseIDIntoInputs("acme/payments", original)
	gotOrg, _ := got.GetOk("org")
	gotName, _ := got.GetOk("name")
	if gotOrg.AsString() != "acme" {
		t.Errorf("org: got %q, want %q (should be filled from ID)", gotOrg.AsString(), "acme")
	}
	if gotName.AsString() != "explicit-name" {
		t.Errorf("name: got %q, want %q (should NOT be overwritten)", gotName.AsString(), "explicit-name")
	}
}
```

Note: this contradicts existing `TestParseIDIntoInputs_NonEmptyInputsPreserved` (check_test.go:246-258), which asserts the import-only contract. That test will need to be deleted (its assertion is the old contract).

- [ ] **Step 2: Delete the obsolete test and run new test to confirm failure**

Delete `TestParseIDIntoInputs_NonEmptyInputsPreserved` from `check_test.go`. Run:

```bash
cd provider/pkg && go test -run TestParseIDIntoInputs -v ./rest/
```
Expected: `TestParseIDIntoInputs_FillsMissingKeys` FAILS (org would be empty), `TestParseIDIntoInputsRecoversPathParams` still PASSES.

- [ ] **Step 3: Implement the new behavior**

Replace `parseIDIntoInputs` in `provider/pkg/rest/resource.go:484-504`:

```go
// parseIDIntoInputs is the inverse of synthesizeIDFromFormat: given a Pulumi
// resource ID and the format template, recover the placeholder values and
// merge them into inputs without overwriting existing keys. Returns the
// original inputs unchanged when no IDFormat is declared or the ID doesn't
// match the format. Used by the import path (empty inputs) and by Read/
// Delete to recover path params no longer carried in state after Phase D.
func (r *Resource) parseIDIntoInputs(id string, inputs property.Map) property.Map {
	if r.meta.IDFormat == "" {
		return inputs
	}
	re, names, err := compileIDFormatRegex(r.meta.IDFormat)
	if err != nil {
		return inputs
	}
	matches := re.FindStringSubmatch(id)
	if matches == nil {
		return inputs
	}
	out := map[string]property.Value{}
	for k, v := range inputs.AllStable {
		out[k] = v
	}
	for i, name := range names {
		if i+1 >= len(matches) {
			break
		}
		if _, exists := out[name]; exists {
			continue
		}
		out[name] = property.New(matches[i+1])
	}
	return property.NewMap(out)
}
```

- [ ] **Step 4: Run all tests**

Run: `cd provider/pkg && go test -v -count=1 ./rest/`
Expected: All pass. The deleted `_NonEmptyInputsPreserved` test was the one asserting old behavior — its absence is the green signal.

- [ ] **Step 5: Commit**

```bash
git add provider/pkg/rest/resource.go provider/pkg/rest/check_test.go
git commit -m "rest: parseIDIntoInputs now fills missing keys, never overwrites"
```

### Task C2: Adjust Create's read-after-create source

**Files:**
- Modify: `provider/pkg/rest/resource.go:352-383` (Create body)
- Modify: `provider/pkg/rest/resource_test.go:310-391` (existing TestCreateReadsAfterCreate)

After Phase D, state will not contain path params. Create's read-after-create currently uses `state` as the URL source, which works only because `populatePathParams` (the thing we're about to remove) injected them. We rebuild the read URL from `req.Properties` (the inputs, which always have path params).

- [ ] **Step 1: Write the failing test**

In `provider/pkg/rest/resource_test.go`, add a new test alongside `TestCreateReadsAfterCreate`:

```go
// TestCreateReadAfterCreateSourcesFromInputs confirms that the read-after-
// create URL is built from the user inputs, not from the (potentially sparse)
// create response. This is what lets Phase D drop path params from state
// without breaking the read-after-create round-trip: the read URL
// substitution still finds path params via inputs.
func TestCreateReadAfterCreateSourcesFromInputs(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "Body":  {"type": "object", "properties": {"name": {"type": "string"}}},
	    "Read":  {"type": "object", "properties": {
	      "id":     {"type": "string"},
	      "name":   {"type": "string"},
	      "status": {"type": "string"}
	    }}
	  }},
	  "paths": {
	    "/things/{org}": {
	      "post": {
	        "operationId": "CreateThing",
	        "parameters": [{"name": "org", "in": "path", "required": true, "schema": {"type": "string"}}],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Body"}}}},
	        "responses": {"200": {"content": {"application/json": {"schema": {"type": "object", "properties": {"id": {"type": "string"}}}}}}}
	      }
	    },
	    "/things/{org}/{id}": {
	      "get": {
	        "operationId": "GetThing",
	        "parameters": [
	          {"name": "org", "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "id",  "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Read"}}}}}
	      }
	    }
	  }
	}`
	spec, _ := ParseSpec([]byte(specJSON))
	r := &Resource{
		spec: spec,
		meta: ResourceMeta{
			Operations: Operations{Create: "CreateThing", Read: "GetThing"},
			IDFormat:   "{org}/{id}",
		},
	}
	mock := &mockTransport{responses: map[string]mockResponse{
		"POST /things/acme":        {status: 200, body: `{"id":"thing-1"}`},
		"GET /things/acme/thing-1": {status: 200, body: `{"id":"thing-1","name":"foo","status":"ready"}`},
	}}
	SetTransportResolver(func(_ context.Context) (Transport, error) { return mock, nil })

	resp, err := r.Create(context.Background(), p.CreateRequest{
		Properties: propMap(map[string]any{"org": "acme", "name": "foo"}),
	})
	if err != nil {
		t.Fatalf("create: %v\n  calls: %v", err, mock.calls)
	}
	// The GET URL must include `acme` even after we (Phase D) stop populating
	// path params into state. Source for the read URL is req.Properties.
	if len(mock.calls) != 2 || mock.calls[1] != "GET /things/acme/thing-1" {
		t.Errorf("expected POST then GET /things/acme/thing-1, got: %v", mock.calls)
	}
	// State carries cloud-returned fields, not the path-param echo.
	if v, ok := resp.Properties.GetOk("status"); !ok || v.AsString() != "ready" {
		t.Errorf("state missing read-only field 'status': ok=%v v=%q", ok, v.AsString())
	}
}
```

- [ ] **Step 2: Run the new test**

Run: `cd provider/pkg && go test -run TestCreateReadAfterCreateSourcesFromInputs -v ./rest/`
Expected: PASS — current behavior already calls populatePathParams which makes the URL build correctly. The test passing today doesn't prove it's robust to Phase D's removal; we keep this test as the contract.

- [ ] **Step 3: Refactor Create to source URL from inputs explicitly**

In `provider/pkg/rest/resource.go:352-383`, replace the body of Create:

```go
func (r *Resource) Create(ctx context.Context, req p.CreateRequest) (p.CreateResponse, error) {
	if req.DryRun {
		return p.CreateResponse{Properties: req.Properties}, nil
	}
	op, err := r.resolveOp("create", r.meta.Operations.Create)
	if err != nil {
		return p.CreateResponse{}, err
	}
	if op == nil {
		return p.CreateResponse{}, fmt.Errorf("create: resource has no create operation declared")
	}
	if r.meta.RequireImport {
		if err := r.checkAlreadyExists(ctx, req.Properties); err != nil {
			return p.CreateResponse{}, err
		}
	}
	_, state, err := r.execAndDecode(ctx, op, req.Properties)
	if err != nil {
		return p.CreateResponse{}, err
	}
	// Read-after-create must source path-parameter values from the inputs
	// (req.Properties), not from state — most Pulumi Cloud create endpoints
	// return sparse bodies (e.g. `{}`), so state alone won't carry path params
	// once we stop echoing them in (Phase D).
	source := mergeMaps(req.Properties, state)
	if fetched, ok, err := r.fetchState(ctx, source, state); err != nil {
		return p.CreateResponse{}, fmt.Errorf("create: read-after-create: %w", err)
	} else if ok {
		state = fetched
	}
	id, err := r.synthesizeID(state, req.Properties)
	if err != nil {
		return p.CreateResponse{}, fmt.Errorf("create: %w", err)
	}
	return p.CreateResponse{ID: id, Properties: state}, nil
}
```

Key change: `source := mergeMaps(req.Properties, state)` replaces the two `populatePathParams` calls. `populatePathParams` is no longer reachable from Create.

- [ ] **Step 4: Run all rest-package tests**

Run: `cd provider/pkg && go test -v -count=1 ./rest/`
Expected: `TestCreateReadsAfterCreate` (line 310-391) FAILS at the path-param assertion (line 384-386: "state lost path-param `org` after read-after-create"). This is the Phase D break — fix in Task D2.

Other tests should pass.

- [ ] **Step 5: Defer commit until Phase D lands**

Hold on committing — the change is incomplete until D2 inverts the failing assertion. Move to Task D1.

### Task D1: Drop path-param echo from outputs schema

**Files:**
- Modify: `provider/pkg/rest/schema.go:132-139`
- Modify: `provider/pkg/rest/schema_test.go`

- [ ] **Step 1: Write the failing test**

Add to `provider/pkg/rest/schema_test.go`:

```go
// TestBuildResourceOmitsPathParamsFromOutputs confirms that path parameters
// appear in InputProperties only, not in ObjectTypeSpec.Properties. After
// strict disjointness, identity-bearing fields are program-owned (inputs +
// resource ID) and not echoed into the cloud-owned output namespace.
func TestBuildResourceOmitsPathParamsFromOutputs(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "Body": {"type": "object", "properties": {"name": {"type": "string"}}},
	    "Read": {"type": "object", "properties": {
	      "id":   {"type": "string"},
	      "name": {"type": "string"}
	    }}
	  }},
	  "paths": {
	    "/things/{org}": {
	      "post": {
	        "operationId": "CreateThing",
	        "parameters": [{"name": "org", "in": "path", "required": true, "schema": {"type": "string"}}],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Body"}}}},
	        "responses": {"200": {"content": {"application/json": {"schema": {"type": "object"}}}}}
	      }
	    },
	    "/things/{org}/{id}": {
	      "get": {
	        "operationId": "GetThing",
	        "parameters": [
	          {"name": "org", "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "id",  "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/Read"}}}}}
	      }
	    }
	  }
	}`
	spec, _ := ParseSpec([]byte(specJSON))
	rm := ResourceMeta{
		Operations: Operations{Create: "CreateThing", Read: "GetThing"},
		IDFormat:   "{org}/{id}",
	}
	rs, err := buildResource(spec, nil, "test:index:Thing", rm)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if _, ok := rs.InputProperties["org"]; !ok {
		t.Errorf("inputs missing 'org' (path param should be input)")
	}
	if _, ok := rs.Properties["org"]; ok {
		t.Errorf("outputs should not include 'org' (path param is input-only after disjointness)")
	}
	if _, ok := rs.Properties["name"]; !ok {
		t.Errorf("outputs missing 'name' (body field returned by read should be output)")
	}
}
```

- [ ] **Step 2: Run the test**

Run: `cd provider/pkg && go test -run TestBuildResourceOmitsPathParamsFromOutputs -v ./rest/`
Expected: FAIL — current `mergePathParamsAsInputs(outputs, …)` injects `org` into outputs.

- [ ] **Step 3: Drop the path-param echo loop**

In `provider/pkg/rest/schema.go`, delete lines 125-139 (the comment + the loop calling `mergePathParamsAsInputs(outputs, …)`):

```go
// DELETE THIS WHOLE BLOCK:
// Path parameters need to round-trip through state so Delete (which
// reads from saved state, not from inputs) can construct its URL.
// Without this, deleting a resource fails with "path parameter X
// missing from inputs" because X never made it into outputs.
if outputs == nil {
	outputs = map[string]schema.PropertySpec{}
}
for _, op := range []*Operation{create, read} {
	if op == nil {
		continue
	}
	if err := mergePathParamsAsInputs(outputs, &requiredOutputs, op, rm); err != nil {
		return nil, fmt.Errorf("outputs (path params): %w", err)
	}
}
```

Replace with a minimal nil-guard so downstream code (the validate-fields-match-outputs loop) doesn't NPE:

```go
if outputs == nil {
	outputs = map[string]schema.PropertySpec{}
}
```

The `mergePathParamsAsInputs` function itself is still used for `inputs` (schema.go:108-111) — keep the function.

- [ ] **Step 4: Run the new test**

Run: `cd provider/pkg && go test -run TestBuildResourceOmitsPathParamsFromOutputs -v ./rest/`
Expected: PASS.

- [ ] **Step 5: Run full test suite**

Run: `cd provider/pkg && go test -v -count=1 ./rest/`
Expected: `TestCreateReadsAfterCreate` (resource_test.go:310-391) still fails at path-param assertion. Fix in Task D2. Other tests pass.

- [ ] **Step 6: Defer commit, move to D2**

### Task D2: Drop runtime path-param injection

**Files:**
- Modify: `provider/pkg/rest/resource.go:603-621` (Update body)
- Modify: `provider/pkg/rest/resource_test.go:310-391` (existing test assertion)
- Delete: `provider/pkg/rest/resource.go:642-677` (`populatePathParams` function)

After Task C2, Create no longer calls `populatePathParams`. After this task, neither does Update — and the function itself goes away.

- [ ] **Step 1: Update Update to drop populatePathParams**

In `provider/pkg/rest/resource.go:603-621`, remove the `state = r.populatePathParams(state, src)` call. Updated body:

```go
// Update fires the update op (if declared).
func (r *Resource) Update(ctx context.Context, req p.UpdateRequest) (p.UpdateResponse, error) {
	op, err := r.resolveOp("update", r.meta.Operations.Update)
	if err != nil {
		return p.UpdateResponse{}, err
	}
	if op == nil {
		return p.UpdateResponse{}, fmt.Errorf("update: resource has no update operation declared")
	}
	if req.DryRun {
		return p.UpdateResponse{Properties: req.Inputs}, nil
	}
	src := mergeMaps(req.Inputs, req.OldInputs, req.State)
	_, state, err := r.execAndDecode(ctx, op, src)
	if err != nil {
		return p.UpdateResponse{}, err
	}
	return p.UpdateResponse{Properties: state}, nil
}
```

- [ ] **Step 2: Delete `populatePathParams`**

Remove `provider/pkg/rest/resource.go:642-677` (the entire function). Verify no callers remain:

Run: `grep -rn "populatePathParams" provider/pkg/rest/`
Expected: zero hits after deletion.

- [ ] **Step 3: Invert the path-param assertion in TestCreateReadsAfterCreate**

In `provider/pkg/rest/resource_test.go:383-386`, replace:

```go
// Path params should still be present (populatePathParams runs after fetchState).
if v, ok := resp.Properties.GetOk("org"); !ok || v.AsString() != "acme" {
	t.Errorf("state lost path-param `org` after read-after-create: ok=%v v=%q", ok, v.AsString())
}
```

with:

```go
// Path params are program-owned: they belong in inputs and the resource ID,
// not in cloud-owned state. After read-after-create, state should carry only
// what the read response returned, plus any emit-on-create preserves.
if _, ok := resp.Properties.GetOk("org"); ok {
	t.Errorf("state should not carry path-param `org` (program owns inputs, cloud owns outputs)")
}
```

- [ ] **Step 4: Run full test suite**

Run: `cd provider/pkg && go test -v -count=1 ./rest/`
Expected: All pass. `TestCreateReadsAfterCreate` now asserts the new contract.

- [ ] **Step 5: Update the now-stale comment in Delete**

In `provider/pkg/rest/resource.go:623-628`, replace:

```go
// Delete fires the delete op (if declared). Resources without a delete op
// quietly succeed; the engine drops the state.
//
// Path parameters are sourced from a union of (state, OldInputs) so that
// stacks created before path params were round-tripped into outputs can
// still be deleted: OldInputs preserves the original user inputs.
```

with:

```go
// Delete fires the delete op (if declared). Resources without a delete op
// quietly succeed; the engine drops the state.
//
// Path parameters are sourced from req.OldInputs (always present at delete
// time) merged with req.Properties as a fallback. State no longer carries
// path params after the disjointness change — they live only in inputs and
// the synthesized ID.
```

The Delete body itself (`mergeMaps(req.Properties, req.OldInputs)`) needs no functional change: with state lacking path params, the merge naturally falls through to `OldInputs`.

- [ ] **Step 6: Run full test suite once more**

Run: `cd provider/pkg && go test -v -count=1 ./rest/`
Expected: All pass.

- [ ] **Step 7: Run linter**

Run: `cd provider && golangci-lint run --timeout 10m ./pkg/rest/`
Expected: PASS. (If `populatePathParams` was the only user of `pathParamRE`, it isn't — `synthesizeIDFromFormat` and `compileIDFormatRegex` use it too. No dead-code warnings expected.)

- [ ] **Step 8: Commit Phase C+D as one logical change**

```bash
git add provider/pkg/rest/resource.go provider/pkg/rest/resource_test.go provider/pkg/rest/schema.go provider/pkg/rest/schema_test.go provider/pkg/rest/check_test.go
git commit -m "rest: path parameters disjoint from outputs; identity via ID

Path parameters are program-owned: they live in inputProperties and inside
the synthesized resource ID. They no longer round-trip through state. Read
recovers them from req.Inputs (refresh) or parses them from req.ID (import);
Delete recovers them from req.OldInputs. Read-after-create sources its URL
from req.Properties (inputs) so the change is invisible to existing
metadata-driven resources.

Drops:
  - mergePathParamsAsInputs(outputs, ...) call in buildResource
  - populatePathParams runtime helper (no callers remain)
  - obsolete TestParseIDIntoInputs_NonEmptyInputsPreserved (replaced by
    TestParseIDIntoInputs_FillsMissingKeys)

Adds:
  - TestBuildResourceOmitsPathParamsFromOutputs
  - TestCreateReadAfterCreateSourcesFromInputs
  - TestParseIDIntoInputs_FillsMissingKeys"
```

---

## Phase E: Integration validation and CHANGELOG

### Task E1: Rebuild SDKs and run the spike harness

**Files:**
- No code edits — verification only

- [ ] **Step 1: Rebuild the provider and SDKs**

Run: `make provider && make build_sdks`
Expected: Clean builds. The SDK regeneration consumes the new schema, so v2 resources should now expose path parameters as input-only in TypeScript / Python / Go / .NET / Java types.

- [ ] **Step 2: Run the Phase A spike harness against the new build**

```bash
cd examples/yaml-spike-disjoint
PULUMI_TEST_OWNER=${PULUMI_TEST_OWNER:-service-provider-test-org} pulumi up --yes --non-interactive
pulumi stack output resolvedOrg
pulumi destroy --yes --non-interactive
```

Expected: PASS. `resolvedOrg` outputs `service-provider-test-org`. This validates that `${upstreamTag.orgName}` resolves at runtime even though `orgName` is now input-only in the schema.

If this fails: revert Phase C+D commit, switch to Option 2 (semantic-only ownership tagging). Document the failure in the spike result file.

- [ ] **Step 3: Run the broader v2 integration suite**

Run: `cd examples && go test -tags=yaml -v -count=1 -timeout 1h -run "TestYaml.*V2"` (adjust if the v2 tests have a different naming convention — check `examples/examples_yaml_test.go`).

Expected: PASS for all v2-shaped resources. If any fail with "missing required input" or path-param resolution errors, that's the disjointness change biting; investigate per resource.

- [ ] **Step 4: Promote the spike example to a permanent regression test**

The yaml-spike-disjoint example is a useful permanent guard against future regressions. Add a test entry in `examples/examples_yaml_test.go`:

```go
//go:build yaml || all
// +build yaml all

func TestYamlDisjointPathParams(t *testing.T) {
	test := pulumitest.NewPulumiTest(t,
		filepath.Join(getCwd(t), "yaml-spike-disjoint"),
		inMemoryProvider(),
		opttest.UseAmbientBackend(),
	)
	test.SetConfig(t, "organizationName", getOrgName())
	runPulumiTest(t, test)
}
```

Rename the directory from `yaml-spike-disjoint` to `yaml-disjoint-path-params` to drop the spike-flavored name now that it's permanent:

```bash
git mv examples/yaml-spike-disjoint examples/yaml-disjoint-path-params
```

Update the test to match the new path. Keep the README's link to `pulumi convert` per the project's YAML-example conventions in CLAUDE.md.

- [ ] **Step 5: Run the new yaml test**

Run: `cd examples && go test -v -run TestYamlDisjointPathParams -tags yaml -timeout 10m`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add examples/yaml-disjoint-path-params examples/examples_yaml_test.go
git rm -r examples/yaml-spike-disjoint
git commit -m "test: promote disjoint-path-params spike to permanent regression test"
```

### Task E2: CHANGELOG

**Files:**
- Modify: `CHANGELOG.md`

This is a user-visible change to the v2 schema (path params no longer appear in resource outputs). v2 is in active development and has no released stable surface yet — but the SDK shapes are observable via published preview builds. Document the change accordingly.

- [ ] **Step 1: Add entry under Unreleased**

In `CHANGELOG.md`, under `## Unreleased` → `### Improvements`, add:

```markdown
- v2 resources: path parameters are now input-only and recovered via the resource ID, removing the redundant echo in resource state outputs.
```

- [ ] **Step 2: Commit**

```bash
git add CHANGELOG.md
git commit -m "changelog: v2 path parameters are input-only"
```

---

## Self-Review

**1. Spec coverage:**
- Phase A spike validates `${parent.pathParam}` resolution → covers DX-risk gate from the brainstorm.
- Phase B validates idFormat at build time → prevents future regressions.
- Phase C+D drops both schema-level (mergePathParamsAsInputs(outputs, …)) and runtime (populatePathParams) path-param injection → fully implements Option 1 from the prior turn.
- Phase E exercises the change end-to-end against real Pulumi Cloud → confirms the spike result holds at scale.

**2. Placeholder scan:** No "TBD", "implement later", or "similar to Task N" patterns. All steps include exact file paths, line ranges, code, commands, and expected output.

**3. Type / signature consistency:**
- `parseIDIntoInputs(id string, inputs property.Map) property.Map` — same signature throughout (Tasks C1, D1).
- `mergePathParamsAsInputs(inputs, required, op, rm)` retained for inputs (schema.go:108-111); only the outputs invocation (schema.go:132-139) is removed.
- `Resource.populatePathParams` deleted entirely — no caller remains after Tasks C2 and D2.
- `IDFormat` field on `ResourceMeta` (metadata.go:90) is the canonical spelling everywhere.

**4. Risks called out:**
- Phase A is the bail-out gate. If it fails, this plan is dead and we switch to Option 2.
- Phase B is independent and useful regardless — land it first.
- Phase C+D is one PR but multiple commits, sequenced so each commit leaves the build green except for the explicit failing-test moments TDD requires.

---

## Execution Handoff

**Plan complete and saved to `tasks/2026-05-06-disjoint-path-params.md`. Two execution options:**

**1. Subagent-Driven (recommended)** — I dispatch a fresh subagent per task, review between tasks, fast iteration.

**2. Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints.

**Which approach?**
