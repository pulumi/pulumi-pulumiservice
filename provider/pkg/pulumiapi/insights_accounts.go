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
	ListInsightsAccounts(ctx context.Context, orgName string) ([]InsightsAccount, error)
	UpdateInsightsAccount(ctx context.Context, orgName, accountName string, req UpdateInsightsAccountRequest) error
	DeleteInsightsAccount(ctx context.Context, orgName, accountName string) error
	TriggerScan(ctx context.Context, orgName, accountName string) (*TriggerScanResponse, error)
	GetScanStatus(ctx context.Context, orgName, accountName string) (*ScanStatusResponse, error)
	GetInsightsAccountTags(ctx context.Context, orgName, accountName string) (map[string]string, error)
	SetInsightsAccountTags(ctx context.Context, orgName, accountName string, tags map[string]string) error
}

type InsightsAccount struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	Provider             string                 `json:"provider"`
	ProviderEnvRef       string                 `json:"providerEnvRef,omitempty"`
	ScheduledScanEnabled bool                   `json:"scheduledScanEnabled"`
	ProviderConfig       map[string]interface{} `json:"providerConfig,omitempty"`
}

type CreateInsightsAccountRequest struct {
	Provider       string                 `json:"provider"`
	Environment    string                 `json:"environment"`
	ScanSchedule   string                 `json:"scanSchedule,omitempty"`
	ProviderConfig map[string]interface{} `json:"providerConfig,omitempty"`
}

type UpdateInsightsAccountRequest struct {
	Environment    string                 `json:"environment"`
	ScanSchedule   string                 `json:"scanSchedule,omitempty"`
	ProviderConfig map[string]interface{} `json:"providerConfig,omitempty"`
}

// WorkflowRun represents a workflow execution (scan) with status and timing information
type WorkflowRun struct {
	ID            string `json:"id"`
	OrgID         string `json:"orgId"`
	UserID        string `json:"userId"`
	Status        string `json:"status"`        // "running", "failed", "succeeded"
	StartedAt     string `json:"startedAt"`     // RFC3339 timestamp
	FinishedAt    string `json:"finishedAt"`    // RFC3339 timestamp
	LastUpdatedAt string `json:"lastUpdatedAt"` // RFC3339 timestamp
}

// TriggerScanResponse is returned when triggering a scan (POST /scan)
type TriggerScanResponse struct {
	WorkflowRun
}

// ScanStatusResponse is returned when getting scan status (GET /scan)
type ScanStatusResponse struct {
	WorkflowRun
	NextScan      string `json:"nextScan,omitempty"`      // Next scheduled scan time
	ResourceCount int    `json:"resourceCount,omitempty"` // Number of resources discovered
}

// InsightsAccountTag represents a tag on an insights account
type InsightsAccountTag struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Created  string `json:"created,omitempty"`
	Modified string `json:"modified,omitempty"`
}

// GetInsightsAccountTagsResponse is returned when getting tags (GET /tags)
type GetInsightsAccountTagsResponse struct {
	Tags map[string]*InsightsAccountTag `json:"tags"`
}

// SetInsightsAccountTagsRequest is the request body for setting tags (PUT /tags)
type SetInsightsAccountTagsRequest struct {
	Tags map[string]string `json:"tags"`
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

type ListInsightsAccountsResponse struct {
	Accounts []InsightsAccount `json:"accounts"`
}

func (c *Client) ListInsightsAccounts(ctx context.Context, orgName string) ([]InsightsAccount, error) {
	if orgName == "" {
		return nil, errors.New("empty orgName")
	}

	apiPath := path.Join("preview", "insights", orgName, "accounts")

	var response ListInsightsAccountsResponse
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list insights accounts: %w", err)
	}

	return response.Accounts, nil
}

func (c *Client) UpdateInsightsAccount(ctx context.Context, orgName, accountName string, req UpdateInsightsAccountRequest) error {
	if orgName == "" {
		return errors.New("empty orgName")
	}

	if accountName == "" {
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
	if orgName == "" {
		return errors.New("empty orgName")
	}

	if accountName == "" {
		return errors.New("empty accountName")
	}

	apiPath := path.Join("preview", "insights", orgName, "accounts", accountName)

	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete insights account %q: %w", accountName, err)
	}

	return nil
}

// ScanOptions contains optional parameters for triggering a scan
type ScanOptions struct {
	ListConcurrency int    `json:"listConcurrency,omitempty"`
	ReadConcurrency int    `json:"readConcurrency,omitempty"`
	BatchSize       int    `json:"batchSize,omitempty"`
	ReadTimeout     string `json:"readTimeout,omitempty"`
}

// TriggerScan initiates an on-demand scan for the insights account
// If a scan is already running, it returns the existing scan details instead of triggering a new one
func (c *Client) TriggerScan(ctx context.Context, orgName, accountName string) (*TriggerScanResponse, error) {
	if orgName == "" {
		return nil, errors.New("empty orgName")
	}

	if accountName == "" {
		return nil, errors.New("empty accountName")
	}

	// First, check if a scan is already running
	currentStatus, err := c.GetScanStatus(ctx, orgName, accountName)
	if err == nil && currentStatus != nil && currentStatus.Status == "running" {
		// A scan is already running - return the existing scan details
		return &TriggerScanResponse{
			WorkflowRun: currentStatus.WorkflowRun,
		}, nil
	}

	// No scan running (or no scan exists yet, currentStatus == nil) - trigger a new scan
	apiPath := path.Join("preview", "insights", orgName, "accounts", accountName, "scan")

	// Send empty ScanOptions (all fields are optional)
	requestBody := ScanOptions{}

	// The API may return either:
	// - HTTP 200 with WorkflowRun JSON body
	// - HTTP 204 No Content (scan queued but no details available yet)
	// We need to handle both cases

	var response TriggerScanResponse
	httpResp, err := c.do(ctx, http.MethodPost, apiPath, requestBody, &response)

	// Handle HTTP 204 No Content - scan triggered but no workflow run details returned yet
	// The sendRequest will return an error when trying to unmarshal empty JSON for 204 responses
	if httpResp != nil && httpResp.StatusCode == http.StatusNoContent {
		// HTTP 204 - scan queued successfully but no details returned
		return &TriggerScanResponse{
			WorkflowRun: WorkflowRun{
				Status: "queued",
			},
		}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to trigger scan for insights account %q: %w", accountName, err)
	}

	return &response, nil
}

// GetScanStatus retrieves the current scan status of the insights account
func (c *Client) GetScanStatus(ctx context.Context, orgName, accountName string) (*ScanStatusResponse, error) {
	if orgName == "" {
		return nil, errors.New("empty orgName")
	}

	if accountName == "" {
		return nil, errors.New("empty accountName")
	}

	apiPath := path.Join("preview", "insights", orgName, "accounts", accountName, "scan")

	var status ScanStatusResponse
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &status)
	if err != nil {
		statusCode := GetErrorStatusCode(err)
		if statusCode == http.StatusNotFound {
			// No scan has been initiated yet - return nil to indicate no scan exists
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get scan status for insights account %q: %w", accountName, err)
	}

	return &status, nil
}

// GetInsightsAccountTags retrieves the tags for an insights account
func (c *Client) GetInsightsAccountTags(ctx context.Context, orgName, accountName string) (map[string]string, error) {
	if orgName == "" {
		return nil, errors.New("empty orgName")
	}

	if accountName == "" {
		return nil, errors.New("empty accountName")
	}

	apiPath := path.Join("preview", "insights", orgName, "accounts", accountName, "tags")

	var response GetInsightsAccountTagsResponse
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags for insights account %q: %w", accountName, err)
	}

	// Convert the response to a simple map[string]string
	tags := make(map[string]string)
	for key, tag := range response.Tags {
		if tag != nil {
			tags[key] = tag.Value
		}
	}

	return tags, nil
}

// SetInsightsAccountTags sets the tags for an insights account
func (c *Client) SetInsightsAccountTags(ctx context.Context, orgName, accountName string, tags map[string]string) error {
	if orgName == "" {
		return errors.New("empty orgName")
	}

	if accountName == "" {
		return errors.New("empty accountName")
	}

	apiPath := path.Join("preview", "insights", orgName, "accounts", accountName, "tags")

	req := SetInsightsAccountTagsRequest{
		Tags: tags,
	}

	_, err := c.do(ctx, http.MethodPut, apiPath, req, nil)
	if err != nil {
		return fmt.Errorf("failed to set tags for insights account %q: %w", accountName, err)
	}

	return nil
}
