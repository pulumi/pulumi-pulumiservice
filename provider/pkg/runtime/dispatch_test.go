// Copyright 2016-2026, Pulumi Corporation.

package runtime

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDispatch_CreateReadDelete exercises the full generic CRUD path
// against a fake Pulumi Cloud API. The resource under test mirrors the
// AgentPool v1 shape to keep the fixture realistic.
func TestDispatch_CreateReadDelete(t *testing.T) {
	// Fake server implementing the AgentPool endpoints.
	var lastCreatedPool map[string]interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Assert the auth/source headers travel correctly.
		assert.Equal(t, "token test-access-token", r.Header.Get("Authorization"))
		assert.Equal(t, "provider", r.Header.Get("X-Pulumi-Source"))
		switch {
		case r.Method == "POST" && r.URL.Path == "/api/orgs/acme-corp/agent-pools":
			body, _ := io.ReadAll(r.Body)
			var req map[string]interface{}
			require.NoError(t, json.Unmarshal(body, &req))
			lastCreatedPool = map[string]interface{}{
				"id":          "pool-abc",
				"name":        req["name"],
				"description": req["description"],
				"tokenValue":  "secret-token-xyz",
			}
			w.WriteHeader(201)
			_ = json.NewEncoder(w).Encode(lastCreatedPool)
		case r.Method == "GET" && strings.HasPrefix(r.URL.Path, "/api/orgs/acme-corp/agent-pools/pool-abc"):
			if lastCreatedPool == nil {
				w.WriteHeader(404)
				return
			}
			_ = json.NewEncoder(w).Encode(lastCreatedPool)
		case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, "/api/orgs/acme-corp/agent-pools/pool-abc"):
			lastCreatedPool = nil
			w.WriteHeader(204)
		default:
			t.Logf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(500)
		}
	}))
	defer srv.Close()

	// Hand-roll the metadata the generator will emit for AgentPool.
	// Kept close to the real resource-map.yaml shape.
	meta := &CloudAPIMetadata{
		Resources: map[string]CloudAPIResource{
			"pulumiservice:orgs/agents:AgentPool": {
				Token:  "pulumiservice:orgs/agents:AgentPool",
				Module: "orgs/agents",
				// Path templates mirror the spec exactly (wire names like
				// `orgName`). ID templates use SDK property names (what
				// users see in `pulumi import`). The dispatcher translates
				// between the two via the property metadata.
				Create: &CloudAPIOperation{
					OperationID:  "CreateOrgAgentPool",
					Method:       "POST",
					PathTemplate: "/api/orgs/{orgName}/agent-pools",
				},
				Read: &CloudAPIOperation{
					OperationID:  "GetAgentPool",
					Method:       "GET",
					PathTemplate: "/api/orgs/{orgName}/agent-pools/{poolId}",
				},
				Delete: &CloudAPIOperation{
					OperationID:  "DeleteOrgAgentPool",
					Method:       "DELETE",
					PathTemplate: "/api/orgs/{orgName}/agent-pools/{poolId}",
				},
				ID: &CloudAPIID{
					Template: "{organizationName}/{poolId}",
					Params:   []string{"organizationName", "poolId"},
				},
				ForceNew: []string{"organizationName"},
				Properties: map[string]CloudAPIProperty{
					"organizationName": {Type: "string", From: "orgName", Source: "path"},
					"name":             {Type: "string", From: "name", Source: "body"},
					"description":      {Type: "string", From: "description", Source: "body"},
					"tokenValue":       {Type: "string", From: "tokenValue", Source: "response", Secret: true, Output: true},
					"poolId":           {Type: "string", From: "id", Source: "response", Output: true},
				},
			},
		},
	}
	client := NewClient(srv.URL, "test-access-token")
	d := NewDispatcher(client, meta)
	ctx := context.Background()
	token := "pulumiservice:orgs/agents:AgentPool"
	inputs := resource.PropertyMap{
		"organizationName": resource.NewStringProperty("acme-corp"),
		"name":             resource.NewStringProperty("vpc-isolated"),
		"description":      resource.NewStringProperty("Runs deployments inside the isolated VPC"),
	}

	// Create.
	id, outputs, err := d.Create(ctx, token, inputs)
	require.NoError(t, err)
	assert.Equal(t, "acme-corp/pool-abc", id)
	// Secret wrapping happened.
	tok := outputs["tokenValue"]
	require.True(t, tok.IsSecret(), "tokenValue should be marked secret in outputs")
	assert.Equal(t, "secret-token-xyz", tok.SecretValue().Element.StringValue())
	assert.Equal(t, "pool-abc", outputs["poolId"].StringValue())

	// Read.
	readID, readOutputs, err := d.Read(ctx, token, id, inputs)
	require.NoError(t, err)
	assert.Equal(t, id, readID)
	assert.Equal(t, "vpc-isolated", readOutputs["name"].StringValue())
	// Path-sourced inputs should be carried forward from priorInputs.
	assert.Equal(t, "acme-corp", readOutputs["organizationName"].StringValue())

	// Delete.
	require.NoError(t, d.Delete(ctx, token, id, outputs))

	// Read after delete → 404 → nil outputs.
	_, missing, err := d.Read(ctx, token, id, inputs)
	require.NoError(t, err)
	assert.Nil(t, missing, "Read of deleted resource should return nil outputs")
}
