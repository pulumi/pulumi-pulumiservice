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

const (
	opCreateWidget  = "CreateWidget"
	widgetRequest   = "WidgetRequest"
	tagCircle       = "circle"
	propsKey        = "properties"
	tagKind         = "kind"
	circleRef       = schemaRefPrefix + "Circle"
	schemaShape     = "Shape"
	schemaSquare    = "Square"
	schemaBlob      = "Blob"
	schemaBoolShape = "BoolShape"
	condProp        = "cond"
	circleTypeRef   = "#/types/pulumiservice:api:Circle"
	widgetTok       = "pulumiservice:api:Widget"
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
					"operationId": opCreateWidget,
					"requestBody": map[string]any{
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{refKey: schemaRefPrefix + widgetRequest},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{refKey: schemaRefPrefix + widgetRequest},
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
			Operations: Operations{Create: opCreateWidget},
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
	out := map[string]any{typeKey: typeObject, propsKey: props}
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
	return map[string]any{refKey: schemaRefPrefix + name}
}

func TestNamedTypeEmission(t *testing.T) {
	pkg := buildSynth(t, map[string]any{
		widgetRequest: obj(map[string]any{
			"config": sref("WidgetConfig"),
		}),
		"WidgetConfig": obj(map[string]any{
			"size":     map[string]any{typeKey: typeInteger},
			"apiKey":   map[string]any{typeKey: typeString},
			"metadata": map[string]any{typeKey: typeObject},
		}, "size"),
	})

	tok := "pulumiservice:api:WidgetConfig"
	ct, ok := pkg.Types[tok]
	if !ok {
		t.Fatalf("named type %q not emitted; have %v", tok, mapKeys(pkg.Types))
	}
	if got := ct.Properties["size"].Type; got != typeInteger {
		t.Errorf("size type: got %q, want integer", got)
	}
	if !ct.Properties["apiKey"].Secret {
		t.Errorf("apiKey should be marked secret by looksSecret")
	}
	if got, want := ct.Required, []string{"size"}; !slicesEqual(got, want) {
		t.Errorf("required: got %v, want %v", got, want)
	}
	in := pkg.Resources["pulumiservice:api:Widget"].InputProperties["config"]
	if got, want := in.Ref, "#/types/"+tok; got != want {
		t.Errorf("config ref: got %q, want %q", got, want)
	}
}

func TestRecursiveTypeTerminates(t *testing.T) {
	pkg := buildSynth(t, map[string]any{
		widgetRequest: obj(map[string]any{"node": sref("TreeNode")}),
		"TreeNode": obj(map[string]any{
			"value":    map[string]any{typeKey: typeString},
			"children": map[string]any{typeKey: typeArray, "items": sref("TreeNode")},
		}),
	})
	ct, ok := pkg.Types["pulumiservice:api:TreeNode"]
	if !ok {
		t.Fatalf("TreeNode not emitted")
	}
	items := ct.Properties["children"].Items
	if items == nil || items.Ref != "#/types/pulumiservice:api:TreeNode" {
		t.Errorf("children items should self-reference, got %+v", items)
	}
}

func TestResourceTokenCollisionSuffixed(t *testing.T) {
	// A component schema named Widget collides with the resource token.
	pkg := buildSynth(t, map[string]any{
		widgetRequest: obj(map[string]any{"inner": sref("Widget")}),
		"Widget":      obj(map[string]any{"name": map[string]any{typeKey: typeString}}),
	})
	if _, clash := pkg.Types["pulumiservice:api:Widget"]; clash {
		t.Fatalf("type token must not collide with the resource token")
	}
	if _, ok := pkg.Types["pulumiservice:api:WidgetProperties"]; !ok {
		t.Fatalf("collision should emit WidgetProperties; have %v", mapKeys(pkg.Types))
	}
	in := pkg.Resources["pulumiservice:api:Widget"].InputProperties["inner"]
	if got := in.Ref; got != "#/types/pulumiservice:api:WidgetProperties" {
		t.Errorf("inner ref: got %q", got)
	}
}

func discBase(tagProp string, mapping map[string]any, extraProps map[string]any) map[string]any {
	props := map[string]any{tagProp: map[string]any{typeKey: typeString}}
	for k, v := range extraProps {
		props[k] = v
	}
	return map[string]any{
		typeKey:    typeObject,
		propsKey:   props,
		"required": []any{tagProp},
		"discriminator": map[string]any{
			"propertyName": tagProp,
			"mapping":      mapping,
		},
	}
}

func variant(base string, extraProps map[string]any, required ...string) map[string]any {
	second := map[string]any{typeKey: typeObject, propsKey: extraProps}
	if len(required) > 0 {
		req := make([]any, len(required))
		for i, r := range required {
			req[i] = r
		}
		second["required"] = req
	}
	return map[string]any{allOfKey: []any{sref(base), second}}
}

func unionSchemas() map[string]any {
	return map[string]any{
		widgetRequest: obj(map[string]any{"shape": sref("Shape")}),
		"Shape": discBase(tagKind, map[string]any{
			tagCircle: circleRef,
			"square":  schemaRefPrefix + "Square",
			"blob":    schemaRefPrefix + "Blob",
		}, map[string]any{"label": map[string]any{typeKey: typeString}}),
		"Circle": variant("Shape", map[string]any{"radius": map[string]any{typeKey: typeNumber}}, "radius"),
		"Square": variant("Shape", map[string]any{"side": map[string]any{typeKey: typeNumber}}),
		"Blob": variant("Shape", map[string]any{
			"points": map[string]any{typeKey: typeArray, "items": map[string]any{typeKey: typeNumber}},
		}),
	}
}

func TestDiscriminatedUnion(t *testing.T) {
	pkg := buildSynth(t, unionSchemas())

	in := pkg.Resources["pulumiservice:api:Widget"].InputProperties["shape"]
	if got, want := len(in.OneOf), 3; got != want {
		t.Fatalf("oneOf members: got %d, want %d", got, want)
	}
	disc := in.Discriminator
	if disc == nil || disc.PropertyName != tagKind {
		t.Fatalf("discriminator: got %+v", disc)
	}
	if got, want := disc.Mapping[tagCircle], circleTypeRef; got != want {
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
	if kind.Const != tagCircle {
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
		tagCircle: circleRef,
	}
	pkg := buildSynth(t, schemas)
	in := pkg.Resources["pulumiservice:api:Widget"].InputProperties["shape"]
	if len(in.OneOf) != 0 {
		t.Fatalf("single-variant base must not produce a oneOf")
	}
	if got, want := in.Ref, circleTypeRef; got != want {
		t.Errorf("shape ref: got %q, want %q", got, want)
	}
}

func TestEmptyMappingFallsBackToNamedType(t *testing.T) {
	schemas := unionSchemas()
	schemas["Shape"].(map[string]any)["discriminator"].(map[string]any)["mapping"] = map[string]any{}
	pkg := buildSynth(t, schemas)
	in := pkg.Resources["pulumiservice:api:Widget"].InputProperties["shape"]
	if got, want := in.Ref, "#/types/pulumiservice:api:Shape"; got != want {
		t.Errorf("shape ref: got %q, want %q", got, want)
	}
}

func TestTwoMemberObjectUnionRejected(t *testing.T) {
	schemas := unionSchemas()
	schemas["Shape"].(map[string]any)["discriminator"].(map[string]any)["mapping"] = map[string]any{
		tagCircle: circleRef,
		"square":  schemaRefPrefix + "Square",
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

// markerSchemas extends unionSchemas with a BoolShape marker over Shape:
// circle, square, and blob chain through the marker; dot descends from Shape
// directly. The request references the marker.
func markerSchemas() map[string]any {
	schemas := unionSchemas()
	mapping := schemas[schemaShape].(map[string]any)["discriminator"].(map[string]any)["mapping"].(map[string]any)
	mapping["dot"] = schemaRefPrefix + "Dot"
	schemas[schemaBoolShape] = map[string]any{allOfKey: []any{
		sref(schemaShape),
		map[string]any{"description": "marker", typeKey: typeObject},
	}}
	for _, n := range []string{"Circle", schemaSquare, schemaBlob} {
		schemas[n] = variant(schemaBoolShape, map[string]any{
			strings.ToLower(n): map[string]any{typeKey: typeNumber},
		})
	}
	schemas["Dot"] = variant(schemaShape, map[string]any{"x": map[string]any{typeKey: typeNumber}})
	schemas[widgetRequest] = obj(map[string]any{condProp: sref(schemaBoolShape)})
	return schemas
}

// repointOutsideMarker moves variants to descend from Shape directly, taking
// them out of the BoolShape marker's subset.
func repointOutsideMarker(schemas map[string]any, names ...string) {
	for _, n := range names {
		schemas[n] = variant(schemaShape, map[string]any{
			strings.ToLower(n): map[string]any{typeKey: typeNumber},
		})
	}
}

func TestMarkerRendersDescendantSubsetUnion(t *testing.T) {
	// A property referencing a pure marker gets the union of the variants
	// that chain through the marker, not the base's full mapping.
	pkg := buildSynth(t, markerSchemas())
	in := pkg.Resources[widgetTok].InputProperties[condProp]
	if got, want := len(in.OneOf), 3; got != want {
		t.Fatalf("marker union members: got %d, want %d (%+v)", got, want, in.TypeSpec)
	}
	if in.Discriminator == nil || in.Discriminator.PropertyName != tagKind {
		t.Fatalf("marker union lost discriminator: %+v", in.Discriminator)
	}
	if _, ok := in.Discriminator.Mapping["dot"]; ok {
		t.Errorf("dot does not descend through the marker and must not appear")
	}
	if _, ok := pkg.Types["pulumiservice:api:BoolShape"]; ok {
		t.Errorf("marker type must not be emitted")
	}
}

func TestMarkerSubsetSingleVariantBecomesDefinite(t *testing.T) {
	schemas := markerSchemas()
	repointOutsideMarker(schemas, schemaSquare, schemaBlob)
	pkg := buildSynth(t, schemas)
	in := pkg.Resources[widgetTok].InputProperties[condProp]
	if got, want := in.Ref, circleTypeRef; got != want {
		t.Errorf("single-descendant marker should be the definite variant, got %+v", in.TypeSpec)
	}
}

func TestMarkerSubsetTwoMembersRejected(t *testing.T) {
	schemas := markerSchemas()
	repointOutsideMarker(schemas, schemaBlob)
	_, err := BuildSchema(synthSpec(t, schemas), widgetMetadata(), "pulumiservice")
	if err == nil || !strings.Contains(err.Error(), "2-member object unions") {
		t.Fatalf("expected 2-member marker subset to fail BuildSchema, got %v", err)
	}
}

func TestMarkerWithNoDescendantsErrors(t *testing.T) {
	schemas := markerSchemas()
	repointOutsideMarker(schemas, "Circle", schemaSquare, schemaBlob)
	_, err := BuildSchema(synthSpec(t, schemas), widgetMetadata(), "pulumiservice")
	if err == nil || !strings.Contains(err.Error(), "no variants") {
		t.Fatalf("expected empty marker subset to fail BuildSchema, got %v", err)
	}
}

func TestVariantReferenceBecomesDefiniteType(t *testing.T) {
	// A property referencing a mapped variant directly gets the definite
	// const-tagged variant type, identical to the union-member emission.
	schemas := unionSchemas()
	schemas[widgetRequest] = obj(map[string]any{"one": sref("Circle")})
	pkg := buildSynth(t, schemas)
	in := pkg.Resources[widgetTok].InputProperties["one"]
	if got, want := in.Ref, circleTypeRef; got != want {
		t.Fatalf("variant ref: got %+v, want %q", in.TypeSpec, want)
	}
	circle := pkg.Types["pulumiservice:api:Circle"]
	if circle.Properties[tagKind].Const != tagCircle {
		t.Errorf("directly-referenced variant must stay const-tagged, got %+v", circle.Properties[tagKind])
	}
}

func TestUnionPropertyDocListsTags(t *testing.T) {
	pkg := buildSynth(t, unionSchemas())
	in := pkg.Resources[widgetTok].InputProperties["shape"]
	if want := "Valid `kind` values: blob, circle, square."; !strings.Contains(in.Description, want) {
		t.Errorf("union property description missing tag list: %q", in.Description)
	}
}

// TestUnderscoreDiscriminatorUsesSuffixForm covers the whole schema-side
// surface of the `__type` → `type__` switch in one build: the const tag, the
// required list, the ordinary sibling property, the DiscriminatorSpec, and
// the generated doc line.
func TestUnderscoreDiscriminatorUsesSuffixForm(t *testing.T) {
	pkg := buildSynth(t, map[string]any{
		widgetRequest: obj(map[string]any{"shape": sref(schemaShape)}),
		schemaShape: discBase(wireType, map[string]any{
			tagCircle: circleRef,
			"square":  schemaRefPrefix + schemaSquare,
			"blob":    schemaRefPrefix + schemaBlob,
		}, map[string]any{"__label": map[string]any{typeKey: typeString}}),
		"Circle":     variant(schemaShape, map[string]any{"radius": map[string]any{typeKey: typeNumber}}),
		schemaSquare: variant(schemaShape, map[string]any{"side": map[string]any{typeKey: typeNumber}}),
		schemaBlob:   variant(schemaShape, map[string]any{"blobby": map[string]any{typeKey: typeBoolean}}),
	})

	circle := pkg.Types["pulumiservice:api:Circle"]
	if _, leaked := circle.Properties[wireType]; leaked {
		t.Errorf("wire name %q must not appear in the schema", wireType)
	}
	tag, ok := circle.Properties[schemaType]
	if !ok {
		t.Fatalf("expected a %q property; have %v", schemaType, mapKeys(circle.Properties))
	}
	if tag.Const != tagCircle {
		t.Errorf("%s const: got %v, want %q", schemaType, tag.Const, tagCircle)
	}
	if !slicesEqual(circle.Required, []string{schemaType}) {
		t.Errorf("required: got %v, want [%s]", circle.Required, schemaType)
	}
	// Non-discriminator `__` properties take the same encoding.
	if _, ok := circle.Properties["label__"]; !ok {
		t.Errorf("inherited __label should surface as label__; have %v", mapKeys(circle.Properties))
	}

	in := pkg.Resources[widgetTok].InputProperties["shape"]
	if in.Discriminator == nil || in.Discriminator.PropertyName != schemaType {
		t.Fatalf("discriminator propertyName: got %+v, want %q", in.Discriminator, schemaType)
	}
	if want := "Valid `" + schemaType + "` values: blob, circle, square."; !strings.Contains(in.Description, want) {
		t.Errorf("union doc should name the schema-side tag: %q", in.Description)
	}
}

func TestMarkerWithOwnPropertiesStaysNamed(t *testing.T) {
	// A subtype that adds its own properties is a real shape, not a
	// marker; it must stay a named type.
	schemas := unionSchemas()
	schemas["FatShape"] = map[string]any{allOfKey: []any{
		sref("Shape"),
		obj(map[string]any{"extra": map[string]any{"type": typeString}}),
	}}
	schemas[widgetRequest] = obj(map[string]any{"cond": sref("FatShape")})
	pkg := buildSynth(t, schemas)
	in := pkg.Resources["pulumiservice:api:Widget"].InputProperties["cond"]
	if got, want := in.Ref, "#/types/pulumiservice:api:FatShape"; got != want {
		t.Fatalf("non-marker subtype should stay named, got %+v", in.TypeSpec)
	}
}

func TestSharedTypeHoistsToApi(t *testing.T) {
	spec := synthSpec(t, map[string]any{
		widgetRequest: obj(map[string]any{"shared": sref("SharedThing")}),
		"SharedThing": obj(map[string]any{"x": map[string]any{typeKey: typeString}}),
	})
	meta := &Metadata{Resources: map[string]ResourceMeta{
		"pulumiservice:api/alpha:Widget": {Operations: Operations{Create: opCreateWidget}},
		"pulumiservice:api/beta:Gadget":  {Operations: Operations{Create: opCreateWidget}},
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
		widgetRequest: obj(map[string]any{"cfg": sref("OnlyHere")}),
		"OnlyHere":    obj(map[string]any{"x": map[string]any{typeKey: typeString}}),
	})
	meta := &Metadata{Resources: map[string]ResourceMeta{
		"pulumiservice:api/alpha:Widget": {Operations: Operations{Create: opCreateWidget}},
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
