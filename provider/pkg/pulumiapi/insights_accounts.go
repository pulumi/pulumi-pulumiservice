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

type InsightsAccountClient interface {
	CreateInsightsAccount(ctx context.Context, orgName, accountName string, req CreateInsightsAccountRequest) error
	GetInsightsAccount(ctx context.Context, orgName, accountName string) (*InsightsAccount, error)
	UpdateInsightsAccount(ctx context.Context, orgName, accountName string, req UpdateInsightsAccountRequest) error
	DeleteInsightsAccount(ctx context.Context, orgName, accountName string) error
}

type InsightsAccount struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	Provider             string                 `json:"provider"`
	ProviderVersion      string                 `json:"providerVersion,omitempty"`
	ProviderEnvRef       string                 `json:"providerEnvRef,omitempty"`
	ScheduledScanEnabled bool                   `json:"scheduledScanEnabled"`
	ProviderConfig       map[string]interface{} `json:"providerConfig,omitempty"`
}

type CreateInsightsAccountRequest struct {
	Provider       string                 `json:"provider"`
	Environment    string                 `json:"environment"`
	Cron           string                 `json:"cron,omitempty"`
	ProviderConfig map[string]interface{} `json:"providerConfig,omitempty"`
}

type UpdateInsightsAccountRequest struct {
	Environment    string                 `json:"environment"`
	Cron           string                 `json:"cron,omitempty"`
	ProviderConfig map[string]interface{} `json:"providerConfig,omitempty"`
}

func (c *Client) CreateInsightsAccount(ctx context.Context, orgName, accountName string, req CreateInsightsAccountRequest) error {
	if len(orgName) == 0 {
		return errors.New("empty orgName")
	}

	if len(accountName) == 0 {
		return errors.New("empty accountName")
	}

	apiPath := path.Join("preview", "insights", orgName, "accounts", accountName)

	_, err := c.do(ctx, http.MethodPost, apiPath, req, nil)
	if err != nil {
		return fmt.Errorf("failed to create insights account: %w", err)
	}

	return nil
}

func (c *Client) GetInsightsAccount(ctx context.Context, orgName, accountName string) (*InsightsAccount, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	if len(accountName) == 0 {
		return nil, errors.New("empty accountName")
	}

	apiPath := path.Join("preview", "insights", orgName, "accounts", accountName)

	var account InsightsAccount
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &account)
	if err != nil {
		statusCode := GetErrorStatusCode(err)
		if statusCode == http.StatusNotFound {
			// Important: we return nil here to hint it was not found
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get insights account: %w", err)
	}

	return &account, nil
}

func (c *Client) UpdateInsightsAccount(ctx context.Context, orgName, accountName string, req UpdateInsightsAccountRequest) error {
	if len(orgName) == 0 {
		return errors.New("empty orgName")
	}

	if len(accountName) == 0 {
		return errors.New("empty accountName")
	}

	apiPath := path.Join("preview", "insights", orgName, "accounts", accountName)

	_, err := c.do(ctx, http.MethodPatch, apiPath, req, nil)
	if err != nil {
		return fmt.Errorf("failed to update insights account: %w", err)
	}

	return nil
}

func (c *Client) DeleteInsightsAccount(ctx context.Context, orgName, accountName string) error {
	if len(orgName) == 0 {
		return errors.New("empty orgName")
	}

	if len(accountName) == 0 {
		return errors.New("empty accountName")
	}

	apiPath := path.Join("preview", "insights", orgName, "accounts", accountName)

	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete insights account %q: %w", accountName, err)
	}

	return nil
}
