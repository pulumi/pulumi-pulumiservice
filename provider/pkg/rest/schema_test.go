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

// TestPathParamsRoundTripIntoOutputs pins the architecture invariant that
// path parameters appear in BOTH inputs and outputs — the latter so Delete
// can reconstruct its URL from saved state alone. Removing path params from
// outputs would silently break refresh-after-restart for resources whose
// engine handle hasn't carried inputs forward.
func TestPathParamsRoundTripIntoOutputs(t *testing.T) {
	spec, meta := loadFixtures(t)
	pkg, err := BuildSchema(spec, meta, "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	// Pick a representative resource with multi-segment path params.
	rm := meta.Resources["pulumiservice:v2:Team"]
	tok := rm.Token
	if tok == "" {
		tok = "pulumiservice:v2:Team"
	}
	rs, ok := pkg.Resources[tok]
	if !ok {
		t.Fatalf("Team resource missing from package: %q", tok)
	}
	for _, name := range []string{"orgName", "name"} {
		if _, ok := rs.InputProperties[name]; !ok {
			t.Errorf("input %q missing from Team", name)
		}
		if _, ok := rs.Properties[name]; !ok {
			t.Errorf("output %q missing from Team — Delete cannot reconstruct URL after process restart", name)
		}
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
