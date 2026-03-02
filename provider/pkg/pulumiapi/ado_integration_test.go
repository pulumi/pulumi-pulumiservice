// Copyright 2016-2026, Pulumi Corporation.
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

const testAdoIntegrationID = "integration-1"

func TestListAzureDevOpsIntegrations(t *testing.T) {
	orgName := testPolicyGroupOrgName
	integrations := []AzureDevOpsIntegration{
		{
			ID: "integration-1",
			Organization: AzureDevOpsOrganization{
				ID:         "ado-org-id",
				Name:       "my-ado-org",
				AccountURL: "https://dev.azure.com/my-ado-org",
			},
			Project: AzureDevOpsProject{
				ID:   "project-id-1",
				Name: "my-project",
			},
			Valid:               true,
			DisablePRComments:   false,
			DisableNeoSummaries: false,
			DisableDetailedDiff: false,
		},
	}

	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/console/orgs/test-org/integrations/azure-devops",
			ResponseCode:      200,
			ResponseBody: listAzureDevOpsIntegrationsResponse{
				Integrations: integrations,
			},
		})
		actual, err := c.ListAzureDevOpsIntegrations(ctx, orgName)
		require.NoError(t, err)
		assert.Equal(t, integrations, actual)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/console/orgs/test-org/integrations/azure-devops",
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		actual, err := c.ListAzureDevOpsIntegrations(ctx, orgName)
		assert.Nil(t, actual, "integrations should be nil since error was returned")
		assert.EqualError(t, err, "failed to list azure devops integrations: 401 API error: unauthorized")
	})
}

func TestGetAzureDevOpsIntegration(t *testing.T) {
	orgName := testPolicyGroupOrgName
	integrationID := testAdoIntegrationID
	integration := AzureDevOpsIntegration{
		ID: integrationID,
		Organization: AzureDevOpsOrganization{
			ID:         "ado-org-id",
			Name:       "my-ado-org",
			AccountURL: "https://dev.azure.com/my-ado-org",
		},
		Project: AzureDevOpsProject{
			ID:   "project-id-1",
			Name: "my-project",
		},
		Valid:               true,
		DisablePRComments:   false,
		DisableNeoSummaries: true,
		DisableDetailedDiff: false,
	}

	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/console/orgs/test-org/integrations/azure-devops/integration-1",
			ResponseCode:      200,
			ResponseBody:      integration,
		})
		actual, err := c.GetAzureDevOpsIntegration(ctx, orgName, integrationID)
		require.NoError(t, err)
		assert.Equal(t, &integration, actual)
	})

	t.Run("Not Found", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/console/orgs/test-org/integrations/azure-devops/integration-1",
			ResponseCode:      404,
			ResponseBody: ErrorResponse{
				StatusCode: 404,
				Message:    "not found",
			},
		})
		actual, err := c.GetAzureDevOpsIntegration(ctx, orgName, integrationID)
		assert.Nil(t, actual, "integration should be nil for 404")
		assert.Nil(t, err, "err should be nil for 404")
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath:   "/api/console/orgs/test-org/integrations/azure-devops/integration-1",
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		actual, err := c.GetAzureDevOpsIntegration(ctx, orgName, integrationID)
		assert.Nil(t, actual, "integration should be nil since error was returned")
		assert.EqualError(t, err, "failed to get azure devops integration: 401 API error: unauthorized")
	})
}

func TestUpdateAzureDevOpsIntegration(t *testing.T) {
	orgName := testPolicyGroupOrgName
	integrationID := testAdoIntegrationID
	updateReq := UpdateAzureDevOpsIntegrationRequest{
		DisablePRComments:   true,
		DisableNeoSummaries: false,
		DisableDetailedDiff: true,
	}

	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/console/orgs/test-org/integrations/azure-devops/integration-1",
			ExpectedReqBody:   updateReq,
			ResponseCode:      204,
		})
		err := c.UpdateAzureDevOpsIntegration(ctx, orgName, integrationID, updateReq)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPatch,
			ExpectedReqPath:   "/api/console/orgs/test-org/integrations/azure-devops/integration-1",
			ExpectedReqBody:   updateReq,
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		err := c.UpdateAzureDevOpsIntegration(ctx, orgName, integrationID, updateReq)
		assert.EqualError(t, err, "failed to update azure devops integration: 401 API error: unauthorized")
	})
}

func TestDeleteAzureDevOpsIntegration(t *testing.T) {
	orgName := testPolicyGroupOrgName
	integrationID := testAdoIntegrationID

	t.Run("Happy Path", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/console/orgs/test-org/integrations/azure-devops/integration-1",
			ResponseCode:      204,
		})
		err := c.DeleteAzureDevOpsIntegration(ctx, orgName, integrationID)
		assert.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodDelete,
			ExpectedReqPath:   "/api/console/orgs/test-org/integrations/azure-devops/integration-1",
			ResponseCode:      401,
			ResponseBody: ErrorResponse{
				Message: "unauthorized",
			},
		})
		err := c.DeleteAzureDevOpsIntegration(ctx, orgName, integrationID)
		assert.EqualError(t, err, "failed to delete azure devops integration: 401 API error: unauthorized")
	})
}
