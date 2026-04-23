// Copyright 2016-2026, Pulumi Corporation.

package gen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEmitSchema_AgentPool verifies that a minimal agent-pool shape round-trips
// from resource-map.yaml through the emitter to a plausible Pulumi schema.json
// with correct property placement (inputs vs outputs), force-new markers, and
// secret wrapping.
func TestEmitSchema_AgentPool(t *testing.T) {
	dir := t.TempDir()
	spec := `{
		"paths": {
			"/api/orgs/{orgName}/agent-pools": {
				"post": {"operationId": "CreateOrgAgentPool"}
			},
			"/api/orgs/{orgName}/agent-pools/{poolId}": {
				"get":    {"operationId": "GetAgentPool"},
				"patch":  {"operationId": "PatchOrgAgentPool"},
				"delete": {"operationId": "DeleteOrgAgentPool"}
			}
		},
		"components": {
			"schemas": {
				"AgentPool": {"description": "Agent Pool for self-hosted deployments."}
			}
		}
	}`
	rmap := `modules:
  orgs/agents:
    resources:
      AgentPool:
        operations:
          create: CreateOrgAgentPool
          read:   GetAgentPool
          update: PatchOrgAgentPool
          delete: DeleteOrgAgentPool
        id:
          template: "{organizationName}/{poolId}"
          params: [organizationName, poolId]
        forceNew: [organizationName]
        properties:
          organizationName: { from: orgName, source: path }
          name:             { from: name,    source: body }
          description:      { from: description, source: body }
          tokenValue:       { from: tokenValue, source: response, secret: true, output: true }
          poolId:           { from: id,      source: response, output: true }
`
	specPath := filepath.Join(dir, "spec.json")
	mapPath := filepath.Join(dir, "map.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(spec), 0644))
	require.NoError(t, os.WriteFile(mapPath, []byte(rmap), 0644))

	raw, err := EmitSchema(specPath, mapPath)
	require.NoError(t, err)

	var sch pulumiSchema
	require.NoError(t, json.Unmarshal(raw, &sch))

	r, ok := sch.Resources["pulumiservice:orgs/agents:AgentPool"]
	require.True(t, ok, "AgentPool missing from emitted schema")
	assert.Equal(t, "Agent Pool for self-hosted deployments.", r.Description)

	// Outputs include everything (including output-only properties).
	assert.Contains(t, r.Properties, "tokenValue")
	assert.Contains(t, r.Properties, "poolId")
	assert.True(t, r.Properties["tokenValue"].Secret, "tokenValue must be marked secret")

	// Inputs exclude output-only properties.
	assert.Contains(t, r.InputProperties, "name")
	assert.NotContains(t, r.InputProperties, "tokenValue", "output-only property must not appear in inputs")
	assert.NotContains(t, r.InputProperties, "poolId", "output-only property must not appear in inputs")

	// force-new propagated.
	assert.True(t, r.InputProperties["organizationName"].WillReplaceOnChanges,
		"organizationName should be flagged willReplaceOnChanges")
	assert.True(t, r.Properties["organizationName"].WillReplaceOnChanges,
		"organizationName should be flagged on the output side too")
}

// TestEmitSchema_SkipsTodoResources confirms resources with TODO create
// operationIds are omitted from the emitted schema (they live under coverage
// report's TODO-markers section instead).
func TestEmitSchema_SkipsTodoResources(t *testing.T) {
	dir := t.TempDir()
	spec := `{"paths": {"/api/x": {"get": {"operationId": "GetX"}}}}`
	rmap := `modules:
  x:
    resources:
      X:
        operations:
          create: TODO:CreateX
          read:   GetX
        id: {template: "{name}", params: [name]}
        properties:
          name: { from: name, source: path }
`
	specPath := filepath.Join(dir, "spec.json")
	mapPath := filepath.Join(dir, "map.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(spec), 0644))
	require.NoError(t, os.WriteFile(mapPath, []byte(rmap), 0644))

	raw, err := EmitSchema(specPath, mapPath)
	require.NoError(t, err)
	var sch pulumiSchema
	require.NoError(t, json.Unmarshal(raw, &sch))
	assert.Empty(t, sch.Resources, "resources with TODO create op should not be emitted")
}
