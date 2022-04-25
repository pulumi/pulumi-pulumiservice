// Copyright 2016-2022, Pulumi Corporation.
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
	"errors"
	"fmt"
	"net/http"
	"path"
)

type Webhook struct {
	Active      bool
	DisplayName string
	PayloadUrl  string
	Secret      string
	Name        string
}

type createWebhookRequest struct {
	OrganizationName string `json:"organizationName"`
	DisplayName      string `json:"displayName"`
	PayloadURL       string `json:"payloadUrl"`
	Secret           string `json:"secret"`
	Active           bool   `json:"active"`
}

type updateWebhookRequest struct {
	Name             string `json:"name"`
	OrganizationName string `json:"organizationName"`
	DisplayName      string `json:"displayName"`
	PayloadURL       string `json:"payloadUrl"`
	Secret           string `json:"secret"`
	Active           bool   `json:"active"`
}

func (c *Client) CreateWebhook(ctx context.Context, orgName, displayName, payloadURL, secret string, active bool) (*Webhook, error) {

	if len(orgName) == 0 {
		return nil, errors.New("orgname must not be empty")
	}

	if len(displayName) == 0 {
		return nil, errors.New("displayname must not be empty")
	}

	if len(payloadURL) == 0 {
		return nil, errors.New("payloadurl must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "hooks")

	createWebhookReq := createWebhookRequest{
		OrganizationName: orgName,
		DisplayName:      displayName,
		PayloadURL:       payloadURL,
		Secret:           secret,
		Active:           active,
	}

	var webhook Webhook
	_, err := c.do(ctx, http.MethodPost, apiPath, createWebhookReq, &webhook)
	if err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	return &webhook, nil

}

func (c *Client) ListWebhooks(ctx context.Context, orgName string) ([]Webhook, error) {
	if len(orgName) == 0 {
		return nil, errors.New("orgName must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "hooks")

	var webhooks []Webhook
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &webhooks)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}

	return webhooks, nil
}

func (c *Client) GetWebhook(ctx context.Context, orgName, webhookName string) (*Webhook, error) {
	if len(orgName) == 0 {
		return nil, errors.New("orgname must not be empty")
	}

	if len(webhookName) == 0 {
		return nil, errors.New("webhookname must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "hooks", webhookName)

	var webhook Webhook
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &webhook)
	if err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}
	return &webhook, nil
}

func (c *Client) UpdateWebhook(ctx context.Context, name, orgName, displayName, payloadUrl, secret string, active bool) error {
	if len(name) == 0 {
		return errors.New("name must not be empty")
	}
	if len(orgName) == 0 {
		return errors.New("orgname must not be empty")
	}
	if len(displayName) == 0 {
		return errors.New("displayname must not be empty")
	}
	if len(payloadUrl) == 0 {
		return errors.New("payloadurl must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "hooks", name)

	updateWebhookReq := updateWebhookRequest{
		OrganizationName: orgName,
		Name:             name,
		DisplayName:      displayName,
		PayloadURL:       payloadUrl,
		Secret:           secret,
		Active:           active,
	}

	_, err := c.do(ctx, http.MethodPatch, apiPath, updateWebhookReq, nil)
	if err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}
	return nil
}

func (c *Client) DeleteWebhook(ctx context.Context, orgName, name string) error {
	if len(name) == 0 {
		return errors.New("name must not be empty")
	}
	if len(orgName) == 0 {
		return errors.New("orgname must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "hooks", name)

	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}
	return nil
}
