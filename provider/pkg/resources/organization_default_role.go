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

package resources

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

type OrganizationDefaultRole struct{}

var (
	_ infer.CustomCreate[OrganizationDefaultRoleInput, OrganizationDefaultRoleState] = &OrganizationDefaultRole{}
	_ infer.CustomCheck[OrganizationDefaultRoleInput]                                = &OrganizationDefaultRole{}
	_ infer.CustomDelete[OrganizationDefaultRoleState]                               = &OrganizationDefaultRole{}
	_ infer.CustomRead[OrganizationDefaultRoleInput, OrganizationDefaultRoleState]   = &OrganizationDefaultRole{}
	_ infer.CustomUpdate[OrganizationDefaultRoleInput, OrganizationDefaultRoleState] = &OrganizationDefaultRole{}
)

var settableDefaultRoles = []string{"reader", "builder", "maintainer"}

func (*OrganizationDefaultRole) Annotate(a infer.Annotator) {
	a.Describe(
		&OrganizationDefaultRole{},
		"Sets the default role of a Pulumi Cloud organization: the role that members without an explicit role "+
			"follow, and that new members receive. Set either `role`, for a built-in role, or `roleId`, for a "+
			"custom role. An organization always has a default role and Pulumi Cloud has no call to clear it, "+
			"so destroying this resource leaves the current default role in place. At most one "+
			"`OrganizationDefaultRole` should manage an organization."+
			util.ImportDocs("pulumiservice:index:OrganizationDefaultRole", "{organization}", "my-org"),
	)
}

type OrganizationDefaultRoleInput struct {
	OrganizationName string  `pulumi:"organizationName" provider:"replaceOnChanges"`
	Role             *string `pulumi:"role,optional"`
	RoleId           *string `pulumi:"roleId,optional"`
}

func (i *OrganizationDefaultRoleInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The Pulumi Cloud organization name.")
	a.Describe(
		&i.Role,
		"The built-in role to use as the default. One of `reader`, `builder` or `maintainer`. "+
			"Exactly one of `role` and `roleId` must be set.",
	)
	a.Describe(
		&i.RoleId,
		"The ID of a custom role to use as the default. Exactly one of `role` and `roleId` must be set.",
	)
}

type OrganizationDefaultRoleState struct {
	OrganizationDefaultRoleInput
	RoleName string `pulumi:"roleName"`
}

func (s *OrganizationDefaultRoleState) Annotate(a infer.Annotator) {
	a.Describe(&s.RoleName, "The name of the current default role.")
}

func (*OrganizationDefaultRole) Check(
	ctx context.Context,
	req infer.CheckRequest,
) (infer.CheckResponse[OrganizationDefaultRoleInput], error) {
	in, failures, err := infer.DefaultCheck[OrganizationDefaultRoleInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[OrganizationDefaultRoleInput]{}, err
	}
	hasRole := in.Role != nil && *in.Role != ""
	hasRoleID := in.RoleId != nil && *in.RoleId != ""
	switch {
	case hasRole == hasRoleID:
		failures = append(failures, p.CheckFailure{
			Property: "role",
			Reason:   "exactly one of role and roleId must be set",
		})
	case hasRole && !slices.Contains(settableDefaultRoles, *in.Role):
		failures = append(failures, p.CheckFailure{
			Property: "role",
			Reason:   fmt.Sprintf("role must be one of %v, got %q", settableDefaultRoles, *in.Role),
		})
	}
	return infer.CheckResponse[OrganizationDefaultRoleInput]{Inputs: in, Failures: failures}, nil
}

func (*OrganizationDefaultRole) Create(
	ctx context.Context,
	req infer.CreateRequest[OrganizationDefaultRoleInput],
) (infer.CreateResponse[OrganizationDefaultRoleState], error) {
	id := req.Inputs.OrganizationName
	if req.DryRun {
		return infer.CreateResponse[OrganizationDefaultRoleState]{
			ID:     id,
			Output: OrganizationDefaultRoleState{OrganizationDefaultRoleInput: req.Inputs},
		}, nil
	}

	state, err := applyDefaultRole(ctx, req.Inputs)
	if err != nil {
		return infer.CreateResponse[OrganizationDefaultRoleState]{}, err
	}
	return infer.CreateResponse[OrganizationDefaultRoleState]{ID: id, Output: state}, nil
}

func (*OrganizationDefaultRole) Update(
	ctx context.Context,
	req infer.UpdateRequest[OrganizationDefaultRoleInput, OrganizationDefaultRoleState],
) (infer.UpdateResponse[OrganizationDefaultRoleState], error) {
	if req.DryRun {
		return infer.UpdateResponse[OrganizationDefaultRoleState]{
			Output: OrganizationDefaultRoleState{
				OrganizationDefaultRoleInput: req.Inputs,
				RoleName:                     req.State.RoleName,
			},
		}, nil
	}

	state, err := applyDefaultRole(ctx, req.Inputs)
	if err != nil {
		return infer.UpdateResponse[OrganizationDefaultRoleState]{}, err
	}
	return infer.UpdateResponse[OrganizationDefaultRoleState]{Output: state}, nil
}

func applyDefaultRole(
	ctx context.Context,
	in OrganizationDefaultRoleInput,
) (OrganizationDefaultRoleState, error) {
	client := config.GetClient(ctx)
	roles, err := client.ListOrgRoles(ctx, in.OrganizationName, apitype.PermissionDescriptorUXPurposeRole)
	if err != nil {
		return OrganizationDefaultRoleState{}, err
	}
	target, err := findDefaultRoleTarget(roles, in)
	if err != nil {
		return OrganizationDefaultRoleState{}, fmt.Errorf("organization %q: %w", in.OrganizationName, err)
	}
	if err := client.SetOrgDefaultRole(ctx, in.OrganizationName, target.ID); err != nil {
		return OrganizationDefaultRoleState{}, err
	}
	return OrganizationDefaultRoleState{OrganizationDefaultRoleInput: in, RoleName: target.Name}, nil
}

func findDefaultRoleTarget(
	roles []apitype.PermissionDescriptorRecord,
	in OrganizationDefaultRoleInput,
) (apitype.PermissionDescriptorRecord, error) {
	if roleID := util.OrZero(in.RoleId); roleID != "" {
		for _, r := range roles {
			if r.ID == roleID {
				return r, nil
			}
		}
		return apitype.PermissionDescriptorRecord{}, fmt.Errorf("role ID %q not found", roleID)
	}
	role := util.OrZero(in.Role)
	for _, r := range roles {
		if r.DefaultIdentifier == role {
			return r, nil
		}
	}
	return apitype.PermissionDescriptorRecord{}, fmt.Errorf("built-in role %q not found", role)
}

// Delete only forgets the resource: the service cannot clear the default role, and picking a replacement
// would change access for members who follow it.
func (*OrganizationDefaultRole) Delete(
	context.Context,
	infer.DeleteRequest[OrganizationDefaultRoleState],
) (infer.DeleteResponse, error) {
	return infer.DeleteResponse{}, nil
}

func (*OrganizationDefaultRole) Read(
	ctx context.Context,
	req infer.ReadRequest[OrganizationDefaultRoleInput, OrganizationDefaultRoleState],
) (infer.ReadResponse[OrganizationDefaultRoleInput, OrganizationDefaultRoleState], error) {
	type response = infer.ReadResponse[OrganizationDefaultRoleInput, OrganizationDefaultRoleState]
	if req.ID == "" {
		return response{}, errors.New("organization name must not be empty")
	}

	roles, err := config.GetClient(ctx).ListOrgRoles(ctx, req.ID, apitype.PermissionDescriptorUXPurposeRole)
	if err != nil {
		return response{}, err
	}
	idx := slices.IndexFunc(roles, func(r apitype.PermissionDescriptorRecord) bool { return r.IsOrgDefault })
	if idx < 0 {
		return response{}, nil
	}
	current := roles[idx]

	in := OrganizationDefaultRoleInput{OrganizationName: req.ID}
	declaredByID := req.Inputs.RoleId != nil && *req.Inputs.RoleId == current.ID
	if !declaredByID && slices.Contains(settableDefaultRoles, current.DefaultIdentifier) {
		role := current.DefaultIdentifier
		in.Role = &role
	} else {
		id := current.ID
		in.RoleId = &id
	}
	return response{
		ID:     req.ID,
		Inputs: in,
		State:  OrganizationDefaultRoleState{OrganizationDefaultRoleInput: in, RoleName: current.Name},
	}, nil
}
