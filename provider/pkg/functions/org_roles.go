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
	"slices"
	"sort"
	"strings"

	_ "embed" // for the //go:embed example files below

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

// The example files are generated from the Pulumi YAML in examples/; edit
// those and run `make generate_examples`.
//
//go:generate go run ../../tools/gen-examples examples .
var (
	//go:embed get_organization_role_examples.md
	getOrganizationRoleExamples string
	//go:embed get_organization_roles_examples.md
	getOrganizationRolesExamples string
)

// GetOrganizationRolesFunction lists the roles defined in a Pulumi Cloud
// organization, built-in and custom.
type GetOrganizationRolesFunction struct{}

type GetOrganizationRolesInput struct {
	OrganizationName string `pulumi:"organizationName"`
}

// OrganizationRoleInfo carries the fields of a single Pulumi Cloud role.
type OrganizationRoleInfo struct {
	RoleId      string  `pulumi:"roleId"`
	Name        string  `pulumi:"name"`
	Description *string `pulumi:"description,optional"`
	BuiltIn     bool    `pulumi:"builtIn"`
}

type GetOrganizationRolesOutput struct {
	Roles []OrganizationRoleInfo `pulumi:"roles"`
}

func (o *GetOrganizationRolesOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.Roles, "The organization's roles, built-in and custom, sorted by name.")
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
		&r.BuiltIn,
		"Whether this is one of the roles every organization has (Admin, Member and Billing Manager), "+
			"rather than a custom role.",
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
		out = append(out, roleInfoFrom(r))
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
// name shown in the console. Returns an error when no role has that name.
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
			"Returns an error when no role has the given name.\n\n"+
			getOrganizationRoleExamples,
	)
	a.SetToken("index", "getOrganizationRole")
}

func (i *GetOrganizationRoleInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The name of the Pulumi organization.")
	a.Describe(
		&i.Name,
		"The role's name, as shown in the Pulumi Cloud console. The match is case-insensitive, "+
			"as role names are unique within an organization regardless of case.",
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

	// Pulumi Cloud keeps role names unique per organization, ignoring case.
	i := slices.IndexFunc(roles, func(r apitype.PermissionDescriptorRecord) bool {
		return strings.EqualFold(r.Name, name)
	})
	if i < 0 {
		return infer.FunctionResponse[GetOrganizationRoleOutput]{}, fmt.Errorf(
			"organization %q has no role named %q",
			orgName,
			name,
		)
	}

	return infer.FunctionResponse[GetOrganizationRoleOutput]{
		Output: GetOrganizationRoleOutput{OrganizationRoleInfo: roleInfoFrom(roles[i])},
	}, nil
}

// listOrgRoles returns the organization's roles, built-in and custom. Pulumi
// Cloud stores roles alongside permission sets and policies; only roles are
// returned, even if the API's uxPurpose filter lets another kind through.
func listOrgRoles(ctx context.Context, orgName string) ([]apitype.PermissionDescriptorRecord, error) {
	records, err := config.GetClient(ctx).ListOrgRoles(ctx, orgName, apitype.PermissionDescriptorUXPurposeRole)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles for organization %q: %w", orgName, err)
	}
	return slices.DeleteFunc(records, func(r apitype.PermissionDescriptorRecord) bool {
		return r.UxPurpose != apitype.PermissionDescriptorUXPurposeRole
	}), nil
}

// roleInfoFrom keeps only the fields that describe a role. The record's
// resourceType and permission descriptor tree are left out: they describe
// permission sets and policies, which these functions don't expose.
func roleInfoFrom(r apitype.PermissionDescriptorRecord) OrganizationRoleInfo {
	info := OrganizationRoleInfo{
		RoleId: r.ID,
		Name:   r.Name,
		// Only built-in roles carry a default identifier (admin, member, ...).
		BuiltIn: r.DefaultIdentifier != "",
	}
	if r.Description != "" {
		description := r.Description
		info.Description = &description
	}
	return info
}
