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
	"errors"
	"fmt"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

// GetOrganizationPermissionSetFunction looks up one permission set, typically
// a built-in one, so a role can grant it via `buildRolePermissions`.
type GetOrganizationPermissionSetFunction struct{}

type GetOrganizationPermissionSetInput struct {
	OrganizationName  string  `pulumi:"organizationName"`
	DefaultIdentifier *string `pulumi:"defaultIdentifier,optional"`
	Name              *string `pulumi:"name,optional"`
}

type GetOrganizationPermissionSetOutput struct {
	PermissionSetId   string   `pulumi:"permissionSetId"` //nolint:revive // matches the SDK property name
	Name              string   `pulumi:"name"`
	Description       string   `pulumi:"description"`
	ResourceType      string   `pulumi:"resourceType"`
	DefaultIdentifier *string  `pulumi:"defaultIdentifier,optional"`
	Permissions       []string `pulumi:"permissions"`
}

func (GetOrganizationPermissionSetFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetOrganizationPermissionSetFunction{},
		"Looks up a permission set in an organization by built-in identifier or by name. Use it to grant "+
			"built-in permission sets, such as \"Stack Read\" (`stack-read`) or \"Read Only\" "+
			"(`org-settings-read-only`), through `buildRolePermissions`: pass its `permissionSetId` and `resourceType` "+
			"in `permissionSets`. "+
			"Exactly one of `defaultIdentifier` or `name` must be set.",
	)
	a.SetToken("index", "getOrganizationPermissionSet")
}

func (i *GetOrganizationPermissionSetInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The Pulumi Cloud organization name.")
	a.Describe(&i.DefaultIdentifier, "The built-in permission set's identifier: `stack-read`, `stack-write`, "+
		"`stack-admin`, `environment-read`, `environment-open`, `environment-write`, `environment-admin`, "+
		"`insights-account-read`, `insights-account-write`, `insights-account-admin`, "+
		"`org-settings-standard`, `org-settings-read-only`, or `org-settings-billing`.")
	a.Describe(&i.Name, "The permission set's display name.")
}

func (o *GetOrganizationPermissionSetOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.PermissionSetId, "The permission set's unique ID. Pass it, with `resourceType`, to "+
		"`buildRolePermissions`.")
	a.Describe(&o.Name, "The permission set's display name.")
	a.Describe(&o.Description, "Human-readable description of what the permission set grants.")
	a.Describe(&o.ResourceType, "The kind of entity the permission set applies to, such as `stack`, "+
		"`environment`, `insights-account`, or `global` for organization-level sets.")
	a.Describe(&o.DefaultIdentifier, "For built-in permission sets, a stable identifier such as `stack-read`. "+
		"Unset for custom permission sets.")
	a.Describe(&o.Permissions, "The scopes the permission set grants. Only populated for permission sets "+
		"that grant a flat list of scopes.")
}

func (GetOrganizationPermissionSetFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetOrganizationPermissionSetInput],
) (infer.FunctionResponse[GetOrganizationPermissionSetOutput], error) {
	in := req.Input
	if (in.DefaultIdentifier == nil) == (in.Name == nil) {
		return infer.FunctionResponse[GetOrganizationPermissionSetOutput]{},
			errors.New("exactly one of `defaultIdentifier` or `name` must be set")
	}
	sets, err := config.GetClient(ctx).ListOrgRoles(
		ctx, in.OrganizationName, string(apitype.PermissionDescriptorUXPurposeSet))
	if err != nil {
		return infer.FunctionResponse[GetOrganizationPermissionSetOutput]{}, err
	}

	what := fmt.Sprintf("name %q", util.OrZero(in.Name))
	if in.DefaultIdentifier != nil {
		what = fmt.Sprintf("defaultIdentifier %q", *in.DefaultIdentifier)
	}
	var matches []apitype.PermissionDescriptorRecord
	for _, s := range sets {
		if (in.DefaultIdentifier != nil && s.DefaultIdentifier == *in.DefaultIdentifier) ||
			(in.Name != nil && s.Name == *in.Name) {
			matches = append(matches, s)
		}
	}
	switch len(matches) {
	case 0:
		return infer.FunctionResponse[GetOrganizationPermissionSetOutput]{}, fmt.Errorf(
			"no permission set with %s found in organization %q", what, in.OrganizationName)
	case 1:
		return infer.FunctionResponse[GetOrganizationPermissionSetOutput]{
			Output: permissionSetOutput(matches[0]),
		}, nil
	default:
		return infer.FunctionResponse[GetOrganizationPermissionSetOutput]{}, fmt.Errorf(
			"%d permission sets with %s found in organization %q", len(matches), what, in.OrganizationName)
	}
}

func permissionSetOutput(s apitype.PermissionDescriptorRecord) GetOrganizationPermissionSetOutput {
	out := GetOrganizationPermissionSetOutput{
		PermissionSetId: s.ID,
		Name:            s.Name,
		Description:     s.Description,
		ResourceType:    s.ResourceType,
		Permissions:     []string{},
	}
	if allow, ok := s.Details.(apitype.PermissionDescriptorAllow); ok {
		for _, p := range allow.Permissions() {
			out.Permissions = append(out.Permissions, string(p))
		}
	}
	if s.DefaultIdentifier != "" {
		id := s.DefaultIdentifier
		out.DefaultIdentifier = &id
	}
	return out
}
