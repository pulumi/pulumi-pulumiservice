// Copyright 2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"encoding/json"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/urn"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

func TestAutoName(t *testing.T) {
	t.Parallel()

	prov, err := infer.NewProviderBuilder().
		WithResources(
			infer.Resource(&Webhook{}),
			infer.Resource(&AgentPool{}),
			infer.Resource(&ApprovalRule{}),
			infer.Resource(&OidcIssuer{}),
			infer.Resource(&OrganizationRole{}),
		).
		WithModuleMap(map[tokens.ModuleName]tokens.ModuleName{"resources": "index"}).
		Build()
	require.NoError(t, err)
	server, err := integration.NewServer(t.Context(), "pulumiservice", semver.MustParse("0.0.0"),
		integration.WithProvider(prov))
	require.NoError(t, err)

	seed := []byte("autoname-seed")
	str := property.New[string]

	for _, tc := range []struct {
		token  tokens.Type
		field  string
		maxLen int
		inputs map[string]property.Value
	}{
		{
			token:  "pulumiservice:index:Webhook",
			field:  gcDisplayName,
			maxLen: webhookDisplayNameMaxLen,
			inputs: map[string]property.Value{
				gcOrganizationName: str(gcMyOrg),
				"active":           property.New(true),
				"payloadUrl":       str("https://example.com/hook"),
			},
		},
		{
			token:  "pulumiservice:index:AgentPool",
			field:  gcName,
			maxLen: agentPoolNameMaxLen,
			inputs: map[string]property.Value{gcOrganizationName: str(gcMyOrg)},
		},
		{
			token:  "pulumiservice:index:ApprovalRule",
			field:  gcName,
			maxLen: approvalRuleNameMaxLen,
			inputs: map[string]property.Value{
				"enabled":           property.New(true),
				"targetActionTypes": property.New(property.NewArray([]property.Value{str("update")})),
				"environmentIdentifier": property.New(property.NewMap(map[string]property.Value{
					gcOrganization: str(gcMyOrg),
					gcProject:      str("my-project"),
					gcName:         str("my-env"),
				})),
				"approvalRuleConfig": property.New(property.NewMap(map[string]property.Value{
					"numApprovalsRequired":      property.New(1.0),
					"allowSelfApproval":         property.New(false),
					"requireReapprovalOnChange": property.New(false),
					"eligibleApprovers": property.New(property.NewArray([]property.Value{
						property.New(property.NewMap(map[string]property.Value{"teamName": str("my-team")})),
					})),
				})),
			},
		},
		{
			token:  "pulumiservice:index:OidcIssuer",
			field:  gcName,
			maxLen: oidcIssuerNameMaxLen,
			inputs: map[string]property.Value{
				gcOrganization: str(gcMyOrg),
				"url":          str("https://token.actions.githubusercontent.com"),
			},
		},
		{
			token:  "pulumiservice:index:OrganizationRole",
			field:  gcName,
			maxLen: organizationRoleNameMaxLen,
			inputs: map[string]property.Value{
				gcOrganizationName: str(gcMyOrg),
				gcPermissions: property.New(property.NewMap(map[string]property.Value{
					gcTypeMeta:    str("PermissionDescriptorAllow"),
					"permissions": property.New(property.NewArray([]property.Value{str("stack:read")})),
				})),
			},
		},
	} {
		t.Run(string(tc.token), func(t *testing.T) {
			t.Parallel()
			inputs := property.NewMap(tc.inputs)

			check := func(t *testing.T, req p.CheckRequest) string {
				t.Helper()
				if req.Urn == "" {
					req.Urn = urn.New("dev", "demo", "", tc.token, "ciThing")
				}
				if req.RandomSeed == nil {
					req.RandomSeed = seed
				}
				resp, err := server.Check(req)
				require.NoError(t, err)
				require.Empty(t, resp.Failures)
				name, ok := resp.Inputs.GetOk(tc.field)
				require.True(t, ok, "check returned no %s", tc.field)
				return name.AsString()
			}

			t.Run("omitted name is generated from the logical name", func(t *testing.T) {
				name := check(t, p.CheckRequest{Inputs: inputs})
				assert.Regexp(t, `^ciThing-[0-9a-f]{8}$`, name)
				assert.Equal(t, name, check(t, p.CheckRequest{Inputs: inputs}), "same seed, same name")
				assert.NotEqual(t, name, check(t, p.CheckRequest{Inputs: inputs, RandomSeed: []byte("other")}))
			})

			t.Run("long logical name is shortened to the API limit", func(t *testing.T) {
				name := check(t, p.CheckRequest{
					Urn:    urn.New("dev", "demo", "", tc.token, strings.Repeat("a", 300)),
					Inputs: inputs,
				})
				assert.Len(t, name, tc.maxLen)
				assert.Regexp(t, `^a+-[0-9a-f]{8}$`, name)
			})

			t.Run("old name is kept on update", func(t *testing.T) {
				old := inputs.Set(tc.field, str("ciThing-1a2b3c4d"))
				assert.Equal(t, "ciThing-1a2b3c4d", check(t, p.CheckRequest{State: old, Inputs: inputs}))
			})

			t.Run("explicit name is kept", func(t *testing.T) {
				old := inputs.Set(tc.field, str("ciThing-1a2b3c4d"))
				news := inputs.Set(tc.field, str("explicit"))
				assert.Equal(t, "explicit", check(t, p.CheckRequest{State: old, Inputs: news}))
			})

			t.Run("proposed name from autonaming config is used", func(t *testing.T) {
				name := check(t, p.CheckRequest{
					Inputs:     inputs,
					Autonaming: &p.AutonamingOptions{Mode: p.AutonamingPropose, ProposedName: "dev-ciThing"},
				})
				assert.Equal(t, "dev-ciThing", name)
			})

			t.Run("name is an optional input and a required output", func(t *testing.T) {
				resp, err := server.GetSchema(p.GetSchemaRequest{})
				require.NoError(t, err)
				var spec struct {
					Resources map[string]struct {
						Required       []string `json:"required"`
						RequiredInputs []string `json:"requiredInputs"`
					} `json:"resources"`
				}
				require.NoError(t, json.Unmarshal([]byte(resp.Schema), &spec))
				res, ok := spec.Resources[string(tc.token)]
				require.True(t, ok, "resource missing from schema")
				assert.NotContains(t, res.RequiredInputs, tc.field)
				assert.Contains(t, res.Required, tc.field)
			})
		})
	}
}

func TestSuffixedAutoName(t *testing.T) {
	t.Parallel()

	generate := func(t *testing.T, cfg infer.AutoNameConfig, logical string) string {
		t.Helper()
		name, err := cfg.Generate(t.Context(), infer.AutoNameGenerateArgs{
			URN:  urn.New("dev", "demo", "", "pulumiservice:index:AgentPool", logical),
			Rand: rand.New(rand.NewPCG(1, 2)), //nolint:gosec // A fixed seed keeps generated names stable.
		})
		require.NoError(t, err)
		return name
	}

	t.Run("multibyte logical name is cut on a rune boundary within the byte limit", func(t *testing.T) {
		name := generate(t, suffixedAutoName(100, nil), strings.Repeat("é", 100))
		assert.LessOrEqual(t, len(name), 100)
		assert.Regexp(t, `^é{45}-[0-9a-f]{8}$`, name)
	})

	t.Run("agent pool names replace disallowed characters with hyphens", func(t *testing.T) {
		name := generate(t, suffixedAutoName(agentPoolNameMaxLen, agentPoolNameRune), "my/pool:é v1.2_x")
		assert.Regexp(t, `^my-pool-- v1\.2_x-[0-9a-f]{8}$`, name)
	})
}
