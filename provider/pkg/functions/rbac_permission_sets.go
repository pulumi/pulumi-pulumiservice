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

package functions

import (
	"context"
	"fmt"
	"sort"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/resources"
)

// RbacPermissionSetInfo describes a built-in or custom permission set.
type RbacPermissionSetInfo struct {
	PermissionSetId   string                     `pulumi:"permissionSetId"`
	Name              string                     `pulumi:"name"`
	Description       string                     `pulumi:"description"`
	ResourceType      resources.RbacResourceType `pulumi:"resourceType"`
	DefaultIdentifier *string                    `pulumi:"defaultIdentifier,optional"`
	Permissions       []string                   `pulumi:"permissions"`
}

func (s *RbacPermissionSetInfo) Annotate(a infer.Annotator) {
	a.Describe(&s.PermissionSetId, "The permission set's unique ID. Reference it from an `RbacRole`.")
	a.Describe(&s.Name, "The permission set's display name.")
	a.Describe(&s.Description, "Human-readable description of what the permission set grants.")
	a.Describe(&s.ResourceType, "The kind of entity the permission set applies to.")
	a.Describe(&s.DefaultIdentifier, "For built-in permission sets, a stable identifier such as `stack-read`, "+
		"`environment-admin`, or `org-settings-read-only`. Unset for custom permission sets.")
	a.Describe(&s.Permissions, "The scopes the permission set grants. Only populated for permission sets "+
		"that grant a flat list of scopes.")
}

func permissionSetInfo(s apitype.PermissionDescriptorRecord) RbacPermissionSetInfo {
	info := RbacPermissionSetInfo{
		PermissionSetId: s.ID,
		Name:            s.Name,
		Description:     s.Description,
		ResourceType:    resources.RbacResourceType(s.ResourceType),
		Permissions:     []string{},
	}
	if allow, ok := s.Details.(apitype.PermissionDescriptorAllow); ok {
		for _, p := range allow.Permissions() {
			info.Permissions = append(info.Permissions, string(p))
		}
	}
	if s.DefaultIdentifier != "" {
		id := s.DefaultIdentifier
		info.DefaultIdentifier = &id
	}
	return info
}

func listPermissionSets(ctx context.Context, orgName string) ([]apitype.PermissionDescriptorRecord, error) {
	return config.GetClient(ctx).ListOrgRoles(ctx, orgName, string(apitype.PermissionDescriptorUXPurposeSet))
}

// GetRbacPermissionSetFunction looks up one permission set, typically a
// built-in one, so a role can grant it without recreating it.
type GetRbacPermissionSetFunction struct{}

type GetRbacPermissionSetInput struct {
	OrganizationName  string  `pulumi:"organizationName"`
	DefaultIdentifier *string `pulumi:"defaultIdentifier,optional"`
	Name              *string `pulumi:"name,optional"`
}

func (GetRbacPermissionSetFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetRbacPermissionSetFunction{},
		"Looks up a permission set in an organization by built-in identifier or by name. Use it to grant "+
			"built-in permission sets, such as \"Stack Read\" (`stack-read`) or \"Read Only\" "+
			"(`org-settings-read-only`), from an `RbacRole`. Exactly one of `defaultIdentifier` or `name` "+
			"must be set.",
	)
	a.SetToken("index", "getRbacPermissionSet")
}

func (i *GetRbacPermissionSetInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The Pulumi Cloud organization name.")
	a.Describe(&i.DefaultIdentifier, "The built-in permission set's identifier: `stack-read`, `stack-write`, "+
		"`stack-admin`, `environment-read`, `environment-open`, `environment-write`, `environment-admin`, "+
		"`insights-account-read`, `insights-account-write`, `insights-account-admin`, "+
		"`org-settings-standard`, `org-settings-read-only`, or `org-settings-billing`.")
	a.Describe(&i.Name, "The permission set's display name.")
}

func (GetRbacPermissionSetFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetRbacPermissionSetInput],
) (infer.FunctionResponse[RbacPermissionSetInfo], error) {
	in := req.Input
	if (in.DefaultIdentifier == nil) == (in.Name == nil) {
		return infer.FunctionResponse[RbacPermissionSetInfo]{},
			fmt.Errorf("exactly one of defaultIdentifier or name must be set")
	}
	sets, err := listPermissionSets(ctx, in.OrganizationName)
	if err != nil {
		return infer.FunctionResponse[RbacPermissionSetInfo]{}, err
	}

	var matches []apitype.PermissionDescriptorRecord
	for _, s := range sets {
		if (in.DefaultIdentifier != nil && s.DefaultIdentifier == *in.DefaultIdentifier) ||
			(in.Name != nil && s.Name == *in.Name) {
			matches = append(matches, s)
		}
	}
	what := func() string {
		if in.DefaultIdentifier != nil {
			return fmt.Sprintf("defaultIdentifier %q", *in.DefaultIdentifier)
		}
		return fmt.Sprintf("name %q", *in.Name)
	}
	switch len(matches) {
	case 0:
		return infer.FunctionResponse[RbacPermissionSetInfo]{}, fmt.Errorf(
			"no permission set with %s found in organization %q", what(), in.OrganizationName)
	case 1:
		return infer.FunctionResponse[RbacPermissionSetInfo]{Output: permissionSetInfo(matches[0])}, nil
	default:
		return infer.FunctionResponse[RbacPermissionSetInfo]{}, fmt.Errorf(
			"%d permission sets with %s found in organization %q", len(matches), what(), in.OrganizationName)
	}
}

// GetRbacPermissionSetsFunction lists an organization's permission sets.
type GetRbacPermissionSetsFunction struct{}

type GetRbacPermissionSetsInput struct {
	OrganizationName string                      `pulumi:"organizationName"`
	ResourceType     *resources.RbacResourceType `pulumi:"resourceType,optional"`
}

type GetRbacPermissionSetsOutput struct {
	PermissionSets []RbacPermissionSetInfo `pulumi:"permissionSets"`
}

func (GetRbacPermissionSetsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetRbacPermissionSetsFunction{},
		"Lists the built-in and custom permission sets in an organization, sorted by resource type and name.",
	)
	a.SetToken("index", "getRbacPermissionSets")
}

func (i *GetRbacPermissionSetsInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The Pulumi Cloud organization name.")
	a.Describe(&i.ResourceType, "Only return permission sets for this kind of entity.")
}

func (o *GetRbacPermissionSetsOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.PermissionSets, "The matching permission sets.")
}

func (GetRbacPermissionSetsFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetRbacPermissionSetsInput],
) (infer.FunctionResponse[GetRbacPermissionSetsOutput], error) {
	sets, err := listPermissionSets(ctx, req.Input.OrganizationName)
	if err != nil {
		return infer.FunctionResponse[GetRbacPermissionSetsOutput]{}, err
	}
	out := make([]RbacPermissionSetInfo, 0, len(sets))
	for _, s := range sets {
		if req.Input.ResourceType != nil && s.ResourceType != string(*req.Input.ResourceType) {
			continue
		}
		out = append(out, permissionSetInfo(s))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ResourceType != out[j].ResourceType {
			return out[i].ResourceType < out[j].ResourceType
		}
		return out[i].Name < out[j].Name
	})
	return infer.FunctionResponse[GetRbacPermissionSetsOutput]{
		Output: GetRbacPermissionSetsOutput{PermissionSets: out},
	}, nil
}
