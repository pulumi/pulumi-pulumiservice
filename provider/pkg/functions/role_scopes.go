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

package functions

import (
	"context"
	"fmt"
	"sort"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

// GetOrganizationRoleScopesFunction exposes the permission-scope catalogue
// that the Pulumi Cloud console uses when editing custom roles. Customers
// use this to discover valid scope names (e.g. "stack:read") before putting
// them into `OrganizationRole.permissions`.
type GetOrganizationRoleScopesFunction struct{}

type GetOrganizationRoleScopesInput struct {
	OrganizationName string `pulumi:"organizationName"`
}

type RoleScopeInfo struct {
	Name         string `pulumi:"name"`
	Description  string `pulumi:"description"`
	GroupName    string `pulumi:"groupName"`
	ResourceType string `pulumi:"resourceType"`
}

type GetOrganizationRoleScopesOutput struct {
	Scopes []RoleScopeInfo `pulumi:"scopes"`
}

func (GetOrganizationRoleScopesFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetOrganizationRoleScopesFunction{},
		"Lists the permission scopes available for custom roles in an organization. Use this to discover "+
			"valid scope names before setting `OrganizationRole.permissions`. The catalogue is flattened "+
			"into a single list with resource type and group context, sorted deterministically.",
	)
	a.SetToken("index", "getOrganizationRoleScopes")
}

func (i *GetOrganizationRoleScopesInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The Pulumi Cloud organization name.")
}

func (s *RoleScopeInfo) Annotate(a infer.Annotator) {
	a.Describe(&s.Name, "The scope name (e.g. `stack:read`).")
	a.Describe(&s.Description, "Human-readable description of what the scope grants.")
	a.Describe(&s.GroupName, "The scope group label as shown in the Pulumi Cloud console (e.g. `Stacks`).")
	a.Describe(&s.ResourceType, "The resource-type bucket the scope belongs to (e.g. `stack`, `team`).")
}

func (GetOrganizationRoleScopesFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetOrganizationRoleScopesInput],
) (infer.FunctionResponse[GetOrganizationRoleScopesOutput], error) {
	client := config.GetClient(ctx)

	scopes, err := client.ListAvailableRoleScopes(ctx, req.Input.OrganizationName)
	if err != nil {
		return infer.FunctionResponse[GetOrganizationRoleScopesOutput]{}, fmt.Errorf(
			"failed to list available role scopes: %w",
			err,
		)
	}

	flat := make([]RoleScopeInfo, 0)
	for resourceType, groups := range scopes {
		for _, g := range groups {
			for _, s := range g.Scopes {
				flat = append(flat, RoleScopeInfo{
					Name:         s.Name,
					Description:  s.Description,
					GroupName:    g.Name,
					ResourceType: resourceType,
				})
			}
		}
	}
	sort.SliceStable(flat, func(i, j int) bool {
		if flat[i].ResourceType != flat[j].ResourceType {
			return flat[i].ResourceType < flat[j].ResourceType
		}
		if flat[i].GroupName != flat[j].GroupName {
			return flat[i].GroupName < flat[j].GroupName
		}
		return flat[i].Name < flat[j].Name
	})

	return infer.FunctionResponse[GetOrganizationRoleScopesOutput]{
		Output: GetOrganizationRoleScopesOutput{Scopes: flat},
	}, nil
}
