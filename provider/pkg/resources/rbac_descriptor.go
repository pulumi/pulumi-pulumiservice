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
// value is HasTag{context, key}; notEquals wraps either in Not{node}; several
// conditions fold left into And{left, right}.

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

const (
	wireAllow        = "PermissionDescriptorAllow"
	wireGroup        = "PermissionDescriptorGroup"
	wireCompose      = "PermissionDescriptorCompose"
	wireCondition    = "PermissionDescriptorCondition"
	wireEqual        = "PermissionExpressionEqual"
	wireNot          = "PermissionExpressionNot"
	wireAnd          = "PermissionExpressionAnd"
	wireHasTag       = "PermissionExpressionHasTag"
	wireTag          = "PermissionExpressionTag"
	wireLiteralStr   = "PermissionLiteralExpressionString"
	wireExprPrefix   = "PermissionExpression"
	wireLiteralPrefx = "PermissionLiteralExpression"
)

// errUnrepresentable marks a descriptor tree that RbacRole's typed model
// cannot express, typically one built with `pulumiservice:api:Role` or
// `OrganizationRole`.
var errUnrepresentable = errors.New("permission descriptor cannot be represented by RbacRole")

// wireNode is a loose view of any descriptor or expression node, enough to
// build and walk the shapes the console uses.
type wireNode struct {
	Type                  string          `json:"__type"`
	PermissionDescriptors []string        `json:"permissionDescriptors,omitempty"`
	Entries               []*wireNode     `json:"entries,omitempty"`
	Condition             *wireNode       `json:"condition,omitempty"`
	SubNode               *wireNode       `json:"subNode,omitempty"`
	Left                  *wireNode       `json:"left,omitempty"`
	Right                 *wireNode       `json:"right,omitempty"`
	Node                  *wireNode       `json:"node,omitempty"`
	Context               *wireNode       `json:"context,omitempty"`
	Key                   string          `json:"key,omitempty"`
	Identity              string          `json:"identity,omitempty"`
	Value                 json.RawMessage `json:"value,omitempty"`
}

// entityKinds maps the entity selector fields to their wire suffix, in the
// order rules are emitted and "all" sets are grouped.
var entityKinds = []struct {
	resourceType RbacResourceType
	wire         string
}{
	{RbacResourceTypeStack, "Stack"},
	{RbacResourceTypeEnvironment, "Environment"},
	{RbacResourceTypeInsightsAccount, "InsightsAccount"},
}

func wireSuffix(rt RbacResourceType) string {
	for _, k := range entityKinds {
		if k.resourceType == rt {
			return k.wire
		}
	}
	return ""
}

func resourceTypeForWire(suffix string) (RbacResourceType, bool) {
	for _, k := range entityKinds {
		if k.wire == suffix {
			return k.resourceType, true
		}
	}
	return "", false
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
	group := &wireNode{Type: wireGroup}

	unconditional := slices.Clone(in.OrganizationPermissionSetIds)
	var conditional []*wireNode
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
		conditional = append(conditional, &wireNode{
			Type:      wireCondition,
			Condition: cond,
			SubNode:   &wireNode{Type: wireCompose, PermissionDescriptors: rule.PermissionSetIds},
		})
	}
	if len(unconditional) > 0 {
		group.Entries = append(group.Entries, &wireNode{Type: wireCompose, PermissionDescriptors: dedupe(unconditional)})
	}
	group.Entries = append(group.Entries, conditional...)

	raw, err := json.Marshal(group)
	if err != nil {
		return nil, fmt.Errorf("marshal policy details: %w", err)
	}
	var details apitype.PermissionDescriptor
	if err := apitype.UnmarshalJSONPermissionDescriptor(raw, &details); err != nil {
		return nil, fmt.Errorf("build policy details: %w", err)
	}
	return details, nil
}

func buildSelectorCondition(rt RbacResourceType, sel *RbacEntitySelector) (*wireNode, error) {
	suffix := wireSuffix(rt)
	context := func() *wireNode { return &wireNode{Type: wireExprPrefix + suffix} }

	if sel.Id != nil {
		return &wireNode{
			Type:  wireEqual,
			Left:  context(),
			Right: &wireNode{Type: wireLiteralPrefx + suffix, Identity: *sel.Id},
		}, nil
	}
	if len(sel.Tags) == 0 {
		return nil, errors.New("one of id, tags, or all must be set")
	}
	var acc *wireNode
	for _, tc := range sel.Tags {
		var leaf *wireNode
		if v := strings.TrimSpace(util.OrZero(tc.Value)); v == "" {
			leaf = &wireNode{Type: wireHasTag, Context: context(), Key: tc.Key}
		} else {
			lit, err := json.Marshal(*tc.Value)
			if err != nil {
				return nil, err
			}
			leaf = &wireNode{
				Type:  wireEqual,
				Left:  &wireNode{Type: wireTag, Context: context(), Key: tc.Key},
				Right: &wireNode{Type: wireLiteralStr, Value: lit},
			}
		}
		if tc.Operator != nil && *tc.Operator == RbacTagOperatorNotEquals {
			leaf = &wireNode{Type: wireNot, Node: leaf}
		}
		if acc == nil {
			acc = leaf
		} else {
			acc = &wireNode{Type: wireAnd, Left: acc, Right: leaf}
		}
	}
	return acc, nil
}

// setTypeLookup resolves a permission set ID to its resource type.
type setTypeLookup func(id string) (RbacResourceType, error)

// parsePolicyDetails is the inverse of buildPolicyDetails. Unconditional set
// references are split by the referenced set's resource type: global sets
// are organization-level access, the rest are "all <entity>" rules.
func parsePolicyDetails(
	details apitype.PermissionDescriptor, lookup setTypeLookup,
) (orgSetIDs []string, rules []RbacEntityRule, err error) {
	raw, err := json.Marshal(details)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal policy details: %w", err)
	}
	var root wireNode
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, nil, fmt.Errorf("parse policy details: %w", err)
	}
	if root.Type != wireGroup {
		return nil, nil, fmt.Errorf("%w: policy is a %s, want %s", errUnrepresentable, root.Type, wireGroup)
	}

	allSets := map[RbacResourceType][]string{}
	for i, entry := range root.Entries {
		switch entry.Type {
		case wireCompose:
			for _, id := range entry.PermissionDescriptors {
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
		case wireCondition:
			if entry.SubNode == nil || entry.SubNode.Type != wireCompose {
				return nil, nil, fmt.Errorf("%w: entry %d grants something other than permission sets", errUnrepresentable, i)
			}
			rt, sel, err := parseCondition(entry.Condition)
			if err != nil {
				return nil, nil, fmt.Errorf("entry %d: %w", i, err)
			}
			rules = append(rules, ruleFor(rt, sel, entry.SubNode.PermissionDescriptors))
		default:
			return nil, nil, fmt.Errorf("%w: entry %d is a %s", errUnrepresentable, i, entry.Type)
		}
	}

	var allRules []RbacEntityRule
	for _, k := range entityKinds {
		if ids := allSets[k.resourceType]; len(ids) > 0 {
			allRules = append(allRules, ruleFor(k.resourceType, &RbacEntitySelector{All: pointerTo(true)}, ids))
		}
	}
	return orgSetIDs, append(allRules, rules...), nil
}

func parseCondition(n *wireNode) (RbacResourceType, *RbacEntitySelector, error) {
	if n == nil {
		return "", nil, fmt.Errorf("%w: condition is empty", errUnrepresentable)
	}
	// A single entity: Equal(<Entity>, Literal<Entity>{identity}).
	if n.Type == wireEqual && n.Left != nil && n.Right != nil &&
		strings.HasPrefix(n.Right.Type, wireLiteralPrefx) && n.Right.Type != wireLiteralStr {
		rt, ok := resourceTypeForWire(strings.TrimPrefix(n.Left.Type, wireExprPrefix))
		if !ok || n.Right.Type != wireLiteralPrefx+wireSuffix(rt) {
			return "", nil, fmt.Errorf("%w: unsupported entity condition %s = %s", errUnrepresentable, n.Left.Type, n.Right.Type)
		}
		return rt, &RbacEntitySelector{Id: pointerTo(n.Right.Identity)}, nil
	}

	var leaves []*wireNode
	var flatten func(*wireNode)
	flatten = func(x *wireNode) {
		if x != nil && x.Type == wireAnd {
			flatten(x.Left)
			flatten(x.Right)
			return
		}
		leaves = append(leaves, x)
	}
	flatten(n)

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

func parseTagCondition(n *wireNode) (RbacResourceType, RbacTagCondition, error) {
	var tc RbacTagCondition
	if n != nil && n.Type == wireNot {
		tc.Operator = pointerTo(RbacTagOperatorNotEquals)
		n = n.Node
	}
	unsupported := func() (RbacResourceType, RbacTagCondition, error) {
		typ := "<nil>"
		if n != nil {
			typ = n.Type
		}
		return "", RbacTagCondition{}, fmt.Errorf("%w: unsupported condition %s", errUnrepresentable, typ)
	}
	if n == nil {
		return unsupported()
	}

	var context *wireNode
	switch {
	case n.Type == wireHasTag:
		context, tc.Key = n.Context, n.Key
	case n.Type == wireEqual && n.Left != nil && n.Left.Type == wireTag &&
		n.Right != nil && n.Right.Type == wireLiteralStr:
		var v string
		if err := json.Unmarshal(n.Right.Value, &v); err != nil {
			return unsupported()
		}
		context, tc.Key, tc.Value = n.Left.Context, n.Left.Key, pointerTo(v)
	default:
		return unsupported()
	}
	if context == nil {
		return unsupported()
	}
	rt, ok := resourceTypeForWire(strings.TrimPrefix(context.Type, wireExprPrefix))
	if !ok {
		return unsupported()
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
