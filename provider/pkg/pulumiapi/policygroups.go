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
	"errors"
	"fmt"
	"net/http"
	"path"
)

type PolicyGroupClient interface {
	ListPolicyGroups(ctx context.Context, orgName string) ([]PolicyGroupSummary, error)
	GetPolicyGroup(ctx context.Context, orgName string, policyGroupName string) (*PolicyGroup, error)
	CreatePolicyGroup(ctx context.Context, orgName, policyGroupName string) error
	UpdatePolicyGroup(ctx context.Context, orgName, policyGroupName string, req UpdatePolicyGroupRequest) error
	DeletePolicyGroup(ctx context.Context, orgName, policyGroupName string) error
}

type PolicyGroupSummary struct {
	Name                   string `json:"name"`
	IsOrgDefault           bool   `json:"isOrgDefault"`
	NumStacks              int    `json:"numStacks"`
	NumAccounts            int    `json:"numAccounts"`
	EntityType             string `json:"entityType"`
	NumEnabledPolicyPacks  int    `json:"numEnabledPolicyPacks"`
}

type PolicyGroup struct {
	Name                string               `json:"name"`
	IsOrgDefault        bool                 `json:"isOrgDefault"`
	Stacks              []StackReference     `json:"stacks"`
	AppliedPolicyPacks  []PolicyPackMetadata `json:"appliedPolicyPacks"`
	Accounts            []string             `json:"accounts"`
}

type StackReference struct {
	Name           string `json:"name"`
	RoutingProject string `json:"routingProject"`
}

type PolicyPackMetadata struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"displayName"`
	Version     int                    `json:"version"`
	VersionTag  string                 `json:"versionTag"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

type listPolicyGroupsResponse struct {
	PolicyGroups []PolicyGroupSummary `json:"policyGroups"`
}

type createPolicyGroupRequest struct {
	Name string `json:"name"`
}

type UpdatePolicyGroupRequest struct {
	NewName          *string               `json:"newName,omitempty"`
	AddStack         *StackReference       `json:"addStack,omitempty"`
	RemoveStack      *StackReference       `json:"removeStack,omitempty"`
	AddPolicyPack    *PolicyPackMetadata   `json:"addPolicyPack,omitempty"`
	RemovePolicyPack *PolicyPackMetadata   `json:"removePolicyPack,omitempty"`
}

func (c *Client) ListPolicyGroups(ctx context.Context, orgName string) ([]PolicyGroupSummary, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	apiPath := path.Join("orgs", orgName, "policygroups")

	var response listPolicyGroupsResponse
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list policy groups for %q: %w", orgName, err)
	}
	return response.PolicyGroups, nil
}

func (c *Client) GetPolicyGroup(ctx context.Context, orgName string, policyGroupName string) (*PolicyGroup, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	if len(policyGroupName) == 0 {
		return nil, errors.New("empty policyGroupName")
	}

	apiPath := path.Join("orgs", orgName, "policygroups", policyGroupName)

	var policyGroup PolicyGroup
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &policyGroup)
	if err != nil {
		statusCode := GetErrorStatusCode(err)
		if statusCode == http.StatusNotFound {
			// Important: we return nil here to hint it was not found
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get policy group: %w", err)
	}

	return &policyGroup, nil
}

func (c *Client) CreatePolicyGroup(ctx context.Context, orgName, policyGroupName string) error {
	if len(orgName) == 0 {
		return errors.New("orgName must not be empty")
	}

	if len(policyGroupName) == 0 {
		return errors.New("policyGroupName must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "policygroups")

	req := createPolicyGroupRequest{
		Name: policyGroupName,
	}

	_, err := c.do(ctx, http.MethodPost, apiPath, req, nil)
	if err != nil {
		return fmt.Errorf("failed to create policy group %q: %w", policyGroupName, err)
	}

	return nil
}

func (c *Client) UpdatePolicyGroup(ctx context.Context, orgName, policyGroupName string, req UpdatePolicyGroupRequest) error {
	if len(orgName) == 0 {
		return errors.New("orgName must not be empty")
	}

	if len(policyGroupName) == 0 {
		return errors.New("policyGroupName must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "policygroups", policyGroupName)

	_, err := c.do(ctx, http.MethodPatch, apiPath, req, nil)
	if err != nil {
		return fmt.Errorf("failed to update policy group %q: %w", policyGroupName, err)
	}

	return nil
}

func (c *Client) DeletePolicyGroup(ctx context.Context, orgName, policyGroupName string) error {
	if len(orgName) == 0 {
		return errors.New("orgName must not be empty")
	}

	if len(policyGroupName) == 0 {
		return errors.New("policyGroupName must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "policygroups", policyGroupName)

	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete policy group %q: %w", policyGroupName, err)
	}

	return nil
}
