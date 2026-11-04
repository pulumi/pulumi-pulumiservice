// Copyright 2026, Pulumi Corporation.
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

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
)

type EntityAssignmentClient interface {
	ListEntityAssignments(
		ctx context.Context,
		orgName string,
		entity apitype.EntityInstance,
	) (*apitype.EntityAssignmentsResponse, error)
	UpdateEntityAssignment(
		ctx context.Context,
		orgName string,
		req apitype.UpdateEntityAssignmentRequest,
	) (*apitype.UpdateEntityAssignmentResponse, error)
}

// ListEntityAssignments returns the grants on one entity and the permission sets that can be granted on it.
// It returns (nil, nil) when the entity does not exist.
func (c *Client) ListEntityAssignments(
	ctx context.Context,
	orgName string,
	entity apitype.EntityInstance,
) (*apitype.EntityAssignmentsResponse, error) {
	if orgName == "" {
		return nil, errors.New("organization name must not be empty")
	}
	if entity.Context == nil || entity.ID == "" {
		return nil, errors.New("entity context and ID must not be empty")
	}

	resp, err := c.SDK.ListEntityAssignments(ctx, orgName, entity)
	if err != nil {
		if GetErrorStatusCode(err) == http.StatusNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to list entity assignments: %w", err)
	}
	return resp, nil
}

// UpdateEntityAssignment sets the permission set one user holds on one entity, replacing only that user's
// grant on that entity. An empty PermissionSetID revokes the grant, and the service answers 404 when there
// is none to revoke.
func (c *Client) UpdateEntityAssignment(
	ctx context.Context,
	orgName string,
	req apitype.UpdateEntityAssignmentRequest,
) (*apitype.UpdateEntityAssignmentResponse, error) {
	if orgName == "" {
		return nil, errors.New("organization name must not be empty")
	}
	if req.UserID == "" {
		return nil, errors.New("user must not be empty")
	}
	if req.Context == nil || req.ID == "" {
		return nil, errors.New("entity context and ID must not be empty")
	}

	resp, err := c.SDK.UpdateEntityAssignment(ctx, orgName, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update entity assignment: %w", err)
	}
	return resp, nil
}
