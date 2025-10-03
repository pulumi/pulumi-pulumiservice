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

type ServiceClient interface {
	CreateService(ctx context.Context, req CreateServiceRequest) (*Service, error)
	GetService(ctx context.Context, orgName, ownerType, ownerName, serviceName string) (*Service, error)
	UpdateService(ctx context.Context, req UpdateServiceRequest) (*Service, error)
	DeleteService(ctx context.Context, orgName, ownerType, ownerName, serviceName string, force bool) error
	AddServiceItem(ctx context.Context, req AddServiceItemRequest) error
	RemoveServiceItem(ctx context.Context, req RemoveServiceItemRequest) error
}

type Service struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Properties  map[string]string      `json:"properties,omitempty"`
	Items       []ServiceItem          `json:"items,omitempty"`
	OwnerType   string                 `json:"ownerType"`
	OwnerName   string                 `json:"ownerName"`
}

type ServiceItem struct {
	ItemType string `json:"itemType"`
	Name     string `json:"name"`
}

type CreateServiceRequest struct {
	OrganizationName string                 `json:"-"`
	OwnerType        string                 `json:"ownerType"`
	OwnerName        string                 `json:"ownerName"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description,omitempty"`
	Properties       map[string]string      `json:"properties,omitempty"`
	Items            []ServiceItem          `json:"items,omitempty"`
}

type UpdateServiceRequest struct {
	OrganizationName string                 `json:"-"`
	OwnerType        string                 `json:"-"`
	OwnerName        string                 `json:"-"`
	ServiceName      string                 `json:"-"`
	Name             *string                `json:"name,omitempty"`
	Description      *string                `json:"description,omitempty"`
	Properties       map[string]string      `json:"properties,omitempty"`
}

type AddServiceItemRequest struct {
	OrganizationName string      `json:"-"`
	OwnerType        string      `json:"-"`
	OwnerName        string      `json:"-"`
	ServiceName      string      `json:"-"`
	Item             ServiceItem `json:"item"`
}

type RemoveServiceItemRequest struct {
	OrganizationName string `json:"-"`
	OwnerType        string `json:"-"`
	OwnerName        string `json:"-"`
	ServiceName      string `json:"-"`
	ItemType         string `json:"-"`
	ItemName         string `json:"-"`
}

func (c *Client) CreateService(ctx context.Context, req CreateServiceRequest) (*Service, error) {
	if len(req.OrganizationName) == 0 {
		return nil, errors.New("organizationName must not be empty")
	}
	if len(req.OwnerType) == 0 {
		return nil, errors.New("ownerType must not be empty")
	}
	if len(req.OwnerName) == 0 {
		return nil, errors.New("ownerName must not be empty")
	}
	if len(req.Name) == 0 {
		return nil, errors.New("name must not be empty")
	}

	apiPath := path.Join("orgs", req.OrganizationName, "services")

	var service Service
	_, err := c.do(ctx, http.MethodPost, apiPath, req, &service)
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}
	return &service, nil
}

func (c *Client) GetService(ctx context.Context, orgName, ownerType, ownerName, serviceName string) (*Service, error) {
	if len(orgName) == 0 {
		return nil, errors.New("organizationName must not be empty")
	}
	if len(ownerType) == 0 {
		return nil, errors.New("ownerType must not be empty")
	}
	if len(ownerName) == 0 {
		return nil, errors.New("ownerName must not be empty")
	}
	if len(serviceName) == 0 {
		return nil, errors.New("serviceName must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "services", ownerType, ownerName, serviceName)

	var service Service
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &service)
	if err != nil {
		statusCode := GetErrorStatusCode(err)
		if statusCode == http.StatusNotFound {
			// Important: we return nil here to hint it was not found
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get service: %w", err)
	}
	return &service, nil
}

func (c *Client) UpdateService(ctx context.Context, req UpdateServiceRequest) (*Service, error) {
	if len(req.OrganizationName) == 0 {
		return nil, errors.New("organizationName must not be empty")
	}
	if len(req.OwnerType) == 0 {
		return nil, errors.New("ownerType must not be empty")
	}
	if len(req.OwnerName) == 0 {
		return nil, errors.New("ownerName must not be empty")
	}
	if len(req.ServiceName) == 0 {
		return nil, errors.New("serviceName must not be empty")
	}

	apiPath := path.Join("orgs", req.OrganizationName, "services", req.OwnerType, req.OwnerName, req.ServiceName)

	var service Service
	_, err := c.do(ctx, http.MethodPatch, apiPath, req, &service)
	if err != nil {
		return nil, fmt.Errorf("failed to update service: %w", err)
	}
	return &service, nil
}

func (c *Client) DeleteService(ctx context.Context, orgName, ownerType, ownerName, serviceName string, force bool) error {
	if len(orgName) == 0 {
		return errors.New("organizationName must not be empty")
	}
	if len(ownerType) == 0 {
		return errors.New("ownerType must not be empty")
	}
	if len(ownerName) == 0 {
		return errors.New("ownerName must not be empty")
	}
	if len(serviceName) == 0 {
		return errors.New("serviceName must not be empty")
	}

	apiPath := path.Join("orgs", orgName, "services", ownerType, ownerName, serviceName)
	if force {
		apiPath += "?force=true"
	}

	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}
	return nil
}

func (c *Client) AddServiceItem(ctx context.Context, req AddServiceItemRequest) error {
	if len(req.OrganizationName) == 0 {
		return errors.New("organizationName must not be empty")
	}
	if len(req.OwnerType) == 0 {
		return errors.New("ownerType must not be empty")
	}
	if len(req.OwnerName) == 0 {
		return errors.New("ownerName must not be empty")
	}
	if len(req.ServiceName) == 0 {
		return errors.New("serviceName must not be empty")
	}

	apiPath := path.Join("orgs", req.OrganizationName, "services", req.OwnerType, req.OwnerName, req.ServiceName, "items")

	_, err := c.do(ctx, http.MethodPost, apiPath, req.Item, nil)
	if err != nil {
		return fmt.Errorf("failed to add service item: %w", err)
	}
	return nil
}

func (c *Client) RemoveServiceItem(ctx context.Context, req RemoveServiceItemRequest) error {
	if len(req.OrganizationName) == 0 {
		return errors.New("organizationName must not be empty")
	}
	if len(req.OwnerType) == 0 {
		return errors.New("ownerType must not be empty")
	}
	if len(req.OwnerName) == 0 {
		return errors.New("ownerName must not be empty")
	}
	if len(req.ServiceName) == 0 {
		return errors.New("serviceName must not be empty")
	}
	if len(req.ItemType) == 0 {
		return errors.New("itemType must not be empty")
	}
	if len(req.ItemName) == 0 {
		return errors.New("itemName must not be empty")
	}

	apiPath := path.Join("orgs", req.OrganizationName, "services", req.OwnerType, req.OwnerName, req.ServiceName, "items", req.ItemType, req.ItemName)

	_, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to remove service item: %w", err)
	}
	return nil
}
