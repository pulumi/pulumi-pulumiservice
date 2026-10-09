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
//	  Compose[<org-level set ids>..., <"all <entity>" set ids>...],        // unconditional
//	  Allow[<org-level scopes>..., <"all <entity>" scopes>...],            // unconditional
//	  Condition(Equal(Stack, LiteralStack{identity}), <grant>),            // one entity
//	  Condition(<tag conditions ANDed>, <grant>),                          // tag rule
//	  <additionalEntries>...,
//	]}
//
// where <grant> is Compose[set ids], Allow[scopes], or a Group of both.
// Tag conditions: Equal(Tag{context, key}, LiteralString{value}); an empty
// value is HasTag{context, key}; notEquals wraps either in Not; several
// conditions fold left into And.
//
// "All" rules are encoded without a condition, so they rely on every scope
// and permission set in them belonging to the rule's entity type: a stack
// scope in an unconditional grant applies to every stack. The per-type rule
// lists, scope enums, and permission-set checks below guarantee that.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// Entity types, as Pulumi Cloud names them in a permission set's
// `resourceType` and in the scope catalog.
const (
	entityTypeOrganization    = "global"
	entityTypeStack           = "stack"
	entityTypeEnvironment     = "environment"
	entityTypeInsightsAccount = "insights-account"
)

// RoleTagOperator is how a RoleTagCondition compares a tag.
type RoleTagOperator string

const (
	RoleTagOperatorEquals    RoleTagOperator = "equals"
	RoleTagOperatorNotEquals RoleTagOperator = "notEquals"
)

func (RoleTagOperator) Values() []infer.EnumValue[RoleTagOperator] {
	return []infer.EnumValue[RoleTagOperator]{
		{
			Name:        "Equals",
			Value:       RoleTagOperatorEquals,
			Description: "Match entities whose tag equals the value (or that have the tag, when no value is set).",
		},
		{
			Name:        "NotEquals",
			Value:       RoleTagOperatorNotEquals,
			Description: "Match entities whose tag does not equal the value (or that lack the tag, when no value is set).",
		},
	}
}

type RoleTagCondition struct {
	Key      string           `pulumi:"key"`
	Value    *string          `pulumi:"value,optional"`
	Operator *RoleTagOperator `pulumi:"operator,optional"`
}

func (c *RoleTagCondition) Annotate(a infer.Annotator) {
	a.Describe(&c.Key, "The tag key.")
	a.Describe(&c.Value, "The tag value. When omitted, the condition matches on whether the tag is present.")
	a.Describe(&c.Operator, "How the tag is compared. Defaults to `equals`.")
}

// RolePermissionSetRef identifies a permission set together with its entity
// type, so buildRolePermissions can check that it fits the rule it's in
// without calling Pulumi Cloud. The output of getOrganizationPermissionSet
// carries both fields.
type RolePermissionSetRef struct {
	PermissionSetId string `pulumi:"permissionSetId"` //nolint:revive // matches the SDK property name
	ResourceType    string `pulumi:"resourceType"`
}

func (r *RolePermissionSetRef) Annotate(a infer.Annotator) {
	a.Describe(&r.PermissionSetId, "The permission set's ID.")
	a.Describe(&r.ResourceType, "The permission set's entity type: `stack`, `environment`, "+
		"`insights-account`, or `global` for organization-level sets. It must match the rule list the set "+
		"is used in.")
}

// RoleRuleSelector chooses which entities a rule applies to.
type RoleRuleSelector struct {
	All  *bool              `pulumi:"all,optional"`
	Id   *string            `pulumi:"id,optional"` //nolint:revive // matches the SDK property name
	Tags []RoleTagCondition `pulumi:"tags,optional"`
}

func (s *RoleRuleSelector) Annotate(a infer.Annotator) {
	a.Describe(&s.All, "Apply to every entity of this type in the organization. Exactly one of `all`, `id`, "+
		"or `tags` must be set.")
	a.Describe(&s.Id, "Apply to a single entity: a stack's `stackId` (from `getStack`), an environment's "+
		"`environmentId`, or an Insights account's `insightsAccountId`.")
	a.Describe(&s.Tags, "Apply to entities whose tags match every condition.")
}

const permissionSetsDoc = "Permission sets to grant, typically the output of `getOrganizationPermissionSet`. " +
	"Each set's `resourceType` must match this rule's entity type. Pulumi Cloud accepts permission sets " +
	"only in a policy: see `buildRolePermissions`."

type RoleStackRule struct {
	RoleRuleSelector
	Scopes         []RbacStackScope       `pulumi:"scopes,optional"`
	PermissionSets []RolePermissionSetRef `pulumi:"permissionSets,optional"`
}

func (r *RoleStackRule) Annotate(a infer.Annotator) {
	a.Describe(&r.Scopes, "Stack scopes to grant on the selected stacks.")
	a.Describe(&r.PermissionSets, permissionSetsDoc)
}

type RoleEnvironmentRule struct {
	RoleRuleSelector
	Scopes         []RbacEnvironmentScope `pulumi:"scopes,optional"`
	PermissionSets []RolePermissionSetRef `pulumi:"permissionSets,optional"`
}

func (r *RoleEnvironmentRule) Annotate(a infer.Annotator) {
	a.Describe(&r.Scopes, "Environment scopes to grant on the selected environments.")
	a.Describe(&r.PermissionSets, permissionSetsDoc)
}

type RoleInsightsAccountRule struct {
	RoleRuleSelector
	Scopes         []RbacInsightsAccountScope `pulumi:"scopes,optional"`
	PermissionSets []RolePermissionSetRef     `pulumi:"permissionSets,optional"`
}

func (r *RoleInsightsAccountRule) Annotate(a infer.Annotator) {
	a.Describe(&r.Scopes, "Insights account scopes to grant on the selected accounts.")
	a.Describe(&r.PermissionSets, permissionSetsDoc)
}

type BuildRolePermissionsFunction struct{}

type BuildRolePermissionsInput struct {
	OrganizationScopes         []RbacOrganizationScope   `pulumi:"organizationScopes,optional"`
	OrganizationPermissionSets []RolePermissionSetRef    `pulumi:"organizationPermissionSets,optional"`
	StackRules                 []RoleStackRule           `pulumi:"stackRules,optional"`
	EnvironmentRules           []RoleEnvironmentRule     `pulumi:"environmentRules,optional"`
	InsightsAccountRules       []RoleInsightsAccountRule `pulumi:"insightsAccountRules,optional"`
	AdditionalEntries          []map[string]any          `pulumi:"additionalEntries,optional"`
}

type BuildRolePermissionsOutput struct {
	Permissions map[string]any `pulumi:"permissions"`
}

func (BuildRolePermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&BuildRolePermissionsFunction{},
		"Builds a complete `OrganizationRole.permissions` descriptor from organization-level access and "+
			"rules for stacks, environments, and Insights accounts, the same model as the role editor in the "+
			"Pulumi Cloud console. Each rule grants scopes and/or permission sets on every entity of its type, "+
			"on one specific entity, or on entities matching tag conditions. Each rule list accepts only scopes "+
			"and permission sets for its own entity type. Use `additionalEntries` to fold in the output of the "+
			"other `build*Permissions` helpers.\n\n"+
			"A result that uses only scopes is directly assignable to `OrganizationRole.permissions`. Pulumi "+
			"Cloud accepts permission sets only in a policy: store such a result with `pulumiservice:api:Role` "+
			"(`uxPurpose: policy`, `details: <result>`) and grant the policy to the role with "+
			"`buildComposePermissions`.",
	)
	a.SetToken("index", "buildRolePermissions")
}

func (i *BuildRolePermissionsInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationScopes, "Organization-level scopes to grant (e.g. `team:read`).")
	a.Describe(&i.OrganizationPermissionSets, "Organization-level (`global`) permission sets to grant, such "+
		"as the built-in `org-settings-read-only` set from `getOrganizationPermissionSet`.")
	a.Describe(&i.StackRules, "Grants on stacks.")
	a.Describe(&i.EnvironmentRules, "Grants on ESC environments.")
	a.Describe(&i.InsightsAccountRules, "Grants on Insights accounts.")
	a.Describe(&i.AdditionalEntries, "Further descriptors to include in the role, typically the "+
		"`permissions` output of `buildAllowPermissions`, `buildStackScopedPermissions`, "+
		"`buildEnvironmentScopedPermissions`, or `buildInsightsAccountScopedPermissions`.")
}

func (o *BuildRolePermissionsOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.Permissions, "A `PermissionDescriptorGroup` combining every grant, ready to assign to "+
		"`OrganizationRole.permissions` (or, when it uses permission sets, to a policy's `details`).")
}

func (BuildRolePermissionsFunction) Invoke(
	_ context.Context,
	req infer.FunctionRequest[BuildRolePermissionsInput],
) (infer.FunctionResponse[BuildRolePermissionsOutput], error) {
	descriptor, err := buildRoleDescriptor(req.Input)
	if err != nil {
		return infer.FunctionResponse[BuildRolePermissionsOutput]{}, err
	}
	out, err := descriptorToSDKMap(descriptor)
	if err != nil {
		return infer.FunctionResponse[BuildRolePermissionsOutput]{}, err
	}
	return infer.FunctionResponse[BuildRolePermissionsOutput]{
		Output: BuildRolePermissionsOutput{Permissions: out},
	}, nil
}

// rule is a stack, environment, or Insights account rule with its scopes
// checked against its entity type.
type rule struct {
	path       string // e.g. "stackRules[2]", for error messages
	entityType string
	sel        RoleRuleSelector
	scopes     []string
	sets       []RolePermissionSetRef
}

func buildRoleDescriptor(in BuildRolePermissionsInput) (apitype.PermissionDescriptor, error) {
	if len(in.OrganizationScopes) == 0 && len(in.OrganizationPermissionSets) == 0 &&
		len(in.StackRules) == 0 && len(in.EnvironmentRules) == 0 && len(in.InsightsAccountRules) == 0 &&
		len(in.AdditionalEntries) == 0 {
		return nil, errors.New("at least one of `organizationScopes`, `organizationPermissionSets`, " +
			"`stackRules`, `environmentRules`, `insightsAccountRules`, or `additionalEntries` must be set")
	}

	rules, err := collectRules(in)
	if err != nil {
		return nil, err
	}

	unconditionalScopes, err := checkScopes(in.OrganizationScopes, rbacOrganizationScopeValues, "an organization")
	if err != nil {
		return nil, fmt.Errorf("organizationScopes: %w", err)
	}
	unconditionalSets, err := checkSets(in.OrganizationPermissionSets, entityTypeOrganization)
	if err != nil {
		return nil, fmt.Errorf("organizationPermissionSets: %w", err)
	}
	var conditional []apitype.PermissionDescriptor
	for _, r := range rules {
		if len(r.scopes) == 0 && len(r.sets) == 0 {
			return nil, fmt.Errorf("%s: at least one of `scopes` or `permissionSets` must be set", r.path)
		}
		sets, err := checkSets(r.sets, r.entityType)
		if err != nil {
			return nil, fmt.Errorf("%s.permissionSets: %w", r.path, err)
		}
		if r.sel.All != nil && *r.sel.All {
			if r.sel.Id != nil || len(r.sel.Tags) > 0 {
				return nil, fmt.Errorf("%s: exactly one of `all`, `id`, or `tags` must be set", r.path)
			}
			unconditionalScopes = append(unconditionalScopes, r.scopes...)
			unconditionalSets = append(unconditionalSets, sets...)
			continue
		}
		cond, err := selectorCondition(r.entityType, &r.sel)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", r.path, err)
		}
		grant, err := grantDescriptor(sets, r.scopes)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", r.path, err)
		}
		conditional = append(conditional, apitype.PermissionDescriptorConditionBuilder{
			Condition: cond,
			SubNode:   grant,
		}.Build())
	}

	var entries []apitype.PermissionDescriptor
	if len(unconditionalSets) > 0 {
		entries = append(entries, compose(dedupe(unconditionalSets)))
	}
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
			r.RoleRuleSelector, scopes, r.PermissionSets})
	}
	for i, r := range in.EnvironmentRules {
		scopes, err := checkScopes(r.Scopes, rbacEnvironmentScopeValues, "an environment")
		if err != nil {
			return nil, fmt.Errorf("environmentRules[%d].scopes: %w", i, err)
		}
		rules = append(rules, rule{fmt.Sprintf("environmentRules[%d]", i), entityTypeEnvironment,
			r.RoleRuleSelector, scopes, r.PermissionSets})
	}
	for i, r := range in.InsightsAccountRules {
		scopes, err := checkScopes(r.Scopes, rbacInsightsAccountScopeValues, "an Insights account")
		if err != nil {
			return nil, fmt.Errorf("insightsAccountRules[%d].scopes: %w", i, err)
		}
		rules = append(rules, rule{fmt.Sprintf("insightsAccountRules[%d]", i), entityTypeInsightsAccount,
			r.RoleRuleSelector, scopes, r.PermissionSets})
	}
	return rules, nil
}

// checkSets returns the sets' IDs, or an error if any set is for a different
// entity type than entityType.
func checkSets(sets []RolePermissionSetRef, entityType string) ([]string, error) {
	ids := make([]string, len(sets))
	for i, s := range sets {
		if strings.TrimSpace(s.PermissionSetId) == "" {
			return nil, fmt.Errorf("[%d]: `permissionSetId` must not be empty", i)
		}
		if s.ResourceType != entityType {
			return nil, fmt.Errorf("[%d]: permission set %q is for `%s`, not `%s`; each rule list accepts only "+
				"permission sets for its own entity type", i, s.PermissionSetId, s.ResourceType, entityType)
		}
		ids[i] = s.PermissionSetId
	}
	return ids, nil
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

	var acc apitype.PermissionBooleanExpression
	for i, tc := range sel.Tags {
		if strings.TrimSpace(tc.Key) == "" {
			return nil, fmt.Errorf("tags[%d]: `key` must not be empty", i)
		}
		var leaf apitype.PermissionBooleanExpression
		if tc.Value == nil || *tc.Value == "" {
			leaf = apitype.PermissionExpressionHasTagBuilder{Context: entityContext(entityType), Key: tc.Key}.Build()
		} else {
			leaf = apitype.PermissionExpressionEqualBuilder{
				Left:  apitype.PermissionExpressionTagBuilder{Context: entityContext(entityType), Key: tc.Key}.Build(),
				Right: apitype.PermissionLiteralExpressionStringBuilder{Value: *tc.Value}.Build(),
			}.Build()
		}
		if tc.Operator != nil {
			switch *tc.Operator {
			case RoleTagOperatorEquals:
			case RoleTagOperatorNotEquals:
				leaf = apitype.PermissionExpressionNotBuilder{
					PermissionBooleanExpressionUnaryBuilder: apitype.PermissionBooleanExpressionUnaryBuilder{
						Node: leaf,
					},
				}.Build()
			default:
				return nil, fmt.Errorf("tags[%d]: unknown operator %q; use `equals` or `notEquals`", i, *tc.Operator)
			}
		}
		if acc == nil {
			acc = leaf
		} else {
			acc = apitype.PermissionExpressionAndBuilder{
				PermissionBooleanExpressionBinaryBuilder: apitype.PermissionBooleanExpressionBinaryBuilder{
					Left: acc, Right: leaf,
				},
			}.Build()
		}
	}
	return acc, nil
}

// grantDescriptor returns what a conditional rule grants: a Compose of
// permission sets, an Allow of scopes, or a Group of both.
func grantDescriptor(setIDs, scopes []string) (apitype.PermissionDescriptor, error) {
	var parts []apitype.PermissionDescriptor
	if len(setIDs) > 0 {
		parts = append(parts, compose(dedupe(setIDs)))
	}
	if len(scopes) > 0 {
		allow, err := allowDescriptor(dedupe(scopes))
		if err != nil {
			return nil, err
		}
		parts = append(parts, allow)
	}
	if len(parts) == 1 {
		return parts[0], nil
	}
	return apitype.PermissionDescriptorGroupBuilder{Entries: parts}.Build(), nil
}

func compose(ids []string) apitype.PermissionDescriptor {
	return apitype.PermissionDescriptorComposeBuilder{PermissionDescriptors: ids}.Build()
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

type BuildComposePermissionsFunction struct{}

type BuildComposePermissionsInput struct {
	Ids []string `pulumi:"ids"` //nolint:revive // matches the SDK property name
}

type BuildComposePermissionsOutput struct {
	Permissions map[string]any `pulumi:"permissions"`
}

func (BuildComposePermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&BuildComposePermissionsFunction{},
		"Builds a permission descriptor that grants everything the referenced descriptors grant. Use it "+
			"to grant a policy, such as one created with `pulumiservice:api:Role` (`uxPurpose: policy`) whose "+
			"`details` come from `buildRolePermissions`, to an `OrganizationRole`. The result is directly "+
			"assignable to `OrganizationRole.permissions`.",
	)
	a.SetToken("index", "buildComposePermissions")
}

func (i *BuildComposePermissionsInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Ids, "The IDs of the descriptors to grant, such as a policy's `roleID`.")
}

func (o *BuildComposePermissionsOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.Permissions, "A `PermissionDescriptorCompose` referencing the supplied IDs.")
}

func (BuildComposePermissionsFunction) Invoke(
	_ context.Context,
	req infer.FunctionRequest[BuildComposePermissionsInput],
) (infer.FunctionResponse[BuildComposePermissionsOutput], error) {
	if len(req.Input.Ids) == 0 {
		return infer.FunctionResponse[BuildComposePermissionsOutput]{}, errors.New("`ids` must not be empty")
	}
	for i, id := range req.Input.Ids {
		if strings.TrimSpace(id) == "" {
			return infer.FunctionResponse[BuildComposePermissionsOutput]{},
				fmt.Errorf("ids[%d] must not be empty", i)
		}
	}
	out, err := descriptorToSDKMap(compose(dedupe(req.Input.Ids)))
	if err != nil {
		return infer.FunctionResponse[BuildComposePermissionsOutput]{}, err
	}
	return infer.FunctionResponse[BuildComposePermissionsOutput]{
		Output: BuildComposePermissionsOutput{Permissions: out},
	}, nil
}
