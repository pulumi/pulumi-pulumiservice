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

package pulumiapi

import (
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	registryPolicyPacksPath = "/api/registry/policypacks"
	orgRegistryAlphaPath    = "/api/orgs/anOrg/registry/policypacks/alpha"
	sourcePulumi            = "pulumi"
	sourcePrivate           = "private"
	cisAwsPolicyPack        = "cis-aws"
	guardPolicyPack         = "guard"
	alphaDisplayName        = "Alpha"
	policyPackVersion110    = "1.1.0"
	policyPackVersion005    = "0.0.5"
)

func TestListRegistryPolicyPacks(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		want := []RegistryPolicyPack{
			{Source: sourcePulumi, Publisher: sourcePulumi, Name: cisAwsPolicyPack, Version: "1.0.2"},
			{Source: sourcePrivate, Publisher: testOrgName, Name: alphaPolicyPack, Version: policyPackVersion005},
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   registryPolicyPacksPath,
			ExpectedReqBody:   listRegistryPolicyPacksRequest{OrgLogin: testOrgName},
			ResponseCode:      http.StatusOK,
			ResponseBody:      listRegistryPolicyPacksResponse{PolicyPacks: want},
		})
		got, err := c.ListRegistryPolicyPacks(ctx, testOrgName)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("empty org rejected", func(t *testing.T) {
		c := &Client{}
		_, err := c.ListRegistryPolicyPacks(ctx, "")
		assert.EqualError(t, err, "empty orgName")
	})

	t.Run("server error surfaces", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPost,
			ExpectedReqPath:   registryPolicyPacksPath,
			ResponseCode:      http.StatusInternalServerError,
			ResponseBody:      ErrorResponse{StatusCode: 500, Message: "boom"},
		})
		_, err := c.ListRegistryPolicyPacks(ctx, testOrgName)
		assert.ErrorContains(t, err, "failed to list registry policy packs")
	})
}

func TestGetRegistryPolicyPack(t *testing.T) {
	t.Run("happy path unwraps the policyPack envelope", func(t *testing.T) {
		want := &RegistryPolicyPack{
			Source: sourcePrivate, Publisher: testOrgName, Name: alphaPolicyPack, Version: policyPackVersion110,
		}
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   orgRegistryAlphaPath,
			ResponseCode:      http.StatusOK,
			ResponseBody:      getRegistryPolicyPackResponse{PolicyPack: *want},
		})
		got, err := c.GetRegistryPolicyPack(ctx, testOrgName, alphaPolicyPack, "")
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("version tag is sent as a query param", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod:   http.MethodGet,
			ExpectedReqPath:     orgRegistryAlphaPath,
			ExpectedQueryParams: url.Values{tagQueryParam: {"latest"}},
			ResponseCode:        http.StatusOK,
			ResponseBody: getRegistryPolicyPackResponse{
				PolicyPack: RegistryPolicyPack{Name: alphaPolicyPack},
			},
		})
		_, err := c.GetRegistryPolicyPack(ctx, testOrgName, alphaPolicyPack, "latest")
		require.NoError(t, err)
	})

	t.Run("404 returns nil pack and nil error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/orgs/anOrg/registry/policypacks/missing",
			ResponseCode:      http.StatusNotFound,
			ResponseBody:      ErrorResponse{StatusCode: 404, Message: notFoundError},
		})
		got, err := c.GetRegistryPolicyPack(ctx, testOrgName, "missing", "")
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("input validation", func(t *testing.T) {
		c := &Client{}
		_, err := c.GetRegistryPolicyPack(ctx, "", alphaPolicyPack, "")
		assert.EqualError(t, err, "empty orgName")
		_, err = c.GetRegistryPolicyPack(ctx, testOrgName, "", "")
		assert.EqualError(t, err, "empty policyPackName")
	})
}

func TestIndexRegistryPolicyPacksByName(t *testing.T) {
	t.Run("unique names are indexed", func(t *testing.T) {
		packs := []RegistryPolicyPack{
			{Name: cisAwsPolicyPack, Source: sourcePulumi, Publisher: sourcePulumi},
			{Name: alphaPolicyPack, Source: sourcePrivate, Publisher: testOrgName},
		}
		got := indexRegistryPolicyPacksByName(packs)
		assert.Len(t, got, 2)
		assert.Equal(t, sourcePulumi, got[cisAwsPolicyPack].Publisher)
		assert.Equal(t, testOrgName, got[alphaPolicyPack].Publisher)
	})

	// Attributing a pack to the wrong publisher is the exact failure this feature
	// exists to eliminate, so an ambiguous name resolves to nothing rather than a guess.
	t.Run("duplicate names are dropped", func(t *testing.T) {
		packs := []RegistryPolicyPack{
			{Name: guardPolicyPack, Source: sourcePulumi, Publisher: sourcePulumi},
			{Name: guardPolicyPack, Source: sourcePrivate, Publisher: testOrgName},
			{Name: alphaPolicyPack, Source: sourcePrivate, Publisher: testOrgName},
		}
		got := indexRegistryPolicyPacksByName(packs)
		_, ok := got[guardPolicyPack]
		assert.False(t, ok, "colliding name should not be indexed")
		assert.Len(t, got, 1)
	})

	// Registry list order is unspecified; a first-wins tie-break would make the
	// output flap between invocations.
	t.Run("result is order independent", func(t *testing.T) {
		packs := []RegistryPolicyPack{
			{Name: guardPolicyPack, Source: sourcePulumi, Publisher: sourcePulumi},
			{Name: guardPolicyPack, Source: sourcePrivate, Publisher: testOrgName},
			{Name: alphaPolicyPack, Source: sourcePrivate, Publisher: testOrgName},
		}
		reversed := slices.Clone(packs)
		slices.Reverse(reversed)
		assert.Equal(t, indexRegistryPolicyPacksByName(packs), indexRegistryPolicyPacksByName(reversed))
	})
}

// registryJoinServer serves the org policy pack list and the registry list off a
// single test server, so the two-call join can be driven end to end.
func registryJoinServer(
	t *testing.T,
	orgPacks []PolicyPackWithVersions,
	registryCode int,
	registryPacks []RegistryPolicyPack,
	registryHit *bool,
) *Client {
	t.Helper()
	return startTestServerMulti(t, func(r *http.Request) (int, any) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == policyPacksPath:
			return http.StatusOK, listPolicyPacksResponse{PolicyPacks: orgPacks}
		case r.Method == http.MethodPost && r.URL.Path == registryPolicyPacksPath:
			if registryHit != nil {
				*registryHit = true
			}
			if registryCode != http.StatusOK {
				return registryCode, ErrorResponse{StatusCode: registryCode, Message: "nope"}
			}
			return http.StatusOK, listRegistryPolicyPacksResponse{PolicyPacks: registryPacks}
		}
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		return http.StatusInternalServerError, nil
	})
}

func TestListPolicyPacksWithRegistryMetadata(t *testing.T) {
	orgPacks := []PolicyPackWithVersions{
		{
			Name: cisAwsPolicyPack, DisplayName: "CIS AWS",
			Versions: []int{1}, VersionTags: []string{"1.0.2"},
		},
		{
			Name: alphaPolicyPack, DisplayName: alphaDisplayName,
			Versions: []int{1, 2}, VersionTags: []string{policyPackVersion005},
		},
	}

	t.Run("enriches packs with source and publisher", func(t *testing.T) {
		c := registryJoinServer(t, orgPacks, http.StatusOK, []RegistryPolicyPack{
			{Name: cisAwsPolicyPack, Source: sourcePulumi, Publisher: sourcePulumi},
			{Name: alphaPolicyPack, Source: sourcePrivate, Publisher: testOrgName},
		}, nil)

		got, degraded, err := c.ListPolicyPacksWithRegistryMetadata(ctx, testOrgName)
		require.NoError(t, err)
		assert.False(t, degraded)
		require.Len(t, got, 2)

		assert.Equal(t, sourcePulumi, got[0].Source)
		assert.Equal(t, sourcePulumi, got[0].Publisher)
		assert.Equal(t, sourcePrivate, got[1].Source)
		assert.Equal(t, testOrgName, got[1].Publisher)

		// The org list stays the source of truth for everything but provenance.
		assert.Equal(t, alphaDisplayName, got[1].DisplayName)
		assert.Equal(t, []int{1, 2}, got[1].Versions)
		assert.Equal(t, []string{policyPackVersion005}, got[1].VersionTags)
	})

	t.Run("pack missing from registry keeps fields unset", func(t *testing.T) {
		c := registryJoinServer(t, orgPacks, http.StatusOK, []RegistryPolicyPack{
			{Name: cisAwsPolicyPack, Source: sourcePulumi, Publisher: sourcePulumi},
		}, nil)

		got, degraded, err := c.ListPolicyPacksWithRegistryMetadata(ctx, testOrgName)
		require.NoError(t, err)
		assert.False(t, degraded)
		require.Len(t, got, 2)
		assert.Equal(t, sourcePulumi, got[0].Publisher)
		assert.Empty(t, got[1].Source)
		assert.Empty(t, got[1].Publisher)
	})

	t.Run("registry entries with no org pack are ignored", func(t *testing.T) {
		c := registryJoinServer(t, orgPacks, http.StatusOK, []RegistryPolicyPack{
			{Name: cisAwsPolicyPack, Source: sourcePulumi, Publisher: sourcePulumi},
			{Name: alphaPolicyPack, Source: sourcePrivate, Publisher: testOrgName},
			{Name: "not-in-org", Source: sourcePulumi, Publisher: sourcePulumi},
		}, nil)

		got, _, err := c.ListPolicyPacksWithRegistryMetadata(ctx, testOrgName)
		require.NoError(t, err)
		assert.Len(t, got, len(orgPacks))
	})

	t.Run("colliding registry names are not attributed", func(t *testing.T) {
		c := registryJoinServer(t, orgPacks, http.StatusOK, []RegistryPolicyPack{
			{Name: alphaPolicyPack, Source: sourcePulumi, Publisher: sourcePulumi},
			{Name: alphaPolicyPack, Source: sourcePrivate, Publisher: testOrgName},
		}, nil)

		got, _, err := c.ListPolicyPacksWithRegistryMetadata(ctx, testOrgName)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Empty(t, got[1].Source)
		assert.Empty(t, got[1].Publisher)
	})

	// A backend that doesn't serve the registry route at all (older or self-hosted
	// Pulumi Cloud) must keep working exactly as it did before this feature.
	t.Run("registry 404 degrades instead of failing", func(t *testing.T) {
		c := registryJoinServer(t, orgPacks, http.StatusNotFound, nil, nil)

		got, degraded, err := c.ListPolicyPacksWithRegistryMetadata(ctx, testOrgName)
		require.NoError(t, err)
		assert.True(t, degraded)
		require.Len(t, got, 2)
		assert.Empty(t, got[0].Publisher)
		assert.Empty(t, got[1].Publisher)
	})

	t.Run("registry 405 degrades instead of failing", func(t *testing.T) {
		c := registryJoinServer(t, orgPacks, http.StatusMethodNotAllowed, nil, nil)

		got, degraded, err := c.ListPolicyPacksWithRegistryMetadata(ctx, testOrgName)
		require.NoError(t, err)
		assert.True(t, degraded)
		assert.Len(t, got, 2)
	})

	// The guard against the silent compliance hole: if the lookup is broken rather
	// than absent, callers filtering on publisher must not quietly get zero matches.
	t.Run("registry 403 fails the call", func(t *testing.T) {
		c := registryJoinServer(t, orgPacks, http.StatusForbidden, nil, nil)

		_, _, err := c.ListPolicyPacksWithRegistryMetadata(ctx, testOrgName)
		assert.ErrorContains(t, err, "registry policy packs")
	})

	t.Run("registry 500 fails the call", func(t *testing.T) {
		c := registryJoinServer(t, orgPacks, http.StatusInternalServerError, nil, nil)

		_, _, err := c.ListPolicyPacksWithRegistryMetadata(ctx, testOrgName)
		assert.ErrorContains(t, err, "registry policy packs")
	})

	t.Run("org list failure fails the call and skips the registry", func(t *testing.T) {
		registryHit := false
		c := startTestServerMulti(t, func(r *http.Request) (int, any) {
			if r.URL.Path == registryPolicyPacksPath {
				registryHit = true
			}
			return http.StatusInternalServerError, ErrorResponse{StatusCode: 500, Message: "boom"}
		})

		_, _, err := c.ListPolicyPacksWithRegistryMetadata(ctx, testOrgName)
		assert.ErrorContains(t, err, "failed to list policy packs")
		assert.False(t, registryHit, "registry should not be queried when the org list fails")
	})

	t.Run("empty org rejected", func(t *testing.T) {
		c := &Client{}
		_, _, err := c.ListPolicyPacksWithRegistryMetadata(ctx, "")
		assert.EqualError(t, err, "empty orgName")
	})
}

// The wire shape must stay flat: source and publisher sit alongside the org list
// fields rather than nested under an embedded struct key.
func TestPolicyPackWithRegistryMetadata_MarshalsFlat(t *testing.T) {
	b, err := json.Marshal(PolicyPackWithRegistryMetadata{
		PolicyPackWithVersions: PolicyPackWithVersions{Name: alphaPolicyPack, DisplayName: alphaDisplayName},
		Source:                 sourcePrivate,
		Publisher:              testOrgName,
	})
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(b, &got))
	assert.Equal(t, alphaPolicyPack, got["name"])
	assert.Equal(t, sourcePrivate, got["source"])
	assert.Equal(t, testOrgName, got["publisher"])
}
