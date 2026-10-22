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

const (
	testAccessTokenID   = "pat-id"
	testAccessTokenDesc = "example token"
)

type accessTokenClientMock struct {
	config.Client
	create func(ctx context.Context, description string, expires int64) (*pulumiapi.AccessToken, error)
	get    func(ctx context.Context, id string) (*pulumiapi.AccessToken, error)
}

func (c *accessTokenClientMock) CreateAccessToken(
	ctx context.Context, description string, expires int64,
) (*pulumiapi.AccessToken, error) {
	return c.create(ctx, description, expires)
}

func (c *accessTokenClientMock) GetAccessToken(ctx context.Context, id string) (*pulumiapi.AccessToken, error) {
	return c.get(ctx, id)
}

func TestAccessTokenCheck(t *testing.T) {
	check := func(extra map[string]property.Value) []string {
		m := map[string]property.Value{gcDescription: property.New(testAccessTokenDesc)}
		for k, v := range extra {
			m[k] = v
		}
		resp, err := (&AccessToken{}).Check(t.Context(), infer.CheckRequest{
			NewInputs: property.NewMap(m),
		})
		require.NoError(t, err)
		var reasons []string
		for _, f := range resp.Failures {
			reasons = append(reasons, f.Reason)
		}
		return reasons
	}

	t.Run("no expiry", func(t *testing.T) {
		assert.Empty(t, check(nil))
	})

	t.Run("valid date-time", func(t *testing.T) {
		assert.Empty(t, check(map[string]property.Value{gcExpires: property.New(testExpiryWithOffset)}))
	})

	t.Run("unparseable", func(t *testing.T) {
		reasons := check(map[string]property.Value{gcExpires: property.New("30d")})
		require.Len(t, reasons, 1)
		assert.Contains(t, reasons[0], "is not a valid expiry")
	})

	t.Run("unknown expiry is skipped", func(t *testing.T) {
		assert.Empty(t, check(map[string]property.Value{gcExpires: property.New(property.Computed)}))
	})
}

func TestAccessTokenCreate(t *testing.T) {
	var gotExpires int64
	var gotDescription string
	mock := &accessTokenClientMock{
		create: func(_ context.Context, description string, expires int64) (*pulumiapi.AccessToken, error) {
			gotDescription, gotExpires = description, expires
			return &pulumiapi.AccessToken{ID: testAccessTokenID, TokenValue: gcTokSecretValue}, nil
		},
	}
	ctx := config.WithMockClient(t.Context(), mock)

	t.Run("with expiry", func(t *testing.T) {
		expires := testExpiryWithOffset
		resp, err := (&AccessToken{}).Create(ctx, infer.CreateRequest[AccessTokenInput]{
			Inputs: AccessTokenInput{Description: testAccessTokenDesc, Expires: &expires},
		})
		require.NoError(t, err)
		assert.Equal(t, testAccessTokenDesc, gotDescription)
		assert.Equal(t, testExpiryEpoch, gotExpires)
		assert.Equal(t, testAccessTokenID, resp.ID)
		assert.Equal(t, gcTokSecretValue, resp.Output.Value)
		require.NotNil(t, resp.Output.Expires)
		assert.Equal(t, expires, *resp.Output.Expires)
	})

	t.Run("without expiry", func(t *testing.T) {
		resp, err := (&AccessToken{}).Create(ctx, infer.CreateRequest[AccessTokenInput]{
			Inputs: AccessTokenInput{Description: testAccessTokenDesc},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(0), gotExpires)
		assert.Nil(t, resp.Output.Expires)
	})

	t.Run("invalid expiry", func(t *testing.T) {
		expires := "30d"
		_, err := (&AccessToken{}).Create(ctx, infer.CreateRequest[AccessTokenInput]{
			Inputs: AccessTokenInput{Description: testAccessTokenDesc, Expires: &expires},
		})
		assert.ErrorContains(t, err, "is not a valid expiry")
	})
}

func TestAccessTokenRead(t *testing.T) {
	mock := &accessTokenClientMock{
		get: func(_ context.Context, id string) (*pulumiapi.AccessToken, error) {
			return &pulumiapi.AccessToken{ID: id, Description: testAccessTokenDesc, Expires: testExpiryEpoch}, nil
		},
	}
	ctx := config.WithMockClient(t.Context(), mock)

	t.Run("keeps expiry spelling", func(t *testing.T) {
		expires := testExpiryWithOffset
		resp, err := (&AccessToken{}).Read(ctx, infer.ReadRequest[AccessTokenInput, AccessTokenState]{
			ID: testAccessTokenID,
			State: AccessTokenState{
				AccessTokenInput: AccessTokenInput{Description: testAccessTokenDesc, Expires: &expires},
				Value:            gcTokSecretValue,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp.Inputs.Expires)
		assert.Equal(t, expires, *resp.Inputs.Expires)
		assert.Equal(t, testAccessTokenDesc, resp.Inputs.Description)
		assert.Equal(t, gcTokSecretValue, resp.State.Value)
	})

	t.Run("import populates expiry", func(t *testing.T) {
		resp, err := (&AccessToken{}).Read(ctx, infer.ReadRequest[AccessTokenInput, AccessTokenState]{
			ID: testAccessTokenID,
		})
		require.NoError(t, err)
		require.NotNil(t, resp.State.Expires)
		assert.Equal(t, "2027-06-30T17:00:00Z", *resp.State.Expires)
	})
}

// State written by the pre-infer AccessToken implementation embedded a
// duplicate `__inputs` map alongside the real outputs. After migrating to
// infer, decoding such state fails with "Unrecognized field '__inputs'" unless
// migrateAccessTokenLegacyInputs runs first.
func TestAccessTokenLegacyInputsMigration(t *testing.T) {
	t.Run("migrates legacy state", func(t *testing.T) {
		legacy := property.NewMap(map[string]property.Value{
			gcInputs: property.New(property.NewMap(map[string]property.Value{
				gcDescription: property.New(testAccessTokenDesc),
			})),
			gcDescription: property.New(testAccessTokenDesc),
			gcValue:       property.New(gcTokSecretValue),
		})

		got, err := migrateAccessTokenLegacyInputs(t.Context(), legacy)
		require.NoError(t, err)
		assert.Equal(t, &AccessTokenState{
			AccessTokenInput: AccessTokenInput{Description: testAccessTokenDesc},
			Value:            gcTokSecretValue,
		}, got.Result)
	})

	t.Run("no-op for already-migrated state", func(t *testing.T) {
		current := property.NewMap(map[string]property.Value{
			gcDescription: property.New(testAccessTokenDesc),
			gcValue:       property.New(gcTokSecretValue),
		})

		got, err := migrateAccessTokenLegacyInputs(t.Context(), current)
		require.NoError(t, err)
		assert.Nil(t, got.Result, "migration must not fire when __inputs is absent")
	})

	t.Run("registered on the resource", func(t *testing.T) {
		migrations := (&AccessToken{}).StateMigrations(t.Context())
		assert.Len(t, migrations, 1)
	})
}
