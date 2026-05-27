// Copyright 2016-2026, Pulumi Corporation.
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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
)

type PolicyPackClient interface {
	ListPolicyPacks(ctx context.Context, orgName string) ([]PolicyPackWithVersions, error)
	GetPolicyPack(ctx context.Context, orgName string, policyPackName string, version int) (*PolicyPackDetail, error)
	GetLatestPolicyPack(ctx context.Context, orgName string, policyPackName string) (*PolicyPackDetail, error)
	PublishPolicyPack(ctx context.Context, orgName string, req CreatePolicyPackRequest, archive io.Reader) (int, error)
	DeletePolicyPack(ctx context.Context, orgName, policyPackName string) error
	DeletePolicyPackVersion(ctx context.Context, orgName, policyPackName, versionTag string) error
}

type CreatePolicyPackRequest struct {
	Name        string           `json:"name"`
	DisplayName string           `json:"displayName,omitempty"`
	VersionTag  string           `json:"versionTag,omitempty"`
	Policies    []apitype.Policy `json:"policies"`
}

type CreatePolicyPackResponse struct {
	Version         int               `json:"version"`
	UploadURI       string            `json:"uploadURI"`
	RequiredHeaders map[string]string `json:"requiredHeaders,omitempty"`
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
	Policies    []apitype.Policy       `json:"policies,omitempty"`
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
	ctx context.Context,
	orgName string,
	policyPackName string,
	version int,
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
	ctx context.Context,
	orgName string,
	policyPackName string,
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

func (c *Client) PublishPolicyPack(
	ctx context.Context,
	orgName string,
	req CreatePolicyPackRequest,
	archive io.Reader,
) (version int, err error) {
	if orgName == "" {
		return 0, errors.New("empty orgName")
	}
	if req.Name == "" {
		return 0, errors.New("empty policy pack name")
	}
	if req.VersionTag == "" {
		return 0, errors.New("empty versionTag")
	}

	var resp CreatePolicyPackResponse
	apiPath := path.Join("orgs", orgName, "policypacks")
	if _, err = c.do(ctx, http.MethodPost, apiPath, req, &resp); err != nil {
		return 0, fmt.Errorf("publish policy pack metadata: %w", err)
	}

	// Metadata POST reserves versionTag in the cloud. If upload or complete
	// fails, the version is orphaned and the next retry hits 409 — so roll it
	// back on the way out. Detached context: caller's ctx may be the reason
	// we're failing.
	defer func() {
		if err == nil {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = c.DeletePolicyPackVersion(cleanupCtx, orgName, req.Name, req.VersionTag)
	}()

	body, readErr := io.ReadAll(archive)
	if readErr != nil {
		return 0, fmt.Errorf("read policy pack archive: %w", readErr)
	}
	if err = c.uploadToSignedURL(ctx, resp.UploadURI, resp.RequiredHeaders, body); err != nil {
		return 0, err
	}

	completePath := path.Join(
		"orgs", orgName, "policypacks", req.Name, "versions", req.VersionTag, "complete",
	)
	if _, err = c.do(ctx, http.MethodPost, completePath, nil, nil); err != nil {
		return 0, fmt.Errorf("signal publish completion: %w", err)
	}
	return resp.Version, nil
}

// uploadToSignedURL PUTs the archive to the pre-signed URL the Cloud returned.
// We retry on transient failures (network errors and 5xx) because signed-URL
// uploads run against blob storage, not our API, and flakes shouldn't burn
// the whole publish.
func (c *Client) uploadToSignedURL(
	ctx context.Context,
	uploadURI string,
	headers map[string]string,
	body []byte,
) error {
	const maxAttempts = 3
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, uploadURI, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("build upload request: %w", err)
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("upload policy pack archive: %w", err)
			if !sleepForRetry(ctx, attempt) {
				return lastErr
			}
			continue
		}
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()
		if resp.StatusCode < 300 {
			return nil
		}
		lastErr = fmt.Errorf("upload failed with status %d: %s",
			resp.StatusCode, strings.TrimSpace(string(respBody)))
		if resp.StatusCode < 500 || !sleepForRetry(ctx, attempt) {
			return lastErr
		}
	}
	return lastErr
}

// sleepForRetry waits before the next retry attempt and returns false if no
// further attempts should be made (context canceled, or attempts exhausted).
func sleepForRetry(ctx context.Context, attempt int) bool {
	if attempt >= 3 {
		return false
	}
	delay := time.Duration(attempt) * 500 * time.Millisecond
	select {
	case <-ctx.Done():
		return false
	case <-time.After(delay):
		return true
	}
}

func (c *Client) DeletePolicyPackVersion(ctx context.Context, orgName, policyPackName, versionTag string) error {
	if orgName == "" {
		return errors.New("empty orgName")
	}
	if policyPackName == "" {
		return errors.New("empty policy pack name")
	}
	if versionTag == "" {
		return errors.New("empty versionTag")
	}
	apiPath := path.Join("orgs", orgName, "policypacks", policyPackName, "versions", versionTag)
	if _, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil); err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("delete policy pack version: %w", err)
	}
	return nil
}

func (c *Client) DeletePolicyPack(ctx context.Context, orgName, policyPackName string) error {
	if orgName == "" {
		return errors.New("empty orgName")
	}
	if policyPackName == "" {
		return errors.New("empty policy pack name")
	}
	apiPath := path.Join("orgs", orgName, "policypacks", policyPackName)
	if _, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil); err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("delete policy pack: %w", err)
	}
	return nil
}
