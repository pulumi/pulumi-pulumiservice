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
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	presource "github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

const (
	testTokenLogicalName = "ci-token"
	testTokenOldName     = "ci-token-1a2b3c4"
	testTeamTokenName    = "test-token"
)

var testTokenSeed = []byte("token-seed")

// autoNameTestContext returns the context a resource's Check sees when the
// engine sends urn and testTokenSeed.
func autoNameTestContext(t *testing.T, urn string) context.Context {
	t.Helper()
	var got context.Context
	prov := util.WithAutoNameContext(p.Provider{
		Check: func(ctx context.Context, _ p.CheckRequest) (p.CheckResponse, error) {
			got = ctx
			return p.CheckResponse{}, nil
		},
	})
	_, err := prov.Check(t.Context(), p.CheckRequest{Urn: presource.URN(urn), RandomSeed: testTokenSeed})
	require.NoError(t, err)
	return got
}

func TestTeamAccessTokenCheckAutoName(t *testing.T) {
	urn := "urn:pulumi:dev::demo::pulumiservice:index:TeamAccessToken::" + testTokenLogicalName
	ctx := autoNameTestContext(t, urn)
	base := map[string]property.Value{
		gcOrganizationName: property.New(gcMyOrg),
		gcTeamName:         property.New("my-team"),
	}
	check := func(news, olds property.Map) TeamAccessTokenInput {
		resp, err := (&TeamAccessToken{}).Check(ctx, infer.CheckRequest{
			Name: testTokenLogicalName, NewInputs: news, OldInputs: olds,
		})
		require.NoError(t, err)
		require.Empty(t, resp.Failures)
		return resp.Inputs
	}

	t.Run("generates a seeded name when omitted", func(t *testing.T) {
		in := check(property.NewMap(base), property.Map{})
		assert.Equal(t, util.GenerateAutoName(urn, testTokenSeed, tokenNameMaxLen), in.Name)
		assert.Equal(t, in.Name, check(property.NewMap(base), property.Map{}).Name)
		assert.LessOrEqual(t, len(in.Name), tokenNameMaxLen)
		assert.Equal(t, "my-team", in.TeamName)
	})

	t.Run("reuses the old name on update", func(t *testing.T) {
		olds := property.NewMap(base).Set(gcName, property.New(testTokenOldName))
		assert.Equal(t, testTokenOldName, check(property.NewMap(base), olds).Name)
	})

	t.Run("keeps an explicit name", func(t *testing.T) {
		news := property.NewMap(base).Set(gcName, property.New("explicit"))
		olds := property.NewMap(base).Set(gcName, property.New(testTokenOldName))
		assert.Equal(t, "explicit", check(news, olds).Name)
	})

	t.Run("preview output carries the generated name", func(t *testing.T) {
		in := check(property.NewMap(base), property.Map{})
		resp, err := (&TeamAccessToken{}).Create(ctx, infer.CreateRequest[TeamAccessTokenInput]{
			Inputs: in, DryRun: true,
		})
		require.NoError(t, err)
		assert.Equal(t, in.Name, resp.Output.Name)
	})

	t.Run("reports invalid inputs", func(t *testing.T) {
		resp, err := (&TeamAccessToken{}).Check(ctx, infer.CheckRequest{
			Name:      testTokenLogicalName,
			NewInputs: property.NewMap(base).Set(gcTeamName, property.New(1.0)),
		})
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Failures)
	})
}

// State written by the pre-infer TeamAccessToken implementation embedded a
// duplicate `__inputs` map alongside the real outputs. After migrating to
// infer, decoding such state fails with "Unrecognized field '__inputs'" unless
// migrateTeamAccessTokenLegacyInputs runs first.
func TestTeamAccessTokenLegacyInputsMigration(t *testing.T) {
	t.Run("migrates legacy state", func(t *testing.T) {
		legacyInputs := property.NewMap(map[string]property.Value{
			gcName:             property.New(testTeamTokenName),
			gcOrganizationName: property.New(gcMyOrg),
			gcTeamName:         property.New("my-team"),
			gcDescription:      property.New("example team token"),
		})
		legacy := property.NewMap(map[string]property.Value{
			gcInputs:           property.New(legacyInputs),
			gcName:             property.New(testTeamTokenName),
			gcOrganizationName: property.New(gcMyOrg),
			gcTeamName:         property.New("my-team"),
			gcDescription:      property.New("example team token"),
			gcValue:            property.New(gcTokSecretValue),
		})

		got, err := migrateTeamAccessTokenLegacyInputs(t.Context(), legacy)
		require.NoError(t, err)
		desc := "example team token"
		assert.Equal(t, &TeamAccessTokenState{
			TeamAccessTokenInput: TeamAccessTokenInput{
				Name:             testTeamTokenName,
				OrganizationName: gcMyOrg,
				TeamName:         "my-team",
				Description:      &desc,
			},
			Name:  testTeamTokenName,
			Value: gcTokSecretValue,
		}, got.Result)
	})

	t.Run("migrates legacy state without description", func(t *testing.T) {
		legacy := property.NewMap(map[string]property.Value{
			gcInputs:           property.New(property.NewMap(nil)),
			gcName:             property.New(testTeamTokenName),
			gcOrganizationName: property.New(gcMyOrg),
			gcTeamName:         property.New("my-team"),
			gcValue:            property.New(gcTokSecretValue),
		})

		got, err := migrateTeamAccessTokenLegacyInputs(t.Context(), legacy)
		require.NoError(t, err)
		assert.Equal(t, &TeamAccessTokenState{
			TeamAccessTokenInput: TeamAccessTokenInput{
				Name:             testTeamTokenName,
				OrganizationName: gcMyOrg,
				TeamName:         "my-team",
			},
			Name:  testTeamTokenName,
			Value: gcTokSecretValue,
		}, got.Result)
	})

	t.Run("no-op for already-migrated state", func(t *testing.T) {
		current := property.NewMap(map[string]property.Value{
			gcName:             property.New(testTeamTokenName),
			gcOrganizationName: property.New(gcMyOrg),
			gcTeamName:         property.New("my-team"),
			gcValue:            property.New(gcTokSecretValue),
		})

		got, err := migrateTeamAccessTokenLegacyInputs(t.Context(), current)
		require.NoError(t, err)
		assert.Nil(t, got.Result, "migration must not fire when __inputs is absent")
	})

	t.Run("registered on the resource", func(t *testing.T) {
		migrations := (&TeamAccessToken{}).StateMigrations(t.Context())
		assert.Len(t, migrations, 1)
	})
}
