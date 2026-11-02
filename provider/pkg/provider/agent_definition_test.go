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

package provider

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

const capabilityOrganizationKey = "organizationName"
const capabilityOrganizationName = "factory-org"
const agentDefinitionRenamedName = "renamed-factory-scout"

const agentDefinitionURN = "urn:pulumi:dev::test::pulumiservice:index:AgentDefinition::scout"

func capabilityProperties(t *testing.T, values map[string]any) *structpb.Struct {
	t.Helper()
	properties, err := structpb.NewStruct(values)
	require.NoError(t, err)
	return properties
}

func TestAgentDefinitionCRUD(t *testing.T) {
	var saved map[string]any
	var writes []map[string]any
	deleted := false
	provider := agentLookupProvider(t, func(w http.ResponseWriter, r *http.Request) {
		const collection = "/api/preview/agents/" + capabilityOrganizationName + "/definitions"
		if r.Method == http.MethodGet && r.URL.Path == collection {
			definitions := []map[string]any{}
			if !deleted {
				definitions = append(definitions, saved)
			}
			assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"definitions": definitions}))
			return
		}
		if r.Method == http.MethodPost {
			assert.Equal(t, collection, r.URL.Path)
		} else {
			assert.Equal(t, collection+"/"+lookupAgentID, r.URL.Path)
		}
		switch r.Method {
		case http.MethodPost, http.MethodPut:
			var body map[string]any
			if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&body)) {
				return
			}
			writes = append(writes, body)
			saved = map[string]any{"id": lookupAgentID, "taskCounts": map[string]any{"running": 2}}
			for key, value := range body {
				saved[key] = value
			}
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
			}
			assert.NoError(t, json.NewEncoder(w).Encode(saved))
		case http.MethodDelete:
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			// Historical GET-by-ID still returns a deleted definition. Refresh
			// must use the active roster instead.
			t.Error("resource read must not use the historical GET-by-ID endpoint")
			assert.NoError(t, json.NewEncoder(w).Encode(saved))
		default:
			t.Errorf("unexpected method %s", r.Method)
		}
	})
	inputs := capabilityProperties(t, map[string]any{
		capabilityOrganizationKey: capabilityOrganizationName,
		lookupNameKey:             lookupAgentName, lookupDisplayNameKey: "Issue scout",
		"roleId": "member-role", "customInstructions": "Review incoming issues.",
		"purpose": "Issue triage", "avatar": "search", "approvalMode": "auto",
		"enabledIntegrations": []any{map[string]any{"id": "honeycomb"}},
	})
	created, err := provider.Create(t.Context(), &pulumirpc.CreateRequest{Urn: agentDefinitionURN, Properties: inputs})
	require.NoError(t, err)
	assert.Equal(t, lookupAgentID, created.Id)
	assert.NotContains(t, created.Properties.AsMap(), "taskCounts")
	assert.Equal(t, "member-role", created.Properties.AsMap()["roleId"])
	assert.NotContains(t, writes[0], capabilityOrganizationKey)

	updatedInputs := capabilityProperties(t, map[string]any{
		capabilityOrganizationKey: capabilityOrganizationName, lookupNameKey: agentDefinitionRenamedName,
		lookupDisplayNameKey: "Renamed scout",
		"roleId":             "updated-role", "customInstructions": "Review issues and propose fixes.",
		"purpose": "Issue triage", "avatar": "search", "approvalMode": "manual", "enabledIntegrations": []any{},
	})
	diff, err := provider.Diff(t.Context(), &pulumirpc.DiffRequest{
		Urn: agentDefinitionURN, Id: created.Id, Olds: created.Properties, OldInputs: inputs, News: updatedInputs,
	})
	require.NoError(t, err)
	assert.Empty(t, diff.Replaces, "renaming and changing configuration must retain the durable ID")
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, diff.Changes)
	updated, err := provider.Update(t.Context(), &pulumirpc.UpdateRequest{
		Urn: agentDefinitionURN, Id: created.Id, Olds: created.Properties, OldInputs: inputs, News: updatedInputs,
	})
	require.NoError(t, err)
	assert.Equal(t, agentDefinitionRenamedName, updated.Properties.AsMap()[lookupNameKey])
	assert.Equal(t, "updated-role", updated.Properties.AsMap()["roleId"])
	assert.Equal(t, []any{}, writes[1]["enabledIntegrations"])

	read, err := provider.Read(t.Context(), &pulumirpc.ReadRequest{
		Urn: agentDefinitionURN, Id: created.Id, Properties: updated.Properties, Inputs: updatedInputs,
	})
	require.NoError(t, err)
	assert.Equal(t, lookupAgentID, read.Id)
	assert.Equal(t, updatedInputs.AsMap(), read.Properties.AsMap())
	imported, err := provider.Read(t.Context(), &pulumirpc.ReadRequest{
		Urn: agentDefinitionURN, Id: capabilityOrganizationName + "/" + lookupAgentID,
	})
	require.NoError(t, err)
	assert.Equal(t, lookupAgentID, imported.Id)
	assert.Equal(t, read.Properties.AsMap(), imported.Properties.AsMap())

	otherOrg := updatedInputs.AsMap()
	otherOrg[capabilityOrganizationKey] = "other-org"
	diff, err = provider.Diff(t.Context(), &pulumirpc.DiffRequest{
		Urn: agentDefinitionURN, Id: read.Id, Olds: read.Properties, OldInputs: updatedInputs,
		News: capabilityProperties(t, otherOrg),
	})
	require.NoError(t, err)
	assert.Contains(t, diff.Replaces, capabilityOrganizationKey)

	_, err = provider.Delete(t.Context(), &pulumirpc.DeleteRequest{
		Urn: agentDefinitionURN, Id: read.Id, Properties: read.Properties,
	})
	require.NoError(t, err)
	missing, err := provider.Read(t.Context(), &pulumirpc.ReadRequest{
		Urn: agentDefinitionURN, Id: read.Id, Properties: read.Properties, Inputs: read.Inputs,
	})
	require.NoError(t, err)
	assert.Empty(t, missing.Id, "deleted definitions must not be resurrected on refresh")
}

func TestAgentDefinitionIntegrationInheritance(t *testing.T) {
	for _, tc := range []struct {
		name         string
		integrations any
	}{
		{name: "inherit"},
		{name: "none", integrations: []any{}},
		{name: "catalog IDs", integrations: []any{map[string]any{"id": "honeycomb"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			provider := agentLookupProvider(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Contains(t, []string{http.MethodPost, http.MethodPut}, r.Method)
				var body map[string]any
				if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&body)) {
					return
				}
				assert.Equal(t, tc.integrations, body["enabledIntegrations"])
				assert.NotContains(t, body, "roleId", "omitting the role must let the API default or preserve it")
				body["id"] = lookupAgentID
				body["roleId"] = "read-only-role"
				body[lookupDisplayNameKey] = body[lookupNameKey]
				// The API omits approvalMode on older definitions; those runs
				// still default to auto approval.
				delete(body, "approvalMode")
				assert.NoError(t, json.NewEncoder(w).Encode(body))
			})
			values := map[string]any{
				capabilityOrganizationKey: capabilityOrganizationName, lookupNameKey: lookupAgentName,
			}
			if tc.integrations != nil {
				values["enabledIntegrations"] = tc.integrations
			}
			inputs := capabilityProperties(t, values)
			checked, err := provider.Check(t.Context(), &pulumirpc.CheckRequest{
				Urn: agentDefinitionURN, News: inputs,
			})
			require.NoError(t, err)
			require.Empty(t, checked.Failures)
			inputs = checked.Inputs
			assert.Equal(t, "auto", inputs.AsMap()["approvalMode"])
			created, err := provider.Create(t.Context(), &pulumirpc.CreateRequest{
				Urn: agentDefinitionURN, Properties: inputs,
			})
			require.NoError(t, err)
			assert.Equal(t, tc.integrations, created.Properties.AsMap()["enabledIntegrations"])
			assert.Equal(t, lookupAgentName, created.Properties.AsMap()[lookupDisplayNameKey])
			assert.Equal(t, "read-only-role", created.Properties.AsMap()["roleId"])
			assert.Equal(t, "auto", created.Properties.AsMap()["approvalMode"])
			values["purpose"] = "Updated purpose"
			_, err = provider.Update(t.Context(), &pulumirpc.UpdateRequest{
				Urn: agentDefinitionURN, Id: created.Id, Olds: created.Properties,
				OldInputs: inputs, News: capabilityProperties(t, values),
			})
			require.NoError(t, err)
		})
	}
}

func TestAgentDefinitionPreviewAndErrors(t *testing.T) {
	provider := agentLookupProvider(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{"code": 403, "message": "access denied"}))
	})
	inputs := capabilityProperties(t, map[string]any{
		capabilityOrganizationKey: capabilityOrganizationName, lookupNameKey: lookupAgentName,
	})
	preview, err := provider.Create(t.Context(), &pulumirpc.CreateRequest{
		Urn: agentDefinitionURN, Properties: inputs, Preview: true,
	})
	require.NoError(t, err)
	assert.Empty(t, preview.Id)
	_, err = provider.Read(t.Context(), &pulumirpc.ReadRequest{
		Urn: agentDefinitionURN, Id: capabilityOrganizationName + "/" + lookupAgentID,
	})
	require.ErrorContains(t, err, "403")
	_, err = provider.Create(t.Context(), &pulumirpc.CreateRequest{Urn: agentDefinitionURN, Properties: inputs})
	require.ErrorContains(t, err, "403")
}
