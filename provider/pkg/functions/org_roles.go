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

package functions

import (
	"context"
	"fmt"
	"sort"
	"strings"

	_ "embed"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

// The examples use the docs `{{% examples %}}` shortcodes, which SDK
// generation and the registry render per language.
var (
	//go:embed get_organization_role_examples.md
	getOrganizationRoleExamples string
	//go:embed get_organization_roles_examples.md
	getOrganizationRolesExamples string
)

// orgRoleUXPurpose is the uxPurpose that selects the roles shown on the
// Roles page of the Pulumi Cloud console: the built-in roles plus the org's
// custom roles.
const orgRoleUXPurpose = "role"

// GetOrganizationRolesFunction lists the roles defined in a Pulumi Cloud
// organization, built-in and custom.
type GetOrganizationRolesFunction struct{}

type GetOrganizationRolesInput struct {
	OrganizationName string `pulumi:"organizationName"`
}

// OrganizationRoleInfo carries the fields of a single Pulumi Cloud role.
type OrganizationRoleInfo struct {
	RoleId            string         `pulumi:"roleId"`
	Name              string         `pulumi:"name"`
	Description       *string        `pulumi:"description,optional"`
	ResourceType      string         `pulumi:"resourceType"`
	DefaultIdentifier *string        `pulumi:"defaultIdentifier,optional"`
	Details           map[string]any `pulumi:"details,optional"`
}

type GetOrganizationRolesOutput struct {
	Roles []OrganizationRoleInfo `pulumi:"roles"`
}

func (GetOrganizationRolesFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetOrganizationRolesFunction{},
		"Lists the roles in a Pulumi Cloud organization, including both built-in and custom roles, "+
			"sorted by name. To look up a single role by the name shown in the Pulumi Cloud console, "+
			"use `getOrganizationRole`.\n\n"+
			getOrganizationRolesExamples,
	)
	a.SetToken("index", "getOrganizationRoles")
}

func (i *GetOrganizationRolesInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The name of the Pulumi organization.")
}

func (r *OrganizationRoleInfo) Annotate(a infer.Annotator) {
	a.Describe(
		&r.RoleId,
		"The role's ID. Pass it to inputs that take a role ID, such as `TeamRoleAssignment.roleId` "+
			"and `OrganizationMember.roleId`.",
	)
	a.Describe(&r.Name, "The role's name, as shown in the Pulumi Cloud console.")
	a.Describe(&r.Description, "The role's description, if it has one.")
	a.Describe(
		&r.ResourceType,
		"The type of resource the role applies to, such as `global`, `stack` or `environment`.",
	)
	a.Describe(
		&r.DefaultIdentifier,
		"The stable identifier of a built-in role, such as `admin` or `member`. Absent for custom roles.",
	)
	a.Describe(
		&r.Details,
		"The role's permission descriptor tree, in the same shape as `OrganizationRole.permissions`.",
	)
}

func (GetOrganizationRolesFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetOrganizationRolesInput],
) (infer.FunctionResponse[GetOrganizationRolesOutput], error) {
	roles, err := listOrgRoles(ctx, req.Input.OrganizationName)
	if err != nil {
		return infer.FunctionResponse[GetOrganizationRolesOutput]{}, err
	}

	out := make([]OrganizationRoleInfo, 0, len(roles))
	for _, r := range roles {
		info, err := roleInfoFrom(r)
		if err != nil {
			return infer.FunctionResponse[GetOrganizationRolesOutput]{}, err
		}
		out = append(out, info)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].RoleId < out[j].RoleId
	})

	return infer.FunctionResponse[GetOrganizationRolesOutput]{
		Output: GetOrganizationRolesOutput{Roles: out},
	}, nil
}

// GetOrganizationRoleFunction looks up a single Pulumi Cloud role by the
// name shown in the console. Returns an error when no role, or more than
// one role, has that name.
type GetOrganizationRoleFunction struct{}

type GetOrganizationRoleInput struct {
	OrganizationName string `pulumi:"organizationName"`
	Name             string `pulumi:"name"`
}

type GetOrganizationRoleOutput struct {
	OrganizationRoleInfo
}

func (GetOrganizationRoleFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetOrganizationRoleFunction{},
		"Looks up a role in a Pulumi Cloud organization by the name shown in the Pulumi Cloud console. "+
			"Finds both built-in and custom roles. Use the returned `roleId` for inputs that take a role ID, "+
			"such as `TeamRoleAssignment.roleId` and `OrganizationMember.roleId`, instead of hardcoding the ID. "+
			"Returns an error when no role has the given name, or when more than one does.\n\n"+
			getOrganizationRoleExamples,
	)
	a.SetToken("index", "getOrganizationRole")
}

func (i *GetOrganizationRoleInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The name of the Pulumi organization.")
	a.Describe(
		&i.Name,
		"The role's name, exactly as shown in the Pulumi Cloud console. The match is case-sensitive.",
	)
}

func (GetOrganizationRoleFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetOrganizationRoleInput],
) (infer.FunctionResponse[GetOrganizationRoleOutput], error) {
	orgName := req.Input.OrganizationName
	name := strings.TrimSpace(req.Input.Name)
	if name == "" {
		return infer.FunctionResponse[GetOrganizationRoleOutput]{}, fmt.Errorf("`name` must not be empty")
	}

	roles, err := listOrgRoles(ctx, orgName)
	if err != nil {
		return infer.FunctionResponse[GetOrganizationRoleOutput]{}, err
	}

	var matches []apitype.PermissionDescriptorRecord
	for _, r := range roles {
		if r.Name == name {
			matches = append(matches, r)
		}
	}
	switch len(matches) {
	case 0:
		return infer.FunctionResponse[GetOrganizationRoleOutput]{}, fmt.Errorf(
			"organization %q has no role named %q",
			orgName,
			name,
		)
	case 1:
	default:
		ids := make([]string, 0, len(matches))
		for _, m := range matches {
			ids = append(ids, m.ID)
		}
		return infer.FunctionResponse[GetOrganizationRoleOutput]{}, fmt.Errorf(
			"organization %q has %d roles named %q (IDs: %s); use the role ID directly instead",
			orgName,
			len(matches),
			name,
			strings.Join(ids, ", "),
		)
	}

	info, err := roleInfoFrom(matches[0])
	if err != nil {
		return infer.FunctionResponse[GetOrganizationRoleOutput]{}, err
	}
	return infer.FunctionResponse[GetOrganizationRoleOutput]{
		Output: GetOrganizationRoleOutput{OrganizationRoleInfo: info},
	}, nil
}

func listOrgRoles(ctx context.Context, orgName string) ([]apitype.PermissionDescriptorRecord, error) {
	roles, err := config.GetClient(ctx).ListOrgRoles(ctx, orgName, orgRoleUXPurpose)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles for organization %q: %w", orgName, err)
	}
	return roles, nil
}

func roleInfoFrom(r apitype.PermissionDescriptorRecord) (OrganizationRoleInfo, error) {
	info := OrganizationRoleInfo{
		RoleId:       r.ID,
		Name:         r.Name,
		ResourceType: r.ResourceType,
	}
	if r.Description != "" {
		description := r.Description
		info.Description = &description
	}
	if r.DefaultIdentifier != "" {
		defaultIdentifier := r.DefaultIdentifier
		info.DefaultIdentifier = &defaultIdentifier
	}
	if r.Details != nil {
		details, err := descriptorToSDKMap(r.Details)
		if err != nil {
			return OrganizationRoleInfo{}, fmt.Errorf("converting details of role %q: %w", r.Name, err)
		}
		info.Details = details
	}
	return info, nil
}
