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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fixtureCatalog = `{
  "global": [
    {"name": "Stack management", "scopes": [{"name": "stack:create", "metadata": {"description": "Create stack"}}]}
  ],
  "insights-account": [
    {"name": "Accounts", "scopes": [{"name": "insights_account:read", "metadata": {"description": "Read account"}}]}
  ],
  "stack": [
    {"name": "Stacks", "scopes": [
      {"name": "stack:read", "metadata": {"description": "Read stack"}},
      {"name": "stack:create", "metadata": {"description": "Create stack"}},
      {"name": "", "metadata": {"description": ""}}
    ]}
  ]
}`

const fixtureSpec = `{"components": {"schemas": {
  "Unrelated": {"properties": {"n": {"enum": [1, 2]}}},
  "RbacScope": {"properties": {"name": {
    "enum": ["", "stack:read", "stack:create", "saml:read"],
    "x-pulumi-model-property": {"enumFieldNames": ["NoPermission", "StackRead", "StackCreate", "SAMLRead"]}
  }}}
}}}`

func loadFixture(t *testing.T) model {
	t.Helper()
	var catalog scopeCatalog
	require.NoError(t, json.Unmarshal([]byte(fixtureCatalog), &catalog))
	names, err := specFieldNames([]byte(fixtureSpec))
	require.NoError(t, err)
	m, err := buildModel(catalog, names)
	require.NoError(t, err)
	return m
}

func TestBuildModel(t *testing.T) {
	m := loadFixture(t)

	require.Len(t, m.Scopes, 3, "empty scope skipped, stack:create deduplicated")
	assert.Equal(t, "insights_account:read", m.Scopes[0].Value)
	assert.Equal(t, "InsightsAccountRead", m.Scopes[0].Name, "name derived when missing from spec")
	assert.Equal(t, "StackCreate", m.Scopes[1].Name, "name taken from spec")
	assert.Equal(t, []string{"global", "stack"}, m.Scopes[1].ResourceTypes)
	assert.Equal(t, []string{"Stack management", "Stacks"}, m.Scopes[1].Groups)

	var rts []string
	for _, rt := range m.ResourceTypes {
		rts = append(rts, rt.Name)
	}
	assert.Equal(t, []string{"Global", "InsightsAccount", "Stack"}, rts)
}

func TestBuildModelNameCollision(t *testing.T) {
	var catalog scopeCatalog
	require.NoError(t, json.Unmarshal([]byte(`{"stack": [{"name": "Stacks", "scopes": [
	  {"name": "stack:read", "metadata": {}},
	  {"name": "stack_read", "metadata": {}}
	]}]}`), &catalog))
	_, err := buildModel(catalog, map[string]string{})
	assert.ErrorContains(t, err, `"StackRead"`)
}

func TestBuildModelEmpty(t *testing.T) {
	_, err := buildModel(scopeCatalog{}, map[string]string{})
	assert.Error(t, err)
}

func TestRender(t *testing.T) {
	src, err := render(loadFixture(t))
	require.NoError(t, err)
	out := string(src)

	assert.Contains(t, out, "DO NOT EDIT")
	assert.Contains(t, out, `RbacResourceTypeInsightsAccount RbacResourceType = "insights-account"`)
	assert.Contains(t, out, `RbacScopeStackCreate         RbacScope = "stack:create"`)
	assert.Contains(t, out, `Description: "Create stack. Applies to: global, stack (Stack management, Stacks)."`)

	// stack:create is listed under both resource types it belongs to.
	stackBlock := out[strings.Index(out, "RbacResourceTypeStack: {"):]
	assert.Contains(t, stackBlock, "RbacScopeStackCreate,")
	globalStart := strings.Index(out, "RbacResourceTypeGlobal: {")
	globalBlock := out[globalStart:strings.Index(out, "RbacResourceTypeInsightsAccount: {")]
	assert.Contains(t, globalBlock, "RbacScopeStackCreate,")
	assert.NotContains(t, globalBlock, "RbacScopeStackRead,")
}

func TestPascal(t *testing.T) {
	assert.Equal(t, "StackDeploymentRead", pascal("stack_deployment:read"))
	assert.Equal(t, "InsightsAccount", pascal("insights-account"))
}
