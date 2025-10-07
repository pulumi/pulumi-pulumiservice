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
	"fmt"
	"net/http"
)

// OrganizationSummary represents basic organization information.
type OrganizationSummary struct {
	Name        string `json:"name"`
	GithubLogin string `json:"githubLogin"`
	AvatarURL   string `json:"avatarUrl"`
}

// OrganizationSummaryWithRole represents an organization with the user's role.
type OrganizationSummaryWithRole struct {
	OrganizationSummary
	Role string `json:"role"` // none, member, admin, potential-member, stack-collaborator, billing-manager
}

// CurrentUser represents the authenticated user's information.
type CurrentUser struct {
	ID                     string                        `json:"id"`
	GithubLogin            string                        `json:"githubLogin"`
	Name                   string                        `json:"name"`
	Email                  string                        `json:"email"`
	AvatarURL              string                        `json:"avatarUrl"`
	Organizations          []OrganizationSummaryWithRole `json:"organizations"`
	PotentialOrganizations []OrganizationSummaryWithRole `json:"potentialOrganizations,omitempty"`
	Identities             []string                      `json:"identities"`
	SiteAdmin              *bool                         `json:"siteAdmin,omitempty"`
	RegistryAdmin          *bool                         `json:"registryAdmin,omitempty"`
	HasMFA                 bool                          `json:"hasMFA"`
	IsOrgManaged           bool                          `json:"isOrgManaged"`
	IsManagedByMultiOrg    bool                          `json:"isManagedByMultiOrg"`
}

// GetCurrentUser retrieves information about the currently authenticated user.
func (c *Client) GetCurrentUser(ctx context.Context) (*CurrentUser, error) {
	var user CurrentUser
	_, err := c.do(ctx, http.MethodGet, "user", nil, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}
	return &user, nil
}

// GetDefaultOrganization gets the name of the user's default organization.
func (user *CurrentUser) GetDefaultOrganization() string {
	// Return the first organization the user is a member of
	for _, org := range user.Organizations {
		if org.Role == "member" || org.Role == "admin" || org.Role == "billing-manager" {
			return org.Name
		}
	}
	// If no organizations found with standard roles, return empty string
	return ""
}
