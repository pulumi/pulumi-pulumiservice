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
)

const (
	wireType   = "__type"
	schemaType = "type__"
)

// The two name resolvers layer the structural `__` rule (util.ToWireName /
// util.ToSchemaName) underneath the metadata's declared renames. A declared
// rename must still win outright.
func TestWireSideNameAppliesStructuralRuleAfterDeclaredRenames(t *testing.T) {
	t.Parallel()
	renames := map[string]string{orgNameKey: organizationNameVal}
	cases := []struct{ pulumi, want string }{
		{orgNameKey, organizationNameVal},
		{schemaType, wireType},
		{nameKey, nameKey},
	}
	for _, c := range cases {
		if got := wireSideName(c.pulumi, renames); got != c.want {
			t.Errorf("wireSideName(%q) = %q, want %q", c.pulumi, got, c.want)
		}
	}
}

func TestPulumiNameAppliesStructuralRuleAfterDeclaredRenames(t *testing.T) {
	t.Parallel()
	renames := map[string]string{orgNameKey: organizationNameVal}
	cases := []struct{ wire, want string }{
		{organizationNameVal, orgNameKey},
		{wireType, schemaType},
		{nameKey, nameKey},
	}
	for _, c := range cases {
		if got := pulumiName(c.wire, renames); got != c.want {
			t.Errorf("pulumiName(%q) = %q, want %q", c.wire, got, c.want)
		}
	}
}

// TestNoWirePropertyEndsInDoubleUnderscore guards the precondition that makes
// the suffix encoding reversible: it can only be injective while no Pulumi
// Cloud property already ends in `__`. A spec refresh that introduces one
// makes `foo__` ambiguous, and must fail here rather than in the field.
func TestNoWirePropertyEndsInDoubleUnderscore(t *testing.T) {
	t.Parallel()
	spec, _ := loadFixtures(t)

	prefixed := 0
	var visit func(node any)
	visit = func(node any) {
		switch n := node.(type) {
		case map[string]any:
			if props, ok := n[propsKey].(map[string]any); ok {
				for name := range props {
					if strings.HasSuffix(name, "__") {
						t.Errorf("wire property %q ends in `__`; the schema-side "+
							"suffix encoding is no longer reversible", name)
					}
					if strings.HasPrefix(name, "__") {
						prefixed++
					}
				}
			}
			for _, v := range n {
				visit(v)
			}
		case []any:
			for _, v := range n {
				visit(v)
			}
		}
	}
	for _, node := range spec.schemas {
		visit(node)
	}
	if prefixed == 0 {
		t.Fatal("no `__`-prefixed wire properties found; the fixture or the walk is wrong")
	}
}
