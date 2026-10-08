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

package pulumiapi

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// futureScope is not in the cloud SDK's RbacPermission enum, whose
// unmarshaller would blank it. The permission set client must not.
const (
	futureScope   = "stack:time_travel"
	deployerSet   = "Deployer"
	stackReadPerm = "stack:read"
)

func TestCreatePermissionSetKeepsUnknownScopes(t *testing.T) {
	record := permissionSetRecord{
		ID: "set-1", Name: deployerSet, ResourceType: stackKey, UxPurpose: uxPurposeSet, Version: 1,
		Details: permissionSetDetails{Type: permissionDescriptorAllow, Permissions: []string{stackReadPerm, futureScope}},
	}
	c := startTestServer(t, testServerConfig{
		ExpectedReqMethod: http.MethodPost,
		ExpectedReqPath:   testOrgRolesPath,
		ExpectedReqBody: permissionSetBody{
			Name:         deployerSet,
			ResourceType: stackKey,
			UxPurpose:    uxPurposeSet,
			Details: &permissionSetDetails{
				Type:        permissionDescriptorAllow,
				Permissions: []string{stackReadPerm, futureScope},
			},
		},
		ResponseCode: http.StatusOK,
		ResponseBody: record,
	})
	set, err := c.CreatePermissionSet(t.Context(), testRoleOrgName, PermissionSetRequest{
		Name: deployerSet, ResourceType: stackKey, Permissions: []string{stackReadPerm, futureScope},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{stackReadPerm, futureScope}, set.Permissions)
	assert.Equal(t, permissionDescriptorAllow, set.DetailsType)
}

func TestGetPermissionSet(t *testing.T) {
	t.Run("keeps unknown scopes", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   testOrgRolesPath + "/set-1",
			ResponseCode:      http.StatusOK,
			ResponseBody: permissionSetRecord{
				ID: "set-1", UxPurpose: uxPurposeSet, DefaultIdentifier: "stack-read",
				Details: permissionSetDetails{Type: permissionDescriptorAllow, Permissions: []string{futureScope}},
			},
		})
		set, err := c.GetPermissionSet(t.Context(), testRoleOrgName, "set-1")
		require.NoError(t, err)
		assert.Equal(t, []string{futureScope}, set.Permissions)
		assert.Equal(t, "stack-read", set.DefaultIdentifier)
	})

	t.Run("not found", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   testOrgRolesPath + "/set-1",
			ResponseCode:      http.StatusNotFound,
			ResponseBody:      ErrorResponse{Message: notFoundError},
		})
		set, err := c.GetPermissionSet(t.Context(), testRoleOrgName, "set-1")
		require.NoError(t, err)
		assert.Nil(t, set)
	})
}

func TestListPermissionSets(t *testing.T) {
	c := startTestServer(t, testServerConfig{
		ExpectedReqMethod:   http.MethodGet,
		ExpectedReqPath:     testOrgRolesPath,
		ExpectedQueryParams: url.Values{"uxPurpose": []string{uxPurposeSet}},
		ResponseCode:        http.StatusOK,
		ResponseBody: map[string][]permissionSetRecord{"roles": {
			{ID: "s1", Name: "Stack Read", ResourceType: stackKey},
			{ID: "s2", Name: "Read Only", ResourceType: "global"},
		}},
	})
	sets, err := c.ListPermissionSets(t.Context(), testRoleOrgName)
	require.NoError(t, err)
	require.Len(t, sets, 2)
	assert.Equal(t, "Read Only", sets[1].Name)
}
