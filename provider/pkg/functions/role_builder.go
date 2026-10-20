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

// buildRolePermissions assembles a full `OrganizationRole.permissions` tree
// from organization-level grants and per-entity-type rules, mirroring the
// console's encoding (cmd/console2/src/app/settings in pulumi-service):
//
//	Group{entries: [
//	  Allow[<org-level scopes>..., <"all <entity>" scopes>...],            // unconditional
//	  Condition(Equal(Stack, LiteralStack{identity}), Allow[scopes]),      // one entity
//	  Condition(<tag conditions ANDed>, Allow[scopes]),                    // tag rule
//	  <additionalEntries>...,
//	]}
//
// Tag conditions are encoded by tagExpression, shared with the policy group
// match rules: equals is Equal(Tag{context, key}, LiteralString{value}),
// exists is HasTag{context, key}, and notEquals and notExists wrap those in
// Not; several conditions fold left into And.
//
// "All" rules are encoded without a condition, so they rely on every scope in
// them belonging to the rule's entity type: a stack scope in an
// unconditional grant applies to every stack. The per-type rule lists and
// scope enums below guarantee that.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// Entity types, as Pulumi Cloud names them in the scope catalog.
const (
	entityTypeStack           = "stack"
	entityTypeEnvironment     = "environment"
	entityTypeInsightsAccount = "insights-account"
)

// RoleRuleSelector chooses which entities a rule applies to.
type RoleRuleSelector struct {
	All  *bool                     `pulumi:"all,optional"`
	Id   *string                   `pulumi:"id,optional"` //nolint:revive // matches the SDK property name
	Tags []PolicyGroupTagCondition `pulumi:"tags,optional"`
}

func (s *RoleRuleSelector) Annotate(a infer.Annotator) {
	a.Describe(&s.All, "Apply to every entity of this type in the organization. Exactly one of `all`, `id`, "+
		"or `tags` must be set.")
	a.Describe(&s.Id, "Apply to a single entity: a stack's `stackId` (from `getStack`), an environment's "+
		"`environmentId`, or a cloud account's `insightsAccountId`.")
	a.Describe(&s.Tags, "Apply to entities whose tags match every condition.")
}

type RoleStackRule struct {
	RoleRuleSelector
	Scopes []RbacStackScope `pulumi:"scopes,optional"`
}

func (r *RoleStackRule) Annotate(a infer.Annotator) {
	a.Describe(&r.Scopes, "Stack scopes to grant on the selected stacks.")
}

type RoleEnvironmentRule struct {
	RoleRuleSelector
	Scopes []RbacEnvironmentScope `pulumi:"scopes,optional"`
}

func (r *RoleEnvironmentRule) Annotate(a infer.Annotator) {
	a.Describe(&r.Scopes, "Environment scopes to grant on the selected environments.")
}

type RoleAccountRule struct {
	RoleRuleSelector
	Scopes []RbacAccountScope `pulumi:"scopes,optional"`
}

func (r *RoleAccountRule) Annotate(a infer.Annotator) {
	a.Describe(&r.Scopes, "Cloud account scopes to grant on the selected accounts.")
}

type BuildRolePermissionsFunction struct{}

type BuildRolePermissionsInput struct {
	OrganizationScopes []RbacOrganizationScope `pulumi:"organizationScopes,optional"`
	StackRules         []RoleStackRule         `pulumi:"stackRules,optional"`
	EnvironmentRules   []RoleEnvironmentRule   `pulumi:"environmentRules,optional"`
	AccountRules       []RoleAccountRule       `pulumi:"accountRules,optional"`
	AdditionalEntries  []map[string]any        `pulumi:"additionalEntries,optional"`
}

type BuildRolePermissionsOutput struct {
	Permissions map[string]any `pulumi:"permissions"`
}

func (BuildRolePermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&BuildRolePermissionsFunction{},
		"Builds a complete `OrganizationRole.permissions` descriptor from organization-level access and "+
			"rules for stacks, environments, and cloud accounts (Discovery; `InsightsAccount` in the API and this "+
			"provider), the same model as the role editor in the "+
			"Pulumi Cloud console. Each rule grants scopes on every entity of its type, on one specific entity, "+
			"or on entities matching tag conditions. Each rule list accepts only scopes for its own entity "+
			"type. The result is directly assignable to `OrganizationRole.permissions`.",
	)
	a.SetToken("index", "buildRolePermissions")
}

func (i *BuildRolePermissionsInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationScopes, "Organization-level scopes to grant (e.g. `team:read`).")
	a.Describe(&i.StackRules, "Grants on stacks.")
	a.Describe(&i.EnvironmentRules, "Grants on ESC environments.")
	a.Describe(&i.AccountRules, "Grants on cloud accounts (`InsightsAccount`).")
	a.Describe(&i.AdditionalEntries, "Output of the other permission helpers to include in the role.")
}

func (o *BuildRolePermissionsOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.Permissions, "A `PermissionDescriptorGroup` combining every grant, ready to assign to "+
		"`OrganizationRole.permissions`.")
}

func (BuildRolePermissionsFunction) Invoke(
	_ context.Context,
	req infer.FunctionRequest[BuildRolePermissionsInput],
) (infer.FunctionResponse[BuildRolePermissionsOutput], error) {
	descriptor, err := buildRoleDescriptor(req.Input)
	if err != nil {
		return infer.FunctionResponse[BuildRolePermissionsOutput]{}, err
	}
	out, err := toSDKMap(descriptor)
	if err != nil {
		return infer.FunctionResponse[BuildRolePermissionsOutput]{}, err
	}
	return infer.FunctionResponse[BuildRolePermissionsOutput]{
		Output: BuildRolePermissionsOutput{Permissions: out},
	}, nil
}

// rule is a stack, environment, or cloud account rule with its scopes
// checked against its entity type.
type rule struct {
	path       string // e.g. "stackRules[2]", for error messages
	entityType string
	sel        RoleRuleSelector
	scopes     []string
}

func buildRoleDescriptor(in BuildRolePermissionsInput) (apitype.PermissionDescriptor, error) {
	if len(in.OrganizationScopes) == 0 && len(in.StackRules) == 0 && len(in.EnvironmentRules) == 0 &&
		len(in.AccountRules) == 0 && len(in.AdditionalEntries) == 0 {
		return nil, errors.New("at least one of `organizationScopes`, `stackRules`, `environmentRules`, " +
			"`accountRules`, or `additionalEntries` must be set")
	}

	rules, err := collectRules(in)
	if err != nil {
		return nil, err
	}

	unconditionalScopes, err := checkScopes(in.OrganizationScopes, rbacOrganizationScopeValues, "an organization")
	if err != nil {
		return nil, fmt.Errorf("organizationScopes: %w", err)
	}
	var conditional []apitype.PermissionDescriptor
	for _, r := range rules {
		if len(r.scopes) == 0 {
			return nil, fmt.Errorf("%s: `scopes` must not be empty", r.path)
		}
		if r.sel.All != nil && *r.sel.All {
			if r.sel.Id != nil || len(r.sel.Tags) > 0 {
				return nil, fmt.Errorf("%s: exactly one of `all`, `id`, or `tags` must be set", r.path)
			}
			unconditionalScopes = append(unconditionalScopes, r.scopes...)
			continue
		}
		cond, err := selectorCondition(r.entityType, &r.sel)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", r.path, err)
		}
		grant, err := allowDescriptor(dedupe(r.scopes))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", r.path, err)
		}
		conditional = append(conditional, apitype.PermissionDescriptorConditionBuilder{
			Condition: cond,
			SubNode:   grant,
		}.Build())
	}

	var entries []apitype.PermissionDescriptor
	if len(unconditionalScopes) > 0 {
		allow, err := allowDescriptor(dedupe(unconditionalScopes))
		if err != nil {
			return nil, err
		}
		entries = append(entries, allow)
	}
	entries = append(entries, conditional...)
	for i, entry := range in.AdditionalEntries {
		d, err := parseDescriptor(entry)
		if err != nil {
			return nil, fmt.Errorf("additionalEntries[%d]: %w", i, err)
		}
		entries = append(entries, d)
	}
	return apitype.PermissionDescriptorGroupBuilder{Entries: entries}.Build(), nil
}

// collectRules flattens the three rule lists, checking each rule's scopes
// against its entity type's enum.
func collectRules(in BuildRolePermissionsInput) ([]rule, error) {
	var rules []rule
	for i, r := range in.StackRules {
		scopes, err := checkScopes(r.Scopes, rbacStackScopeValues, "a stack")
		if err != nil {
			return nil, fmt.Errorf("stackRules[%d].scopes: %w", i, err)
		}
		rules = append(rules, rule{fmt.Sprintf("stackRules[%d]", i), entityTypeStack,
			r.RoleRuleSelector, scopes})
	}
	for i, r := range in.EnvironmentRules {
		scopes, err := checkScopes(r.Scopes, rbacEnvironmentScopeValues, "an environment")
		if err != nil {
			return nil, fmt.Errorf("environmentRules[%d].scopes: %w", i, err)
		}
		rules = append(rules, rule{fmt.Sprintf("environmentRules[%d]", i), entityTypeEnvironment,
			r.RoleRuleSelector, scopes})
	}
	for i, r := range in.AccountRules {
		scopes, err := checkScopes(r.Scopes, rbacAccountScopeValues, "a cloud account")
		if err != nil {
			return nil, fmt.Errorf("accountRules[%d].scopes: %w", i, err)
		}
		rules = append(rules, rule{fmt.Sprintf("accountRules[%d]", i), entityTypeInsightsAccount,
			r.RoleRuleSelector, scopes})
	}
	return rules, nil
}

// entityContext returns the expression that refers to the entity being
// authorized, e.g. "the stack".
func entityContext(entityType string) apitype.PermissionContextExpression {
	switch entityType {
	case entityTypeEnvironment:
		return apitype.PermissionExpressionEnvironmentBuilder{}.Build()
	case entityTypeInsightsAccount:
		return apitype.PermissionExpressionInsightsAccountBuilder{}.Build()
	default:
		return apitype.PermissionExpressionStackBuilder{}.Build()
	}
}

// entityLiteral returns the expression for one specific entity.
func entityLiteral(entityType, id string) apitype.PermissionExpression {
	switch entityType {
	case entityTypeEnvironment:
		return apitype.PermissionLiteralExpressionEnvironmentBuilder{Identity: id}.Build()
	case entityTypeInsightsAccount:
		return apitype.PermissionLiteralExpressionInsightsAccountBuilder{Identity: id}.Build()
	default:
		return apitype.PermissionLiteralExpressionStackBuilder{Identity: id}.Build()
	}
}

func selectorCondition(entityType string, sel *RoleRuleSelector) (apitype.PermissionBooleanExpression, error) {
	switch {
	case sel.Id != nil && len(sel.Tags) == 0:
		if *sel.Id == "" {
			return nil, errors.New("`id` must not be empty")
		}
		return apitype.PermissionExpressionEqualBuilder{
			Left:  entityContext(entityType),
			Right: entityLiteral(entityType, *sel.Id),
		}.Build(), nil
	case sel.Id == nil && len(sel.Tags) > 0:
	default:
		return nil, errors.New("exactly one of `all`, `id`, or `tags` must be set")
	}

	conditions := make([]apitype.PermissionBooleanExpression, len(sel.Tags))
	for i, tc := range sel.Tags {
		expr, err := tagExpression(entityContext(entityType), tc)
		if err != nil {
			return nil, fmt.Errorf("tags[%d]: %w", i, err)
		}
		conditions[i] = expr
	}
	return foldExpressions(conditions, andExpression), nil
}

func allowDescriptor(scopes []string) (apitype.PermissionDescriptor, error) {
	slice, err := rbacPermissionSlice(scopes)
	if err != nil {
		return nil, err
	}
	return apitype.PermissionDescriptorAllowBuilder{Permissions: slice}.Build(), nil
}

// parseDescriptor decodes an SDK-boundary descriptor map, as returned by
// the other helpers, with the same typed unmarshaller OrganizationRole uses.
func parseDescriptor(m map[string]any) (apitype.PermissionDescriptor, error) {
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshalling descriptor: %w", err)
	}
	var d apitype.PermissionDescriptor
	if err := apitype.UnmarshalJSONPermissionDescriptor(raw, &d); err != nil {
		return nil, fmt.Errorf("parsing descriptor: %w", err)
	}
	if d == nil {
		return nil, errors.New("descriptor parsed to nil — missing or unknown __type")
	}
	return d, nil
}

// dedupe drops repeated values, keeping first occurrences in order.
func dedupe[T comparable](values []T) []T {
	seen := make(map[T]bool, len(values))
	out := make([]T, 0, len(values))
	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
