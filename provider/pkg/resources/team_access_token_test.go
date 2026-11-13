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

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

// State written by the pre-infer TeamAccessToken implementation embedded a
// duplicate `__inputs` map alongside the real outputs. After migrating to
// infer, decoding such state fails with "Unrecognized field '__inputs'" unless
// migrateTeamAccessTokenLegacyInputs runs first.
func TestTeamAccessTokenLegacyInputsMigration(t *testing.T) {
	t.Run("migrates legacy state", func(t *testing.T) {
		legacyInputs := property.NewMap(map[string]property.Value{
			gcName:             property.New("test-token"),
			gcOrganizationName: property.New(gcMyOrg),
			gcTeamName:         property.New("my-team"),
			gcDescription:      property.New("example team token"),
		})
		legacy := property.NewMap(map[string]property.Value{
			gcInputs:           property.New(legacyInputs),
			gcName:             property.New("test-token"),
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
				Name:             "test-token",
				OrganizationName: gcMyOrg,
				TeamName:         "my-team",
				Description:      &desc,
			},
			Value: gcTokSecretValue,
		}, got.Result)
	})

	t.Run("migrates legacy state without description", func(t *testing.T) {
		legacy := property.NewMap(map[string]property.Value{
			gcInputs:           property.New(property.NewMap(nil)),
			gcName:             property.New("test-token"),
			gcOrganizationName: property.New(gcMyOrg),
			gcTeamName:         property.New("my-team"),
			gcValue:            property.New(gcTokSecretValue),
		})

		got, err := migrateTeamAccessTokenLegacyInputs(t.Context(), legacy)
		require.NoError(t, err)
		assert.Equal(t, &TeamAccessTokenState{
			TeamAccessTokenInput: TeamAccessTokenInput{
				Name:             "test-token",
				OrganizationName: gcMyOrg,
				TeamName:         "my-team",
			},
			Value: gcTokSecretValue,
		}, got.Result)
	})

	t.Run("no-op for already-migrated state", func(t *testing.T) {
		current := property.NewMap(map[string]property.Value{
			gcName:             property.New("test-token"),
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

const (
	testTeamTokenID   = "tok-id"
	testTeamTokenName = "ci-token"
	testTeamTokenTeam = "my-team"
)

type teamAccessTokenClientMock struct {
	config.Client
	create func(ctx context.Context, name, orgName, teamName, description string, expires int64) (
		*pulumiapi.AccessToken, error)
	get func(ctx context.Context, tokenID, orgName, teamName string) (*pulumiapi.AccessToken, error)
}

func (c *teamAccessTokenClientMock) CreateTeamAccessToken(
	ctx context.Context, name, orgName, teamName, description string, expires int64,
) (*pulumiapi.AccessToken, error) {
	return c.create(ctx, name, orgName, teamName, description, expires)
}

func (c *teamAccessTokenClientMock) GetTeamAccessToken(
	ctx context.Context, tokenID, orgName, teamName string,
) (*pulumiapi.AccessToken, error) {
	return c.get(ctx, tokenID, orgName, teamName)
}

func TestTeamAccessTokenCheck(t *testing.T) {
	check := func(expires property.Value) []string {
		resp, err := (&TeamAccessToken{}).Check(t.Context(), infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				gcName:             property.New(testTeamTokenName),
				gcOrganizationName: property.New(gcMyOrg),
				gcTeamName:         property.New(testTeamTokenTeam),
				gcExpires:          expires,
			}),
		})
		require.NoError(t, err)
		var reasons []string
		for _, f := range resp.Failures {
			reasons = append(reasons, f.Reason)
		}
		return reasons
	}

	t.Run("valid date-time", func(t *testing.T) {
		assert.Empty(t, check(property.New("2027-06-30T17:00:00Z")))
	})

	t.Run("unparseable", func(t *testing.T) {
		reasons := check(property.New("30d"))
		require.Len(t, reasons, 1)
		assert.Contains(t, reasons[0], "is not a valid expiry")
	})

	t.Run("unknown expiry is skipped", func(t *testing.T) {
		assert.Empty(t, check(property.New(property.Computed)))
	})
}

func TestTeamAccessTokenCreateSendsExpiry(t *testing.T) {
	var gotExpires int64
	mock := &teamAccessTokenClientMock{
		create: func(_ context.Context, _, _, _, _ string, expires int64) (*pulumiapi.AccessToken, error) {
			gotExpires = expires
			return &pulumiapi.AccessToken{ID: testTeamTokenID, TokenValue: gcTokSecretValue}, nil
		},
	}
	ctx := config.WithMockClient(t.Context(), mock)

	expires := testExpiryWithOffset
	resp, err := (&TeamAccessToken{}).Create(ctx, infer.CreateRequest[TeamAccessTokenInput]{
		Inputs: TeamAccessTokenInput{
			Name:             testTeamTokenName,
			OrganizationName: gcMyOrg,
			TeamName:         testTeamTokenTeam,
			Expires:          &expires,
		},
	})
	require.NoError(t, err)
	assert.Equal(t, testExpiryEpoch, gotExpires)
	assert.Equal(t, expires, *resp.Output.Expires)
}

func TestTeamAccessTokenReadKeepsExpirySpelling(t *testing.T) {
	mock := &teamAccessTokenClientMock{
		get: func(_ context.Context, _, _, _ string) (*pulumiapi.AccessToken, error) {
			return &pulumiapi.AccessToken{ID: testTeamTokenID, Expires: testExpiryEpoch}, nil
		},
	}
	ctx := config.WithMockClient(t.Context(), mock)

	expires := testExpiryWithOffset
	resp, err := (&TeamAccessToken{}).Read(ctx, infer.ReadRequest[TeamAccessTokenInput, TeamAccessTokenState]{
		ID: teamAccessTokenID(gcMyOrg, testTeamTokenTeam, testTeamTokenName, testTeamTokenID),
		State: TeamAccessTokenState{
			TeamAccessTokenInput: TeamAccessTokenInput{Expires: &expires},
			Value:                gcTokSecretValue,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp.Inputs.Expires)
	assert.Equal(t, expires, *resp.Inputs.Expires)
	assert.Equal(t, gcTokSecretValue, resp.State.Value)
}
