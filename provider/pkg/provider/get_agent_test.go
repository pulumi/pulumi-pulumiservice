// Copyright 2026, Pulumi Corporation.

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

const (
	lookupAgentID         = "4f1d6f44-bf29-435b-af3c-c6309c0eacb4"
	lookupAgentIDKey      = "agentId"
	lookupDisplayNameKey  = "displayName"
	lookupAgentName       = "scout"
	lookupOrganizationKey = "organization"
	lookupNameKey         = "name"
)

func agentLookupProvider(t *testing.T, handler http.HandlerFunc) pulumirpc.ResourceProviderServer {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	provider, err := MakeProvider(nil, "pulumiservice", "1.0.0")
	require.NoError(t, err)
	args, err := structpb.NewStruct(map[string]any{"accessToken": "lookup-test-token", "apiUrl": server.URL})
	require.NoError(t, err)
	_, err = provider.Configure(t.Context(), &pulumirpc.ConfigureRequest{Args: args})
	require.NoError(t, err)
	return provider
}

func invokeAgentLookup(
	ctx context.Context, provider pulumirpc.ResourceProviderServer, org, id string,
) (*pulumirpc.InvokeResponse, error) {
	args, err := structpb.NewStruct(map[string]any{lookupOrganizationKey: org, lookupAgentIDKey: id})
	if err != nil {
		return nil, err
	}
	return provider.Invoke(ctx, &pulumirpc.InvokeRequest{Tok: "pulumiservice:index:getAgent", Args: args})
}

func TestGetAgentLookupUsesDurableID(t *testing.T) {
	var requests atomic.Int32
	provider := agentLookupProvider(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/preview/agents/example-org/definitions", r.URL.Path)
		assert.Empty(t, r.URL.RawQuery)
		assert.Equal(t, "token lookup-test-token", r.Header.Get("Authorization"))
		assert.Equal(t, "provider", r.Header.Get("X-Pulumi-Source"))
		name := lookupAgentName
		if requests.Add(1) > 1 {
			name = "renamed-scout"
		}
		_, err := fmt.Fprintf(w, `{"definitions":[
			{"id":"other-id","name":%q,"displayName":"Wrong agent"},
			{"id":%q,"name":%q,"displayName":"Issue scout","customInstructions":"private instructions","roleId":"private-role"}
		]}`, lookupAgentID, lookupAgentID, name)
		assert.NoError(t, err)
	})
	for _, name := range []string{lookupAgentName, "renamed-scout"} {
		response, err := invokeAgentLookup(t.Context(), provider, "example-org", lookupAgentID)
		require.NoError(t, err)
		require.Empty(t, response.Failures)
		assert.Equal(t, map[string]any{
			lookupOrganizationKey: "example-org",
			lookupAgentIDKey:      lookupAgentID,
			lookupNameKey:         name,
			lookupDisplayNameKey:  "Issue scout",
		}, response.Return.AsMap(), "lookup returns identity metadata only")
	}
	assert.EqualValues(t, 2, requests.Load())
}

func TestGetAgentLookupDoesNotResolveNamesOrUnavailableIDs(t *testing.T) {
	provider := agentLookupProvider(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/preview/agents/example-org/definitions", r.URL.Path)
		_, err := fmt.Fprintf(w, `{"definitions":[{"id":%q,"name":"scout","displayName":"Scout"}]}`, lookupAgentID)
		assert.NoError(t, err)
	})
	for _, id := range []string{lookupAgentName, "deleted-agent-id", "foreign-org-agent-id", "missing-agent-id"} {
		t.Run(id, func(t *testing.T) {
			_, err := invokeAgentLookup(t.Context(), provider, "example-org", id)
			require.ErrorContains(t, err, "not found or not visible")
		})
	}
}

func TestGetAgentLookupRejectsInvalidInputsWithoutRequests(t *testing.T) {
	provider := agentLookupProvider(t, func(http.ResponseWriter, *http.Request) {
		t.Error("invalid lookup must not call the API")
	})
	for _, input := range []struct{ org, id, message string }{
		{"", lookupAgentID, lookupOrganizationKey},
		{"../other", lookupAgentID, lookupOrganizationKey},
		{"org/other", lookupAgentID, lookupOrganizationKey},
		{"..", lookupAgentID, lookupOrganizationKey},
		{"org?filter=other", lookupAgentID, lookupOrganizationKey},
		{"example-org", " ", "id"},
	} {
		_, err := invokeAgentLookup(t.Context(), provider, input.org, input.id)
		require.ErrorContains(t, err, input.message)
	}
}

func TestGetAgentLookupPropagatesAPIErrors(t *testing.T) {
	for _, code := range []int{
		http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusServiceUnavailable,
	} {
		t.Run(http.StatusText(code), func(t *testing.T) {
			var requests atomic.Int32
			provider := agentLookupProvider(t, func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(code)
				_, err := fmt.Fprintf(w, `{"code":%d,"message":"lookup refused"}`, code)
				assert.NoError(t, err)
			})
			_, err := invokeAgentLookup(t.Context(), provider, "example-org", lookupAgentID)
			require.ErrorContains(t, err, fmt.Sprint(code))
			assert.EqualValues(t, 1, requests.Load(), "there is no permission or endpoint fallback")
		})
	}
}

func TestGetAgentLookupRejectsMalformedResponses(t *testing.T) {
	for _, body := range []string{`not-json`, `{"definitions":"wrong shape"}`, `{"definitions":[{"id":1}]}`} {
		provider := agentLookupProvider(t, func(w http.ResponseWriter, _ *http.Request) {
			_, err := fmt.Fprint(w, body)
			assert.NoError(t, err)
		})
		_, err := invokeAgentLookup(t.Context(), provider, "example-org", lookupAgentID)
		require.ErrorContains(t, err, "parse response")
	}
}

func TestGetAgentLookupHonorsCancellation(t *testing.T) {
	provider := agentLookupProvider(t, func(http.ResponseWriter, *http.Request) {
		t.Error("a canceled lookup must not send an HTTP request")
	})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := invokeAgentLookup(ctx, provider, "example-org", lookupAgentID)
	require.ErrorContains(t, err, "context canceled")
}

func TestGetAgentLookupSchema(t *testing.T) {
	provider, err := MakeProvider(nil, "pulumiservice", "1.0.0")
	require.NoError(t, err)
	response, err := provider.GetSchema(t.Context(), &pulumirpc.GetSchemaRequest{})
	require.NoError(t, err)
	var spec schema.PackageSpec
	require.NoError(t, json.Unmarshal([]byte(response.Schema), &spec))
	function, ok := spec.Functions["pulumiservice:index:getAgent"]
	require.True(t, ok)
	assert.ElementsMatch(t, []string{lookupOrganizationKey, lookupAgentIDKey}, function.Inputs.Required)
	require.NotNil(t, function.ReturnType)
	require.NotNil(t, function.ReturnType.ObjectTypeSpec)
	assert.ElementsMatch(t,
		[]string{lookupOrganizationKey, lookupAgentIDKey, lookupNameKey, lookupDisplayNameKey},
		function.ReturnType.ObjectTypeSpec.Required,
	)
	assert.Len(t, function.ReturnType.ObjectTypeSpec.Properties, 4)
}
