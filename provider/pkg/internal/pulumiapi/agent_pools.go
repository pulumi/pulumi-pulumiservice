// Copyright 2016-2022, Pulumi Corporation.
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

type AgentPoolClient interface {
	CreateAgentPool(ctx context.Context, orgName, name, description string) (*AgentPool, error)
	UpdateAgentPool(ctx context.Context, agentPoolId, orgName, name, description string) error
	DeleteAgentPool(ctx context.Context, agentPoolId, orgName string) error
	GetAgentPool(ctx context.Context, agentPoolId, orgName string) (*AgentPool, error)
}

type AgentPool struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TokenValue  string `json:"tokenValue"`
}

type agentPoolResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Created        int    `json:"created"`
	LastSeen       int    `json:"lastSeen"`
	LastDeployment int    `json:"lastDeployment"`
	Status         string `json:"status"`
}

type listAgentPoolResponse struct {
	AgentPools []agentPoolResponse `json:"agentPools"`
}

type createAgentPoolResponse struct {
	ID         string `json:"id"`
	TokenValue string `json:"tokenValue"`
}

type createUpdateAgentPoolRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *Client) CreateAgentPool(ctx context.Context, orgName, name, description string) (*AgentPool, error) {

	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	if len(name) == 0 {
		return nil, errors.New("empty name")
	}

	apiPath := path.Join("orgs", orgName, "agent-pools")

	createReq := createUpdateAgentPoolRequest{
		Name:        name,
		Description: description,
	}

	var createRes createAgentPoolResponse

	_, err := c.do(ctx, http.MethodPost, apiPath, createReq, &createRes)

	if err != nil {
		return nil, fmt.Errorf("failed to create agent pool: %w", err)
	}

	return &AgentPool{
		ID:          createRes.ID,
		Name:        createReq.Name,
		Description: createReq.Description,
		TokenValue:  createRes.TokenValue,
	}, nil

}

func (c *Client) UpdateAgentPool(ctx context.Context, agentPoolId, orgName, name, description string) error {
	if len(agentPoolId) == 0 {
		return errors.New("agentPoolId length must be greater than zero")
	}

	if len(orgName) == 0 {
		return errors.New("empty orgName")
	}

	if len(name) == 0 {
		return errors.New("empty name")
	}

	apiPath := path.Join("orgs", orgName, "agent-pools", agentPoolId)

	updateReq := createUpdateAgentPoolRequest{
		Name:        name,
		Description: description,
	}

	_, err := c.do(ctx, http.MethodPatch, apiPath, updateReq, nil)
	if err != nil {
		return fmt.Errorf("failed to update agent pool: %w", err)
	}
	return nil
}

func (c *Client) DeleteAgentPool(ctx context.Context, agentPoolId, orgName string) error {
	if len(agentPoolId) == 0 {
		return errors.New("agentPoolId length must be greater than zero")
	}

	if len(orgName) == 0 {
		return errors.New("orgname length must be greater than zero")
	}

	apiPath := path.Join("orgs", orgName, "agent-pools", agentPoolId)

	fmt.Println(apiPath)

	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete agent pool %q: %w", agentPoolId, err)
	}

	return nil
}

func (c *Client) GetAgentPool(ctx context.Context, agentPoolId, orgName string) (*AgentPool, error) {
	apiPath := path.Join("orgs", orgName, "agent-pools")

	var listRes listAgentPoolResponse

	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &listRes)

	if err != nil {
		return nil, fmt.Errorf("failed to list agent pools: %w", err)
	}

	for i := 0; i < len(listRes.AgentPools); i++ {
		agentPool := listRes.AgentPools[i]
		if agentPool.ID == agentPoolId {
			return &AgentPool{
				ID:          agentPool.ID,
				Name:        agentPool.Name,
				Description: agentPool.Description,
			}, nil
		}
	}

	return nil, nil
}
