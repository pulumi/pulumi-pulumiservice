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
	"context"
	"strings"
	"testing"

	p "github.com/pulumi/pulumi-go-provider"
	presource "github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

// syntheticSpec constructs a minimal in-memory Spec covering one create op
// with a request body that has an enum field, an array field, and a name
// field — enough to exercise Check normalization paths.
func syntheticSpec(t *testing.T) *Spec {
	t.Helper()
	const specJSON = `{
	  "openapi": "3.0.0",
	  "servers": [{"url": "https://api.example.com"}],
	  "components": {
	    "schemas": {
	      "FooBody": {
	        "type": "object",
	        "properties": {
	          "name":   {"type": "string"},
	          "mode":   {"type": "string", "enum": ["UP", "DOWN"]},
	          "scopes": {"type": "array", "items": {"type": "string"}}
	        }
	      },
	      "FooResp": {
	        "type": "object",
	        "properties": {
	          "id": {"type": "string"},
	          "tokenValue": {"type": "string"}
	        }
	      }
	    }
	  },
	  "paths": {
	    "/api/foos/{org}": {
	      "post": {
	        "operationId": "CreateFoo",
	        "parameters": [
	          {"name": "org", "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "requestBody": {
	          "content": {"application/json": {"schema": {"$ref": "#/components/schemas/FooBody"}}}
	        },
	        "responses": {
	          "200": {
	            "content": {"application/json": {"schema": {"$ref": "#/components/schemas/FooResp"}}}
	          }
	        }
	      },
	      "get": {
	        "operationId": "GetFoo",
	        "parameters": [
	          {"name": "org", "in": "path", "required": true, "schema": {"type": "string"}},
	          {"name": "name", "in": "path", "required": true, "schema": {"type": "string"}}
	        ],
	        "responses": {
	          "200": {
	            "content": {"application/json": {"schema": {"type": "object", "properties": {"id": {"type": "string"}}}}}
	          }
	        }
	      }
	    }
	  }
	}`
	spec, err := ParseSpec([]byte(specJSON))
	if err != nil {
		t.Fatalf("synthetic spec: %v", err)
	}
	return spec
}

// fooResource builds a Resource against the synthetic spec with custom
// FieldMeta — used by each test to express what's being checked.
func fooResource(spec *Spec, fields map[string]FieldMeta, idFormat string, dbr bool) *Resource {
	return &Resource{
		spec: spec,
		meta: ResourceMeta{
			Operations:          Operations{Create: "CreateFoo", Read: "GetFoo"},
			Fields:              fields,
			IDFormat:            idFormat,
			DeleteBeforeReplace: dbr,
		},
	}
}

func TestCheckNormalizesEnumCase(t *testing.T) {
	spec := syntheticSpec(t)
	r := fooResource(spec, nil, "", false)
	resp, err := r.Check(context.Background(), p.CheckRequest{
		Inputs: property.NewMap(map[string]property.Value{
			"mode": property.New("up"),
		}),
	})
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	got, _ := resp.Inputs.GetOk("mode")
	if got.AsString() != "UP" {
		t.Errorf("mode: got %q, want \"UP\"", got.AsString())
	}
}

func TestCheckSortsUnorderedArray(t *testing.T) {
	spec := syntheticSpec(t)
	r := fooResource(spec, map[string]FieldMeta{
		"scopes": {Unordered: true},
	}, "", false)
	resp, err := r.Check(context.Background(), p.CheckRequest{
		Inputs: property.NewMap(map[string]property.Value{
			"scopes": property.New(property.NewArray([]property.Value{
				property.New("write"),
				property.New("read"),
				property.New("admin"),
			})),
		}),
	})
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	got, _ := resp.Inputs.GetOk("scopes")
	values := got.AsArray().AsSlice()
	want := []string{"admin", "read", "write"}
	for i, v := range values {
		if v.AsString() != want[i] {
			t.Errorf("scopes[%d]: got %q, want %q", i, v.AsString(), want[i])
		}
	}
}

func TestCheckGeneratesAutoName(t *testing.T) {
	spec := syntheticSpec(t)
	r := fooResource(spec, map[string]FieldMeta{
		"name": {AutoName: 24},
	}, "", false)
	resp, err := r.Check(context.Background(), p.CheckRequest{
		Urn:        presource.URN("urn:pulumi:dev::demo::pulumiservice:v2:Foo::myresource"),
		RandomSeed: []byte("deterministic-seed"),
		Inputs:     property.NewMap(map[string]property.Value{}),
	})
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	got, ok := resp.Inputs.GetOk("name")
	if !ok {
		t.Fatalf("name: missing from inputs")
	}
	name := got.AsString()
	if !strings.HasPrefix(name, "myresource-") {
		t.Errorf("name: got %q, want prefix %q", name, "myresource-")
	}
	if len(name) > 24 {
		t.Errorf("name: length %d exceeds maxLen 24 (%q)", len(name), name)
	}
	// Determinism: same seed → same name.
	resp2, _ := r.Check(context.Background(), p.CheckRequest{
		Urn:        presource.URN("urn:pulumi:dev::demo::pulumiservice:v2:Foo::myresource"),
		RandomSeed: []byte("deterministic-seed"),
	})
	got2, _ := resp2.Inputs.GetOk("name")
	if got2.AsString() != name {
		t.Errorf("name: not deterministic (%q vs %q)", name, got2.AsString())
	}
}

func TestCheckPreservesUserSuppliedName(t *testing.T) {
	spec := syntheticSpec(t)
	r := fooResource(spec, map[string]FieldMeta{
		"name": {AutoName: 24},
	}, "", false)
	resp, _ := r.Check(context.Background(), p.CheckRequest{
		Urn:        presource.URN("urn:pulumi:dev::demo::pulumiservice:v2:Foo::myresource"),
		RandomSeed: []byte("seed"),
		Inputs: property.NewMap(map[string]property.Value{
			"name": property.New("explicit-name"),
		}),
	})
	got, _ := resp.Inputs.GetOk("name")
	if got.AsString() != "explicit-name" {
		t.Errorf("user-supplied name overwritten: got %q", got.AsString())
	}
}

func TestSynthesizeIDFromFormat(t *testing.T) {
	spec := syntheticSpec(t)
	r := fooResource(spec, nil, "{org}/{name}", false)
	id, err := r.synthesizeID(
		property.NewMap(map[string]property.Value{
			"org":  property.New("acme"),
			"name": property.New("payments"),
		}),
		property.Map{},
	)
	if err != nil {
		t.Fatalf("synthesize: %v", err)
	}
	if id != "acme/payments" {
		t.Errorf("id: got %q, want %q", id, "acme/payments")
	}
}

func TestSynthesizeIDFromFormat_MissingValue(t *testing.T) {
	spec := syntheticSpec(t)
	r := fooResource(spec, nil, "{org}/{name}", false)
	_, err := r.synthesizeID(
		property.NewMap(map[string]property.Value{
			"org": property.New("acme"),
		}),
		property.Map{},
	)
	if err == nil || !strings.Contains(err.Error(), "name") {
		t.Errorf("expected missing-value error mentioning \"name\", got: %v", err)
	}
}

func TestParseIDIntoInputsRecoversPathParams(t *testing.T) {
	spec := syntheticSpec(t)
	r := fooResource(spec, nil, "{org}/{name}", false)
	got := r.parseIDIntoInputs("acme/payments", property.Map{})
	wantOrg, _ := got.GetOk("org")
	wantName, _ := got.GetOk("name")
	if wantOrg.AsString() != "acme" || wantName.AsString() != "payments" {
		t.Errorf("parsed inputs: org=%q name=%q", wantOrg.AsString(), wantName.AsString())
	}
}

func TestParseIDIntoInputs_NonEmptyInputsPreserved(t *testing.T) {
	spec := syntheticSpec(t)
	r := fooResource(spec, nil, "{org}/{name}", false)
	original := property.NewMap(map[string]property.Value{
		"org":   property.New("existing-org"),
		"extra": property.New("user-data"),
	})
	got := r.parseIDIntoInputs("other/different", original)
	gotOrg, _ := got.GetOk("org")
	if gotOrg.AsString() != "existing-org" {
		t.Errorf("non-empty inputs were overwritten: %q", gotOrg.AsString())
	}
}

func TestPreserveEmitOnCreate(t *testing.T) {
	spec := syntheticSpec(t)
	r := fooResource(spec, map[string]FieldMeta{
		"tokenValue": {EmitOnCreate: true},
	}, "", false)
	prior := property.NewMap(map[string]property.Value{
		"id":         property.New("foo-1"),
		"tokenValue": property.New("secret-from-create"),
	})
	fresh := property.NewMap(map[string]property.Value{
		"id": property.New("foo-1"),
	})
	merged := r.preserveEmitOnCreate(fresh, prior)
	got, ok := merged.GetOk("tokenValue")
	if !ok || got.AsString() != "secret-from-create" {
		t.Errorf("tokenValue: got %v, want \"secret-from-create\"", got.AsString())
	}
}

func TestDiffSetsDeleteBeforeReplaceWhenMarked(t *testing.T) {
	spec := syntheticSpec(t)
	r := fooResource(spec, nil, "", true)
	resp, err := r.Diff(context.Background(), p.DiffRequest{
		OldInputs: property.NewMap(map[string]property.Value{"name": property.New("a")}),
		Inputs:    property.NewMap(map[string]property.Value{"name": property.New("b")}),
	})
	if err != nil {
		t.Fatalf("diff: %v", err)
	}
	if !resp.HasChanges {
		t.Errorf("HasChanges: got false, want true")
	}
	if !resp.DeleteBeforeReplace {
		t.Errorf("DeleteBeforeReplace: got false, want true")
	}
}

func TestDiffOmitsDeleteBeforeReplaceWhenUnmarked(t *testing.T) {
	spec := syntheticSpec(t)
	r := fooResource(spec, nil, "", false)
	resp, _ := r.Diff(context.Background(), p.DiffRequest{
		OldInputs: property.NewMap(map[string]property.Value{"name": property.New("a")}),
		Inputs:    property.NewMap(map[string]property.Value{"name": property.New("b")}),
	})
	if resp.DeleteBeforeReplace {
		t.Errorf("DeleteBeforeReplace: got true, want false")
	}
}
