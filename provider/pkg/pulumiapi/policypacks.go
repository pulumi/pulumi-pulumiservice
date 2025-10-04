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
	"strconv"
)

type PolicyPackClient interface {
	ListPolicyPacks(ctx context.Context, orgName string) ([]PolicyPackWithVersions, error)
	GetPolicyPack(ctx context.Context, orgName string, policyPackName string, version int) (*PolicyPackDetail, error)
	GetLatestPolicyPack(ctx context.Context, orgName string, policyPackName string) (*PolicyPackDetail, error)
}

type PolicyPackWithVersions struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	Versions    []int    `json:"versions"`
	VersionTags []string `json:"versionTags"`
}

type PolicyPackDetail struct {
	Name        string                 `json:"name"`
	DisplayName string                 `json:"displayName"`
	Version     int                    `json:"version"`
	VersionTag  string                 `json:"versionTag,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	Policies    []Policy               `json:"policies,omitempty"`
}

type PolicyComplianceFramework struct {
	Name          string `json:"name,omitempty"`
	Version       string `json:"version,omitempty"`
	Reference     string `json:"reference,omitempty"`
	Specification string `json:"specification,omitempty"`
}

type Policy struct {
	Name             string                     `json:"name"`
	DisplayName      string                     `json:"displayName,omitempty"`
	Description      string                     `json:"description,omitempty"`
	EnforcementLevel string                     `json:"enforcementLevel,omitempty"`
	Message          string                     `json:"message,omitempty"`
	ConfigSchema     map[string]interface{}     `json:"configSchema,omitempty"`
	Severity         string                     `json:"severity,omitempty"`
	Framework        *PolicyComplianceFramework `json:"framework,omitempty"`
	Tags             []string                   `json:"tags,omitempty"`
	RemediationSteps string                     `json:"remediationSteps,omitempty"`
	URL              string                     `json:"url,omitempty"`
}

type listPolicyPacksResponse struct {
	PolicyPacks []PolicyPackWithVersions `json:"policyPacks"`
}

func (c *Client) ListPolicyPacks(ctx context.Context, orgName string) ([]PolicyPackWithVersions, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	apiPath := path.Join("orgs", orgName, "policypacks")

	var response listPolicyPacksResponse
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list policy packs for %q: %w", orgName, err)
	}
	return response.PolicyPacks, nil
}

func (c *Client) GetPolicyPack(
	ctx context.Context, orgName string, policyPackName string, version int,
) (*PolicyPackDetail, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	if len(policyPackName) == 0 {
		return nil, errors.New("empty policyPackName")
	}

	apiPath := path.Join("orgs", orgName, "policypacks", policyPackName, "versions", strconv.Itoa(version))

	var policyPack PolicyPackDetail
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &policyPack)
	if err != nil {
		statusCode := GetErrorStatusCode(err)
		if statusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get policy pack: %w", err)
	}

	return &policyPack, nil
}

func (c *Client) GetLatestPolicyPack(
	ctx context.Context, orgName string, policyPackName string,
) (*PolicyPackDetail, error) {
	if len(orgName) == 0 {
		return nil, errors.New("empty orgName")
	}

	if len(policyPackName) == 0 {
		return nil, errors.New("empty policyPackName")
	}

	apiPath := path.Join("orgs", orgName, "policypacks", policyPackName, "latest")

	var policyPack PolicyPackDetail
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &policyPack)
	if err != nil {
		statusCode := GetErrorStatusCode(err)
		if statusCode == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest policy pack: %w", err)
	}

	return &policyPack, nil
}
