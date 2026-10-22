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

package functions

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const (
	ppTestOrg           = "anOrg"
	ppOrgListPath       = "/api/orgs/anOrg/policypacks"
	ppRegistryListPath  = "/api/registry/policypacks"
	ppOrgRegistryPath   = "/api/orgs/anOrg/registry/policypacks/alpha"
	ppAlphaVersionsPath = "/api/orgs/anOrg/policypacks/alpha/versions/"
	ppSourcePulumi      = "pulumi"
	ppSourcePrivate     = "private"
	ppCisAws            = "cis-aws"
	ppAlpha             = "alpha"
	ppAlphaDisplayName  = "Alpha"
	ppAlphaLatestTag    = "0.0.5"
	ppAlphaOlderTag     = "0.0.4"

	ppKeyName        = "name"
	ppKeyDisplayName = "displayName"
	ppKeyVersion     = "version"
	ppKeyVersions    = "versions"
	ppKeyVersionTag  = "versionTag"
	ppKeyVersionTags = "versionTags"
	ppKeyPolicyPacks = "policyPacks"
	ppKeySource      = "source"
	ppKeyPublisher   = "publisher"
	ppKeyMessage     = "message"
	ppKeyEnforcement = "enforcementLevel"

	getPolicyPacksToken = "pulumiservice:index:getPolicyPacks"
	getPolicyPackToken  = "pulumiservice:index:getPolicyPack"
)

// newPolicyPackServer runs both functions through the infer provider against a
// real API client pointed at a test HTTP server, so each test covers the routes
// the client calls, the two-call registry join and the wire encoding of the
// result.
func newPolicyPackServer(t *testing.T, handler http.HandlerFunc) integration.Server {
	t.Helper()
	httpServer := httptest.NewServer(handler)
	t.Cleanup(httpServer.Close)

	client, err := pulumiapi.NewClient(&http.Client{}, "tok", httpServer.URL)
	require.NoError(t, err)

	prov, err := infer.NewProviderBuilder().
		WithFunctions(
			infer.Function(&GetPolicyPacksFunction{}),
			infer.Function(&GetPolicyPackFunction{}),
		).
		Build()
	require.NoError(t, err)

	server, err := integration.NewServer(config.WithMockClient(t.Context(), client),
		"pulumiservice", semver.MustParse("0.0.0"), integration.WithProvider(prov))
	require.NoError(t, err)
	return server
}

func writeJSON(t *testing.T, w http.ResponseWriter, code int, body any) {
	t.Helper()
	w.WriteHeader(code)
	if body != nil {
		require.NoError(t, json.NewEncoder(w).Encode(body))
	}
}

// policyPacksHandler serves the org list plus a registry list that responds with
// registryCode. registryPacks is the registry payload when registryCode is 200.
func policyPacksHandler(
	t *testing.T,
	orgPacks []map[string]any,
	registryCode int,
	registryPacks []map[string]any,
) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ppOrgListPath:
			writeJSON(t, w, http.StatusOK, map[string]any{ppKeyPolicyPacks: orgPacks})
		case ppRegistryListPath:
			if registryCode != http.StatusOK {
				writeJSON(t, w, registryCode, map[string]any{ppKeyMessage: "nope"})
				return
			}
			writeJSON(t, w, http.StatusOK, map[string]any{ppKeyPolicyPacks: registryPacks})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func orgPackFixtures() []map[string]any {
	return []map[string]any{
		{ppKeyName: ppCisAws, ppKeyDisplayName: "CIS AWS", ppKeyVersions: []int{1}, ppKeyVersionTags: []string{"1.0.2"}},
		{
			ppKeyName: ppAlpha, ppKeyDisplayName: ppAlphaDisplayName,
			ppKeyVersions: []int{1, 2}, ppKeyVersionTags: []string{ppAlphaLatestTag},
		},
	}
}

func invokeGetPolicyPacks(t *testing.T, server integration.Server) (p.InvokeResponse, error) {
	t.Helper()
	return server.Invoke(p.InvokeRequest{
		Token: getPolicyPacksToken,
		Args:  property.NewMap(map[string]property.Value{"organizationName": property.New(ppTestOrg)}),
	})
}

// returnedPacks pulls the policyPacks array out of an invoke response, which is
// what proves the fields actually crossed the RPC boundary.
func returnedPacks(t *testing.T, resp p.InvokeResponse) []property.Map {
	t.Helper()
	require.Empty(t, resp.Failures)
	packs, ok := resp.Return.GetOk(ppKeyPolicyPacks)
	require.True(t, ok, "policyPacks should be present")
	require.True(t, packs.IsArray(), "policyPacks should be an array")
	out := make([]property.Map, 0, packs.AsArray().Len())
	for _, v := range packs.AsArray().All {
		require.True(t, v.IsMap(), "each policy pack should be an object")
		out = append(out, v.AsMap())
	}
	return out
}

func stringField(t *testing.T, m property.Map, key string) string {
	t.Helper()
	v, ok := m.GetOk(key)
	require.Truef(t, ok, "%q should be present", key)
	require.Truef(t, v.IsString(), "%q should be a string", key)
	return v.AsString()
}

func numberField(t *testing.T, m property.Map, key string) float64 {
	t.Helper()
	v, ok := m.GetOk(key)
	require.Truef(t, ok, "%q should be present", key)
	require.Truef(t, v.IsNumber(), "%q should be a number", key)
	return v.AsNumber()
}

// assertAbsent checks that a key is missing or null on the wire, never "".
// Absence keeps a degraded result distinguishable from real data in typed SDKs.
func assertAbsent(t *testing.T, m property.Map, key string) {
	t.Helper()
	v, ok := m.GetOk(key)
	if ok {
		assert.Truef(t, v.IsNull(), "%q should be absent, got %v", key, v)
	}
}

func TestGetPolicyPacks_IncludesSourceAndPublisher(t *testing.T) {
	t.Parallel()
	server := newPolicyPackServer(t, policyPacksHandler(t, orgPackFixtures(), http.StatusOK, []map[string]any{
		{ppKeyName: ppCisAws, ppKeySource: ppSourcePulumi, ppKeyPublisher: ppSourcePulumi},
		{ppKeyName: ppAlpha, ppKeySource: ppSourcePrivate, ppKeyPublisher: ppTestOrg},
	}))

	resp, err := invokeGetPolicyPacks(t, server)
	require.NoError(t, err)

	packs := returnedPacks(t, resp)
	require.Len(t, packs, 2)

	assert.Equal(t, ppSourcePulumi, stringField(t, packs[0], ppKeySource))
	assert.Equal(t, ppSourcePulumi, stringField(t, packs[0], ppKeyPublisher))

	alpha := packs[1]
	assert.Equal(t, ppSourcePrivate, stringField(t, alpha, ppKeySource))
	assert.Equal(t, ppTestOrg, stringField(t, alpha, ppKeyPublisher))

	// Provenance must not disturb the fields that were already there.
	assert.Equal(t, ppAlphaDisplayName, stringField(t, alpha, ppKeyDisplayName))
	versions, ok := alpha.GetOk(ppKeyVersions)
	require.True(t, ok)
	assert.Equal(t, 2, versions.AsArray().Len())
	tags, ok := alpha.GetOk(ppKeyVersionTags)
	require.True(t, ok)
	assert.Equal(t, ppAlphaLatestTag, tags.AsArray().Get(0).AsString())
}

// A backend without the registry route must keep serving getPolicyPacks exactly
// as it did before provenance existed.
func TestGetPolicyPacks_OmitsFieldsWhenRegistryRouteMissing(t *testing.T) {
	t.Parallel()
	server := newPolicyPackServer(t, policyPacksHandler(t, orgPackFixtures(), http.StatusNotFound, nil))

	resp, err := invokeGetPolicyPacks(t, server)
	require.NoError(t, err)

	packs := returnedPacks(t, resp)
	require.Len(t, packs, 2)
	for _, pack := range packs {
		assertAbsent(t, pack, ppKeySource)
		assertAbsent(t, pack, ppKeyPublisher)
		assert.NotEmpty(t, stringField(t, pack, ppKeyName))
	}
}

// The guard against a silent compliance hole: a broken (rather than absent)
// registry lookup must not quietly yield packs with no publisher, or a program
// filtering on publisher attaches nothing and still succeeds.
func TestGetPolicyPacks_FailsWhenRegistryErrors(t *testing.T) {
	t.Parallel()
	for _, code := range []int{http.StatusForbidden, http.StatusInternalServerError} {
		server := newPolicyPackServer(t, policyPacksHandler(t, orgPackFixtures(), code, nil))

		_, err := invokeGetPolicyPacks(t, server)
		assert.Errorf(t, err, "registry status %d should fail the invoke", code)
	}
}

func TestGetPolicyPacks_UnmatchedPackOmitsFields(t *testing.T) {
	t.Parallel()
	server := newPolicyPackServer(t, policyPacksHandler(t, orgPackFixtures(), http.StatusOK, []map[string]any{
		{ppKeyName: ppCisAws, ppKeySource: ppSourcePulumi, ppKeyPublisher: ppSourcePulumi},
	}))

	resp, err := invokeGetPolicyPacks(t, server)
	require.NoError(t, err)

	packs := returnedPacks(t, resp)
	require.Len(t, packs, 2)

	// Degradation is per pack, not all or nothing.
	assert.Equal(t, ppSourcePulumi, stringField(t, packs[0], ppKeyPublisher))
	assertAbsent(t, packs[1], ppKeyPublisher)
}

func TestGetPolicyPacks_EmptyVersionListsAreArrays(t *testing.T) {
	t.Parallel()
	server := newPolicyPackServer(t, policyPacksHandler(t,
		[]map[string]any{{ppKeyName: ppAlpha, ppKeyDisplayName: ppAlphaDisplayName}},
		http.StatusOK, nil))

	resp, err := invokeGetPolicyPacks(t, server)
	require.NoError(t, err)

	packs := returnedPacks(t, resp)
	require.Len(t, packs, 1)
	// versions and versionTags are required in the schema, so they must be sent
	// even when the API omits them.
	for _, key := range []string{ppKeyVersions, ppKeyVersionTags} {
		v, ok := packs[0].GetOk(key)
		require.Truef(t, ok, "%q should be present", key)
		assert.Truef(t, v.IsArray(), "%q should be an array", key)
	}
}

// invokeGetPolicyPack calls getPolicyPack for the alpha pack. An empty
// versionTag omits the input, which asks for the latest version.
func invokeGetPolicyPack(t *testing.T, server integration.Server, versionTag string) (p.InvokeResponse, error) {
	t.Helper()
	args := map[string]property.Value{
		"organizationName": property.New(ppTestOrg),
		"policyPackName":   property.New(ppAlpha),
	}
	if versionTag != "" {
		args[ppKeyVersionTag] = property.New(versionTag)
	}
	return server.Invoke(p.InvokeRequest{Token: getPolicyPackToken, Args: property.NewMap(args)})
}

// singlePolicyPackHandler serves the routes the real service exposes for one
// pack: the org list, the tag-keyed version route and the registry lookup. alpha
// has versions 2 (`0.0.5`) and 1 (`0.0.4`), listed newest first. There is
// deliberately no numeric-version or `/latest` route, because the service has
// neither (#1012).
func singlePolicyPackHandler(t *testing.T, registryCode int, detail map[string]any) http.HandlerFunc {
	t.Helper()
	versions := map[string]int{ppAlphaLatestTag: 2, ppAlphaOlderTag: 1}
	var fetchedTag string
	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == ppOrgListPath:
			writeJSON(t, w, http.StatusOK, map[string]any{ppKeyPolicyPacks: []map[string]any{
				{
					ppKeyName: ppAlpha, ppKeyDisplayName: ppAlphaDisplayName,
					ppKeyVersions: []int{2, 1}, ppKeyVersionTags: []string{ppAlphaLatestTag, ppAlphaOlderTag},
				},
			}})
		case strings.HasPrefix(r.URL.Path, ppAlphaVersionsPath):
			tag := strings.TrimPrefix(r.URL.Path, ppAlphaVersionsPath)
			version, ok := versions[tag]
			if !ok {
				writeJSON(t, w, http.StatusNotFound, map[string]any{ppKeyMessage: "not found"})
				return
			}
			fetchedTag = tag
			body := map[string]any{
				ppKeyName: ppAlpha, ppKeyDisplayName: ppAlphaDisplayName, ppKeyVersion: version, ppKeyVersionTag: tag,
			}
			for k, v := range detail {
				body[k] = v
			}
			writeJSON(t, w, http.StatusOK, body)
		case r.URL.Path == ppOrgRegistryPath:
			// The resolved tag must be forwarded (#1196). Omitting it makes the
			// real service look for the literal tag `latest`, which a pack tagged
			// only `0.0.5` does not have, and provenance silently disappears.
			assert.Equal(t, fetchedTag, r.URL.Query().Get("tag"),
				"registry lookup must carry the pack's resolved version tag")
			if registryCode != http.StatusOK {
				writeJSON(t, w, registryCode, map[string]any{ppKeyMessage: "nope"})
				return
			}
			writeJSON(t, w, http.StatusOK, map[string]any{
				"policyPack": map[string]any{
					ppKeyName: ppAlpha, ppKeySource: ppSourcePrivate, ppKeyPublisher: ppTestOrg,
				},
			})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func TestGetPolicyPack_LatestVersion(t *testing.T) {
	t.Parallel()
	server := newPolicyPackServer(t, singlePolicyPackHandler(t, http.StatusOK, nil))

	resp, err := invokeGetPolicyPack(t, server, "")
	require.NoError(t, err)

	assert.Equal(t, ppAlpha, stringField(t, resp.Return, ppKeyName))
	assert.EqualValues(t, 2, numberField(t, resp.Return, ppKeyVersion))
	assert.Equal(t, ppAlphaLatestTag, stringField(t, resp.Return, ppKeyVersionTag))
}

func TestGetPolicyPack_ByVersionTag(t *testing.T) {
	t.Parallel()
	server := newPolicyPackServer(t, singlePolicyPackHandler(t, http.StatusOK, nil))

	resp, err := invokeGetPolicyPack(t, server, ppAlphaOlderTag)
	require.NoError(t, err)

	assert.Equal(t, ppAlpha, stringField(t, resp.Return, ppKeyName))
	assert.EqualValues(t, 1, numberField(t, resp.Return, ppKeyVersion))
	assert.Equal(t, ppAlphaOlderTag, stringField(t, resp.Return, ppKeyVersionTag))
}

func TestGetPolicyPack_UnknownVersionTag(t *testing.T) {
	t.Parallel()
	server := newPolicyPackServer(t, singlePolicyPackHandler(t, http.StatusOK, nil))

	_, err := invokeGetPolicyPack(t, server, "9.9.9")
	assert.ErrorContains(t, err, "policy pack not found")
}

func TestGetPolicyPack_IncludesSourceAndPublisher(t *testing.T) {
	t.Parallel()
	server := newPolicyPackServer(t, singlePolicyPackHandler(t, http.StatusOK, nil))

	resp, err := invokeGetPolicyPack(t, server, "")
	require.NoError(t, err)

	assert.Equal(t, ppSourcePrivate, stringField(t, resp.Return, ppKeySource))
	assert.Equal(t, ppTestOrg, stringField(t, resp.Return, ppKeyPublisher))
	assert.Equal(t, ppAlpha, stringField(t, resp.Return, ppKeyName))
}

// The single-pack path soft-fails by design: the caller already named the pack,
// so a failed provenance lookup can't empty a filter, and failing would deny a
// result the provider could fully produce without it.
func TestGetPolicyPack_SucceedsWhenRegistryFails(t *testing.T) {
	t.Parallel()
	for _, code := range []int{http.StatusNotFound, http.StatusForbidden} {
		server := newPolicyPackServer(t, singlePolicyPackHandler(t, code, nil))

		resp, err := invokeGetPolicyPack(t, server, "")
		require.NoErrorf(t, err, "registry status %d should not fail the invoke", code)

		assert.Equal(t, ppAlpha, stringField(t, resp.Return, ppKeyName))
		assert.EqualValues(t, 2, numberField(t, resp.Return, ppKeyVersion))
		assertAbsent(t, resp.Return, ppKeySource)
		assertAbsent(t, resp.Return, ppKeyPublisher)
	}
}

func TestGetPolicyPack_ConfigAndPolicies(t *testing.T) {
	t.Parallel()
	server := newPolicyPackServer(t, singlePolicyPackHandler(t, http.StatusOK, map[string]any{
		"config": map[string]any{"all": "mandatory", "nested": map[string]any{"n": 3}},
		"policies": []map[string]any{
			{
				ppKeyName:        "no-public-buckets",
				ppKeyDisplayName: "No public buckets",
				"description":    "Buckets must be private.",
				ppKeyEnforcement: "mandatory",
				"message":        "Make it private.",
				"configSchema": map[string]any{
					"type":       "object",
					"required":   []string{"allowed"},
					"properties": map[string]any{"allowed": map[string]any{"type": "array"}},
				},
			},
			{ppKeyName: "bare"},
		},
	}))

	resp, err := invokeGetPolicyPack(t, server, "")
	require.NoError(t, err)

	cfg, ok := resp.Return.GetOk("config")
	require.True(t, ok)
	require.True(t, cfg.IsMap())
	assert.Equal(t, "mandatory", stringField(t, cfg.AsMap(), "all"))
	nested, ok := cfg.AsMap().GetOk("nested")
	require.True(t, ok)
	assert.EqualValues(t, 3, numberField(t, nested.AsMap(), "n"))

	policies, ok := resp.Return.GetOk("policies")
	require.True(t, ok)
	require.Equal(t, 2, policies.AsArray().Len())

	full := policies.AsArray().Get(0).AsMap()
	assert.Equal(t, "no-public-buckets", stringField(t, full, ppKeyName))
	assert.Equal(t, "No public buckets", stringField(t, full, ppKeyDisplayName))
	assert.Equal(t, "Buckets must be private.", stringField(t, full, "description"))
	assert.Equal(t, "mandatory", stringField(t, full, ppKeyEnforcement))
	assert.Equal(t, "Make it private.", stringField(t, full, "message"))
	schemaVal, ok := full.GetOk("configSchema")
	require.True(t, ok)
	schema := schemaVal.AsMap()
	assert.Equal(t, "object", stringField(t, schema, "type"))
	required, ok := schema.GetOk("required")
	require.True(t, ok)
	assert.Equal(t, "allowed", required.AsArray().Get(0).AsString())
	props, ok := schema.GetOk("properties")
	require.True(t, ok)
	allowed, ok := props.AsMap().GetOk("allowed")
	require.True(t, ok)
	assert.Equal(t, "array", stringField(t, allowed.AsMap(), "type"))

	// Empty policy fields are omitted, not sent as "".
	bare := policies.AsArray().Get(1).AsMap()
	assert.Equal(t, "bare", stringField(t, bare, ppKeyName))
	for _, key := range []string{ppKeyDisplayName, "description", ppKeyEnforcement, ppKeyMessage, "configSchema"} {
		assertAbsent(t, bare, key)
	}
}

func TestGetPolicyPack_OmitsEmptyOptionalFields(t *testing.T) {
	t.Parallel()
	// No versionTag, config or policies on the API response.
	got := policyPackOutputFromAPI(&pulumiapi.PolicyPackDetail{Name: ppAlpha, Version: 1}, nil)
	assert.Nil(t, got.VersionTag)
	assert.Nil(t, got.Source)
	assert.Nil(t, got.Publisher)
	assert.Nil(t, got.Config)
	assert.Nil(t, got.Policies)
}

func TestPolicyPackSummaryFromAPI_OmitsEmptyRegistryFields(t *testing.T) {
	t.Parallel()
	populated := policyPackSummaryFromAPI(pulumiapi.PolicyPackWithRegistryMetadata{
		PolicyPackWithVersions: pulumiapi.PolicyPackWithVersions{Name: ppCisAws, Versions: []int{1}},
		Source:                 ppSourcePulumi,
		Publisher:              ppSourcePulumi,
	})
	require.NotNil(t, populated.Source)
	require.NotNil(t, populated.Publisher)
	assert.Equal(t, ppSourcePulumi, *populated.Source)
	assert.Equal(t, ppSourcePulumi, *populated.Publisher)

	empty := policyPackSummaryFromAPI(pulumiapi.PolicyPackWithRegistryMetadata{
		PolicyPackWithVersions: pulumiapi.PolicyPackWithVersions{Name: ppAlpha, Versions: []int{1}},
	})
	assert.Nil(t, empty.Source)
	assert.Nil(t, empty.Publisher)
}

func TestPolicyToMap_EnforcementLevel(t *testing.T) {
	t.Parallel()
	got := policyToMap(apitype.Policy{Name: "p", EnforcementLevel: apitype.Advisory})
	assert.Equal(t, map[string]any{"name": "p", "enforcementLevel": "advisory"}, got)
}
