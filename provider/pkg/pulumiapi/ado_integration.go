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
	"context"
	"fmt"
	"net/http"
	"path"
)

type AzureDevOpsIntegrationClient interface {
	GetAzureDevOpsIntegration(
		ctx context.Context, orgName, integrationID string,
	) (*AzureDevOpsIntegration, error)
	UpdateAzureDevOpsIntegration(
		ctx context.Context, orgName, integrationID string, req UpdateAzureDevOpsIntegrationRequest,
	) error
	DeleteAzureDevOpsIntegration(ctx context.Context, orgName, integrationID string) error
	ListAzureDevOpsIntegrations(ctx context.Context, orgName string) ([]AzureDevOpsIntegration, error)
}

type UpdateAzureDevOpsIntegrationRequest struct {
	DisablePRComments   bool `json:"disablePRComments"`
	DisableNeoSummaries bool `json:"disableNeoSummaries"`
	DisableDetailedDiff bool `json:"disableDetailedDiff"`
}

type AzureDevOpsIntegration struct {
	ID                  string                  `json:"id"`
	Organization        AzureDevOpsOrganization `json:"organization"`
	Project             AzureDevOpsProject      `json:"project"`
	Valid               bool                    `json:"valid"`
	DisablePRComments   bool                    `json:"disablePRComments"`
	DisableNeoSummaries bool                    `json:"disableNeoSummaries"`
	DisableDetailedDiff bool                    `json:"disableDetailedDiff"`
}

type AzureDevOpsOrganization struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	AccountURL string `json:"accountUrl"`
}

type AzureDevOpsProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type listAzureDevOpsIntegrationsResponse struct {
	Integrations []AzureDevOpsIntegration `json:"integrations"`
}

func adoIntegrationBasePath(orgName string) string {
	return path.Join("console", "orgs", orgName, "integrations", "azure-devops")
}

func (c *Client) GetAzureDevOpsIntegration(
	ctx context.Context, orgName, integrationID string,
) (*AzureDevOpsIntegration, error) {
	apiPath := path.Join(adoIntegrationBasePath(orgName), integrationID)

	var integration AzureDevOpsIntegration
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &integration)
	if err != nil {
		statusCode := GetErrorStatusCode(err)
		if statusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get azure devops integration: %w", err)
	}

	return &integration, nil
}

func (c *Client) UpdateAzureDevOpsIntegration(
	ctx context.Context, orgName, integrationID string, req UpdateAzureDevOpsIntegrationRequest,
) error {
	apiPath := path.Join(adoIntegrationBasePath(orgName), integrationID)

	_, err := c.do(ctx, http.MethodPatch, apiPath, req, nil)
	if err != nil {
		return fmt.Errorf("failed to update azure devops integration: %w", err)
	}

	return nil
}

func (c *Client) DeleteAzureDevOpsIntegration(ctx context.Context, orgName, integrationID string) error {
	apiPath := path.Join(adoIntegrationBasePath(orgName), integrationID)

	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete azure devops integration: %w", err)
	}

	return nil
}

func (c *Client) ListAzureDevOpsIntegrations(ctx context.Context, orgName string) ([]AzureDevOpsIntegration, error) {
	apiPath := adoIntegrationBasePath(orgName)

	var listRes listAzureDevOpsIntegrationsResponse
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &listRes)
	if err != nil {
		return nil, fmt.Errorf("failed to list azure devops integrations: %w", err)
	}

	return listRes.Integrations, nil
}
