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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
)

// PermissionSetClient manages permission sets (uxPurpose=set descriptors).
//
// It speaks raw JSON rather than going through the cloud SDK's typed
// descriptors: the SDK's RbacPermission unmarshaller blanks any scope it does
// not know, which would silently drop scopes Pulumi Cloud added after the
// pinned SDK version.
type PermissionSetClient interface {
	CreatePermissionSet(ctx context.Context, orgName string, req PermissionSetRequest) (*PermissionSet, error)
	GetPermissionSet(ctx context.Context, orgName, id string) (*PermissionSet, error)
	UpdatePermissionSet(ctx context.Context, orgName, id string, req PermissionSetRequest) (*PermissionSet, error)
	ListPermissionSets(ctx context.Context, orgName string) ([]PermissionSet, error)
}

// PermissionSetRequest is the user-settable part of a permission set.
// ResourceType is ignored on update; the service cannot change it.
type PermissionSetRequest struct {
	Name         string
	Description  string
	ResourceType string
	Permissions  []string
}

// PermissionSet is a permission set as stored by Pulumi Cloud.
type PermissionSet struct {
	ID                string
	Name              string
	Description       string
	ResourceType      string
	UxPurpose         string
	DefaultIdentifier string
	Version           int
	// DetailsType is the descriptor type of the set's details. Sets created
	// in the console are always PermissionDescriptorAllow; Permissions is only
	// meaningful in that case.
	DetailsType string
	Permissions []string
}

type permissionSetDetails struct {
	Type        string   `json:"__type"`
	Permissions []string `json:"permissions,omitempty"`
}

type permissionSetBody struct {
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	ResourceType string                `json:"resourceType,omitempty"`
	UxPurpose    string                `json:"uxPurpose,omitempty"`
	Details      *permissionSetDetails `json:"details"`
}

type permissionSetRecord struct {
	ID                string               `json:"id"`
	Name              string               `json:"name"`
	Description       string               `json:"description"`
	ResourceType      string               `json:"resourceType"`
	UxPurpose         string               `json:"uxPurpose"`
	DefaultIdentifier string               `json:"defaultIdentifier"`
	Version           int                  `json:"version"`
	Details           permissionSetDetails `json:"details"`
}

func (r permissionSetRecord) toPermissionSet() *PermissionSet {
	return &PermissionSet{
		ID:                r.ID,
		Name:              r.Name,
		Description:       r.Description,
		ResourceType:      r.ResourceType,
		UxPurpose:         r.UxPurpose,
		DefaultIdentifier: r.DefaultIdentifier,
		Version:           r.Version,
		DetailsType:       r.Details.Type,
		Permissions:       r.Details.Permissions,
	}
}

const (
	permissionDescriptorAllow = "PermissionDescriptorAllow"
	uxPurposeSet              = "set"
)

func (c *Client) CreatePermissionSet(
	ctx context.Context, orgName string, req PermissionSetRequest,
) (*PermissionSet, error) {
	if len(orgName) == 0 {
		return nil, errors.New("organization name must not be empty")
	}
	if req.Name == "" {
		return nil, errors.New("permission set name must not be empty")
	}
	body := permissionSetBody{
		Name:         req.Name,
		Description:  req.Description,
		ResourceType: req.ResourceType,
		UxPurpose:    uxPurposeSet,
		Details:      &permissionSetDetails{Type: permissionDescriptorAllow, Permissions: req.Permissions},
	}
	var rec permissionSetRecord
	if _, err := c.do(ctx, http.MethodPost, path.Join("orgs", orgName, "roles"), body, &rec); err != nil {
		return nil, fmt.Errorf("failed to create permission set: %w", err)
	}
	return rec.toPermissionSet(), nil
}

// GetPermissionSet returns (nil, nil) if the descriptor does not exist.
func (c *Client) GetPermissionSet(ctx context.Context, orgName, id string) (*PermissionSet, error) {
	if len(orgName) == 0 {
		return nil, errors.New("organization name must not be empty")
	}
	if len(id) == 0 {
		return nil, errors.New("permission set id must not be empty")
	}
	var rec permissionSetRecord
	if _, err := c.do(ctx, http.MethodGet, path.Join("orgs", orgName, "roles", id), nil, &rec); err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get permission set: %w", err)
	}
	return rec.toPermissionSet(), nil
}

func (c *Client) UpdatePermissionSet(
	ctx context.Context, orgName, id string, req PermissionSetRequest,
) (*PermissionSet, error) {
	if len(orgName) == 0 {
		return nil, errors.New("organization name must not be empty")
	}
	if len(id) == 0 {
		return nil, errors.New("permission set id must not be empty")
	}
	body := permissionSetBody{
		Name:        req.Name,
		Description: req.Description,
		Details:     &permissionSetDetails{Type: permissionDescriptorAllow, Permissions: req.Permissions},
	}
	var rec permissionSetRecord
	if _, err := c.do(ctx, http.MethodPatch, path.Join("orgs", orgName, "roles", id), body, &rec); err != nil {
		return nil, fmt.Errorf("failed to update permission set: %w", err)
	}
	return rec.toPermissionSet(), nil
}

// ListPermissionSets returns the organization's built-in and custom
// permission sets.
func (c *Client) ListPermissionSets(ctx context.Context, orgName string) ([]PermissionSet, error) {
	if len(orgName) == 0 {
		return nil, errors.New("organization name must not be empty")
	}
	var resp struct {
		Roles []permissionSetRecord `json:"roles"`
	}
	query := url.Values{"uxPurpose": []string{uxPurposeSet}}
	if _, err := c.doWithQuery(ctx, http.MethodGet, path.Join("orgs", orgName, "roles"), query, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to list permission sets: %w", err)
	}
	out := make([]PermissionSet, 0, len(resp.Roles))
	for _, r := range resp.Roles {
		out = append(out, *r.toPermissionSet())
	}
	return out, nil
}
