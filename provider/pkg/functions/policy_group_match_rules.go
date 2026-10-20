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
	"regexp"
	"strings"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// Pulumi Cloud manages tags with this prefix itself and rejects match rules
// that read them.
const teamAccessTagKeyPrefix = "pulumi:team-access:"

// The tag key pattern Pulumi Cloud enforces on stack and Insights account tags.
var tagKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9-_.:]+$`)

type PolicyGroupTagOperator string

const (
	PolicyGroupTagOperatorEquals    PolicyGroupTagOperator = "equals"
	PolicyGroupTagOperatorNotEquals PolicyGroupTagOperator = "notEquals"
	PolicyGroupTagOperatorExists    PolicyGroupTagOperator = "exists"
	PolicyGroupTagOperatorNotExists PolicyGroupTagOperator = "notExists"
)

func (PolicyGroupTagOperator) Values() []infer.EnumValue[PolicyGroupTagOperator] {
	return []infer.EnumValue[PolicyGroupTagOperator]{
		{Name: "Equals", Value: PolicyGroupTagOperatorEquals, Description: "The tag is set to `value`."},
		{
			Name:        "NotEquals",
			Value:       PolicyGroupTagOperatorNotEquals,
			Description: "The tag isn't set to `value`. Also matches entities that don't have the tag.",
		},
		{Name: "Exists", Value: PolicyGroupTagOperatorExists, Description: "The entity has the tag, with any value."},
		{Name: "NotExists", Value: PolicyGroupTagOperatorNotExists, Description: "The entity doesn't have the tag."},
	}
}

type PolicyGroupTagCondition struct {
	Key      string                  `pulumi:"key"`
	Operator *PolicyGroupTagOperator `pulumi:"operator,optional"`
	Value    *string                 `pulumi:"value,optional"`
}

func (c *PolicyGroupTagCondition) Annotate(a infer.Annotator) {
	a.Describe(&c.Key, "The tag key.")
	a.Describe(&c.Operator, "How the tag is compared. Defaults to `equals`.")
	a.Describe(&c.Value,
		"The tag value. Required with `equals` and `notEquals`, and not allowed with `exists` and `notExists`.")
}

type PolicyGroupMatchRule struct {
	Conditions []PolicyGroupTagCondition `pulumi:"conditions"`
}

func (r *PolicyGroupMatchRule) Annotate(a infer.Annotator) {
	a.Describe(&r.Conditions, "Tag conditions that must all hold for the rule to match. At least one is required.")
}

type BuildMatchRuleInput struct {
	Rules []PolicyGroupMatchRule `pulumi:"rules"`
}

func (i *BuildMatchRuleInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Rules, "The rules. An entity matches if it matches any rule.")
}

type BuildMatchRuleOutput struct {
	MatchRule map[string]any `pulumi:"matchRule"`
}

func (o *BuildMatchRuleOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.MatchRule, "The match rule, as the Pulumi Cloud expression a policy group's match rule field takes.")
}

type BuildStackMatchRuleFunction struct{}

func (BuildStackMatchRuleFunction) Annotate(a infer.Annotator) {
	a.Describe(&BuildStackMatchRuleFunction{},
		"Builds a policy group's stack match rule from tag rules, without writing Pulumi Cloud's "+
			"expression format by hand. Use the result as a policy group's `stackMatchRule`. Pulumi "+
			"Cloud adds the stacks the rule matches to the group, alongside the stacks added "+
			"individually, and keeps membership up to date as stack tags change. Match rules are a "+
			"preview feature of Pulumi Cloud.")
	a.SetToken("index", "buildStackMatchRule")
}

func (BuildStackMatchRuleFunction) Invoke(
	_ context.Context,
	req infer.FunctionRequest[BuildMatchRuleInput],
) (infer.FunctionResponse[BuildMatchRuleOutput], error) {
	out, err := buildMatchRule(apitype.PermissionExpressionStackBuilder{}.Build(), req.Input.Rules)
	return infer.FunctionResponse[BuildMatchRuleOutput]{Output: BuildMatchRuleOutput{MatchRule: out}}, err
}

type BuildAccountMatchRuleFunction struct{}

func (BuildAccountMatchRuleFunction) Annotate(a infer.Annotator) {
	a.Describe(&BuildAccountMatchRuleFunction{},
		"Builds a policy group's Insights account match rule from tag rules, without writing Pulumi "+
			"Cloud's expression format by hand. Use the result as a policy group's `accountMatchRule`. "+
			"Pulumi Cloud adds the accounts the rule matches to the group, alongside the accounts added "+
			"individually, and keeps membership up to date as account tags change. Only `audit` policy "+
			"groups support account rules. Match rules are a preview feature of Pulumi Cloud.")
	a.SetToken("index", "buildAccountMatchRule")
}

func (BuildAccountMatchRuleFunction) Invoke(
	_ context.Context,
	req infer.FunctionRequest[BuildMatchRuleInput],
) (infer.FunctionResponse[BuildMatchRuleOutput], error) {
	out, err := buildMatchRule(apitype.PermissionExpressionInsightsAccountBuilder{}.Build(), req.Input.Rules)
	return infer.FunctionResponse[BuildMatchRuleOutput]{Output: BuildMatchRuleOutput{MatchRule: out}}, err
}

// buildMatchRule encodes rules the way the Pulumi Cloud console does: a
// left-folded Or of left-folded Ands of tag conditions. Matching it keeps a
// rule saved in the console from showing up as a diff against the same rule
// built here.
func buildMatchRule(
	entity apitype.PermissionContextExpression,
	rules []PolicyGroupMatchRule,
) (map[string]any, error) {
	if len(rules) == 0 {
		return nil, errors.New("`rules` must not be empty")
	}
	ruleExprs := make([]apitype.PermissionBooleanExpression, len(rules))
	for i, rule := range rules {
		if len(rule.Conditions) == 0 {
			return nil, fmt.Errorf("rules[%d]: `conditions` must not be empty", i)
		}
		conditions := make([]apitype.PermissionBooleanExpression, len(rule.Conditions))
		for j, c := range rule.Conditions {
			expr, err := tagConditionExpression(entity, c)
			if err != nil {
				return nil, fmt.Errorf("rules[%d].conditions[%d]: %w", i, j, err)
			}
			conditions[j] = expr
		}
		ruleExprs[i] = foldExpressions(conditions, andExpression)
	}

	return toSDKMap(foldExpressions(ruleExprs, orExpression))
}

func tagConditionExpression(
	entity apitype.PermissionContextExpression,
	c PolicyGroupTagCondition,
) (apitype.PermissionBooleanExpression, error) {
	switch {
	case c.Key == "":
		return nil, errors.New("`key` must not be empty")
	case !tagKeyPattern.MatchString(c.Key):
		return nil, fmt.Errorf("`key` %q may only contain alphanumerics, hyphens, underscores, periods, or colons", c.Key)
	case strings.HasPrefix(c.Key, teamAccessTagKeyPrefix):
		return nil, fmt.Errorf("match rules can't use team-access tags (%s*)", teamAccessTagKeyPrefix)
	}
	return tagExpression(entity, c)
}

// tagExpression encodes one tag condition the way the console does. Match
// rules and buildRolePermissions' tag rules share it, so a condition means the
// same thing in both.
func tagExpression(
	entity apitype.PermissionContextExpression,
	c PolicyGroupTagCondition,
) (apitype.PermissionBooleanExpression, error) {
	if strings.TrimSpace(c.Key) == "" {
		return nil, errors.New("`key` must not be empty")
	}
	operator := PolicyGroupTagOperatorEquals
	if c.Operator != nil {
		operator = *c.Operator
	}

	var positive apitype.PermissionBooleanExpression
	switch operator {
	case PolicyGroupTagOperatorEquals, PolicyGroupTagOperatorNotEquals:
		if c.Value == nil || *c.Value == "" {
			return nil, fmt.Errorf("`value` is required with the %q operator", operator)
		}
		positive = apitype.PermissionExpressionEqualBuilder{
			Left:  apitype.PermissionExpressionTagBuilder{Context: entity, Key: c.Key}.Build(),
			Right: apitype.PermissionLiteralExpressionStringBuilder{Value: *c.Value}.Build(),
		}.Build()
	case PolicyGroupTagOperatorExists, PolicyGroupTagOperatorNotExists:
		if c.Value != nil {
			return nil, fmt.Errorf("`value` can't be set with the %q operator", operator)
		}
		positive = apitype.PermissionExpressionHasTagBuilder{Context: entity, Key: c.Key}.Build()
	default:
		return nil, fmt.Errorf("unknown operator %q; use `equals`, `notEquals`, `exists`, or `notExists`", operator)
	}

	if operator == PolicyGroupTagOperatorEquals || operator == PolicyGroupTagOperatorExists {
		return positive, nil
	}
	return apitype.PermissionExpressionNotBuilder{
		PermissionBooleanExpressionUnaryBuilder: apitype.PermissionBooleanExpressionUnaryBuilder{Node: positive},
	}.Build(), nil
}

func andExpression(l, r apitype.PermissionBooleanExpression) apitype.PermissionBooleanExpression {
	return apitype.PermissionExpressionAndBuilder{
		PermissionBooleanExpressionBinaryBuilder: apitype.PermissionBooleanExpressionBinaryBuilder{Left: l, Right: r},
	}.Build()
}

func orExpression(l, r apitype.PermissionBooleanExpression) apitype.PermissionBooleanExpression {
	return apitype.PermissionExpressionOrBuilder{
		PermissionBooleanExpressionBinaryBuilder: apitype.PermissionBooleanExpressionBinaryBuilder{Left: l, Right: r},
	}.Build()
}

func foldExpressions(
	exprs []apitype.PermissionBooleanExpression,
	join func(l, r apitype.PermissionBooleanExpression) apitype.PermissionBooleanExpression,
) apitype.PermissionBooleanExpression {
	acc := exprs[0]
	for _, e := range exprs[1:] {
		acc = join(acc, e)
	}
	return acc
}
