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

package resources

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

type ServicePrincipal struct{}

var (
	_ infer.CustomCreate[ServicePrincipalInput, ServicePrincipalState] = &ServicePrincipal{}
	_ infer.CustomDelete[ServicePrincipalState]                        = &ServicePrincipal{}
	_ infer.CustomRead[ServicePrincipalInput, ServicePrincipalState]   = &ServicePrincipal{}
	_ infer.CustomUpdate[ServicePrincipalInput, ServicePrincipalState] = &ServicePrincipal{}
)

func (*ServicePrincipal) Annotate(a infer.Annotator) {
	a.Describe(
		&ServicePrincipal{},
		"Creates a service principal in a Pulumi Cloud organization. A service principal is a named, non-human "+
			"member of the organization for automation to authenticate as. It has no access of its own: its "+
			"access comes from the organization-wide roles in `roleIds` and from entity access granted with "+
			"`OrganizationMemberEntityAccess`, and it never receives the organization's default role. A "+
			"stack's deployments can run as it: set `operationContext.servicePrincipal.login` on "+
			"`DeploymentSettings`.\n\n"+
			"Renaming the service principal or changing its roles updates it in place. Destroying it also "+
			"deletes its access tokens. Pulumi Cloud refuses to delete a service principal while a stack's "+
			"deployment settings run as it."+
			util.ImportDocs("pulumiservice:index:ServicePrincipal", "{organization}/{login}",
				"my-org/service-principal:d333a711-4aa0-402f-be6d-72af9665fc37",
				util.APIIDHint("service principal's login", "/api/orgs/my-org/members?type=service-principal",
					".members[] | {name: .user.name, login: .user.githubLogin}")),
	)
}

type ServicePrincipalInput struct {
	OrganizationName string   `pulumi:"organizationName" provider:"replaceOnChanges"`
	Name             string   `pulumi:"name"`
	RoleIDs          []string `pulumi:"roleIds,optional"`
}

func (i *ServicePrincipalInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The organization's name.")
	a.Describe(
		&i.Name,
		"The service principal's display name, unique (ignoring case) among the organization's service "+
			"principals. Up to 255 characters: ASCII letters, digits, spaces, hyphens, underscores and periods, "+
			"starting with a letter or digit. Changing it renames the service principal in place.\n\n"+
			autoNamedDescription,
	)
	a.Describe(
		&i.RoleIDs,
		"The IDs of the organization-wide roles the service principal holds. Use `getOrganizationRole` to look "+
			"up a role's ID by its name. Organizations with custom roles can assign any number of custom roles "+
			"and the built-in Admin, Maintainer, Builder and Reader roles. Without custom roles, a service "+
			"principal holds at most one of the Reader, Maintainer and Admin roles. Omit it, or leave it empty, "+
			"for a service principal whose only access is its entity access.",
	)
}

func (i *ServicePrincipalInput) AutoName(autoname infer.SetAutoName) {
	autoname(&i.Name, suffixedAutoName(servicePrincipalNameMaxLen, agentPoolNameRune))
}

// servicePrincipalNameMaxLen is the longest service principal name Pulumi
// Cloud accepts. Names allow the same characters as agent pool names.
const servicePrincipalNameMaxLen = 255

type ServicePrincipalState struct {
	ServicePrincipalInput
	Login string `pulumi:"login"`
}

func (s *ServicePrincipalState) Annotate(a infer.Annotator) {
	a.Describe(
		&s.Login,
		"The service principal's login, such as `service-principal:d333a711-4aa0-402f-be6d-72af9665fc37`. It "+
			"doesn't change when the service principal is renamed. Pass it as the `username` of "+
			"`OrganizationMemberEntityAccess`, or as `operationContext.servicePrincipal.login` of "+
			"`DeploymentSettings`.",
	)
}

func (*ServicePrincipal) Create(
	ctx context.Context,
	req infer.CreateRequest[ServicePrincipalInput],
) (infer.CreateResponse[ServicePrincipalState], error) {
	if req.DryRun {
		return infer.CreateResponse[ServicePrincipalState]{
			Output: ServicePrincipalState{ServicePrincipalInput: req.Inputs},
		}, nil
	}
	sp, err := config.GetClient(ctx).CreateServicePrincipal(
		ctx, req.Inputs.OrganizationName, req.Inputs.Name, req.Inputs.RoleIDs,
	)
	if err != nil {
		return infer.CreateResponse[ServicePrincipalState]{}, err
	}
	return infer.CreateResponse[ServicePrincipalState]{
		ID: servicePrincipalID(req.Inputs.OrganizationName, sp.Login),
		Output: ServicePrincipalState{
			ServicePrincipalInput: req.Inputs,
			Login:                 sp.Login,
		},
	}, nil
}

func (*ServicePrincipal) Update(
	ctx context.Context,
	req infer.UpdateRequest[ServicePrincipalInput, ServicePrincipalState],
) (infer.UpdateResponse[ServicePrincipalState], error) {
	out := ServicePrincipalState{ServicePrincipalInput: req.Inputs, Login: req.State.Login}
	if req.DryRun {
		return infer.UpdateResponse[ServicePrincipalState]{Output: out}, nil
	}
	org, login, err := splitServicePrincipalID(req.ID)
	if err != nil {
		return infer.UpdateResponse[ServicePrincipalState]{}, err
	}
	client := config.GetClient(ctx)
	if req.Inputs.Name != req.State.Name {
		if err := client.RenameServicePrincipal(ctx, org, login, req.Inputs.Name); err != nil {
			return infer.UpdateResponse[ServicePrincipalState]{}, err
		}
	}
	if !sameRoleIDs(req.Inputs.RoleIDs, req.State.RoleIDs) {
		if _, err := client.UpdateServicePrincipalRoles(ctx, org, login, req.Inputs.RoleIDs); err != nil {
			return infer.UpdateResponse[ServicePrincipalState]{}, err
		}
	}
	out.Login = login
	return infer.UpdateResponse[ServicePrincipalState]{Output: out}, nil
}

func (*ServicePrincipal) Delete(
	ctx context.Context,
	req infer.DeleteRequest[ServicePrincipalState],
) (infer.DeleteResponse, error) {
	org, login, err := splitServicePrincipalID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteServicePrincipal(ctx, org, login)
}

func (*ServicePrincipal) Read(
	ctx context.Context,
	req infer.ReadRequest[ServicePrincipalInput, ServicePrincipalState],
) (infer.ReadResponse[ServicePrincipalInput, ServicePrincipalState], error) {
	type response = infer.ReadResponse[ServicePrincipalInput, ServicePrincipalState]
	org, login, err := splitServicePrincipalID(req.ID)
	if err != nil {
		return response{}, err
	}
	sp, err := config.GetClient(ctx).GetServicePrincipal(ctx, org, login)
	if err != nil {
		return response{}, err
	}
	if sp == nil {
		return response{}, nil
	}
	inputs := ServicePrincipalInput{
		OrganizationName: org,
		Name:             sp.Name,
		RoleIDs:          readRoleIDs(sp.RoleIDs, req.State.RoleIDs),
	}
	return response{
		ID:     req.ID,
		Inputs: inputs,
		State:  ServicePrincipalState{ServicePrincipalInput: inputs, Login: login},
	}, nil
}

// sameRoleIDs reports whether a and b hold the same role IDs, in any order.
func sameRoleIDs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sa, sb := slices.Clone(a), slices.Clone(b)
	slices.Sort(sa)
	slices.Sort(sb)
	return slices.Equal(sa, sb)
}

// readRoleIDs converts the roles Pulumi Cloud reports to the input value. When
// they are the prior roles in another order, or both are empty, it keeps the
// prior value so refresh doesn't report a change.
func readRoleIDs(actual, prior []string) []string {
	if sameRoleIDs(actual, prior) {
		return prior
	}
	if len(actual) == 0 {
		return nil
	}
	return actual
}

func servicePrincipalID(org, login string) string {
	return org + "/" + login
}

func splitServicePrincipalID(id string) (string, string, error) {
	org, login, ok := strings.Cut(id, "/")
	if !ok || org == "" || login == "" || strings.Contains(login, "/") {
		return "", "", fmt.Errorf("%q is invalid, must be of the form organization/login", id)
	}
	return org, login, nil
}
