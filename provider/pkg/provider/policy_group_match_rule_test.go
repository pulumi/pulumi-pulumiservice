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

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

func TestPolicyGroupStackMatchRuleLifecycle(t *testing.T) {
	for _, mode := range []string{"preventative", "audit"} {
		t.Run(mode, func(t *testing.T) {
			const groupURN = "urn:pulumi:dev::test::pulumiservice:index:PolicyGroup::review"
			const collection = "/api/orgs/" + capabilityOrganizationName + "/policygroups"
			var saved map[string]any
			var operations []map[string]any
			var createRule any
			provider := agentLookupProvider(t, func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodPost:
					assert.Equal(t, collection, r.URL.Path)
					if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&saved)) {
						return
					}
					createRule = saved["stackMatchRule"]
					saved["stacks"] = []any{}
					saved["accounts"] = []any{}
					saved["appliedPolicyPacks"] = []any{}
					w.WriteHeader(http.StatusCreated)
				case http.MethodPatch:
					assert.Equal(t, collection+"/review/batch", r.URL.Path)
					var batch []map[string]any
					if !assert.NoError(t, json.NewDecoder(r.Body).Decode(&batch)) {
						return
					}
					operations = append(operations, batch...)
					for _, op := range batch {
						if pack, ok := op["addPolicyPack"].(map[string]any); ok {
							pack["version"] = 7
							saved["appliedPolicyPacks"] = append(saved["appliedPolicyPacks"].([]any), pack)
						}
						if stack, ok := op["addStack"]; ok {
							saved["stacks"] = append(saved["stacks"].([]any), stack)
						}
						if rule, ok := op["stackMatchRule"]; ok {
							assert.NotContains(t, op, "removeStackMatchRule")
							saved["stackMatchRule"] = rule
						}
						if op["removeStackMatchRule"] == true {
							assert.NotContains(t, op, "stackMatchRule")
							delete(saved, "stackMatchRule")
						}
					}
					w.WriteHeader(http.StatusNoContent)
				case http.MethodGet:
					assert.Equal(t, collection+"/review", r.URL.Path)
					assert.NoError(t, json.NewEncoder(w).Encode(saved))
				default:
					t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				}
			})
			buildRule := func(template string) any {
				t.Helper()
				result, err := provider.Invoke(t.Context(), &pulumirpc.InvokeRequest{
					Tok: "pulumiservice:index:buildStackMatchRule",
					Args: capabilityProperties(t, map[string]any{"rules": []any{map[string]any{
						"conditions": []any{map[string]any{
							"key": "pulumi:from-pull-request-template", "value": template,
						}},
					}}}),
				})
				require.NoError(t, err)
				return result.Return.AsMap()["matchRule"]
			}
			firstRule := buildRule("pulumi_local/review-project/template")
			values := map[string]any{
				capabilityOrganizationKey: capabilityOrganizationName,
				lookupNameKey:             "review", "entityType": "stacks", "mode": mode,
				"stackMatchRule": firstRule,
				"stacks":         []any{map[string]any{lookupNameKey: "static", "routingProject": "review-project"}},
				"policyPacks": []any{map[string]any{
					lookupNameKey: "demo-policy", "versionTag": "v1", "config": map[string]any{
						"all": map[string]any{"enforcementLevel": "mandatory"},
					},
				}},
			}
			inputs := capabilityProperties(t, values)
			created, err := provider.Create(t.Context(), &pulumirpc.CreateRequest{Urn: groupURN, Properties: inputs})
			require.NoError(t, err)
			assert.Equal(t, firstRule, createRule, "attach the rule in the create request")
			assert.Equal(t, firstRule, created.Properties.AsMap()["stackMatchRule"])
			assert.Len(t, created.Properties.AsMap()["policyPacks"], 1)
			assert.Len(t, operations, 2, "only static stack and policy pack attachments follow create")

			state := created.Properties
			for _, rule := range []any{buildRule("pulumi_local/review-project/next-template"), nil, firstRule} {
				oldInputs := inputs
				delete(values, "stackMatchRule")
				if rule != nil {
					values["stackMatchRule"] = rule
				}
				inputs = capabilityProperties(t, values)
				diff, err := provider.Diff(t.Context(), &pulumirpc.DiffRequest{
					Urn: groupURN, Id: created.Id, Olds: state, OldInputs: oldInputs, News: inputs,
				})
				require.NoError(t, err)
				assert.Empty(t, diff.Replaces)
				assert.Contains(t, diff.DetailedDiff, "stackMatchRule")
				assert.NotContains(t, diff.DetailedDiff, "policyPacks")
				operations = nil
				updated, err := provider.Update(t.Context(), &pulumirpc.UpdateRequest{
					Urn: groupURN, Id: created.Id, Olds: state, OldInputs: oldInputs, News: inputs,
				})
				require.NoError(t, err)
				require.Len(t, operations, 1, "rule edits must leave static stacks and packs attached")
				if rule == nil {
					assert.Equal(t, map[string]any{"removeStackMatchRule": true}, operations[0])
				} else {
					assert.Equal(t, map[string]any{"stackMatchRule": rule}, operations[0])
				}
				read, err := provider.Read(t.Context(), &pulumirpc.ReadRequest{
					Urn: groupURN, Id: created.Id, Properties: updated.Properties, Inputs: inputs,
				})
				require.NoError(t, err)
				assert.Equal(t, rule, read.Properties.AsMap()["stackMatchRule"])
				assert.Equal(t, rule, read.Inputs.AsMap()["stackMatchRule"])
				assert.Equal(t, created.Properties.AsMap()["policyPacks"], read.Properties.AsMap()["policyPacks"])
				assert.Equal(t, created.Properties.AsMap()["stacks"], read.Properties.AsMap()["stacks"])
				state = read.Properties
				noDiff, err := provider.Diff(t.Context(), &pulumirpc.DiffRequest{
					Urn: groupURN, Id: created.Id, Olds: state, OldInputs: read.Inputs, News: inputs,
				})
				require.NoError(t, err)
				assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, noDiff.Changes)
			}
		})
	}
}
