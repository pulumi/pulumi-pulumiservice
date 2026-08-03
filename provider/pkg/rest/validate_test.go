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
	"strings"
	"testing"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

// pv converts plain Go data into a property.Value for readable test tables.
func pv(x any) property.Value {
	switch t := x.(type) {
	case property.Value:
		return t
	case string:
		return property.New(t)
	case bool:
		return property.New(t)
	case int:
		return property.New(float64(t))
	case float64:
		return property.New(t)
	case map[string]any:
		out := map[string]property.Value{}
		for k, v := range t {
			out[k] = pv(v)
		}
		return property.New(out)
	case []any:
		out := make([]property.Value, len(t))
		for i, v := range t {
			out[i] = pv(v)
		}
		return property.New(out)
	default:
		return property.New(property.Computed)
	}
}

func validateWidget(t *testing.T, schemas map[string]any, typeMeta map[string]TypeMeta, inputs map[string]any) []p.CheckFailure {
	t.Helper()
	spec := synthSpec(t, schemas)
	op, ok := spec.Op(opCreateWidget)
	if !ok {
		t.Fatalf("op %s not found", opCreateWidget)
	}
	vals := map[string]property.Value{}
	for k, v := range inputs {
		vals[k] = pv(v)
	}
	return ValidateInputs(spec, typeMeta, op, ResourceMeta{}, property.NewMap(vals))
}

func circleShape() map[string]any {
	return map[string]any{tagKind: tagCircle, "radius": 3, "label": "a"}
}

func TestValidateAcceptsValidUnion(t *testing.T) {
	failures := validateWidget(t, unionSchemas(), nil, map[string]any{"shape": circleShape()})
	if len(failures) != 0 {
		t.Fatalf("unexpected failures: %+v", failures)
	}
}

func TestValidateUnknownTagSuggests(t *testing.T) {
	failures := validateWidget(t, unionSchemas(), nil, map[string]any{
		"shape": map[string]any{tagKind: "circl", "radius": 3},
	})
	if len(failures) != 1 {
		t.Fatalf("want 1 failure, got %+v", failures)
	}
	f := failures[0]
	if f.Property != "shape.kind" {
		t.Errorf("property: got %q", f.Property)
	}
	for _, want := range []string{"expected one of: blob, circle, square", `did you mean "circle"`} {
		if !strings.Contains(f.Reason, want) {
			t.Errorf("reason %q missing %q", f.Reason, want)
		}
	}
}

func TestValidateMissingTag(t *testing.T) {
	failures := validateWidget(t, unionSchemas(), nil, map[string]any{
		"shape": map[string]any{"radius": 3},
	})
	if len(failures) != 1 || !strings.Contains(failures[0].Reason, `missing "kind"`) {
		t.Fatalf("want missing-tag failure, got %+v", failures)
	}
}

func TestValidateUnknownFieldSuggests(t *testing.T) {
	failures := validateWidget(t, unionSchemas(), nil, map[string]any{
		"shape": map[string]any{tagKind: tagCircle, "radiuss": 3},
	})
	if len(failures) != 1 {
		t.Fatalf("want 1 failure, got %+v", failures)
	}
	f := failures[0]
	if f.Property != "shape.radiuss" {
		t.Errorf("property: got %q", f.Property)
	}
	for _, want := range []string{"unknown field", "drops unknown fields silently", `did you mean "radius"`} {
		if !strings.Contains(f.Reason, want) {
			t.Errorf("reason %q missing %q", f.Reason, want)
		}
	}
}

func TestValidateSkipsComputed(t *testing.T) {
	failures := validateWidget(t, unionSchemas(), nil, map[string]any{
		"shape": property.New(property.Computed),
	})
	if len(failures) != 0 {
		t.Fatalf("computed value must skip validation, got %+v", failures)
	}
	failures = validateWidget(t, unionSchemas(), nil, map[string]any{
		"shape": map[string]any{tagKind: property.New(property.Computed), "radius": 3},
	})
	if len(failures) != 0 {
		t.Fatalf("computed tag must skip validation, got %+v", failures)
	}
}

func TestValidateSecretsAreWalked(t *testing.T) {
	failures := validateWidget(t, unionSchemas(), nil, map[string]any{
		"shape": pv(map[string]any{tagKind: "circl", "radius": 3}).WithSecret(true),
	})
	if len(failures) != 1 {
		t.Fatalf("secret values must still validate, got %+v", failures)
	}
}

func TestValidateWrongShapeKind(t *testing.T) {
	failures := validateWidget(t, unionSchemas(), nil, map[string]any{"shape": "circle"})
	if len(failures) != 1 || !strings.Contains(failures[0].Reason, `expected an object with a "kind" discriminator`) {
		t.Fatalf("want shape-kind failure, got %+v", failures)
	}
}

func TestValidateMarkerSubsetRejectsOutOfSubsetTag(t *testing.T) {
	failures := validateWidget(t, markerSchemas(), nil, map[string]any{
		condProp: map[string]any{tagKind: "dot", "x": 1},
	})
	if len(failures) != 1 {
		t.Fatalf("want 1 failure, got %+v", failures)
	}
	if !strings.Contains(failures[0].Reason, "expected one of: blob, circle, square") {
		t.Errorf("subset list wrong: %q", failures[0].Reason)
	}
}

func TestValidateDefiniteVariantTag(t *testing.T) {
	schemas := unionSchemas()
	schemas[widgetRequest] = obj(map[string]any{"one": sref("Circle")})
	failures := validateWidget(t, schemas, nil, map[string]any{
		"one": map[string]any{tagKind: "square", "radius": 3},
	})
	if len(failures) != 1 || !strings.Contains(failures[0].Reason, `"kind" must be "circle" here`) {
		t.Fatalf("want definite-variant tag failure, got %+v", failures)
	}
}

func TestValidateArrayElementPaths(t *testing.T) {
	schemas := unionSchemas()
	schemas[widgetRequest] = obj(map[string]any{
		"shapes": map[string]any{typeKey: typeArray, "items": sref(schemaShape)},
	})
	failures := validateWidget(t, schemas, nil, map[string]any{
		"shapes": []any{circleShape(), map[string]any{tagKind: "nope"}},
	})
	if len(failures) != 1 {
		t.Fatalf("want 1 failure, got %+v", failures)
	}
	if failures[0].Property != "shapes[1].kind" {
		t.Errorf("property: got %q", failures[0].Property)
	}
}

func TestValidateFreeFormObjectAccepted(t *testing.T) {
	failures := validateWidget(t, map[string]any{
		widgetRequest: obj(map[string]any{
			"metadata": map[string]any{typeKey: typeObject},
		}),
	}, nil, map[string]any{
		"metadata": map[string]any{"anything": "goes", "nested": map[string]any{"too": 1}},
	})
	if len(failures) != 0 {
		t.Fatalf("free-form objects must accept anything, got %+v", failures)
	}
}

func TestValidateTypeMetaAnySkips(t *testing.T) {
	failures := validateWidget(t, unionSchemas(), map[string]TypeMeta{
		schemaShape: {Any: true},
	}, map[string]any{
		"shape": map[string]any{tagKind: "nope"},
	})
	if len(failures) != 0 {
		t.Fatalf("TypeMeta.Any must skip validation, got %+v", failures)
	}
}

func TestValidateScalarShorthandAcceptsBothForms(t *testing.T) {
	schemas := map[string]any{
		widgetRequest: obj(map[string]any{"image": sref("Image")}),
		"Image":       obj(map[string]any{"name": map[string]any{typeKey: typeString}}),
	}
	tm := map[string]TypeMeta{"Image": {ScalarShorthand: typeString}}
	if failures := validateWidget(t, schemas, tm, map[string]any{"image": "nginx:latest"}); len(failures) != 0 {
		t.Fatalf("scalar form must pass, got %+v", failures)
	}
	if failures := validateWidget(t, schemas, tm, map[string]any{"image": map[string]any{"nam": "x"}}); len(failures) != 1 {
		t.Fatalf("object form must still validate fields, got %+v", failures)
	}
}

func TestCheckSurfacesValidationFailures(t *testing.T) {
	spec := synthSpec(t, unionSchemas())
	r := &Resource{
		spec: spec,
		meta: ResourceMeta{Operations: Operations{Create: opCreateWidget}},
	}
	resp, err := r.Check(t.Context(), p.CheckRequest{
		Inputs: property.NewMap(map[string]property.Value{
			"shape": pv(map[string]any{tagKind: "circl", "radius": 3}),
		}),
	})
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if len(resp.Failures) != 1 || resp.Failures[0].Property != "shape.kind" {
		t.Fatalf("check must surface validation failures, got %+v", resp.Failures)
	}
}
