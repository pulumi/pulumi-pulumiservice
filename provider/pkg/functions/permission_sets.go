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
	"strings"

	_ "embed" // for the //go:embed example files below

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

var (
	//go:embed get_stack_permission_set_examples.md
	getStackPermissionSetExamples string
	//go:embed get_environment_permission_set_examples.md
	getEnvironmentPermissionSetExamples string
	//go:embed get_account_permission_set_examples.md
	getAccountPermissionSetExamples string
	//go:embed get_organization_permission_set_examples.md
	getOrganizationPermissionSetExamples string
)

// There is one lookup per entity type, rather than one that returns the set's
// entity type, so that a rule's permissionSetScopes can take the lookup's
// scopes directly: the function a set came from already says which rule list
// it belongs in.

type (
	BuiltInStackPermissionSet        string
	BuiltInEnvironmentPermissionSet  string
	BuiltInAccountPermissionSet      string
	BuiltInOrganizationPermissionSet string
)

func (BuiltInStackPermissionSet) Values() []infer.EnumValue[BuiltInStackPermissionSet] {
	return builtInStackPermissionSetValues
}

func (BuiltInEnvironmentPermissionSet) Values() []infer.EnumValue[BuiltInEnvironmentPermissionSet] {
	return builtInEnvironmentPermissionSetValues
}

func (BuiltInAccountPermissionSet) Values() []infer.EnumValue[BuiltInAccountPermissionSet] {
	return builtInAccountPermissionSetValues
}

func (BuiltInOrganizationPermissionSet) Values() []infer.EnumValue[BuiltInOrganizationPermissionSet] {
	return builtInOrganizationPermissionSetValues
}

// PermissionSetOutput is the result of every permission set lookup.
type PermissionSetOutput struct {
	PermissionSetId string   `pulumi:"permissionSetId"` //nolint:revive // matches the SDK property name
	Name            string   `pulumi:"name"`
	Description     string   `pulumi:"description"`
	Scopes          []string `pulumi:"scopes"`
	BuiltIn         *string  `pulumi:"builtIn,optional"`
}

func (o *PermissionSetOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.PermissionSetId, "The permission set's ID.")
	a.Describe(&o.Name, "The permission set's name, as shown in the Pulumi Cloud console.")
	a.Describe(&o.Description, "The permission set's description.")
	a.Describe(&o.Scopes, "The scopes the set grants, including those of any sets it contains. Scopes this "+
		"version of the provider doesn't recognize are left out, with a warning.")
	a.Describe(&o.BuiltIn, "The built-in permission set's identifier, such as `stack-write`. Unset for custom "+
		"permission sets.")
}

// permissionSetKind describes one entity type's lookup.
type permissionSetKind struct {
	entityType string // as Pulumi Cloud names it
	noun       string // for messages, e.g. "stack"
	ruleInput  string // where buildRolePermissions takes the scopes
}

var (
	stackPermissionSets = permissionSetKind{entityTypeStack, "stack",
		"`permissionSetScopes` in `stackRules`"}
	environmentPermissionSets = permissionSetKind{entityTypeEnvironment, "environment",
		"`permissionSetScopes` in `environmentRules`"}
	accountPermissionSets = permissionSetKind{entityTypeInsightsAccount, "cloud account",
		"`permissionSetScopes` in `accountRules`"}
	organizationPermissionSets = permissionSetKind{entityTypeOrganization, "organization-level",
		"`organizationPermissionSetScopes`"}
)

func (k permissionSetKind) describe(
	a infer.Annotator, fn any, token tokens.TypeName, builtInExample, examples string,
) {
	a.Describe(fn, fmt.Sprintf(
		"Looks up a %s permission set in a Pulumi Cloud organization, built-in (such as %s) or custom, and "+
			"returns the scopes it grants. Pass `scopes` to %s of `buildRolePermissions` to grant the set, "+
			"the way access levels do in the Pulumi Cloud console. Exactly one of `builtIn` or `name` must "+
			"be set.\n\n"+
			"The role gets a copy of the set's scopes, not a link to the set. If the set changes in Pulumi "+
			"Cloud, the next `pulumi up` updates the role to match. If the set contains scopes this version "+
			"of the provider doesn't recognize, the lookup leaves them out and warns; upgrade the provider to "+
			"grant them.\n\n%s",
		k.noun, builtInExample, k.ruleInput, examples))
	a.SetToken("index", token)
}

func (k permissionSetKind) describeInput(a infer.Annotator, org, builtIn, name any) {
	a.Describe(org, "The name of the Pulumi organization.")
	a.Describe(builtIn, fmt.Sprintf("The built-in %s permission set to look up.", k.noun))
	a.Describe(name, "The permission set's name, as shown in the Pulumi Cloud console. The match is "+
		"case-insensitive.")
}

// invoke looks up the set, then warns about any scopes left out because the
// provider doesn't recognize them.
func (k permissionSetKind) invoke(
	ctx context.Context, orgName string, builtIn, name *string,
) (infer.FunctionResponse[PermissionSetOutput], error) {
	out, unknown, err := k.lookup(ctx, orgName, builtIn, name)
	if err != nil {
		return infer.FunctionResponse[PermissionSetOutput]{}, err
	}
	if len(unknown) > 0 {
		p.GetLogger(ctx).Warning(unknownScopesWarning(out.Name, unknown))
	}
	return infer.FunctionResponse[PermissionSetOutput]{Output: out}, nil
}

func unknownScopesWarning(setName string, unknown []string) string {
	return fmt.Sprintf("Permission set %q contains %d scope(s) that this version of the pulumiservice provider "+
		"doesn't recognize: %s. Roles built from this set won't grant them. Every other scope in the set is "+
		"granted as usual. Upgrade the pulumiservice provider to grant them.",
		setName, len(unknown), strings.Join(unknown, ", "))
}

// lookup returns the set and the scopes it left out of the output because
// the provider's Pulumi Cloud SDK doesn't recognize them.
func (k permissionSetKind) lookup(
	ctx context.Context, orgName string, builtIn, name *string,
) (PermissionSetOutput, []string, error) {
	var match func(apitype.PermissionDescriptorRecord) bool
	var what string
	switch {
	case builtIn != nil && name == nil:
		what = fmt.Sprintf("built-in %s permission set %q", k.noun, *builtIn)
		match = func(s apitype.PermissionDescriptorRecord) bool { return s.DefaultIdentifier == *builtIn }
	case builtIn == nil && name != nil && strings.TrimSpace(*name) != "":
		n := strings.TrimSpace(*name)
		what = fmt.Sprintf("%s permission set named %q", k.noun, n)
		match = func(s apitype.PermissionDescriptorRecord) bool { return strings.EqualFold(s.Name, n) }
	default:
		return PermissionSetOutput{}, nil, errors.New("exactly one of `builtIn` or `name` must be set")
	}

	sets, err := config.GetClient(ctx).ListOrgRoles(ctx, orgName, apitype.PermissionDescriptorUXPurposeSet)
	if err != nil {
		return PermissionSetOutput{}, nil, fmt.Errorf(
			"failed to list permission sets for organization %q: %w", orgName, err)
	}
	byID := make(map[string]apitype.PermissionDescriptorRecord, len(sets))
	var matches, otherTypes []apitype.PermissionDescriptorRecord
	for _, s := range sets {
		// The uxPurpose filter is the API's; keep other kinds out regardless.
		if s.UxPurpose != apitype.PermissionDescriptorUXPurposeSet {
			continue
		}
		byID[s.ID] = s
		switch {
		case !match(s):
		case s.ResourceType == k.entityType:
			matches = append(matches, s)
		default:
			otherTypes = append(otherTypes, s)
		}
	}
	switch {
	case len(matches) == 1:
	case len(matches) > 1:
		return PermissionSetOutput{}, nil, fmt.Errorf(
			"organization %q has %d of %s; rename one in the Pulumi Cloud console", orgName, len(matches), what)
	case len(otherTypes) > 0:
		return PermissionSetOutput{}, nil, fmt.Errorf(
			"permission set %q in organization %q applies to %s, not %s; use the lookup for that entity type",
			otherTypes[0].Name, orgName, otherTypes[0].ResourceType, k.entityType)
	default:
		return PermissionSetOutput{}, nil, fmt.Errorf("organization %q has no %s", orgName, what)
	}

	s := matches[0]
	all, err := setScopes(s.Details, byID, map[string]bool{s.ID: true})
	if err != nil {
		return PermissionSetOutput{}, nil, fmt.Errorf("permission set %q: %w", s.Name, err)
	}
	var known, unknown []string
	for _, scope := range dedupe(all) {
		if apitype.RbacPermission(scope).IsValid() {
			known = append(known, scope)
		} else {
			unknown = append(unknown, scope)
		}
	}
	out := PermissionSetOutput{
		PermissionSetId: s.ID,
		Name:            s.Name,
		Description:     s.Description,
		Scopes:          known,
	}
	if s.DefaultIdentifier != "" {
		out.BuiltIn = new(s.DefaultIdentifier)
	}
	return out, unknown, nil
}

// setScopes flattens a permission set's descriptor into the scopes it
// grants, following references to other sets. buildRolePermissions grants a
// set by copying its scopes into a rule, so a set whose grants depend on a
// condition can't be expressed and is rejected instead of being widened.
func setScopes(
	d apitype.PermissionDescriptor,
	byID map[string]apitype.PermissionDescriptorRecord,
	visiting map[string]bool,
) ([]string, error) {
	switch d := d.(type) {
	case nil:
		return nil, nil
	case apitype.PermissionDescriptorAllow:
		if d.Constraints() != nil {
			return nil, errors.New("it constrains its scopes, so it can't be expanded to a list of scopes")
		}
		out := make([]string, len(d.Permissions()))
		for i, perm := range d.Permissions() {
			out[i] = string(perm)
		}
		return out, nil
	case apitype.PermissionDescriptorGroup:
		var out []string
		for _, e := range d.Entries() {
			scopes, err := setScopes(e, byID, visiting)
			if err != nil {
				return nil, err
			}
			out = append(out, scopes...)
		}
		return out, nil
	case apitype.PermissionDescriptorCompose:
		var out []string
		for _, id := range d.PermissionDescriptors() {
			if visiting[id] {
				return nil, fmt.Errorf("it contains itself through permission set %q", id)
			}
			nested, ok := byID[id]
			if !ok {
				return nil, fmt.Errorf("it contains permission set %q, which wasn't found", id)
			}
			visiting[id] = true
			scopes, err := setScopes(nested.Details, byID, visiting)
			delete(visiting, id)
			if err != nil {
				return nil, err
			}
			out = append(out, scopes...)
		}
		return out, nil
	default:
		kind, _ := d.GetDiscriminatorValue()
		return nil, fmt.Errorf("it contains a %s, so it can't be expanded to a list of scopes", kind)
	}
}

// ----------------------------------------------------------------------------
// One function per entity type
// ----------------------------------------------------------------------------

type GetStackPermissionSetFunction struct{}

type GetStackPermissionSetInput struct {
	OrganizationName string                     `pulumi:"organizationName"`
	BuiltIn          *BuiltInStackPermissionSet `pulumi:"builtIn,optional"`
	Name             *string                    `pulumi:"name,optional"`
}

func (GetStackPermissionSetFunction) Annotate(a infer.Annotator) {
	stackPermissionSets.describe(a, &GetStackPermissionSetFunction{}, "getStackPermissionSet", "Stack Write",
		getStackPermissionSetExamples)
}

func (i *GetStackPermissionSetInput) Annotate(a infer.Annotator) {
	stackPermissionSets.describeInput(a, &i.OrganizationName, &i.BuiltIn, &i.Name)
}

func (GetStackPermissionSetFunction) Invoke(
	ctx context.Context, req infer.FunctionRequest[GetStackPermissionSetInput],
) (infer.FunctionResponse[PermissionSetOutput], error) {
	return stackPermissionSets.invoke(ctx, req.Input.OrganizationName, (*string)(req.Input.BuiltIn), req.Input.Name)
}

type GetEnvironmentPermissionSetFunction struct{}

type GetEnvironmentPermissionSetInput struct {
	OrganizationName string                           `pulumi:"organizationName"`
	BuiltIn          *BuiltInEnvironmentPermissionSet `pulumi:"builtIn,optional"`
	Name             *string                          `pulumi:"name,optional"`
}

func (GetEnvironmentPermissionSetFunction) Annotate(a infer.Annotator) {
	environmentPermissionSets.describe(a, &GetEnvironmentPermissionSetFunction{},
		"getEnvironmentPermissionSet", "Environment Open", getEnvironmentPermissionSetExamples)
}

func (i *GetEnvironmentPermissionSetInput) Annotate(a infer.Annotator) {
	environmentPermissionSets.describeInput(a, &i.OrganizationName, &i.BuiltIn, &i.Name)
}

func (GetEnvironmentPermissionSetFunction) Invoke(
	ctx context.Context, req infer.FunctionRequest[GetEnvironmentPermissionSetInput],
) (infer.FunctionResponse[PermissionSetOutput], error) {
	return environmentPermissionSets.invoke(ctx, req.Input.OrganizationName,
		(*string)(req.Input.BuiltIn), req.Input.Name)
}

type GetAccountPermissionSetFunction struct{}

type GetAccountPermissionSetInput struct {
	OrganizationName string                       `pulumi:"organizationName"`
	BuiltIn          *BuiltInAccountPermissionSet `pulumi:"builtIn,optional"`
	Name             *string                      `pulumi:"name,optional"`
}

func (GetAccountPermissionSetFunction) Annotate(a infer.Annotator) {
	accountPermissionSets.describe(a, &GetAccountPermissionSetFunction{}, "getAccountPermissionSet", "Account Read",
		getAccountPermissionSetExamples)
}

func (i *GetAccountPermissionSetInput) Annotate(a infer.Annotator) {
	accountPermissionSets.describeInput(a, &i.OrganizationName, &i.BuiltIn, &i.Name)
}

func (GetAccountPermissionSetFunction) Invoke(
	ctx context.Context, req infer.FunctionRequest[GetAccountPermissionSetInput],
) (infer.FunctionResponse[PermissionSetOutput], error) {
	return accountPermissionSets.invoke(ctx, req.Input.OrganizationName, (*string)(req.Input.BuiltIn), req.Input.Name)
}

type GetOrganizationPermissionSetFunction struct{}

type GetOrganizationPermissionSetInput struct {
	OrganizationName string                            `pulumi:"organizationName"`
	BuiltIn          *BuiltInOrganizationPermissionSet `pulumi:"builtIn,optional"`
	Name             *string                           `pulumi:"name,optional"`
}

func (GetOrganizationPermissionSetFunction) Annotate(a infer.Annotator) {
	organizationPermissionSets.describe(a, &GetOrganizationPermissionSetFunction{},
		"getOrganizationPermissionSet", "Read Only", getOrganizationPermissionSetExamples)
}

func (i *GetOrganizationPermissionSetInput) Annotate(a infer.Annotator) {
	organizationPermissionSets.describeInput(a, &i.OrganizationName, &i.BuiltIn, &i.Name)
}

func (GetOrganizationPermissionSetFunction) Invoke(
	ctx context.Context, req infer.FunctionRequest[GetOrganizationPermissionSetInput],
) (infer.FunctionResponse[PermissionSetOutput], error) {
	return organizationPermissionSets.invoke(ctx, req.Input.OrganizationName,
		(*string)(req.Input.BuiltIn), req.Input.Name)
}
