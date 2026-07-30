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
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

// synthSpec builds a minimal OpenAPI spec with one POST /widget operation
// whose request and response bodies reference WidgetRequest, plus the given
// component schemas.
func synthSpec(t *testing.T, schemas map[string]any) *Spec {
	t.Helper()
	doc := map[string]any{
		"openapi": "3.0.0",
		"paths": map[string]any{
			"/widget": map[string]any{
				"post": map[string]any{
					"operationId": "CreateWidget",
					"requestBody": map[string]any{
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{"$ref": "#/components/schemas/WidgetRequest"},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{"$ref": "#/components/schemas/WidgetRequest"},
								},
							},
						},
					},
				},
			},
		},
		"components": map[string]any{"schemas": schemas},
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal synth spec: %v", err)
	}
	spec, err := ParseSpec(raw)
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}
	return spec
}

func widgetMetadata() *Metadata {
	return &Metadata{Resources: map[string]ResourceMeta{
		"pulumiservice:api:Widget": {
			Operations: Operations{Create: "CreateWidget"},
		},
	}}
}

func buildSynth(t *testing.T, schemas map[string]any) *schema.PackageSpec {
	t.Helper()
	pkg, err := BuildSchema(synthSpec(t, schemas), widgetMetadata(), "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	return pkg
}

func obj(props map[string]any, required ...string) map[string]any {
	out := map[string]any{"type": "object", "properties": props}
	if len(required) > 0 {
		req := make([]any, len(required))
		for i, r := range required {
			req[i] = r
		}
		out["required"] = req
	}
	return out
}

func sref(name string) map[string]any {
	return map[string]any{"$ref": "#/components/schemas/" + name}
}

func TestNamedTypeEmission(t *testing.T) {
	pkg := buildSynth(t, map[string]any{
		"WidgetRequest": obj(map[string]any{
			"config": sref("WidgetConfig"),
		}),
		"WidgetConfig": obj(map[string]any{
			"size":     map[string]any{"type": "integer"},
			"apiKey":   map[string]any{"type": "string"},
			"metadata": map[string]any{"type": "object"},
		}, "size"),
	})

	tok := "pulumiservice:api:WidgetConfig"
	ct, ok := pkg.Types[tok]
	if !ok {
		t.Fatalf("named type %q not emitted; have %v", tok, mapKeys(pkg.Types))
	}
	if got := ct.Properties["size"].TypeSpec.Type; got != "integer" {
		t.Errorf("size type: got %q, want integer", got)
	}
	if !ct.Properties["apiKey"].Secret {
		t.Errorf("apiKey should be marked secret by looksSecret")
	}
	if got, want := ct.Required, []string{"size"}; !slicesEqual(got, want) {
		t.Errorf("required: got %v, want %v", got, want)
	}
	in := pkg.Resources["pulumiservice:api:Widget"].InputProperties["config"]
	if got, want := in.TypeSpec.Ref, "#/types/"+tok; got != want {
		t.Errorf("config ref: got %q, want %q", got, want)
	}
}

func TestRecursiveTypeTerminates(t *testing.T) {
	pkg := buildSynth(t, map[string]any{
		"WidgetRequest": obj(map[string]any{"node": sref("TreeNode")}),
		"TreeNode": obj(map[string]any{
			"value":    map[string]any{"type": "string"},
			"children": map[string]any{"type": "array", "items": sref("TreeNode")},
		}),
	})
	ct, ok := pkg.Types["pulumiservice:api:TreeNode"]
	if !ok {
		t.Fatalf("TreeNode not emitted")
	}
	items := ct.Properties["children"].TypeSpec.Items
	if items == nil || items.Ref != "#/types/pulumiservice:api:TreeNode" {
		t.Errorf("children items should self-reference, got %+v", items)
	}
}

func TestResourceTokenCollisionSuffixed(t *testing.T) {
	// A component schema named Widget collides with the resource token.
	pkg := buildSynth(t, map[string]any{
		"WidgetRequest": obj(map[string]any{"inner": sref("Widget")}),
		"Widget":        obj(map[string]any{"name": map[string]any{"type": "string"}}),
	})
	if _, clash := pkg.Types["pulumiservice:api:Widget"]; clash {
		t.Fatalf("type token must not collide with the resource token")
	}
	if _, ok := pkg.Types["pulumiservice:api:WidgetProperties"]; !ok {
		t.Fatalf("collision should emit WidgetProperties; have %v", mapKeys(pkg.Types))
	}
	in := pkg.Resources["pulumiservice:api:Widget"].InputProperties["inner"]
	if got := in.TypeSpec.Ref; got != "#/types/pulumiservice:api:WidgetProperties" {
		t.Errorf("inner ref: got %q", got)
	}
}

func discBase(tagProp string, mapping map[string]any, extraProps map[string]any) map[string]any {
	props := map[string]any{tagProp: map[string]any{"type": "string"}}
	for k, v := range extraProps {
		props[k] = v
	}
	return map[string]any{
		"type":       "object",
		"properties": props,
		"required":   []any{tagProp},
		"discriminator": map[string]any{
			"propertyName": tagProp,
			"mapping":      mapping,
		},
	}
}

func variant(base string, extraProps map[string]any, required ...string) map[string]any {
	second := map[string]any{"type": "object", "properties": extraProps}
	if len(required) > 0 {
		req := make([]any, len(required))
		for i, r := range required {
			req[i] = r
		}
		second["required"] = req
	}
	return map[string]any{"allOf": []any{sref(base), second}}
}

func unionSchemas() map[string]any {
	return map[string]any{
		"WidgetRequest": obj(map[string]any{"shape": sref("Shape")}),
		"Shape": discBase("kind", map[string]any{
			"circle": "#/components/schemas/Circle",
			"square": "#/components/schemas/Square",
			"blob":   "#/components/schemas/Blob",
		}, map[string]any{"label": map[string]any{"type": "string"}}),
		"Circle": variant("Shape", map[string]any{"radius": map[string]any{"type": "number"}}, "radius"),
		"Square": variant("Shape", map[string]any{"side": map[string]any{"type": "number"}}),
		"Blob":   variant("Shape", map[string]any{"points": map[string]any{"type": "array", "items": map[string]any{"type": "number"}}}),
	}
}

func TestDiscriminatedUnion(t *testing.T) {
	pkg := buildSynth(t, unionSchemas())

	in := pkg.Resources["pulumiservice:api:Widget"].InputProperties["shape"]
	if got, want := len(in.TypeSpec.OneOf), 3; got != want {
		t.Fatalf("oneOf members: got %d, want %d", got, want)
	}
	disc := in.TypeSpec.Discriminator
	if disc == nil || disc.PropertyName != "kind" {
		t.Fatalf("discriminator: got %+v", disc)
	}
	if got, want := disc.Mapping["circle"], "#/types/pulumiservice:api:Circle"; got != want {
		t.Errorf("mapping[circle]: got %q, want %q", got, want)
	}

	// The base type itself must not be registered.
	if _, ok := pkg.Types["pulumiservice:api:Shape"]; ok {
		t.Errorf("union base Shape must not be emitted as a type")
	}

	circle, ok := pkg.Types["pulumiservice:api:Circle"]
	if !ok {
		t.Fatalf("Circle variant not emitted")
	}
	// Base properties flatten into the variant; the tag is const + required.
	if _, ok := circle.Properties["label"]; !ok {
		t.Errorf("base property label should flatten into Circle")
	}
	kind := circle.Properties["kind"]
	if kind.Const != "circle" {
		t.Errorf("kind const: got %v, want circle", kind.Const)
	}
	if !strings.Contains(kind.Description, "Expected value is 'circle'.") {
		t.Errorf("kind description missing expected-value hint: %q", kind.Description)
	}
	if got, want := circle.Required, []string{"kind", "radius"}; !slicesEqual(got, want) {
		t.Errorf("Circle required: got %v, want %v", got, want)
	}
}

func TestSingleVariantBecomesDefiniteType(t *testing.T) {
	schemas := unionSchemas()
	schemas["Shape"].(map[string]any)["discriminator"].(map[string]any)["mapping"] = map[string]any{
		"circle": "#/components/schemas/Circle",
	}
	pkg := buildSynth(t, schemas)
	in := pkg.Resources["pulumiservice:api:Widget"].InputProperties["shape"]
	if len(in.TypeSpec.OneOf) != 0 {
		t.Fatalf("single-variant base must not produce a oneOf")
	}
	if got, want := in.TypeSpec.Ref, "#/types/pulumiservice:api:Circle"; got != want {
		t.Errorf("shape ref: got %q, want %q", got, want)
	}
}

func TestEmptyMappingFallsBackToNamedType(t *testing.T) {
	schemas := unionSchemas()
	schemas["Shape"].(map[string]any)["discriminator"].(map[string]any)["mapping"] = map[string]any{}
	pkg := buildSynth(t, schemas)
	in := pkg.Resources["pulumiservice:api:Widget"].InputProperties["shape"]
	if got, want := in.TypeSpec.Ref, "#/types/pulumiservice:api:Shape"; got != want {
		t.Errorf("shape ref: got %q, want %q", got, want)
	}
}

func TestTwoMemberObjectUnionRejected(t *testing.T) {
	schemas := unionSchemas()
	schemas["Shape"].(map[string]any)["discriminator"].(map[string]any)["mapping"] = map[string]any{
		"circle": "#/components/schemas/Circle",
		"square": "#/components/schemas/Square",
	}
	_, err := BuildSchema(synthSpec(t, schemas), widgetMetadata(), "pulumiservice")
	if err == nil {
		t.Fatalf("expected 2-member union to fail BuildSchema")
	}
	if !strings.Contains(err.Error(), "2-member object unions") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRecursiveUnionTerminates(t *testing.T) {
	// A variant whose property references the union base again.
	schemas := unionSchemas()
	schemas["Blob"] = variant("Shape", map[string]any{"inner": sref("Shape")})
	pkg := buildSynth(t, schemas)
	blob := pkg.Types["pulumiservice:api:Blob"]
	inner := blob.Properties["inner"].TypeSpec
	if got, want := len(inner.OneOf), 3; got != want {
		t.Fatalf("recursive union members: got %d, want %d", got, want)
	}
}

func TestSharedTypeHoistsToApi(t *testing.T) {
	spec := synthSpec(t, map[string]any{
		"WidgetRequest": obj(map[string]any{"shared": sref("SharedThing")}),
		"SharedThing":   obj(map[string]any{"x": map[string]any{"type": "string"}}),
	})
	meta := &Metadata{Resources: map[string]ResourceMeta{
		"pulumiservice:api/alpha:Widget": {Operations: Operations{Create: "CreateWidget"}},
		"pulumiservice:api/beta:Gadget":  {Operations: Operations{Create: "CreateWidget"}},
	}}
	pkg, err := BuildSchema(spec, meta, "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	if _, ok := pkg.Types["pulumiservice:api:SharedThing"]; !ok {
		t.Fatalf("shared type should hoist to the api module; have %v", mapKeys(pkg.Types))
	}
}

func TestSoleModuleTypeStaysInModule(t *testing.T) {
	spec := synthSpec(t, map[string]any{
		"WidgetRequest": obj(map[string]any{"cfg": sref("OnlyHere")}),
		"OnlyHere":      obj(map[string]any{"x": map[string]any{"type": "string"}}),
	})
	meta := &Metadata{Resources: map[string]ResourceMeta{
		"pulumiservice:api/alpha:Widget": {Operations: Operations{Create: "CreateWidget"}},
	}}
	pkg, err := BuildSchema(spec, meta, "pulumiservice")
	if err != nil {
		t.Fatalf("BuildSchema: %v", err)
	}
	if _, ok := pkg.Types["pulumiservice:api/alpha:OnlyHere"]; !ok {
		t.Fatalf("sole-module type should stay in api/alpha; have %v", mapKeys(pkg.Types))
	}
}

func slicesEqual(a, b []string) bool {
	return fmt.Sprint(a) == fmt.Sprint(b)
}
