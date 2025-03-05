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

type WebhookClient interface {
	CreateWebhook(ctx context.Context, req WebhookRequest) (*Webhook, error)
	ListWebhooks(ctx context.Context, orgName string, projectName, stackName, environmentName *string) ([]Webhook, error)
	GetWebhook(ctx context.Context, orgName string, projectName, stackName, environmentName *string, webhookName string) (*Webhook, error)
	UpdateWebhook(ctx context.Context, req UpdateWebhookRequest) (*Webhook, error)
	DeleteWebhook(ctx context.Context, orgName string, projectName, stackName, environmentName *string, name string) error
}

type Webhook struct {
	Active           bool
	DisplayName      string
	PayloadUrl       string
	Secret           *string
	Name             string
	Format           string
	Filters          []string
	Groups           []string
	HasSecret        bool
	SecretCiphertext string
}

type WebhookRequest struct {
	OrganizationName string   `json:"organizationName"`
	ProjectName      *string  `json:"projectName,omitempty"`
	StackName        *string  `json:"stackName,omitempty"`
	EnvironmentName  *string  `json:"envName,omitempty"`
	DisplayName      string   `json:"displayName"`
	PayloadURL       string   `json:"payloadUrl"`
	Secret           *string  `json:"secret,omitempty"`
	Active           bool     `json:"active"`
	Format           *string  `json:"format,omitempty"`
	Filters          []string `json:"filters,omitempty"`
	Groups           []string `json:"groups,omitempty"`
}

type UpdateWebhookRequest struct {
	WebhookRequest
	Name string `json:"name"`
}

func (c *Client) CreateWebhook(ctx context.Context, req WebhookRequest) (*Webhook, error) {
	if len(req.OrganizationName) == 0 {
		return nil, errors.New("orgname must not be empty")
	}

	if len(req.DisplayName) == 0 {
		return nil, errors.New("displayname must not be empty")
	}

	if len(req.PayloadURL) == 0 {
		return nil, errors.New("payloadurl must not be empty")
	}

	apiPath := constructApiPath(req.OrganizationName, req.ProjectName, req.StackName, req.EnvironmentName)

	var webhook Webhook
	_, err := c.do(ctx, http.MethodPost, apiPath, req, &webhook)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	return &webhook, nil
}

func (c *Client) ListWebhooks(ctx context.Context, orgName string, projectName, stackName, environmentName *string) ([]Webhook, error) {
	if len(orgName) == 0 {
		return nil, errors.New("orgName must not be empty")
	}

	apiPath := constructApiPath(orgName, projectName, stackName, environmentName)

	var webhooks []Webhook
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &webhooks)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}

	return webhooks, nil
}

func (c *Client) GetWebhook(ctx context.Context,
	orgName string, projectName, stackName, environmentName *string, webhookName string) (*Webhook, error) {
	if len(orgName) == 0 {
		return nil, errors.New("orgname must not be empty")
	}

	if len(webhookName) == 0 {
		return nil, errors.New("webhookname must not be empty")
	}

	apiPath := constructApiPath(orgName, projectName, stackName, environmentName) + "/" + webhookName

	var webhook Webhook
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &webhook)
	if err != nil {
		statusCode := GetErrorStatusCode(err)
		if statusCode == http.StatusNotFound {
			// Important: we return nil here to hint it was not found
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}
	return &webhook, nil
}

func (c *Client) UpdateWebhook(ctx context.Context, req UpdateWebhookRequest) (*Webhook, error) {
	if len(req.Name) == 0 {
		return nil, errors.New("name must not be empty")
	}
	if len(req.OrganizationName) == 0 {
		return nil, errors.New("orgname must not be empty")
	}
	if len(req.DisplayName) == 0 {
		return nil, errors.New("displayname must not be empty")
	}
	if len(req.PayloadURL) == 0 {
		return nil, errors.New("payloadurl must not be empty")
	}

	apiPath := constructApiPath(req.OrganizationName, req.ProjectName, req.StackName, req.EnvironmentName) + "/" + req.Name

	var webhook Webhook
	_, err := c.do(ctx, http.MethodPatch, apiPath, req, &webhook)
	if err != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}
	return &webhook, nil
}

func (c *Client) DeleteWebhook(ctx context.Context, orgName string, projectName, stackName, environmentName *string, name string) error {
	if len(name) == 0 {
		return errors.New("name must not be empty")
	}
	if len(orgName) == 0 {
		return errors.New("orgname must not be empty")
	}

	apiPath := constructApiPath(orgName, projectName, stackName, environmentName) + "/" + name

	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}
	return nil
}

func constructApiPath(orgName string, projectName, stackName, environmentName *string) string {
	if projectName != nil && stackName != nil {
		return path.Join("stacks", orgName, *projectName, *stackName, "hooks")
	} else if projectName != nil && environmentName != nil {
		return path.Join("esc", "environments", orgName, *projectName, *environmentName, "hooks")
	}
	return path.Join("orgs", orgName, "hooks")
}
