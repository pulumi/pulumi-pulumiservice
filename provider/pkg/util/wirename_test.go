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

package util

import (
	"reflect"
	"testing"
)

const (
	wireType   = "__type"
	schemaType = "type__"
	allowTag   = "PermissionDescriptorAllow"
)

func TestToSchemaNameRewritesOnlyPrefixedNames(t *testing.T) {
	t.Parallel()
	cases := []struct{ wire, want string }{
		{wireType, schemaType},
		{"__kind", "kind__"},
		{"name", "name"},
		{"__", "__"},           // no base left over; not a rename candidate
		{"___type", "_type__"}, // one prefix stripped, remainder is the base
		{"type__", "type__"},   // already schema-side; unchanged
	}
	for _, c := range cases {
		if got := ToSchemaName(c.wire); got != c.want {
			t.Errorf("ToSchemaName(%q) = %q, want %q", c.wire, got, c.want)
		}
	}
}

func TestToWireNameInvertsToSchemaName(t *testing.T) {
	t.Parallel()
	for _, wire := range []string{wireType, "__kind", "name", "__"} {
		if got := ToWireName(ToSchemaName(wire)); got != wire {
			t.Errorf("ToWireName(ToSchemaName(%q)) = %q, want %q", wire, got, wire)
		}
	}
}

// The request path applies ToWireName twice on some keys: wireSideName resolves
// declared renames (which may already yield a wire name) and marshalWireBody
// then walks the whole tree. A second application must be a no-op.
func TestToWireNameIsIdempotent(t *testing.T) {
	t.Parallel()
	for _, name := range []string{schemaType, wireType, "name", "__"} {
		once := ToWireName(name)
		if twice := ToWireName(once); twice != once {
			t.Errorf("ToWireName(%q) not idempotent: %q then %q", name, once, twice)
		}
	}
}

// ToSchemaName is likewise applied twice on the response path — pulumiName
// covers the top level and ToSchemaTree covers every depth.
func TestToSchemaNameIsIdempotent(t *testing.T) {
	t.Parallel()
	for _, name := range []string{wireType, schemaType, "name", "__"} {
		once := ToSchemaName(name)
		if twice := ToSchemaName(once); twice != once {
			t.Errorf("ToSchemaName(%q) not idempotent: %q then %q", name, once, twice)
		}
	}
}

func TestRenameTreeRecursesThroughObjectsAndArrays(t *testing.T) {
	t.Parallel()
	schemaSide := map[string]any{
		schemaType: "PermissionDescriptorGroup",
		"entries": []any{
			map[string]any{
				schemaType:   "PermissionDescriptorCondition",
				"condition":  map[string]any{schemaType: "PermissionExpressionNot"},
				"subNode":    map[string]any{schemaType: allowTag, "permissions": []any{"stack:read"}},
				"unaffected": "value",
			},
			map[string]any{schemaType: allowTag},
		},
	}
	wireSide := map[string]any{
		wireType: "PermissionDescriptorGroup",
		"entries": []any{
			map[string]any{
				wireType:     "PermissionDescriptorCondition",
				"condition":  map[string]any{wireType: "PermissionExpressionNot"},
				"subNode":    map[string]any{wireType: allowTag, "permissions": []any{"stack:read"}},
				"unaffected": "value",
			},
			map[string]any{wireType: allowTag},
		},
	}

	if got := ToWireTree(schemaSide); !reflect.DeepEqual(got, wireSide) {
		t.Errorf("ToWireTree =\n%#v\nwant\n%#v", got, wireSide)
	}
	if got := ToSchemaTree(wireSide); !reflect.DeepEqual(got, schemaSide) {
		t.Errorf("ToSchemaTree =\n%#v\nwant\n%#v", got, schemaSide)
	}
}

func TestRenameTreeRoundTrips(t *testing.T) {
	t.Parallel()
	original := map[string]any{
		schemaType: "PermissionExpressionEqual",
		"left":     map[string]any{schemaType: "PermissionExpressionEnvironment"},
		"right": map[string]any{
			schemaType: "PermissionLiteralExpressionEnvironment",
			"value":    "acme/prod",
		},
	}
	if got := ToSchemaTree(ToWireTree(original)); !reflect.DeepEqual(got, original) {
		t.Errorf("round-trip lost fidelity:\n%#v\nwant\n%#v", got, original)
	}
}

// Nothing but the discriminator moves: scalars, nulls and free-form map keys
// that don't carry the marker survive both directions untouched.
func TestRenameTreeLeavesOrdinaryKeysAlone(t *testing.T) {
	t.Parallel()
	in := map[string]any{
		"name":        "infra",
		"count":       float64(3),
		"enabled":     true,
		"missing":     nil,
		"tags":        map[string]any{"owner": "platform", "_internal": "x"},
		"permissions": []any{"stack:read", "stack:write"},
	}
	if got := ToWireTree(in); !reflect.DeepEqual(got, in) {
		t.Errorf("ToWireTree mutated an unrelated tree:\n%#v", got)
	}
	if got := ToSchemaTree(in); !reflect.DeepEqual(got, in) {
		t.Errorf("ToSchemaTree mutated an unrelated tree:\n%#v", got)
	}
}

// renameTree must not alias the input: callers hand it maps decoded from a
// response and then keep using the original.
func TestRenameTreeDoesNotMutateInput(t *testing.T) {
	t.Parallel()
	nested := map[string]any{wireType: allowTag}
	in := map[string]any{wireType: "PermissionDescriptorGroup", "entries": []any{nested}}
	ToSchemaTree(in)
	if _, ok := in[wireType]; !ok {
		t.Error("ToSchemaTree rewrote the caller's top-level key in place")
	}
	if _, ok := nested[wireType]; !ok {
		t.Error("ToSchemaTree rewrote a nested map in place")
	}
}
