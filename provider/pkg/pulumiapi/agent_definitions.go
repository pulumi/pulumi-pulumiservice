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
}

// AgentDefinition contains only the identity metadata needed to reference a
// saved Custom Agent. Instructions and permissions are not exposed by the lookup.
type AgentDefinition struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

func (c *Client) GetAgentDefinition(ctx context.Context, organization, id string) (*AgentDefinition, error) {
	if organization == "" || organization == "." || organization == ".." ||
		url.PathEscape(organization) != organization {
		return nil, fmt.Errorf("organization must be a non-empty organization name")
	}
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("id must be a non-empty Custom Agent definition ID")
	}

	// The direct GET-by-ID endpoint retains deleted definitions for historical
	// task attribution. The active roster excludes them and applies the caller's
	// role visibility, so it is the appropriate source for a new workflow member.
	var response struct {
		Definitions []AgentDefinition `json:"definitions"`
	}
	_, err := c.do(ctx, http.MethodGet, "preview/agents/"+organization+"/definitions", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to look up Custom Agent: %w", err)
	}
	for _, definition := range response.Definitions {
		if definition.ID == id {
			return &definition, nil
		}
	}
	return nil, fmt.Errorf("custom agent %q in organization %q not found or not visible", id, organization)
}
