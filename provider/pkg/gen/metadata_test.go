// Copyright 2016-2026, Pulumi Corporation.

package gen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/runtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEmitMetadata_AgentPool confirms that an AgentPool-shaped map entry
// translates into a complete runtime.CloudAPIResource record (path templates
// filled in from the spec, properties carried through with their flags).
func TestEmitMetadata_AgentPool(t *testing.T) {
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
          tokenValue:       { from: tokenValue, source: response, secret: true, output: true }
          poolId:           { from: id,      source: response, output: true }
`
	specPath := filepath.Join(dir, "spec.json")
	mapPath := filepath.Join(dir, "map.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(spec), 0644))
	require.NoError(t, os.WriteFile(mapPath, []byte(rmap), 0644))

	raw, err := EmitMetadata(specPath, mapPath)
	require.NoError(t, err)

	var md runtime.CloudAPIMetadata
	require.NoError(t, json.Unmarshal(raw, &md))

	r, ok := md.Resources["pulumiservice:orgs/agents:AgentPool"]
	require.True(t, ok, "AgentPool missing from emitted metadata")
	require.NotNil(t, r.Create, "Create operation should be populated")
	assert.Equal(t, "POST", r.Create.Method)
	assert.Equal(t, "/api/orgs/{orgName}/agent-pools", r.Create.PathTemplate)
	require.NotNil(t, r.Read)
	assert.Equal(t, "/api/orgs/{orgName}/agent-pools/{poolId}", r.Read.PathTemplate)
	assert.True(t, r.Properties["tokenValue"].Secret, "tokenValue must be marked secret")
	assert.True(t, r.Properties["tokenValue"].Output, "tokenValue must be output-only")
	assert.Equal(t, []string{"organizationName"}, r.ForceNew)
	require.NotNil(t, r.ID)
	assert.Equal(t, "{organizationName}/{poolId}", r.ID.Template)
}

// TestEmitMetadata_Polymorphic exercises the webhook-style scope discriminator.
func TestEmitMetadata_Polymorphic(t *testing.T) {
	dir := t.TempDir()
	spec := `{"paths": {
		"/api/orgs/{orgName}/hooks":      {"post": {"operationId": "CreateOrgHook"}},
		"/api/stacks/{orgName}/{projectName}/{stackName}/hooks": {"post": {"operationId": "CreateStackHook"}}
	}}`
	rmap := `modules:
  stacks/hooks:
    resources:
      Webhook:
        operations:
          case: scope
          scopes:
            org:   { create: CreateOrgHook }
            stack: { create: CreateStackHook }
        id:
          case: scope
          templates:
            org:   "{organizationName}/{name}"
            stack: "{organizationName}/{projectName}/{stackName}/{name}"
        properties:
          name: { source: body }
`
	specPath := filepath.Join(dir, "spec.json")
	mapPath := filepath.Join(dir, "map.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(spec), 0644))
	require.NoError(t, os.WriteFile(mapPath, []byte(rmap), 0644))

	raw, err := EmitMetadata(specPath, mapPath)
	require.NoError(t, err)
	var md runtime.CloudAPIMetadata
	require.NoError(t, json.Unmarshal(raw, &md))
	r, ok := md.Resources["pulumiservice:stacks/hooks:Webhook"]
	require.True(t, ok)
	require.NotNil(t, r.PolymorphicScopes)
	assert.Contains(t, r.PolymorphicScopes.Scopes, "org")
	assert.Contains(t, r.PolymorphicScopes.Scopes, "stack")
	assert.Equal(t, "CreateOrgHook", r.PolymorphicScopes.Scopes["org"].Create.OperationID)
	assert.Equal(t, "/api/stacks/{orgName}/{projectName}/{stackName}/hooks",
		r.PolymorphicScopes.Scopes["stack"].Create.PathTemplate)
}
