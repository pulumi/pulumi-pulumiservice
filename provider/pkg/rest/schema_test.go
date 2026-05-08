// Copyright 2016-2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rest

import (
	"slices"
	"strings"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

// TestBuildSchemaSucceeds ensures the live spec.json + metadata.json pair
// builds a complete schema with no per-resource errors. Catches drift like
// metadata referring to operationIds the spec no longer carries.
func TestBuildSchemaSucceeds(t *testing.T) {
	spec, meta := loadFixtures(t)
	pkg, err := BuildSchema(spec, meta, "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	if got, want := len(pkg.Resources), len(meta.Resources); got != want {
		t.Errorf("Resources: got %d, want %d (one per metadata entry)", got, want)
	}
}

// TestPathParamsAreInputOnly pins the architecture invariant that path
// parameters appear ONLY in inputs, not in outputs. Path params are program-
// owned (the user types them); they live in inputProperties and inside the
// synthesized resource ID. Read recovers them from req.Inputs (refresh) or
// from req.ID (import); Delete recovers them from req.OldInputs.
//
// Note: a Pulumi-side name that is both a path param (via rename) AND a body
// field — like Team's `name`, which renames to wire-side `teamName` for path
// substitution but is also a legitimate body field echoed by Read — stays in
// outputs because the read response carries it. The invariant only fails when
// a name is purely a path param.
func TestPathParamsAreInputOnly(t *testing.T) {
	spec, meta := loadFixtures(t)
	pkg, err := BuildSchema(spec, meta, "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	// Pick a representative resource with multi-segment path params. orgName
	// is purely a path param on Team (no body field of that name); `name` is
	// a path param AND a body field, so it doesn't fit this assertion.
	rm := meta.Resources["pulumiservice:v2:Team"]
	tok := rm.Token
	if tok == "" {
		tok = "pulumiservice:v2:Team"
	}
	rs, ok := pkg.Resources[tok]
	if !ok {
		t.Fatalf("Team resource missing from package: %q", tok)
	}
	if _, ok := rs.InputProperties["orgName"]; !ok {
		t.Errorf("input \"orgName\" missing from Team")
	}
	if _, ok := rs.Properties["orgName"]; ok {
		t.Errorf("output \"orgName\" should not appear on Team — purely-path-param fields are input-only")
	}
}

// TestPathParamInputsAreReplaceOnChanges pins that path params are marked
// replaceOnChanges so the engine triggers a replace (not an in-place update)
// when their values change. Without it, mutating a path param would update
// against the new URL and 404 on a non-existent resource.
func TestPathParamInputsAreReplaceOnChanges(t *testing.T) {
	spec, meta := loadFixtures(t)
	pkg, err := BuildSchema(spec, meta, "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	rm := meta.Resources["pulumiservice:v2:Team"]
	tok := rm.Token
	if tok == "" {
		tok = "pulumiservice:v2:Team"
	}
	rs := pkg.Resources[tok]
	for _, name := range []string{"orgName", "name"} {
		ps, ok := rs.InputProperties[name]
		if !ok {
			t.Fatalf("input %q missing from Team", name)
		}
		if !ps.WillReplaceOnChanges {
			t.Errorf("input %q: WillReplaceOnChanges = false (expected true for path param)", name)
		}
	}
}

// TestIDIsSkippedFromOutputs pins that any response field literally named
// "id" is dropped from outputs, since Pulumi reserves that name for the
// resource ID. Resources whose API returns server-generated identifiers under
// that name expose them through path-parameter renames instead.
func TestIDIsSkippedFromOutputs(t *testing.T) {
	spec, meta := loadFixtures(t)
	pkg, err := BuildSchema(spec, meta, "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	for tok, rs := range pkg.Resources {
		if _, ok := rs.Properties["id"]; ok {
			t.Errorf("%s exposes a top-level `id` output; the schema builder must skip it", tok)
		}
	}
}

// TestSecretFieldsAreMarkedSecret pins the secret heuristic for known
// sensitive field names (tokenValue, secret, password, apiKey, accessToken,
// ciphertext). Catches metadata edits that drop the override AND
// resources whose schema name matches but isn't yet known.
func TestSecretFieldsAreMarkedSecret(t *testing.T) {
	spec, meta := loadFixtures(t)
	pkg, err := BuildSchema(spec, meta, "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	cases := []struct {
		token string
		field string
	}{
		// Token-value resources rely on the heuristic.
		{"pulumiservice:v2:OrgToken", "tokenValue"},
		{"pulumiservice:v2:TeamToken", "tokenValue"},
		{"pulumiservice:v2:PersonalToken", "tokenValue"},
	}
	for _, tc := range cases {
		rm := meta.Resources[tc.token]
		tok := rm.Token
		if tok == "" {
			tok = tc.token
		}
		rs, ok := pkg.Resources[tok]
		if !ok {
			t.Errorf("%s: resource missing from package", tc.token)
			continue
		}
		ps, ok := rs.Properties[tc.field]
		if !ok {
			t.Errorf("%s: output %q missing", tc.token, tc.field)
			continue
		}
		if !ps.Secret {
			t.Errorf("%s.%s: Secret = false (expected true via looksSecret heuristic)", tc.token, tc.field)
		}
	}
}

// TestIDFormatRoundTrip verifies synthesizeIDFromFormat → parseIDIntoInputs
// recovers the same path-param values across every metadata entry. Catches
// idFormat templates whose placeholders don't survive escape-and-recompile,
// or whose values would collide with the slash separator.
func TestIDFormatRoundTrip(t *testing.T) {
	_, meta := loadFixtures(t)
	for tok, rm := range meta.Resources {
		if rm.IDFormat == "" {
			t.Errorf("%s: idFormat is empty (every v2 resource should declare one)", tok)
			continue
		}
		// Build inputs that match every placeholder in the format string.
		re, names, err := compileIDFormatRegex(rm.IDFormat)
		if err != nil {
			t.Errorf("%s: compile idFormat %q: %v", tok, rm.IDFormat, err)
			continue
		}
		inputs := map[string]property.Value{}
		for _, n := range names {
			inputs[n] = property.New("v-" + n)
		}
		id, err := synthesizeIDFromFormat(rm.IDFormat, property.NewMap(inputs), property.Map{})
		if err != nil {
			t.Errorf("%s: synthesize from %q: %v", tok, rm.IDFormat, err)
			continue
		}
		matches := re.FindStringSubmatch(id)
		if matches == nil {
			t.Errorf("%s: synthesized ID %q does not match its own idFormat %q", tok, id, rm.IDFormat)
			continue
		}
		for i, n := range names {
			if got, want := matches[i+1], "v-"+n; got != want {
				t.Errorf("%s: idFormat %q placeholder %q round-trip: got %q, want %q",
					tok, rm.IDFormat, n, got, want)
			}
		}
	}
}

// TestRequiredInputsExist pins that every name listed in RequiredInputs is
// also present in InputProperties. A drift here means the schema declares a
// required field the SDK can't surface — Pulumi engine errors out.
func TestRequiredInputsExist(t *testing.T) {
	spec, meta := loadFixtures(t)
	pkg, err := BuildSchema(spec, meta, "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	for tok, rs := range pkg.Resources {
		for _, req := range rs.RequiredInputs {
			if _, ok := rs.InputProperties[req]; !ok {
				t.Errorf("%s: required input %q not in InputProperties", tok, req)
			}
		}
	}
}

// TestEmitOnCreateSurfacesAsOutput pins that fields marked emitOnCreate in
// metadata appear as outputs even when the read response schema doesn't
// carry them. Token resources rely on this — the create response is the
// only place tokenValue ever appears.
func TestEmitOnCreateSurfacesAsOutput(t *testing.T) {
	spec, meta := loadFixtures(t)
	pkg, err := BuildSchema(spec, meta, "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	for tok, rm := range meta.Resources {
		for fieldName, fm := range rm.Fields {
			if !fm.EmitOnCreate {
				continue
			}
			ulookup := rm.Token
			if ulookup == "" {
				ulookup = tok
			}
			rs, ok := pkg.Resources[ulookup]
			if !ok {
				continue
			}
			if _, present := rs.Properties[fieldName]; !present {
				t.Errorf("%s.%s: emitOnCreate field missing from outputs", tok, fieldName)
			}
		}
	}
}

// TestAutoNameNotRequired pins that fields marked autoName>0 are not in the
// required-inputs list — users must be able to leave them unset and let
// Check generate one.
func TestAutoNameNotRequired(t *testing.T) {
	spec, meta := loadFixtures(t)
	pkg, err := BuildSchema(spec, meta, "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	for tok, rm := range meta.Resources {
		for fieldName, fm := range rm.Fields {
			if fm.AutoName <= 0 {
				continue
			}
			ulookup := rm.Token
			if ulookup == "" {
				ulookup = tok
			}
			rs, ok := pkg.Resources[ulookup]
			if !ok {
				continue
			}
			if slices.Contains(rs.RequiredInputs, fieldName) {
				t.Errorf("%s.%s: autoName field appears in RequiredInputs", tok, fieldName)
			}
		}
	}
}

// TestLooksSecret covers the substring set the heuristic recognizes; callers
// outside the cloud spec rely on this contract too.
func TestLooksSecret(t *testing.T) {
	cases := map[string]bool{
		"tokenValue":       true,
		"webhookSecret":    true,
		"password":         true,
		"apiKey":           true,
		"accessToken":      true,
		"secretCiphertext": true,
		"name":             false,
		"description":      false,
		"orgName":          false,
	}
	for name, want := range cases {
		if got := looksSecret(name); got != want {
			t.Errorf("looksSecret(%q): got %v, want %v", name, got, want)
		}
	}
}

// TestPulumiNameRoundTrip pins the rename helper's contract: wire-side names
// translate to their Pulumi-side counterpart, and unmapped names pass
// through. wireSideName is the inverse direction.
func TestPulumiNameRoundTrip(t *testing.T) {
	renames := map[string]string{
		"hookName":         "name",
		"organizationName": "orgName",
	}
	cases := map[string]string{
		"name":     "hookName",
		"orgName":  "organizationName",
		"untouched": "untouched",
	}
	for wire, wantPulumi := range cases {
		if got := pulumiName(wire, renames); got != wantPulumi {
			t.Errorf("pulumiName(%q): got %q, want %q", wire, got, wantPulumi)
		}
	}
	// Inverse: wireSideName(pul) → wire.
	for wire, pul := range cases {
		if got := wireSideName(pul, renames); got != wire && wire != pul {
			// Unmapped names round-trip through both helpers unchanged; only
			// check explicit renames here.
			if _, mapped := renames[pul]; mapped {
				t.Errorf("wireSideName(%q): got %q, want %q", pul, got, wire)
			}
		}
	}
}

// TestCheckEnumCasePreservedForUnknownValue pins that values not present in
// the spec's enum list pass through unchanged (rather than being silently
// rewritten). A bad input should surface at the API call, not be papered
// over by Check.
func TestCheckEnumCasePreservedForUnknownValue(t *testing.T) {
	spec := syntheticSpec(t)
	r := fooResource(spec, nil, "", false)
	for _, in := range []string{"sideways", "Up "} { // not in enum, even after fold
		out := normalizeValue(property.New(in), "mode", flattenedRequestProperties(spec, mustOp(t, spec, "CreateFoo")), r.meta)
		if !out.IsString() || out.AsString() != in {
			t.Errorf("normalizeValue(%q) = %v; expected unchanged passthrough", in, out)
		}
	}
}

func mustOp(t *testing.T, spec *Spec, id string) *Operation {
	t.Helper()
	op, ok := spec.Op(id)
	if !ok {
		t.Fatalf("spec missing op %q", id)
	}
	return op
}

// TestBuildResourceOmitsPathParamsFromOutputs confirms that path parameters
// appear in InputProperties only, not in ObjectTypeSpec.Properties. Identity-
// bearing fields are program-owned (inputs + resource ID) and not echoed into
// the cloud-owned output namespace.
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
		t.Errorf("outputs should not include 'org' (path param is input-only)")
	}
	if _, ok := rs.Properties["name"]; !ok {
		t.Errorf("outputs missing 'name' (body field returned by read should be output)")
	}
}

// TestBuildResourceRejectsPathParamsWithoutIDFormat confirms that any resource
// declaring path parameters but no idFormat fails build with a clear error.
// idFormat is the canonical identity carrier; without it, import is broken
// and Delete cannot recover path params from state once disjointness lands.
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

// TestBuildResourceSurfacesYamlBody covers the ESC-shaped pattern: create
// takes a structured JSON body (project+name) and update takes a raw yaml
// body. The dispatch should expose a single string "yaml" input on the
// resource and let users supply the body through that field.
func TestBuildResourceSurfacesYamlBody(t *testing.T) {
	const specJSON = `{
	  "openapi": "3.0.0",
	  "components": {"schemas": {
	    "CreateBody": {"type": "object", "properties": {"project": {"type": "string"}, "name": {"type": "string"}}}
	  }},
	  "paths": {
	    "/things/{org}": {
	      "post": {
	        "operationId": "CreateThing",
	        "parameters": [{"name": "org", "in": "path", "required": true, "schema": {"type": "string"}}],
	        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/CreateBody"}}}},
	        "responses": {"200": {"description": "OK"}}
	      }
	    },
	    "/things/{org}/{project}/{name}": {
	      "patch": {
	        "operationId": "UpdateThing",
	        "parameters": [
	          {"name": "org",     "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "project", "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "name",    "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "requestBody": {"content": {"application/x-yaml": {"schema": {"type": "string"}}}},
	        "responses": {"204": {"description": "no content"}}
	      }
	    }
	  }
	}`
	spec, err := ParseSpec([]byte(specJSON))
	if err != nil {
		t.Fatalf("spec: %v", err)
	}
	rm := ResourceMeta{
		Operations: Operations{Create: "CreateThing", Update: "UpdateThing"},
		IDFormat:   "{org}/{project}/{name}",
	}
	rs, err := buildResource(spec, nil, "test:index:Thing", rm)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	yaml, ok := rs.InputProperties["yaml"]
	if !ok {
		var got []string
		for k := range rs.InputProperties {
			got = append(got, k)
		}
		t.Fatalf("expected synthesized 'yaml' input field, got inputs: %v", got)
	}
	if yaml.Type != "string" {
		t.Errorf("yaml input type: got %q, want string", yaml.Type)
	}
	if !yaml.Secret {
		t.Errorf("yaml input should default to Secret=true")
	}
}

// TestBuildResourceAcceptsResourceWithoutPathParams covers the other side of
// the validation: a resource whose operations have no path params shouldn't
// require idFormat.
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

// TestExamplesAppendedToDescription guards the format that SDK codegen relies
// on for auto-translating PCL example snippets to per-language code blocks.
func TestExamplesAppendedToDescription(t *testing.T) {
	got := appendExamples("Manages a Foo.", []string{`resource "foo" "pulumiservice:v2:Foo" {}`})
	if !strings.Contains(got, "## Example Usage") {
		t.Errorf("missing `## Example Usage` heading:\n%s", got)
	}
	if !strings.Contains(got, "```pulumi") {
		t.Errorf("missing pulumi-tagged fence:\n%s", got)
	}
}
