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
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type AgentDefinitionClient interface {
	GetAgentDefinition(ctx context.Context, organization, id string) (*AgentDefinition, error)
	GetActiveAgentDefinition(ctx context.Context, organization, id string) (*AgentDefinition, error)
	CreateAgentDefinition(
		ctx context.Context, organization string, request AgentDefinitionRequest,
	) (*AgentDefinition, error)
	UpdateAgentDefinition(
		ctx context.Context, organization, id string, request AgentDefinitionRequest,
	) (*AgentDefinition, error)
	DeleteAgentDefinition(ctx context.Context, organization, id string) error
}

// AgentDefinition contains the saved configuration, without task histories.
// The getAgent invoke continues to expose only identity metadata.
type AgentDefinition struct {
	ID string `json:"id"`
	AgentDefinitionRequest
}

type AgentTaskIntegrationRef struct {
	ID string `json:"id"`
}

// EnabledIntegrations deliberately has no omitempty: nil encodes as null
// (inherit all organization integrations), while an empty slice encodes as []
// (enable none). RoleID is omitted when unset, preserving the role on updates.
type AgentDefinitionRequest struct {
	Name                string                    `json:"name"`
	DisplayName         string                    `json:"displayName,omitempty"`
	RoleID              *string                   `json:"roleId,omitempty"`
	CustomInstructions  string                    `json:"customInstructions,omitempty"`
	EnabledIntegrations []AgentTaskIntegrationRef `json:"enabledIntegrations"`
	Purpose             string                    `json:"purpose,omitempty"`
	Avatar              *string                   `json:"avatar,omitempty"`
	ApprovalMode        string                    `json:"approvalMode,omitempty"`
}

func agentDefinitionsPath(organization string) (string, error) {
	if organization == "" || organization == "." || organization == ".." ||
		url.PathEscape(organization) != organization {
		return "", fmt.Errorf("organization must be a non-empty organization name")
	}
	return "preview/agents/" + organization + "/definitions", nil
}

func agentDefinitionPath(organization, id string) (string, error) {
	collection, err := agentDefinitionsPath(organization)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(id) == "" || id == "." || id == ".." || url.PathEscape(id) != id {
		return "", fmt.Errorf("id must be a non-empty Custom Agent definition ID")
	}
	return collection + "/" + id, nil
}

func (c *Client) GetAgentDefinition(ctx context.Context, organization, id string) (*AgentDefinition, error) {
	definition, err := c.GetActiveAgentDefinition(ctx, organization, id)
	if err != nil {
		return nil, err
	}
	if definition == nil {
		return nil, fmt.Errorf("custom agent %q in organization %q not found or not visible", id, organization)
	}
	return definition, nil
}

// GetActiveAgentDefinition returns nil for a definition missing from the active
// roster, so resource refresh can remove deleted definitions from state.
func (c *Client) GetActiveAgentDefinition(ctx context.Context, organization, id string) (*AgentDefinition, error) {
	if _, err := agentDefinitionPath(organization, id); err != nil {
		return nil, err
	}
	collection, err := agentDefinitionsPath(organization)
	if err != nil {
		return nil, err
	}

	// The direct GET-by-ID endpoint retains deleted definitions for historical
	// task attribution. The active roster excludes them and applies the caller's
	// role visibility, so it is the appropriate source for a new workflow member.
	var response struct {
		Definitions []AgentDefinition `json:"definitions"`
	}
	_, err = c.do(ctx, http.MethodGet, collection, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to look up Custom Agent: %w", err)
	}
	for _, definition := range response.Definitions {
		if definition.ID == id {
			return &definition, nil
		}
	}
	return nil, nil
}

func (c *Client) CreateAgentDefinition(
	ctx context.Context, organization string, request AgentDefinitionRequest,
) (*AgentDefinition, error) {
	apiPath, err := agentDefinitionsPath(organization)
	if err != nil {
		return nil, err
	}
	var definition AgentDefinition
	if _, err := c.do(ctx, http.MethodPost, apiPath, request, &definition); err != nil {
		return nil, fmt.Errorf("failed to create Custom Agent: %w", err)
	}
	return &definition, nil
}

func (c *Client) UpdateAgentDefinition(
	ctx context.Context, organization, id string, request AgentDefinitionRequest,
) (*AgentDefinition, error) {
	apiPath, err := agentDefinitionPath(organization, id)
	if err != nil {
		return nil, err
	}
	var definition AgentDefinition
	if _, err := c.do(ctx, http.MethodPut, apiPath, request, &definition); err != nil {
		return nil, fmt.Errorf("failed to update Custom Agent: %w", err)
	}
	return &definition, nil
}

func (c *Client) DeleteAgentDefinition(ctx context.Context, organization, id string) error {
	apiPath, err := agentDefinitionPath(organization, id)
	if err != nil {
		return err
	}
	_, err = c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil && GetErrorStatusCode(err) != http.StatusNotFound {
		return fmt.Errorf("failed to delete Custom Agent: %w", err)
	}
	return nil
}
