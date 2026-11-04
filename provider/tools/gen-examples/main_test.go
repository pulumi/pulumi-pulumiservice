// Copyright 2026, Pulumi Corporation.
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

package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

const testSchema = `{
  "name": "pulumiservice",
  "functions": {
    "pulumiservice:index:buildRolePermissions": {
      "inputs": {"properties": {
        "organizationScopes": {"type": "array", "items": {"$ref": "#/types/pulumiservice:index:RbacOrganizationScope"}},
        "stackRules": {"type": "array", "items": {"$ref": "#/types/pulumiservice:index:RoleStackRule"}}
      }},
      "outputs": {"properties": {}}
    }
  },
  "types": {
    "pulumiservice:index:RbacOrganizationScope": {
      "type": "string", "enum": [{"name": "StackRead", "value": "stack:read"}]
    },
    "pulumiservice:index:RbacStackScope": {
      "type": "string", "enum": [{"name": "StackRead", "value": "stack:read"}]
    },
    "pulumiservice:index:RoleStackRule": {"type": "object", "properties": {
      "scopes": {"type": "array", "items": {"$ref": "#/types/pulumiservice:index:RbacStackScope"}}
    }}
  }
}`

const testProgram = `
variables:
  permissions:
    fn::invoke:
      function: pulumiservice:buildRolePermissions
      arguments:
        organizationScopes: [stack:read]
        stackRules:
          - scopes: [stack:read]
`

var (
	orgStackRead   = enumValue{value: "stack:read", class: "RbacOrganizationScope", constant: "StackRead"}
	stackStackRead = enumValue{value: "stack:read", class: "RbacStackScope", constant: "StackRead"}
)

func TestEnumValuesFollowProgramOrderAndSchemaTypes(t *testing.T) {
	t.Parallel()
	var spec schema.PackageSpec
	require.NoError(t, json.Unmarshal([]byte(testSchema), &spec))
	pkg, err := schema.ImportSpec(spec, nil, schema.NewNullLoader(), schema.ValidationOptions{})
	require.NoError(t, err)
	var doc yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte(testProgram), &doc))

	enums, err := enumValues(pkg, doc.Content[0])
	require.NoError(t, err)
	assert.Equal(t, []enumValue{orgStackRead, stackStackRead}, enums)
}

func TestFixJavaEnums(t *testing.T) {
	t.Parallel()
	enums := []enumValue{orgStackRead, stackStackRead}

	t.Run("rewrites literals in order and imports each enum class", func(t *testing.T) {
		t.Parallel()
		code := "package p;\n\nimport com.pulumi.Context;\n\n" +
			`x.name("prod").organizationScopes("stack:read").scopes("stack:read");`
		got, err := fixJavaEnums(code, enums)
		require.NoError(t, err)
		assert.Equal(t, "package p;\n\n"+
			"import com.pulumi.pulumiservice.enums.RbacOrganizationScope;\n"+
			"import com.pulumi.pulumiservice.enums.RbacStackScope;\n"+
			"import com.pulumi.Context;\n\n"+
			`x.name("prod").organizationScopes(RbacOrganizationScope.StackRead).scopes(RbacStackScope.StackRead);`,
			got)
	})

	t.Run("shortens enums the generator already qualified", func(t *testing.T) {
		t.Parallel()
		code := "package p;\n\nimport com.pulumi.Context;\n\n" +
			`x.organizationScopes(com.pulumi.pulumiservice.enums.RbacOrganizationScope.StackRead).scopes("stack:read");`
		got, err := fixJavaEnums(code, enums)
		require.NoError(t, err)
		assert.Equal(t, "package p;\n\n"+
			"import com.pulumi.pulumiservice.enums.RbacOrganizationScope;\n"+
			"import com.pulumi.pulumiservice.enums.RbacStackScope;\n"+
			"import com.pulumi.Context;\n\n"+
			`x.organizationScopes(RbacOrganizationScope.StackRead).scopes(RbacStackScope.StackRead);`,
			got)
	})

	t.Run("fails when an enum value is missing from the output", func(t *testing.T) {
		t.Parallel()
		_, err := fixJavaEnums("package p;\n\nimport com.pulumi.Context;\n", enums)
		assert.ErrorContains(t, err, "found 0 of 2 enum values")
	})

	t.Run("leaves programs without enums unchanged", func(t *testing.T) {
		t.Parallel()
		got, err := fixJavaEnums(`x("stack:read");`, nil)
		require.NoError(t, err)
		assert.Equal(t, `x("stack:read");`, got)
	})
}
