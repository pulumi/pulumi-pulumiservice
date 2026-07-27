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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const (
	invokeTestOrg      = "anOrg"
	orgPolicyPacksPath = "/api/orgs/anOrg/policypacks"
	registryListPath   = "/api/registry/policypacks"
	orgRegistryPath    = "/api/orgs/anOrg/registry/policypacks/alpha"
	latestPolicyPack   = "/api/orgs/anOrg/policypacks/alpha/latest"
	testSourcePulumi   = "pulumi"
	testSourcePrivate  = "private"
	cisAwsPack         = "cis-aws"
	alphaPack          = "alpha"
)

// newInvokeTestProvider wires a provider against a test server that serves both
// the org policy pack endpoints and the registry endpoints, so an invoke can be
// driven end to end across the two-call join.
//
// host is left nil: the degradation warning guards on it, and these tests assert
// on the returned property map rather than on emitted diagnostics.
func newInvokeTestProvider(t *testing.T, handler http.HandlerFunc) *pulumiserviceProvider {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := pulumiapi.NewClient(&http.Client{}, "tok", server.URL)
	require.NoError(t, err)
	return &pulumiserviceProvider{client: client}
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
		case orgPolicyPacksPath:
			writeJSON(t, w, http.StatusOK, map[string]any{policyPacksKey: orgPacks})
		case registryListPath:
			if registryCode != http.StatusOK {
				writeJSON(t, w, registryCode, map[string]any{"message": "nope"})
				return
			}
			writeJSON(t, w, http.StatusOK, map[string]any{policyPacksKey: registryPacks})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func invokeGetPolicyPacks(
	t *testing.T,
	k *pulumiserviceProvider,
) (*pulumirpc.InvokeResponse, error) {
	t.Helper()
	args, err := plugin.MarshalProperties(resource.PropertyMap{
		"organizationName": resource.NewStringProperty(invokeTestOrg),
	}, plugin.MarshalOptions{})
	require.NoError(t, err)

	return k.Invoke(context.Background(), &pulumirpc.InvokeRequest{
		Tok:  "pulumiservice:index:getPolicyPacks",
		Args: args,
	})
}

// returnedPacks unmarshals an invoke response back into a property map, which is
// what proves the fields actually crossed the RPC boundary.
func returnedPacks(t *testing.T, resp *pulumirpc.InvokeResponse) []resource.PropertyValue {
	t.Helper()
	out, err := plugin.UnmarshalProperties(resp.GetReturn(), plugin.MarshalOptions{})
	require.NoError(t, err)
	require.True(t, out[policyPacksKey].IsArray(), "policyPacks should be an array")
	return out[policyPacksKey].ArrayValue()
}

func orgPackFixtures() []map[string]any {
	return []map[string]any{
		{nameKey: cisAwsPack, displayNameKey: "CIS AWS", versionsKey: []int{1}, versionTagsKey: []string{"1.0.2"}},
		{nameKey: alphaPack, displayNameKey: "Alpha", versionsKey: []int{1, 2}, versionTagsKey: []string{"0.0.5"}},
	}
}

func TestInvokeGetPolicyPacks_IncludesSourceAndPublisher(t *testing.T) {
	k := newInvokeTestProvider(t, policyPacksHandler(t, orgPackFixtures(), http.StatusOK, []map[string]any{
		{nameKey: cisAwsPack, sourceKey: testSourcePulumi, publisherKey: testSourcePulumi},
		{nameKey: alphaPack, sourceKey: testSourcePrivate, publisherKey: invokeTestOrg},
	}))

	resp, err := invokeGetPolicyPacks(t, k)
	require.NoError(t, err)

	packs := returnedPacks(t, resp)
	require.Len(t, packs, 2)

	cis := packs[0].ObjectValue()
	assert.Equal(t, testSourcePulumi, cis[sourceKey].StringValue())
	assert.Equal(t, testSourcePulumi, cis[publisherKey].StringValue())

	alpha := packs[1].ObjectValue()
	assert.Equal(t, testSourcePrivate, alpha[sourceKey].StringValue())
	assert.Equal(t, invokeTestOrg, alpha[publisherKey].StringValue())

	// Provenance must not disturb the fields that were already there.
	assert.Equal(t, "Alpha", alpha[displayNameKey].StringValue())
	assert.Len(t, alpha[versionsKey].ArrayValue(), 2)
}

// A backend without the registry route must keep serving getPolicyPacks exactly
// as it did before this feature existed.
func TestInvokeGetPolicyPacks_OmitsFieldsWhenRegistryRouteMissing(t *testing.T) {
	k := newInvokeTestProvider(t, policyPacksHandler(t, orgPackFixtures(), http.StatusNotFound, nil))

	resp, err := invokeGetPolicyPacks(t, k)
	require.NoError(t, err)

	packs := returnedPacks(t, resp)
	require.Len(t, packs, 2)

	for _, pack := range packs {
		obj := pack.ObjectValue()
		// Absence, not "": the schema marks these optional so a degraded result
		// is distinguishable from a real one in typed SDKs.
		_, hasSource := obj[sourceKey]
		_, hasPublisher := obj[publisherKey]
		assert.False(t, hasSource, "source should be absent, not empty")
		assert.False(t, hasPublisher, "publisher should be absent, not empty")
		assert.NotEmpty(t, obj[nameKey].StringValue())
	}
}

// The end-to-end guard against the silent compliance hole: a broken (rather than
// absent) registry lookup must not quietly yield packs with no publisher, or a
// program filtering on publisher attaches nothing and still succeeds.
func TestInvokeGetPolicyPacks_FailsWhenRegistryErrors(t *testing.T) {
	for _, code := range []int{http.StatusForbidden, http.StatusInternalServerError} {
		k := newInvokeTestProvider(t, policyPacksHandler(t, orgPackFixtures(), code, nil))

		_, err := invokeGetPolicyPacks(t, k)
		assert.Errorf(t, err, "registry status %d should fail the invoke", code)
	}
}

func TestInvokeGetPolicyPacks_UnmatchedPackOmitsFields(t *testing.T) {
	k := newInvokeTestProvider(t, policyPacksHandler(t, orgPackFixtures(), http.StatusOK, []map[string]any{
		{nameKey: cisAwsPack, sourceKey: testSourcePulumi, publisherKey: testSourcePulumi},
	}))

	resp, err := invokeGetPolicyPacks(t, k)
	require.NoError(t, err)

	packs := returnedPacks(t, resp)
	require.Len(t, packs, 2)

	// Degradation is per pack, not all or nothing.
	assert.Equal(t, testSourcePulumi, packs[0].ObjectValue()[publisherKey].StringValue())
	_, hasPublisher := packs[1].ObjectValue()[publisherKey]
	assert.False(t, hasPublisher)
}

func invokeGetPolicyPack(t *testing.T, k *pulumiserviceProvider) (*pulumirpc.InvokeResponse, error) {
	t.Helper()
	args, err := plugin.MarshalProperties(resource.PropertyMap{
		"organizationName": resource.NewStringProperty(invokeTestOrg),
		"policyPackName":   resource.NewStringProperty(alphaPack),
	}, plugin.MarshalOptions{})
	require.NoError(t, err)

	return k.Invoke(context.Background(), &pulumirpc.InvokeRequest{
		Tok:  "pulumiservice:index:getPolicyPack",
		Args: args,
	})
}

func singlePolicyPackHandler(t *testing.T, registryCode int) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case latestPolicyPack:
			writeJSON(t, w, http.StatusOK, map[string]any{
				nameKey: alphaPack, displayNameKey: "Alpha", versionKey: 2, "versionTag": "0.0.5",
			})
		case orgRegistryPath:
			// The resolved tag must be forwarded. Omitting it makes the real
			// service look for the literal tag `latest`, which a pack tagged
			// only `0.0.5` does not have, and provenance silently disappears.
			assert.Equal(t, "0.0.5", r.URL.Query().Get("tag"),
				"registry lookup must carry the pack's resolved version tag")
			if registryCode != http.StatusOK {
				writeJSON(t, w, registryCode, map[string]any{"message": "nope"})
				return
			}
			writeJSON(t, w, http.StatusOK, map[string]any{
				"policyPack": map[string]any{
					nameKey: alphaPack, sourceKey: testSourcePrivate, publisherKey: invokeTestOrg,
				},
			})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func TestInvokeGetPolicyPack_IncludesSourceAndPublisher(t *testing.T) {
	k := newInvokeTestProvider(t, singlePolicyPackHandler(t, http.StatusOK))

	resp, err := invokeGetPolicyPack(t, k)
	require.NoError(t, err)

	out, err := plugin.UnmarshalProperties(resp.GetReturn(), plugin.MarshalOptions{})
	require.NoError(t, err)
	assert.Equal(t, testSourcePrivate, out[sourceKey].StringValue())
	assert.Equal(t, invokeTestOrg, out[publisherKey].StringValue())
	assert.Equal(t, alphaPack, out[nameKey].StringValue())
}

// The single-pack path soft-fails by design: the caller already named the pack,
// so a failed provenance lookup can't empty a filter, and failing would deny a
// result the provider could fully produce before this feature.
func TestInvokeGetPolicyPack_SucceedsWhenRegistryFails(t *testing.T) {
	for _, code := range []int{http.StatusNotFound, http.StatusForbidden} {
		k := newInvokeTestProvider(t, singlePolicyPackHandler(t, code))

		resp, err := invokeGetPolicyPack(t, k)
		require.NoErrorf(t, err, "registry status %d should not fail the invoke", code)

		out, err := plugin.UnmarshalProperties(resp.GetReturn(), plugin.MarshalOptions{})
		require.NoError(t, err)
		assert.Equal(t, alphaPack, out[nameKey].StringValue())
		assert.EqualValues(t, 2, out[versionKey].NumberValue())

		_, hasSource := out[sourceKey]
		_, hasPublisher := out[publisherKey]
		assert.False(t, hasSource)
		assert.False(t, hasPublisher)
	}
}

func TestConvertPolicyPacksToProperties_OmitsEmptyRegistryFields(t *testing.T) {
	got := convertPolicyPacksToProperties([]pulumiapi.PolicyPackWithRegistryMetadata{
		{
			PolicyPackWithVersions: pulumiapi.PolicyPackWithVersions{Name: cisAwsPack, Versions: []int{1}},
			Source:                 testSourcePulumi,
			Publisher:              testSourcePulumi,
		},
		{
			PolicyPackWithVersions: pulumiapi.PolicyPackWithVersions{Name: alphaPack, Versions: []int{1}},
		},
	})
	require.Len(t, got, 2)

	populated := got[0].ObjectValue()
	assert.Equal(t, testSourcePulumi, populated[sourceKey].StringValue())
	assert.Equal(t, testSourcePulumi, populated[publisherKey].StringValue())

	empty := got[1].ObjectValue()
	_, hasSource := empty[sourceKey]
	_, hasPublisher := empty[publisherKey]
	assert.False(t, hasSource)
	assert.False(t, hasPublisher)
}

func TestConvertPolicyPackDetailToProperties_NilRegistryOmitsFields(t *testing.T) {
	got := convertPolicyPackDetailToProperties(
		&pulumiapi.PolicyPackDetail{Name: alphaPack, Version: 1}, nil)

	_, hasSource := got[sourceKey]
	_, hasPublisher := got[publisherKey]
	assert.False(t, hasSource)
	assert.False(t, hasPublisher)
}
