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

func TestTokenAutoName(t *testing.T) {
	t.Parallel()

	prov, err := infer.NewProviderBuilder().
		WithResources(infer.Resource(&OrgAccessToken{}), infer.Resource(&TeamAccessToken{})).
		WithModuleMap(map[tokens.ModuleName]tokens.ModuleName{"resources": "index"}).
		Build()
	require.NoError(t, err)
	server, err := integration.NewServer(t.Context(), "pulumiservice", semver.MustParse("0.0.0"),
		integration.WithProvider(prov))
	require.NoError(t, err)

	seed := []byte("token-seed")

	for _, tc := range []struct {
		token  tokens.Type
		inputs property.Map
	}{
		{
			token:  "pulumiservice:index:OrgAccessToken",
			inputs: property.NewMap(map[string]property.Value{gcOrganizationName: property.New(gcMyOrg)}),
		},
		{
			token: "pulumiservice:index:TeamAccessToken",
			inputs: property.NewMap(map[string]property.Value{
				gcOrganizationName: property.New(gcMyOrg),
				gcTeamName:         property.New("my-team"),
			}),
		},
	} {
		t.Run(string(tc.token), func(t *testing.T) {
			check := func(t *testing.T, req p.CheckRequest) string {
				t.Helper()
				if req.Urn == "" {
					req.Urn = urn.New("dev", "demo", "", tc.token, "ciToken")
				}
				if req.RandomSeed == nil {
					req.RandomSeed = seed
				}
				resp, err := server.Check(req)
				require.NoError(t, err)
				require.Empty(t, resp.Failures)
				name, ok := resp.Inputs.GetOk(gcName)
				require.True(t, ok, "check returned no name")
				return name.AsString()
			}

			t.Run("omitted name is generated from the logical name", func(t *testing.T) {
				name := check(t, p.CheckRequest{Inputs: tc.inputs})
				assert.Regexp(t, `^ciToken-[0-9a-f]{8}$`, name)
				assert.Equal(t, name, check(t, p.CheckRequest{Inputs: tc.inputs}), "same seed, same name")
				assert.NotEqual(t, name, check(t, p.CheckRequest{Inputs: tc.inputs, RandomSeed: []byte("other")}))
			})

			t.Run("long logical name is shortened to the API limit", func(t *testing.T) {
				logical := strings.Repeat("é", 50)
				name := check(t, p.CheckRequest{
					Urn:    urn.New("dev", "demo", "", tc.token, logical),
					Inputs: tc.inputs,
				})
				assert.Len(t, []rune(name), tokenNameMaxLen)
				assert.Regexp(t, `^é+-[0-9a-f]{8}$`, name)
			})

			t.Run("old name is kept on update", func(t *testing.T) {
				old := tc.inputs.Set(gcName, property.New("ciToken-1a2b3c4d"))
				assert.Equal(t, "ciToken-1a2b3c4d", check(t, p.CheckRequest{State: old, Inputs: tc.inputs}))
			})

			t.Run("explicit name is kept", func(t *testing.T) {
				old := tc.inputs.Set(gcName, property.New("ciToken-1a2b3c4d"))
				news := tc.inputs.Set(gcName, property.New("explicit"))
				assert.Equal(t, "explicit", check(t, p.CheckRequest{State: old, Inputs: news}))
			})

			t.Run("proposed name from autonaming config is used", func(t *testing.T) {
				name := check(t, p.CheckRequest{
					Inputs:     tc.inputs,
					Autonaming: &p.AutonamingOptions{Mode: p.AutonamingPropose, ProposedName: "dev-ciToken"},
				})
				assert.Equal(t, "dev-ciToken", name)
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
				assert.NotContains(t, res.RequiredInputs, gcName)
				assert.Contains(t, res.Required, gcName)
			})
		})
	}
}
