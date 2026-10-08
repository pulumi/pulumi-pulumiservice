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

// This file converts between the typed RbacRole model and the permission
// descriptor tree the Pulumi Cloud console writes for a custom role's policy.
// The encodings mirror the console (cmd/console2/src/app/settings in
// pulumi-service) so roles created here look and edit like console roles:
//
//	Group{entries: [
//	  Compose[<org-level set ids>..., <"all <entity>" set ids>...],     // unconditional
//	  Condition(Equal(Stack, LiteralStack{identity}), Compose[<set ids>]), // one entity
//	  Condition(<tag conditions ANDed>, Compose[<set ids>]),              // tag rule
//	]}
//
// Tag conditions: Equal(Tag{context, key}, LiteralString{value}); an empty
// value is HasTag{context, key}; notEquals wraps either in Not; several
// conditions fold left into And.

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

// errUnrepresentable marks a descriptor tree that RbacRole's typed model
// cannot express, typically one built with `pulumiservice:api:Role` or
// `OrganizationRole`.
var errUnrepresentable = errors.New("permission descriptor cannot be represented by RbacRole")

// entityKinds lists the entity selector kinds in the order rules are emitted
// and "all" sets are grouped.
var entityKinds = []RbacResourceType{
	RbacResourceTypeStack,
	RbacResourceTypeEnvironment,
	RbacResourceTypeInsightsAccount,
}

// entityContext returns the expression that refers to the entity being
// authorized, e.g. "the stack".
func entityContext(rt RbacResourceType) apitype.PermissionContextExpression {
	switch rt {
	case RbacResourceTypeEnvironment:
		return apitype.PermissionExpressionEnvironmentBuilder{}.Build()
	case RbacResourceTypeInsightsAccount:
		return apitype.PermissionExpressionInsightsAccountBuilder{}.Build()
	default:
		return apitype.PermissionExpressionStackBuilder{}.Build()
	}
}

// entityLiteral returns the expression for one specific entity.
func entityLiteral(rt RbacResourceType, id string) apitype.PermissionExpression {
	switch rt {
	case RbacResourceTypeEnvironment:
		return apitype.PermissionLiteralExpressionEnvironmentBuilder{Identity: id}.Build()
	case RbacResourceTypeInsightsAccount:
		return apitype.PermissionLiteralExpressionInsightsAccountBuilder{Identity: id}.Build()
	default:
		return apitype.PermissionLiteralExpressionStackBuilder{Identity: id}.Build()
	}
}

// contextResourceType is the inverse of entityContext.
func contextResourceType(e apitype.PermissionExpression) (RbacResourceType, bool) {
	switch e.(type) {
	case apitype.PermissionExpressionStack:
		return RbacResourceTypeStack, true
	case apitype.PermissionExpressionEnvironment:
		return RbacResourceTypeEnvironment, true
	case apitype.PermissionExpressionInsightsAccount:
		return RbacResourceTypeInsightsAccount, true
	}
	return "", false
}

// literalEntity is the inverse of entityLiteral.
func literalEntity(e apitype.PermissionExpression) (RbacResourceType, string, bool) {
	switch lit := e.(type) {
	case apitype.PermissionLiteralExpressionStack:
		return RbacResourceTypeStack, lit.Identity(), true
	case apitype.PermissionLiteralExpressionEnvironment:
		return RbacResourceTypeEnvironment, lit.Identity(), true
	case apitype.PermissionLiteralExpressionInsightsAccount:
		return RbacResourceTypeInsightsAccount, lit.Identity(), true
	}
	return "", "", false
}

func compose(ids []string) apitype.PermissionDescriptorCompose {
	return apitype.PermissionDescriptorComposeBuilder{PermissionDescriptors: ids}.Build()
}

// selector returns the rule's entity type and selector. Check guarantees
// exactly one is set before this is called.
func (r RbacEntityRule) selector() (RbacResourceType, *RbacEntitySelector) {
	switch {
	case r.Stack != nil:
		return RbacResourceTypeStack, r.Stack
	case r.Environment != nil:
		return RbacResourceTypeEnvironment, r.Environment
	case r.InsightsAccount != nil:
		return RbacResourceTypeInsightsAccount, r.InsightsAccount
	}
	return "", nil
}

func ruleFor(rt RbacResourceType, sel *RbacEntitySelector, setIDs []string) RbacEntityRule {
	r := RbacEntityRule{PermissionSetIds: setIDs}
	switch rt {
	case RbacResourceTypeStack:
		r.Stack = sel
	case RbacResourceTypeEnvironment:
		r.Environment = sel
	case RbacResourceTypeInsightsAccount:
		r.InsightsAccount = sel
	}
	return r
}

// buildPolicyDetails returns the policy descriptor tree for a role.
func buildPolicyDetails(in RbacRoleCore) (apitype.PermissionDescriptor, error) {
	unconditional := slices.Clone(in.OrganizationPermissionSetIds)
	var conditional []apitype.PermissionDescriptor
	for i, rule := range in.EntityRules {
		rt, sel := rule.selector()
		if sel == nil {
			return nil, fmt.Errorf("entityRules[%d]: one of stack, environment, or insightsAccount must be set", i)
		}
		if sel.All != nil && *sel.All {
			unconditional = append(unconditional, rule.PermissionSetIds...)
			continue
		}
		cond, err := buildSelectorCondition(rt, sel)
		if err != nil {
			return nil, fmt.Errorf("entityRules[%d]: %w", i, err)
		}
		conditional = append(conditional, apitype.PermissionDescriptorConditionBuilder{
			Condition: cond,
			SubNode:   compose(rule.PermissionSetIds),
		}.Build())
	}

	var entries []apitype.PermissionDescriptor
	if len(unconditional) > 0 {
		entries = append(entries, compose(dedupe(unconditional)))
	}
	entries = append(entries, conditional...)
	return apitype.PermissionDescriptorGroupBuilder{Entries: entries}.Build(), nil
}

func buildSelectorCondition(rt RbacResourceType, sel *RbacEntitySelector) (apitype.PermissionBooleanExpression, error) {
	if sel.Id != nil {
		return apitype.PermissionExpressionEqualBuilder{
			Left:  entityContext(rt),
			Right: entityLiteral(rt, *sel.Id),
		}.Build(), nil
	}
	if len(sel.Tags) == 0 {
		return nil, errors.New("one of id, tags, or all must be set")
	}

	var acc apitype.PermissionBooleanExpression
	for _, tc := range sel.Tags {
		var leaf apitype.PermissionBooleanExpression
		if v := strings.TrimSpace(util.OrZero(tc.Value)); v == "" {
			leaf = apitype.PermissionExpressionHasTagBuilder{Context: entityContext(rt), Key: tc.Key}.Build()
		} else {
			leaf = apitype.PermissionExpressionEqualBuilder{
				Left:  apitype.PermissionExpressionTagBuilder{Context: entityContext(rt), Key: tc.Key}.Build(),
				Right: apitype.PermissionLiteralExpressionStringBuilder{Value: *tc.Value}.Build(),
			}.Build()
		}
		if tc.Operator != nil && *tc.Operator == RbacTagOperatorNotEquals {
			leaf = apitype.PermissionExpressionNotBuilder{
				PermissionBooleanExpressionUnaryBuilder: apitype.PermissionBooleanExpressionUnaryBuilder{Node: leaf},
			}.Build()
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

// setTypeLookup resolves a permission set ID to its resource type.
type setTypeLookup func(id string) (RbacResourceType, error)

// typeName names a descriptor or expression node for error messages.
func typeName(n interface{ GetDiscriminatorValue() (string, error) }) string {
	if n == nil {
		return "<nil>"
	}
	name, _ := n.GetDiscriminatorValue()
	return name
}

// parsePolicyDetails is the inverse of buildPolicyDetails. Unconditional set
// references are split by the referenced set's resource type: global sets
// are organization-level access, the rest are "all <entity>" rules.
func parsePolicyDetails(
	details apitype.PermissionDescriptor, lookup setTypeLookup,
) (orgSetIDs []string, rules []RbacEntityRule, err error) {
	group, ok := details.(apitype.PermissionDescriptorGroup)
	if !ok {
		return nil, nil, fmt.Errorf("%w: policy is a %s, want PermissionDescriptorGroup",
			errUnrepresentable, typeName(details))
	}

	allSets := map[RbacResourceType][]string{}
	for i, entry := range group.Entries() {
		switch entry := entry.(type) {
		case apitype.PermissionDescriptorCompose:
			for _, id := range entry.PermissionDescriptors() {
				rt, err := lookup(id)
				if err != nil {
					return nil, nil, err
				}
				if rt == RbacResourceTypeGlobal {
					orgSetIDs = append(orgSetIDs, id)
				} else {
					allSets[rt] = append(allSets[rt], id)
				}
			}
		case apitype.PermissionDescriptorCondition:
			sub, ok := entry.SubNode().(apitype.PermissionDescriptorCompose)
			if !ok {
				return nil, nil, fmt.Errorf("%w: entry %d grants a %s instead of permission sets",
					errUnrepresentable, i, typeName(entry.SubNode()))
			}
			rt, sel, err := parseCondition(entry.Condition())
			if err != nil {
				return nil, nil, fmt.Errorf("entry %d: %w", i, err)
			}
			rules = append(rules, ruleFor(rt, sel, sub.PermissionDescriptors()))
		default:
			return nil, nil, fmt.Errorf("%w: entry %d is a %s", errUnrepresentable, i, typeName(entry))
		}
	}

	var allRules []RbacEntityRule
	for _, rt := range entityKinds {
		if ids := allSets[rt]; len(ids) > 0 {
			allRules = append(allRules, ruleFor(rt, &RbacEntitySelector{All: pointerTo(true)}, ids))
		}
	}
	return orgSetIDs, append(allRules, rules...), nil
}

func parseCondition(c apitype.PermissionBooleanExpression) (RbacResourceType, *RbacEntitySelector, error) {
	// A single entity: Equal(<Entity>, Literal<Entity>{identity}).
	if eq, ok := c.(apitype.PermissionExpressionEqual); ok {
		if rt, id, ok := literalEntity(eq.Right()); ok {
			if ctxRT, ok := contextResourceType(eq.Left()); !ok || ctxRT != rt {
				return "", nil, fmt.Errorf("%w: unsupported entity condition %s = %s",
					errUnrepresentable, typeName(eq.Left()), typeName(eq.Right()))
			}
			return rt, &RbacEntitySelector{Id: pointerTo(id)}, nil
		}
	}

	var leaves []apitype.PermissionBooleanExpression
	var flatten func(apitype.PermissionBooleanExpression)
	flatten = func(x apitype.PermissionBooleanExpression) {
		if and, ok := x.(apitype.PermissionExpressionAnd); ok {
			flatten(and.Left())
			flatten(and.Right())
			return
		}
		leaves = append(leaves, x)
	}
	flatten(c)

	var rt RbacResourceType
	tags := make([]RbacTagCondition, 0, len(leaves))
	for _, leaf := range leaves {
		leafRT, tc, err := parseTagCondition(leaf)
		if err != nil {
			return "", nil, err
		}
		if rt != "" && leafRT != rt {
			return "", nil, fmt.Errorf("%w: tag conditions mix %s and %s", errUnrepresentable, rt, leafRT)
		}
		rt = leafRT
		tags = append(tags, tc)
	}
	return rt, &RbacEntitySelector{Tags: tags}, nil
}

func parseTagCondition(n apitype.PermissionBooleanExpression) (RbacResourceType, RbacTagCondition, error) {
	var tc RbacTagCondition
	if not, ok := n.(apitype.PermissionExpressionNot); ok {
		tc.Operator = pointerTo(RbacTagOperatorNotEquals)
		n = not.Node()
	}
	unsupported := fmt.Errorf("%w: unsupported condition %s", errUnrepresentable, typeName(n))

	var context apitype.PermissionContextExpression
	switch leaf := n.(type) {
	case apitype.PermissionExpressionHasTag:
		context, tc.Key = leaf.Context(), leaf.Key()
	case apitype.PermissionExpressionEqual:
		tag, okTag := leaf.Left().(apitype.PermissionExpressionTag)
		str, okStr := leaf.Right().(apitype.PermissionLiteralExpressionString)
		if !okTag || !okStr {
			return "", RbacTagCondition{}, unsupported
		}
		context, tc.Key, tc.Value = tag.Context(), tag.Key(), pointerTo(str.Value())
	default:
		return "", RbacTagCondition{}, unsupported
	}
	rt, ok := contextResourceType(context)
	if !ok {
		return "", RbacTagCondition{}, unsupported
	}
	return rt, tc, nil
}

// canonicalRoleKey renders the permissions a role grants in an
// order-independent form. Read uses it to keep the user's own layout of
// inputs when it grants exactly what the service reports.
func canonicalRoleKey(orgSetIDs []string, rules []RbacEntityRule) string {
	type canonRule struct {
		Type     RbacResourceType
		Selector string
		Sets     []string
	}
	org := dedupe(orgSetIDs)
	sort.Strings(org)

	all := map[RbacResourceType][]string{}
	var conds []canonRule
	for _, r := range rules {
		rt, sel := r.selector()
		if sel == nil {
			continue
		}
		if sel.All != nil && *sel.All {
			all[rt] = append(all[rt], r.PermissionSetIds...)
			continue
		}
		s := RbacEntitySelector{Id: sel.Id}
		for _, tc := range sel.Tags {
			op := RbacTagOperatorEquals
			if tc.Operator != nil {
				op = *tc.Operator
			}
			v := strings.TrimSpace(util.OrZero(tc.Value))
			s.Tags = append(s.Tags, RbacTagCondition{Key: tc.Key, Value: &v, Operator: &op})
		}
		selJSON, _ := json.Marshal(s)
		sets := slices.Clone(r.PermissionSetIds)
		sort.Strings(sets)
		conds = append(conds, canonRule{Type: rt, Selector: string(selJSON), Sets: sets})
	}
	for rt, ids := range all {
		ids = dedupe(ids)
		sort.Strings(ids)
		all[rt] = ids
	}
	sort.Slice(conds, func(i, j int) bool {
		a, _ := json.Marshal(conds[i])
		b, _ := json.Marshal(conds[j])
		return string(a) < string(b)
	})
	out, _ := json.Marshal(struct {
		Org   []string
		All   map[RbacResourceType][]string
		Conds []canonRule
	}{org, all, conds})
	return string(out)
}

func dedupe(ids []string) []string {
	seen := make(map[string]bool, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

func pointerTo[T any](v T) *T { return &v }
