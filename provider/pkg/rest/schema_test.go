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

// TestBuildSchemaSucceeds catches drift between metadata and spec.json.
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

// TestPathParamsAreInputOnly pins that purely-path-param fields appear only
// in inputs. Names that are both a renamed path param and a body field
// (like Team's `name`) stay in outputs because the read response carries them.
func TestPathParamsAreInputOnly(t *testing.T) {
	spec, meta := loadFixtures(t)
	pkg, err := BuildSchema(spec, meta, "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	rm := meta.Resources["pulumiservice:v1:Team"]
	tok := rm.Token
	if tok == "" {
		tok = "pulumiservice:v1:Team"
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

// TestPathParamInputsAreReplaceOnChanges pins that path params trigger
// replace; without it mutating one would update against a non-existent URL.
func TestPathParamInputsAreReplaceOnChanges(t *testing.T) {
	spec, meta := loadFixtures(t)
	pkg, err := BuildSchema(spec, meta, "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	rm := meta.Resources["pulumiservice:v1:Team"]
	tok := rm.Token
	if tok == "" {
		tok = "pulumiservice:v1:Team"
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

// TestIDIsSkippedFromOutputs: response fields literally named "id" must not
// appear as outputs (Pulumi reserves the name).
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

// TestSecretFieldsAreMarkedSecret pins the looksSecret heuristic against
// representative resources, on both outputs and inputs.
func TestSecretFieldsAreMarkedSecret(t *testing.T) {
	spec, meta := loadFixtures(t)
	pkg, err := BuildSchema(spec, meta, "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	cases := []struct {
		token   string
		field   string
		surface string // "output" or "input"
	}{
		{"pulumiservice:v1:OrgToken", "tokenValue", "output"},
		{"pulumiservice:v1:TeamToken", "tokenValue", "output"},
		{"pulumiservice:v1:PersonalToken", "tokenValue", "output"},
		{"pulumiservice:v1:OrganizationWebhook", "secret", "input"},
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
		props := rs.Properties
		if tc.surface == "input" {
			props = rs.InputProperties
		}
		ps, ok := props[tc.field]
		if !ok {
			t.Errorf("%s: %s %q missing", tc.token, tc.surface, tc.field)
			continue
		}
		if !ps.Secret {
			t.Errorf("%s.%s (%s): Secret = false (expected true via looksSecret heuristic)", tc.token, tc.field, tc.surface)
		}
	}
}

// TestIDFormatRoundTrip pins synthesize → parse round-trip across every
// metadata entry.
func TestIDFormatRoundTrip(t *testing.T) {
	_, meta := loadFixtures(t)
	for tok, rm := range meta.Resources {
		if rm.IDFormat == "" {
			t.Errorf("%s: idFormat is empty (every v1 resource should declare one)", tok)
			continue
		}
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

// TestRequiredInputsExist pins that every RequiredInputs entry exists in
// InputProperties (drift would cause engine errors).
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

// TestEmitOnCreateSurfacesAsOutput: emitOnCreate fields appear as outputs
// even when the read schema doesn't carry them.
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

// TestAutoNameNotRequired: autoName>0 fields must not be in RequiredInputs.
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

// TestLooksSecret pins the substring set the heuristic recognizes.
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

// TestPulumiNameRoundTrip pins pulumiName ↔ wireSideName.
func TestPulumiNameRoundTrip(t *testing.T) {
	renames := map[string]string{
		"hookName":         "name",
		"organizationName": "orgName",
	}
	cases := map[string]string{
		"name":      "hookName",
		"orgName":   "organizationName",
		"untouched": "untouched",
	}
	for wire, wantPulumi := range cases {
		if got := pulumiName(wire, renames); got != wantPulumi {
			t.Errorf("pulumiName(%q): got %q, want %q", wire, got, wantPulumi)
		}
	}
	for wire, pul := range cases {
		if got := wireSideName(pul, renames); got != wire && wire != pul {
			if _, mapped := renames[pul]; mapped {
				t.Errorf("wireSideName(%q): got %q, want %q", pul, got, wire)
			}
		}
	}
}

// TestCheckEnumCasePreservedForUnknownValue: values absent from the spec's
// enum list pass through unchanged so bad input surfaces at the API call.
func TestCheckEnumCasePreservedForUnknownValue(t *testing.T) {
	spec := syntheticSpec(t)
	r := fooResource(spec, nil, "", false)
	for _, in := range []string{"sideways", "Up "} { // not in enum, even after fold
		bodyProps := flattenedRequestProperties(spec, mustOp(t, spec, "CreateFoo"))
		out := normalizeValue(property.New(in), "mode", bodyProps, r.meta)
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

// TestBuildResourceOmitsPathParamsFromOutputs: path params live only in
// InputProperties, never in ObjectTypeSpec.Properties.
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

// TestBuildResourceRejectsPathParamsWithoutIDFormat: a resource with path
// params but no idFormat must fail build (otherwise import would be broken).
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

// TestBuildResourceSurfacesYamlBody covers the ESC-shaped pattern: JSON
// create + yaml update fuses into a single "yaml" input.
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
	if yaml.Type != "string" { //nolint:goconst // matches over JSON-literal-embedded "string" counts
		t.Errorf("yaml input type: got %q, want string", yaml.Type)
	}
	if !yaml.Secret {
		t.Errorf("yaml input should default to Secret=true")
	}
}

// TestBuildResourceAcceptsResourceWithoutPathParams: a path-param-free
// resource builds without an idFormat.
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
	        "responses": {"200": {"content": {"application/json": {
	          "schema": {"type": "object", "properties": {"id": {"type": "string"}}}
	        }}}}
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

// TestExamplesAppendedToDescription pins the format SDK codegen relies on
// for auto-translating PCL examples per language.
func TestExamplesAppendedToDescription(t *testing.T) {
	got := appendExamples("Manages a Foo.", []string{`resource "foo" "pulumiservice:v1:Foo" {}`})
	if !strings.Contains(got, "## Example Usage") {
		t.Errorf("missing `## Example Usage` heading:\n%s", got)
	}
	if !strings.Contains(got, "```pulumi") {
		t.Errorf("missing pulumi-tagged fence:\n%s", got)
	}
}
