// Copyright 2016-2025, Pulumi Corporation.
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
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListPolicyPacks_Sorting(t *testing.T) {
	// Create test data with unsorted policy packs
	testResponse := listPolicyPacksResponse{
		PolicyPacks: []PolicyPackWithVersions{
			{Name: "zebra-pack", DisplayName: "Zebra Pack", Versions: []int{1}, VersionTags: []string{"latest"}},
			{Name: "alpha-pack", DisplayName: "Alpha Pack", Versions: []int{1, 2}, VersionTags: []string{"v1", "v2"}},
			{Name: "middle-pack", DisplayName: "Middle Pack", Versions: []int{1}, VersionTags: []string{"latest"}},
			{Name: "beta-pack", DisplayName: "Beta Pack", Versions: []int{3}, VersionTags: []string{"stable"}},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/orgs/test-org/policypacks", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(testResponse)
		require.NoError(t, err)
	}))
	defer server.Close()

	client, err := NewClient(server.Client(), "test-token", server.URL)
	require.NoError(t, err)

	result, err := client.ListPolicyPacks(context.Background(), "test-org")
	require.NoError(t, err)
	require.Len(t, result, 4)

	// Verify policy packs are sorted by name
	expectedOrder := []string{"alpha-pack", "beta-pack", "middle-pack", "zebra-pack"}
	actualOrder := make([]string, len(result))
	for i, pack := range result {
		actualOrder[i] = pack.Name
	}

	assert.Equal(t, expectedOrder, actualOrder, "Policy packs should be sorted by name")

	// Verify other properties are preserved
	assert.Equal(t, "Alpha Pack", result[0].DisplayName)
	assert.Equal(t, []int{1, 2}, result[0].Versions)
	assert.Equal(t, []string{"v1", "v2"}, result[0].VersionTags)
}

func TestListPolicyPacks_EmptyOrgName(t *testing.T) {
	client, err := NewClient(&http.Client{}, "test-token", "http://example.com")
	require.NoError(t, err)

	result, err := client.ListPolicyPacks(context.Background(), "")
	assert.Nil(t, result)
	assert.EqualError(t, err, "empty orgName")
}

func TestListPolicyPacks_EmptyResult(t *testing.T) {
	testResponse := listPolicyPacksResponse{
		PolicyPacks: []PolicyPackWithVersions{},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(testResponse)
		require.NoError(t, err)
	}))
	defer server.Close()

	client, err := NewClient(server.Client(), "test-token", server.URL)
	require.NoError(t, err)

	result, err := client.ListPolicyPacks(context.Background(), "test-org")
	require.NoError(t, err)
	assert.Empty(t, result)
}