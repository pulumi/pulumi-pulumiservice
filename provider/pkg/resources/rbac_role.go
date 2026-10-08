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

package resources

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

const (
	gcNameNotEmpty                 = "name must not be empty"
	gcInsightsAccount              = "insightsAccount"
	gcEntityRules                  = "entityRules"
	gcOrganizationPermissionSetIds = "organizationPermissionSetIds"
)

type RbacRole struct{}

var (
	_ infer.CustomCreate[RbacRoleInput, RbacRoleState] = &RbacRole{}
	_ infer.CustomCheck[RbacRoleInput]                 = &RbacRole{}
	_ infer.CustomDelete[RbacRoleState]                = &RbacRole{}
	_ infer.CustomRead[RbacRoleInput, RbacRoleState]   = &RbacRole{}
	_ infer.CustomUpdate[RbacRoleInput, RbacRoleState] = &RbacRole{}
)

func (*RbacRole) Annotate(a infer.Annotator) {
	a.Describe(
		&RbacRole{},
		"A custom role, modeled the way the Pulumi Cloud console builds one under **Settings > Access "+
			"management > Roles**: organization-level access plus entity rules. Each grants one or more "+
			"`RbacPermissionSet`s (or built-in permission sets, looked up with `getRbacPermissionSet`).\n\n"+
			"- **Organization-level access** (`organizationPermissionSetIds`) grants `global` permission "+
			"sets across the organization.\n"+
			"- **Entity rules** (`entityRules`) grant `stack`, `environment`, or `insights-account` "+
			"permission sets on all entities of that kind, on one entity by ID, or on entities whose tags "+
			"match.\n\n"+
			"Assign the role to teams with `TeamRoleAssignment` or to members with "+
			"`OrganizationMember.roleId`. Requires the Custom Roles feature to be enabled on the "+
			"organization.\n\n"+rbacExampleDocs,
	)
	a.SetToken("index", "RbacRole")
}

type RbacRoleCore struct {
	//nolint:lll // struct tags align with the fields below
	OrganizationName             string           `pulumi:"organizationName"                      provider:"replaceOnChanges"`
	Name                         string           `pulumi:"name"`
	Description                  *string          `pulumi:"description,optional"`
	OrganizationPermissionSetIds []string         `pulumi:"organizationPermissionSetIds,optional"`
	EntityRules                  []RbacEntityRule `pulumi:"entityRules,optional"`
}

func (c *RbacRoleCore) Annotate(a infer.Annotator) {
	a.Describe(&c.OrganizationName, "The Pulumi Cloud organization name.")
	a.Describe(&c.Name, "The role's display name. Must be unique within the organization.")
	a.Describe(&c.Description, "Human-readable description of what the role grants.")
	a.Describe(
		&c.OrganizationPermissionSetIds,
		"IDs of `global` permission sets granted across the organization (the console's "+
			"\"Organization-level access\"), for example the built-in \"Standard\" or \"Read Only\" sets.",
	)
	a.Describe(
		&c.EntityRules,
		"Rules granting `stack`, `environment`, or `insights-account` permission sets on matching "+
			"entities (the console's \"Entity rules\").",
	)
}

// RbacEntityRule grants permission sets on the entities its selector matches.
type RbacEntityRule struct {
	PermissionSetIds []string            `pulumi:"permissionSetIds"`
	Stack            *RbacEntitySelector `pulumi:"stack,optional"`
	Environment      *RbacEntitySelector `pulumi:"environment,optional"`
	InsightsAccount  *RbacEntitySelector `pulumi:"insightsAccount,optional"`
}

func (r *RbacEntityRule) Annotate(a infer.Annotator) {
	a.Describe(
		&r.PermissionSetIds,
		"IDs of the permission sets to grant. Their `resourceType` must match the selected entity kind.",
	)
	a.Describe(&r.Stack, "Select stacks. Exactly one of `stack`, `environment`, or `insightsAccount` must be set.")
	a.Describe(&r.Environment, "Select ESC environments.")
	a.Describe(&r.InsightsAccount, "Select Insights accounts.")
}

// RbacEntitySelector picks the entities a rule applies to.
type RbacEntitySelector struct {
	All  *bool              `pulumi:"all,optional"`
	Id   *string            `pulumi:"id,optional"`
	Tags []RbacTagCondition `pulumi:"tags,optional"`
}

func (s *RbacEntitySelector) Annotate(a infer.Annotator) {
	a.Describe(&s.All, "Apply to every entity of this kind in the organization. Exactly one of `all`, `id`, "+
		"or `tags` must be set.")
	a.Describe(&s.Id, "Apply to a single entity: a `Stack.stackId`, `Environment.environmentId`, or "+
		"`InsightsAccount.insightsAccountId`.")
	a.Describe(&s.Tags, "Apply to entities whose tags match every condition.")
}

// RbacTagCondition matches entities on one tag.
type RbacTagCondition struct {
	Key      string           `pulumi:"key"`
	Value    *string          `pulumi:"value,optional"`
	Operator *RbacTagOperator `pulumi:"operator,optional"`
}

func (c *RbacTagCondition) Annotate(a infer.Annotator) {
	a.Describe(&c.Key, "The tag key.")
	a.Describe(&c.Value, "The tag value. When omitted, the condition matches on whether the tag is present "+
		"at all.")
	a.Describe(&c.Operator, "How the tag is compared. Defaults to `equals`.")
}

// RbacTagOperator compares an entity's tag against a condition. These are
// the operators the Pulumi Cloud console offers.
type RbacTagOperator string

const (
	RbacTagOperatorEquals    RbacTagOperator = "equals"
	RbacTagOperatorNotEquals RbacTagOperator = "notEquals"
)

func (RbacTagOperator) Values() []infer.EnumValue[RbacTagOperator] {
	return []infer.EnumValue[RbacTagOperator]{
		{Name: "Equals", Value: RbacTagOperatorEquals,
			Description: "The tag equals the value (or, with no value, the tag is present)."},
		{Name: "NotEquals", Value: RbacTagOperatorNotEquals,
			Description: "The tag does not equal the value (or, with no value, the tag is absent)."},
	}
}

type RbacRoleInput struct {
	RbacRoleCore
}

type RbacRoleState struct {
	RbacRoleCore
	RoleId   string `pulumi:"roleId"`
	PolicyId string `pulumi:"policyId"`
	Version  int    `pulumi:"version"`
}

func (s *RbacRoleState) Annotate(a infer.Annotator) {
	a.Describe(&s.RoleId, "The role's unique ID. Assign it with `TeamRoleAssignment.roleId` or "+
		"`OrganizationMember.roleId`.")
	a.Describe(&s.PolicyId, "The ID of the policy that holds the role's organization-level access and "+
		"entity rules. Pulumi Cloud creates and deletes it with the role.")
	a.Describe(&s.Version, "The service-maintained version number of the role's policy, which increments on "+
		"every update.")
}

func (*RbacRole) Check(
	ctx context.Context,
	req infer.CheckRequest,
) (infer.CheckResponse[RbacRoleInput], error) {
	in, failures, err := infer.DefaultCheck[RbacRoleInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[RbacRoleInput]{}, err
	}
	if !isUnknownInput(req.NewInputs, gcName) && in.Name == "" {
		failures = append(failures, p.CheckFailure{Property: gcName, Reason: gcNameNotEmpty})
	}
	// Rules routinely reference IDs that are unknown at preview (a new
	// permission set or stack); the typed fields decode those as nil, so
	// only validate structure once every value is known.
	if !hasUnknownInput(req.NewInputs, gcEntityRules) &&
		!hasUnknownInput(req.NewInputs, gcOrganizationPermissionSetIds) {
		failures = append(failures, checkRoleRules(in.RbacRoleCore)...)
	}
	return infer.CheckResponse[RbacRoleInput]{Inputs: in, Failures: failures}, nil
}

func checkRoleRules(core RbacRoleCore) []p.CheckFailure {
	var failures []p.CheckFailure
	fail := func(prop, reason string) {
		failures = append(failures, p.CheckFailure{Property: prop, Reason: reason})
	}
	if len(core.OrganizationPermissionSetIds) == 0 && len(core.EntityRules) == 0 {
		fail(gcOrganizationPermissionSetIds,
			"a role must grant something: set organizationPermissionSetIds, entityRules, or both")
	}
	for i, rule := range core.EntityRules {
		prop := fmt.Sprintf("%s[%d]", gcEntityRules, i)
		if len(rule.PermissionSetIds) == 0 {
			fail(prop+".permissionSetIds", "at least one permission set ID is required")
		}
		selected := 0
		for _, s := range []*RbacEntitySelector{rule.Stack, rule.Environment, rule.InsightsAccount} {
			if s != nil {
				selected++
			}
		}
		if selected != 1 {
			fail(prop, "exactly one of stack, environment, or insightsAccount must be set")
			continue
		}
		rt, sel := rule.selector()
		selProp := prop + "." + map[RbacResourceType]string{
			RbacResourceTypeStack:           gcStack,
			RbacResourceTypeEnvironment:     gcEnvironment,
			RbacResourceTypeInsightsAccount: gcInsightsAccount,
		}[rt]
		modes := 0
		if sel.All != nil && *sel.All {
			modes++
		}
		if sel.Id != nil {
			modes++
			if *sel.Id == "" {
				fail(selProp+".id", "id must not be empty")
			}
		}
		if len(sel.Tags) > 0 {
			modes++
		}
		if modes != 1 {
			fail(selProp, "exactly one of all (true), id, or tags must be set")
		}
		for j, tc := range sel.Tags {
			if strings.TrimSpace(tc.Key) == "" {
				fail(fmt.Sprintf("%s.tags[%d].key", selProp, j), "tag key must not be empty")
			}
		}
	}
	return failures
}

// hasUnknownInput reports whether the value at key contains an unknown
// anywhere inside it.
func hasUnknownInput(m property.Map, key string) bool {
	v, ok := m.GetOk(key)
	return ok && v.HasComputed()
}

func (*RbacRole) Create(
	ctx context.Context,
	req infer.CreateRequest[RbacRoleInput],
) (infer.CreateResponse[RbacRoleState], error) {
	core := req.Inputs.RbacRoleCore
	if req.DryRun {
		return infer.CreateResponse[RbacRoleState]{
			ID:     fmt.Sprintf("%s/%s", core.OrganizationName, core.Name),
			Output: RbacRoleState{RbacRoleCore: core},
		}, nil
	}

	details, err := buildPolicyDetails(core)
	if err != nil {
		return infer.CreateResponse[RbacRoleState]{}, fmt.Errorf("invalid role %q: %w", core.Name, err)
	}
	client := config.GetClient(ctx)
	role, err := client.CreateRoleWithPolicy(ctx, core.OrganizationName, apitype.PermissionDescriptorBase{
		Name:        core.Name,
		Description: util.OrZero(core.Description),
		UxPurpose:   apitype.PermissionDescriptorUXPurposeRole,
		Details:     details,
	})
	if err != nil {
		return infer.CreateResponse[RbacRoleState]{}, fmt.Errorf("failed to create role %q: %w", core.Name, err)
	}
	policyID, err := rolePolicyID(role)
	if err != nil {
		return infer.CreateResponse[RbacRoleState]{}, err
	}
	policy, err := client.GetRole(ctx, core.OrganizationName, policyID)
	if err != nil || policy == nil {
		return infer.CreateResponse[RbacRoleState]{}, fmt.Errorf(
			"role %q was created but its policy %q could not be read: %w", role.ID, policyID, err)
	}

	return infer.CreateResponse[RbacRoleState]{
		ID: fmt.Sprintf("%s/%s", core.OrganizationName, role.ID),
		Output: RbacRoleState{
			RbacRoleCore: core,
			RoleId:       role.ID,
			PolicyId:     policyID,
			Version:      int(policy.Version),
		},
	}, nil
}

func (*RbacRole) Update(
	ctx context.Context,
	req infer.UpdateRequest[RbacRoleInput, RbacRoleState],
) (infer.UpdateResponse[RbacRoleState], error) {
	core := req.Inputs.RbacRoleCore
	state := RbacRoleState{
		RbacRoleCore: core,
		RoleId:       req.State.RoleId,
		PolicyId:     req.State.PolicyId,
		Version:      req.State.Version,
	}
	if req.DryRun {
		return infer.UpdateResponse[RbacRoleState]{Output: state}, nil
	}

	details, err := buildPolicyDetails(core)
	if err != nil {
		return infer.UpdateResponse[RbacRoleState]{}, fmt.Errorf("invalid role %q: %w", core.Name, err)
	}
	client := config.GetClient(ctx)
	org := req.State.OrganizationName
	name := core.Name
	description := util.OrZero(core.Description)

	// Like the console: the policy carries the permissions, and both
	// descriptors carry the role's name and description.
	policy, err := client.UpdateRole(ctx, org, req.State.PolicyId, apitype.UpdateRoleRequest{
		Name: &name, Description: &description, Details: details,
	})
	if err != nil {
		return infer.UpdateResponse[RbacRoleState]{}, fmt.Errorf(
			"failed to update policy %q of role %q: %w", req.State.PolicyId, req.State.RoleId, err)
	}
	if name != req.State.Name || description != util.OrZero(req.State.Description) {
		if _, err := client.UpdateRole(ctx, org, req.State.RoleId, apitype.UpdateRoleRequest{
			Name: &name, Description: &description,
		}); err != nil {
			return infer.UpdateResponse[RbacRoleState]{}, fmt.Errorf(
				"failed to update role %q: %w", req.State.RoleId, err)
		}
	}
	state.Version = int(policy.Version)
	return infer.UpdateResponse[RbacRoleState]{Output: state}, nil
}

func (*RbacRole) Delete(
	ctx context.Context,
	req infer.DeleteRequest[RbacRoleState],
) (infer.DeleteResponse, error) {
	// Pulumi Cloud deletes the role's policy along with the role.
	return infer.DeleteResponse{}, deletePermissionDescriptor(ctx, req.State.OrganizationName, req.State.RoleId)
}

func (*RbacRole) Read(
	ctx context.Context,
	req infer.ReadRequest[RbacRoleInput, RbacRoleState],
) (infer.ReadResponse[RbacRoleInput, RbacRoleState], error) {
	orgName, roleID, err := splitOrgRoleID(req.ID)
	if err != nil {
		return infer.ReadResponse[RbacRoleInput, RbacRoleState]{}, err
	}
	client := config.GetClient(ctx)
	role, err := client.GetRole(ctx, orgName, roleID)
	if err != nil {
		return infer.ReadResponse[RbacRoleInput, RbacRoleState]{}, fmt.Errorf("failed to read role %q: %w", req.ID, err)
	}
	if role == nil {
		return infer.ReadResponse[RbacRoleInput, RbacRoleState]{}, nil
	}
	if role.UxPurpose != apitype.PermissionDescriptorUXPurposeRole {
		return infer.ReadResponse[RbacRoleInput, RbacRoleState]{}, fmt.Errorf(
			"descriptor %q is not a role (uxPurpose=%q)", role.ID, role.UxPurpose)
	}
	policyID, err := rolePolicyID(role)
	if err != nil {
		return infer.ReadResponse[RbacRoleInput, RbacRoleState]{}, err
	}
	policy, err := client.GetRole(ctx, orgName, policyID)
	if err != nil {
		return infer.ReadResponse[RbacRoleInput, RbacRoleState]{}, fmt.Errorf(
			"failed to read policy %q of role %q: %w", policyID, req.ID, err)
	}
	if policy == nil || policy.UxPurpose != apitype.PermissionDescriptorUXPurposePolicy {
		return infer.ReadResponse[RbacRoleInput, RbacRoleState]{}, unrepresentableRole(role.ID,
			fmt.Errorf("%w: role does not compose a policy", errUnrepresentable))
	}

	orgSetIDs, rules, err := parsePolicyDetails(policy.Details, newSetTypeLookup(ctx, orgName))
	if err != nil {
		return infer.ReadResponse[RbacRoleInput, RbacRoleState]{}, unrepresentableRole(role.ID, err)
	}

	prior := req.Inputs.RbacRoleCore
	core := RbacRoleCore{
		OrganizationName:             orgName,
		Name:                         role.Name,
		Description:                  prior.Description,
		OrganizationPermissionSetIds: orgSetIDs,
		EntityRules:                  rules,
	}
	if role.Description != "" || prior.Description != nil {
		core.Description = &role.Description
	}
	// Keep the user's layout when it grants exactly what the service reports.
	if canonicalRoleKey(prior.OrganizationPermissionSetIds, prior.EntityRules) == canonicalRoleKey(orgSetIDs, rules) {
		core.OrganizationPermissionSetIds = prior.OrganizationPermissionSetIds
		core.EntityRules = prior.EntityRules
	}

	return infer.ReadResponse[RbacRoleInput, RbacRoleState]{
		ID:     req.ID,
		Inputs: RbacRoleInput{RbacRoleCore: core},
		State: RbacRoleState{
			RbacRoleCore: core,
			RoleId:       role.ID,
			PolicyId:     policyID,
			Version:      int(policy.Version),
		},
	}, nil
}

// rolePolicyID returns the single policy a console-style role composes.
func rolePolicyID(role *apitype.PermissionDescriptorRecord) (string, error) {
	compose, ok := role.Details.(apitype.PermissionDescriptorCompose)
	if !ok || len(compose.PermissionDescriptors()) != 1 {
		return "", unrepresentableRole(role.ID,
			fmt.Errorf("%w: role does not compose exactly one policy", errUnrepresentable))
	}
	return compose.PermissionDescriptors()[0], nil
}

func unrepresentableRole(roleID string, err error) error {
	if !errors.Is(err, errUnrepresentable) {
		return err
	}
	return fmt.Errorf("role %q uses permissions `RbacRole` cannot represent (%w). Manage it with "+
		"`OrganizationRole` or `pulumiservice:api:Role` instead", roleID, err)
}

// newSetTypeLookup resolves permission set IDs to resource types, listing the
// organization's sets once and falling back to a direct read for sets the
// listing omits. A set that no longer exists is treated as global so the
// dangling reference shows up as drift rather than failing the refresh.
func newSetTypeLookup(ctx context.Context, orgName string) setTypeLookup {
	client := config.GetClient(ctx)
	var types map[string]RbacResourceType
	return func(id string) (RbacResourceType, error) {
		if types == nil {
			sets, err := client.ListPermissionSets(ctx, orgName)
			if err != nil {
				return "", err
			}
			types = make(map[string]RbacResourceType, len(sets))
			for _, s := range sets {
				types[s.ID] = RbacResourceType(s.ResourceType)
			}
		}
		if rt, ok := types[id]; ok {
			return rt, nil
		}
		set, err := client.GetPermissionSet(ctx, orgName, id)
		if err != nil {
			return "", fmt.Errorf("failed to read permission set %q: %w", id, err)
		}
		rt := RbacResourceTypeGlobal
		if set != nil && set.ResourceType != "" {
			rt = RbacResourceType(set.ResourceType)
		}
		types[id] = rt
		return rt, nil
	}
}
