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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const (
	testSPTokID     = "a81d4c3e-52f0-4b7a-b9e2-6d0f1c8e2a47"
	testSPTokenName = "ci-token"
	testSPTokDesc   = "for CI"
)

var testSPTokenResourceID = spTokenID(gcMyOrg, testServicePrincipalLogin, testSPTokID)

type spTokenClientMock struct {
	config.Client
	create func(orgName, login string, req pulumiapi.CreateServicePrincipalTokenRequest) (*pulumiapi.AccessToken, error)
	get    func(orgName, login, tokenID string) (*pulumiapi.AccessToken, error)
	del    func(orgName, login, tokenID string) error
}

func (c *spTokenClientMock) CreateServicePrincipalToken(
	_ context.Context, orgName, login string, req pulumiapi.CreateServicePrincipalTokenRequest,
) (*pulumiapi.AccessToken, error) {
	return c.create(orgName, login, req)
}

func (c *spTokenClientMock) GetServicePrincipalToken(
	_ context.Context, orgName, login, tokenID string,
) (*pulumiapi.AccessToken, error) {
	return c.get(orgName, login, tokenID)
}

func (c *spTokenClientMock) DeleteServicePrincipalToken(_ context.Context, orgName, login, tokenID string) error {
	return c.del(orgName, login, tokenID)
}

func TestServicePrincipalAccessTokenCheck(t *testing.T) {
	check := func(expires property.Value) []string {
		resp, err := (&ServicePrincipalAccessToken{}).Check(t.Context(), infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				gcName:                  property.New(testSPTokenName),
				gcOrganizationName:      property.New(gcMyOrg),
				"servicePrincipalLogin": property.New(testServicePrincipalLogin),
				gcExpires:               expires,
			}),
		})
		require.NoError(t, err)
		var reasons []string
		for _, f := range resp.Failures {
			reasons = append(reasons, f.Reason)
		}
		return reasons
	}

	assert.Empty(t, check(property.New("2027-06-30T17:00:00Z")))
	assert.Empty(t, check(property.New(property.Computed)), "an unknown expiry is not checked")
	reasons := check(property.New("30d"))
	require.Len(t, reasons, 1)
	assert.Contains(t, reasons[0], "is not a valid expiry")
}

func TestServicePrincipalAccessTokenCreate(t *testing.T) {
	t.Run("sends the request and keeps the secret value", func(t *testing.T) {
		var got pulumiapi.CreateServicePrincipalTokenRequest
		mock := &spTokenClientMock{
			create: func(orgName, login string, req pulumiapi.CreateServicePrincipalTokenRequest) (
				*pulumiapi.AccessToken, error,
			) {
				assert.Equal(t, gcMyOrg, orgName)
				assert.Equal(t, testServicePrincipalLogin, login)
				got = req
				return &pulumiapi.AccessToken{ID: testSPTokID, TokenValue: gcTokSecretValue}, nil
			},
		}
		ctx := config.WithMockClient(t.Context(), mock)
		desc := testSPTokDesc
		expires := testExpiryWithOffset
		in := ServicePrincipalAccessTokenInput{
			Name:                  testSPTokenName,
			OrganizationName:      gcMyOrg,
			ServicePrincipalLogin: testServicePrincipalLogin,
			Description:           &desc,
			Expires:               &expires,
		}
		resp, err := (&ServicePrincipalAccessToken{}).Create(ctx, infer.CreateRequest[spTokenInput]{Inputs: in})
		require.NoError(t, err)

		assert.Equal(t, testSPTokenName, got.Name)
		assert.Equal(t, desc, got.Description)
		require.NotNil(t, got.Expires)
		assert.Equal(t, time.Unix(testExpiryEpoch, 0).UTC(), *got.Expires)
		assert.Equal(t, time.UTC, got.Expires.Location(), "the expiry is sent in UTC")

		assert.Equal(t, testSPTokenResourceID, resp.ID)
		assert.Equal(t, spTokenState{ServicePrincipalAccessTokenInput: in, Value: gcTokSecretValue}, resp.Output)
	})

	t.Run("no expiry", func(t *testing.T) {
		mock := &spTokenClientMock{
			create: func(_, _ string, req pulumiapi.CreateServicePrincipalTokenRequest) (*pulumiapi.AccessToken, error) {
				assert.Nil(t, req.Expires)
				assert.Empty(t, req.Description)
				return &pulumiapi.AccessToken{ID: testSPTokID, TokenValue: gcTokSecretValue}, nil
			},
		}
		ctx := config.WithMockClient(t.Context(), mock)
		_, err := (&ServicePrincipalAccessToken{}).Create(ctx, infer.CreateRequest[spTokenInput]{
			Inputs: ServicePrincipalAccessTokenInput{
				Name:                  testSPTokenName,
				OrganizationName:      gcMyOrg,
				ServicePrincipalLogin: testServicePrincipalLogin,
			},
		})
		require.NoError(t, err)
	})

	t.Run("preview makes no call", func(t *testing.T) {
		ctx := config.WithMockClient(t.Context(), &spTokenClientMock{})
		resp, err := (&ServicePrincipalAccessToken{}).Create(ctx, infer.CreateRequest[spTokenInput]{
			Inputs: ServicePrincipalAccessTokenInput{Name: testSPTokenName, OrganizationName: gcMyOrg},
			DryRun: true,
		})
		require.NoError(t, err)
		assert.Empty(t, resp.Output.Value)
	})
}

func TestServicePrincipalAccessTokenRead(t *testing.T) {
	read := func(
		t *testing.T, token *pulumiapi.AccessToken, state spTokenState,
	) infer.ReadResponse[spTokenInput, spTokenState] {
		t.Helper()
		mock := &spTokenClientMock{
			get: func(orgName, login, tokenID string) (*pulumiapi.AccessToken, error) {
				assert.Equal(t, gcMyOrg, orgName)
				assert.Equal(t, testServicePrincipalLogin, login)
				assert.Equal(t, testSPTokID, tokenID)
				return token, nil
			},
		}
		ctx := config.WithMockClient(t.Context(), mock)
		resp, err := (&ServicePrincipalAccessToken{}).Read(ctx, infer.ReadRequest[spTokenInput, spTokenState]{
			ID:    testSPTokenResourceID,
			State: state,
		})
		require.NoError(t, err)
		return resp
	}

	t.Run("import", func(t *testing.T) {
		resp := read(t, &pulumiapi.AccessToken{
			ID:          testSPTokID,
			Name:        testSPTokenName,
			Description: testSPTokDesc,
			Expires:     testExpiryEpoch,
		}, spTokenState{})
		desc := testSPTokDesc
		expires := "2027-06-30T17:00:00Z"
		want := ServicePrincipalAccessTokenInput{
			Name:                  testSPTokenName,
			OrganizationName:      gcMyOrg,
			ServicePrincipalLogin: testServicePrincipalLogin,
			Description:           &desc,
			Expires:               &expires,
		}
		assert.Equal(t, testSPTokenResourceID, resp.ID)
		assert.Equal(t, want, resp.Inputs)
		assert.Empty(t, resp.State.Value, "the value can't be read back")
	})

	t.Run("refresh keeps the value and the expiry's spelling", func(t *testing.T) {
		expires := testExpiryWithOffset
		resp := read(t, &pulumiapi.AccessToken{
			ID:      testSPTokID,
			Name:    testSPTokenName,
			Expires: testExpiryEpoch,
		}, spTokenState{
			ServicePrincipalAccessTokenInput: ServicePrincipalAccessTokenInput{Expires: &expires},
			Value:                            gcTokSecretValue,
		})
		assert.Equal(t, gcTokSecretValue, resp.State.Value)
		require.NotNil(t, resp.Inputs.Expires)
		assert.Equal(t, testExpiryWithOffset, *resp.Inputs.Expires)
		assert.Nil(t, resp.Inputs.Description)
	})

	t.Run("deleted", func(t *testing.T) {
		resp := read(t, nil, spTokenState{Value: gcTokSecretValue})
		assert.Empty(t, resp.ID)
	})
}

func TestServicePrincipalAccessTokenDelete(t *testing.T) {
	var deleted string
	mock := &spTokenClientMock{
		del: func(orgName, login, tokenID string) error {
			assert.Equal(t, gcMyOrg, orgName)
			assert.Equal(t, testServicePrincipalLogin, login)
			deleted = tokenID
			return nil
		},
	}
	ctx := config.WithMockClient(t.Context(), mock)
	_, err := (&ServicePrincipalAccessToken{}).Delete(ctx, infer.DeleteRequest[spTokenState]{ID: testSPTokenResourceID})
	require.NoError(t, err)
	assert.Equal(t, testSPTokID, deleted)
}

func TestSplitSPTokenID(t *testing.T) {
	org, login, tokenID, err := splitSPTokenID(testSPTokenResourceID)
	require.NoError(t, err)
	assert.Equal(t, []string{gcMyOrg, testServicePrincipalLogin, testSPTokID}, []string{org, login, tokenID})

	for _, id := range []string{"", "my-org/login", "my-org//tok", "my-org/login/tok/extra"} {
		_, _, _, err := splitSPTokenID(id)
		assert.Error(t, err, id)
	}
}
