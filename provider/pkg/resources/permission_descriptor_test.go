// Copyright 2016-2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nestedDescriptorKind is a fixture exercising every kind the translator
// handles. Mirrors the env-scoped role shape used in examples/*-rbac.
func nestedDescriptorKind() map[string]interface{} {
	return map[string]interface{}{
		"kind": "group",
		"entries": []interface{}{
			map[string]interface{}{
				"kind": "condition",
				"condition": map[string]interface{}{
					"kind": "equal",
					"left": map[string]interface{}{
						"kind": "expressionEnvironment",
					},
					"right": map[string]interface{}{
						"kind":     "literalEnvironment",
						"identity": "env-uuid-1",
					},
				},
				"subNode": map[string]interface{}{
					"kind":        "allow",
					"permissions": []interface{}{"environment:read"},
				},
			},
		},
	}
}

func nestedDescriptorWire() map[string]interface{} {
	return map[string]interface{}{
		"__type": "PermissionDescriptorGroup",
		"entries": []interface{}{
			map[string]interface{}{
				"__type": "PermissionDescriptorCondition",
				"condition": map[string]interface{}{
					"__type": "PermissionExpressionEqual",
					"left": map[string]interface{}{
						"__type": "PermissionExpressionEnvironment",
					},
					"right": map[string]interface{}{
						"__type":   "PermissionLiteralExpressionEnvironment",
						"identity": "env-uuid-1",
					},
				},
				"subNode": map[string]interface{}{
					"__type":      "PermissionDescriptorAllow",
					"permissions": []interface{}{"environment:read"},
				},
			},
		},
	}
}

func TestPermissionsKindToWire_RoundTrip(t *testing.T) {
	t.Parallel()
	wire, err := permissionsKindToWire(nestedDescriptorKind())
	require.NoError(t, err)
	assert.Equal(t, nestedDescriptorWire(), wire)

	back, err := permissionsWireToKind(wire)
	require.NoError(t, err)
	assert.Equal(t, nestedDescriptorKind(), back)
}

func TestPermissionsKindToWire_AllKindsCovered(t *testing.T) {
	t.Parallel()
	// Every kind ↔ __type pair must round-trip. If a new variant is added,
	// add it here so the table stays exhaustive.
	for kind, wireType := range kindToWireType {
		input := map[string]interface{}{"kind": kind}
		wire, err := permissionsKindToWire(input)
		require.NoError(t, err)
		assert.Equal(t, wireType, wire["__type"], "kind=%q", kind)
		back, err := permissionsWireToKind(wire)
		require.NoError(t, err)
		assert.Equal(t, kind, back["kind"], "wire=%q", wireType)
	}
}

func TestPermissionsKindToWire_RejectsUnknownKind(t *testing.T) {
	t.Parallel()
	_, err := permissionsKindToWire(map[string]interface{}{"kind": "bogus"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bogus")
}

func TestPermissionsKindToWire_RejectsMissingKind(t *testing.T) {
	t.Parallel()
	_, err := permissionsKindToWire(map[string]interface{}{
		"permissions": []interface{}{"stack:read"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "kind")
}

func TestPermissionsWireToKind_RejectsUnknownType(t *testing.T) {
	t.Parallel()
	_, err := permissionsWireToKind(map[string]interface{}{
		"__type": "PermissionDescriptorBogus",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PermissionDescriptorBogus")
}

func TestPermissionsKindToWire_PreservesNonDescriptorFields(t *testing.T) {
	// `permissions` (the string list inside Allow), `identity` (inside Literal),
	// and arbitrary leaf scalars must pass through untouched.
	t.Parallel()
	in := map[string]interface{}{
		"kind":        "allow",
		"permissions": []interface{}{"stack:read", "stack:edit"},
	}
	out, err := permissionsKindToWire(in)
	require.NoError(t, err)
	assert.Equal(t, []interface{}{"stack:read", "stack:edit"}, out["permissions"])
}
