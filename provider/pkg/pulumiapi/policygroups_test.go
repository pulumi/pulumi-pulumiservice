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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreatePolicyGroup_HappyPath tests that CreatePolicyGroup sends all required fields
func TestCreatePolicyGroup_HappyPath(t *testing.T) {
	orgName := "test-org"
	policyGroupName := "test-policy-group"
	entityType := "stacks"
	mode := "audit"

	expectedReqBody := createPolicyGroupRequest{
		Name:       policyGroupName,
		EntityType: entityType,
		Mode:       mode,
	}

	c, cleanup := startTestServer(t, testServerConfig{
		ExpectedReqMethod: http.MethodPost,
		ExpectedReqPath:   "/api/orgs/test-org/policygroups",
		ExpectedReqBody:   expectedReqBody,
		ResponseCode:      201,
		ResponseBody:      nil,
	})
	defer cleanup()

	err := c.CreatePolicyGroup(ctx, orgName, policyGroupName, entityType, mode)
	assert.NoError(t, err)
}

// TestCreatePolicyGroup_AccountsPreventative tests creating a policy group with accounts and preventative mode
func TestCreatePolicyGroup_AccountsPreventative(t *testing.T) {
	orgName := "test-org"
	policyGroupName := "test-policy-group"
	entityType := "accounts"
	mode := "preventative"

	expectedReqBody := createPolicyGroupRequest{
		Name:       policyGroupName,
		EntityType: entityType,
		Mode:       mode,
	}

	c, cleanup := startTestServer(t, testServerConfig{
		ExpectedReqMethod: http.MethodPost,
		ExpectedReqPath:   "/api/orgs/test-org/policygroups",
		ExpectedReqBody:   expectedReqBody,
		ResponseCode:      201,
		ResponseBody:      nil,
	})
	defer cleanup()

	err := c.CreatePolicyGroup(ctx, orgName, policyGroupName, entityType, mode)
	assert.NoError(t, err)
}

// TestCreatePolicyGroup_EmptyOrgName tests that empty orgName returns validation error
func TestCreatePolicyGroup_EmptyOrgName(t *testing.T) {
	c := &Client{}

	err := c.CreatePolicyGroup(ctx, "", "policy-group", "stacks", "audit")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "orgName must not be empty")
}

// TestCreatePolicyGroup_EmptyPolicyGroupName tests that empty policyGroupName returns validation error
func TestCreatePolicyGroup_EmptyPolicyGroupName(t *testing.T) {
	c := &Client{}

	err := c.CreatePolicyGroup(ctx, "test-org", "", "stacks", "audit")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "policyGroupName must not be empty")
}

// TestCreatePolicyGroup_EmptyEntityType tests that empty entityType returns validation error
func TestCreatePolicyGroup_EmptyEntityType(t *testing.T) {
	c := &Client{}

	err := c.CreatePolicyGroup(ctx, "test-org", "policy-group", "", "audit")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "entityType must not be empty")
}

// TestCreatePolicyGroup_EmptyMode tests that empty mode returns validation error
func TestCreatePolicyGroup_EmptyMode(t *testing.T) {
	c := &Client{}

	err := c.CreatePolicyGroup(ctx, "test-org", "policy-group", "stacks", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mode must not be empty")
}

// TestCreatePolicyGroup_APIError tests that API errors are properly propagated
func TestCreatePolicyGroup_APIError(t *testing.T) {
	orgName := "test-org"
	policyGroupName := "test-policy-group"
	entityType := "invalid"
	mode := "audit"

	expectedReqBody := createPolicyGroupRequest{
		Name:       policyGroupName,
		EntityType: entityType,
		Mode:       mode,
	}

	c, cleanup := startTestServer(t, testServerConfig{
		ExpectedReqMethod: http.MethodPost,
		ExpectedReqPath:   "/api/orgs/test-org/policygroups",
		ExpectedReqBody:   expectedReqBody,
		ResponseCode:      400,
		ResponseBody: ErrorResponse{
			Message: "Invalid entity type",
		},
	})
	defer cleanup()

	err := c.CreatePolicyGroup(ctx, orgName, policyGroupName, entityType, mode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create policy group")
	assert.Contains(t, err.Error(), "Invalid entity type")
}

// TestCreatePolicyGroup_Unauthorized tests that 401 errors are properly handled
func TestCreatePolicyGroup_Unauthorized(t *testing.T) {
	orgName := "test-org"
	policyGroupName := "test-policy-group"
	entityType := "stacks"
	mode := "audit"

	expectedReqBody := createPolicyGroupRequest{
		Name:       policyGroupName,
		EntityType: entityType,
		Mode:       mode,
	}

	c, cleanup := startTestServer(t, testServerConfig{
		ExpectedReqMethod: http.MethodPost,
		ExpectedReqPath:   "/api/orgs/test-org/policygroups",
		ExpectedReqBody:   expectedReqBody,
		ResponseCode:      401,
		ResponseBody: ErrorResponse{
			Message: "unauthorized",
		},
	})
	defer cleanup()

	err := c.CreatePolicyGroup(ctx, orgName, policyGroupName, entityType, mode)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create policy group")
	assert.Contains(t, err.Error(), "unauthorized")
}

// TestGetPolicyGroup_IncludesEntityTypeAndMode tests that GetPolicyGroup response includes new fields
func TestGetPolicyGroup_IncludesEntityTypeAndMode(t *testing.T) {
	orgName := "test-org"
	policyGroupName := "test-policy-group"

	expectedResponse := PolicyGroup{
		Name:                policyGroupName,
		IsOrgDefault:        false,
		EntityType:          "accounts",
		Mode:                "preventative",
		Stacks:              []StackReference{},
		AppliedPolicyPacks:  []PolicyPackMetadata{},
		Accounts:            []string{},
	}

	c, cleanup := startTestServer(t, testServerConfig{
		ExpectedReqMethod: http.MethodGet,
		ExpectedReqPath:   "/api/orgs/test-org/policygroups/test-policy-group",
		ResponseCode:      200,
		ResponseBody:      expectedResponse,
	})
	defer cleanup()

	result, err := c.GetPolicyGroup(ctx, orgName, policyGroupName)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "accounts", result.EntityType)
	assert.Equal(t, "preventative", result.Mode)
}
