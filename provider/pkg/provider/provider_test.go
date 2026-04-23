// Copyright 2016-2026, Pulumi Corporation.

package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	pbempty "google.golang.org/protobuf/types/known/emptypb"
	structpb "google.golang.org/protobuf/types/known/structpb"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServer_GetSchema confirms the embedded schema is returned verbatim.
func TestServer_GetSchema(t *testing.T) {
	schemaJSON := `{"name":"pulumiservice"}`
	metadataJSON := `{"resources":{}}`

	srv, err := New("pulumiservice", "1.0.0-test", []byte(schemaJSON), []byte(metadataJSON))
	require.NoError(t, err)
	resp, err := srv.GetSchema(context.Background(), &pulumirpc.GetSchemaRequest{})
	require.NoError(t, err)
	assert.Equal(t, schemaJSON, resp.Schema)
}

// TestServer_GetPluginInfo reports the provider version.
func TestServer_GetPluginInfo(t *testing.T) {
	srv, err := New("pulumiservice", "2.0.0-alpha", []byte(`{}`), []byte(`{}`))
	require.NoError(t, err)
	resp, err := srv.GetPluginInfo(context.Background(), &pbempty.Empty{})
	require.NoError(t, err)
	assert.Equal(t, "2.0.0-alpha", resp.Version)
}

// TestServer_ConfigureAndCreate exercises the full path: Configure stashes
// credentials, then a Create dispatches through the metadata-driven runtime
// against a fake Pulumi Cloud server. Proves the gRPC → dispatcher → HTTP
// wiring end-to-end.
func TestServer_ConfigureAndCreate(t *testing.T) {
	// Fake Pulumi Cloud.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "token test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "provider", r.Header.Get("X-Pulumi-Source"))
		body, _ := io.ReadAll(r.Body)
		var req map[string]interface{}
		_ = json.Unmarshal(body, &req)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          "pool-abc",
			"name":        req["name"],
			"description": req["description"],
			"tokenValue":  "secret-xyz",
		})
	}))
	defer srv.Close()

	// Metadata mirroring AgentPool v1 shape.
	metadata := `{
		"resources": {
			"pulumiservice:orgs/agents:AgentPool": {
				"token": "pulumiservice:orgs/agents:AgentPool",
				"module": "orgs/agents",
				"create": {
					"operationId": "CreateOrgAgentPool",
					"method": "POST",
					"path": "/api/orgs/{orgName}/agent-pools"
				},
				"id": {
					"template": "{organizationName}/{agentPoolId}",
					"params": ["organizationName", "agentPoolId"]
				},
				"properties": {
					"organizationName": {"from": "orgName", "source": "path"},
					"name":             {"from": "name", "source": "body"},
					"description":      {"from": "description", "source": "body"},
					"tokenValue":       {"from": "tokenValue", "source": "response", "secret": true, "output": true},
					"agentPoolId":      {"from": "id", "source": "response", "output": true}
				}
			}
		}
	}`
	ps, err := New("pulumiservice", "2.0.0-test", []byte(`{}`), []byte(metadata))
	require.NoError(t, err)

	ctx := context.Background()
	_, err = ps.Configure(ctx, &pulumirpc.ConfigureRequest{
		Args: &structpb.Struct{Fields: map[string]*structpb.Value{
			"accessToken": structpb.NewStringValue("test-token"),
			"apiUrl":      structpb.NewStringValue(srv.URL),
		}},
	})
	require.NoError(t, err)

	props, err := structpb.NewStruct(map[string]interface{}{
		"organizationName": "acme-corp",
		"name":             "vpc-pool",
		"description":      "VPC-isolated agent pool",
	})
	require.NoError(t, err)

	resp, err := ps.Create(ctx, &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:orgs/agents:AgentPool::my-pool",
		Properties: props,
	})
	require.NoError(t, err)
	assert.Equal(t, "acme-corp/pool-abc", resp.GetId())

	out := resp.GetProperties().GetFields()
	require.Contains(t, out, "tokenValue")
	// Secret-wrapped values are serialized as a sentinel object (secret=true + value).
	sv := out["tokenValue"]
	// Unwrap the structpb secret form if present.
	if s := sv.GetStructValue(); s != nil {
		if _, isSecret := s.GetFields()["4dabf18193072939515e22adb298388d"]; isSecret {
			inner := s.GetFields()["value"]
			assert.Equal(t, "secret-xyz", inner.GetStringValue())
			return
		}
	}
	t.Fatalf("tokenValue in response should be a wrapped secret, got %v", sv)
}

// TestServer_TokenFromURN confirms the URN parsing.
func TestServer_TokenFromURN(t *testing.T) {
	urn := "urn:pulumi:dev::test::pulumiservice:orgs/agents:AgentPool::pool"
	assert.Equal(t, "pulumiservice:orgs/agents:AgentPool", tokenFromURN(urn))
}
