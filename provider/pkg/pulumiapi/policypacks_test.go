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