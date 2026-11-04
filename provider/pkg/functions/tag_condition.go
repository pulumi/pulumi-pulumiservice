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
	"errors"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"
)

type TagOperator string

const (
	TagOperatorEquals    TagOperator = "equals"
	TagOperatorNotEquals TagOperator = "notEquals"
	TagOperatorExists    TagOperator = "exists"
	TagOperatorNotExists TagOperator = "notExists"
)

func (TagOperator) Values() []infer.EnumValue[TagOperator] {
	return []infer.EnumValue[TagOperator]{
		{Name: "Equals", Value: TagOperatorEquals, Description: "The tag is set to `value`."},
		{
			Name:        "NotEquals",
			Value:       TagOperatorNotEquals,
			Description: "The tag isn't set to `value`. Also matches entities that don't have the tag.",
		},
		{Name: "Exists", Value: TagOperatorExists, Description: "The entity has the tag, with any value."},
		{Name: "NotExists", Value: TagOperatorNotExists, Description: "The entity doesn't have the tag."},
	}
}

type TagCondition struct {
	Key      string       `pulumi:"key"`
	Operator *TagOperator `pulumi:"operator,optional"`
	Value    *string      `pulumi:"value,optional"`
}

func (c *TagCondition) Annotate(a infer.Annotator) {
	a.Describe(&c.Key, "The tag key.")
	a.Describe(&c.Operator, "How the tag is compared. Defaults to `equals`.")
	a.Describe(&c.Value,
		"The tag value. Required with `equals` and `notEquals`, and not allowed with `exists` and `notExists`.")
}

// tagExpression encodes one tag condition the way the console does. Match
// rules and buildRolePermissions' tag rules share it, so a condition means the
// same thing in both.
func tagExpression(
	entity apitype.PermissionContextExpression,
	c TagCondition,
) (apitype.PermissionBooleanExpression, error) {
	if strings.TrimSpace(c.Key) == "" {
		return nil, errors.New("`key` must not be empty")
	}
	operator := TagOperatorEquals
	if c.Operator != nil {
		operator = *c.Operator
	}

	var positive apitype.PermissionBooleanExpression
	switch operator {
	case TagOperatorEquals, TagOperatorNotEquals:
		if c.Value == nil || *c.Value == "" {
			return nil, fmt.Errorf("`value` is required with the %q operator", operator)
		}
		positive = apitype.PermissionExpressionEqualBuilder{
			Left:  apitype.PermissionExpressionTagBuilder{Context: entity, Key: c.Key}.Build(),
			Right: apitype.PermissionLiteralExpressionStringBuilder{Value: *c.Value}.Build(),
		}.Build()
	case TagOperatorExists, TagOperatorNotExists:
		if c.Value != nil {
			return nil, fmt.Errorf("`value` can't be set with the %q operator", operator)
		}
		positive = apitype.PermissionExpressionHasTagBuilder{Context: entity, Key: c.Key}.Build()
	default:
		return nil, fmt.Errorf("unknown operator %q; use `equals`, `notEquals`, `exists`, or `notExists`", operator)
	}

	if operator == TagOperatorEquals || operator == TagOperatorExists {
		return positive, nil
	}
	return apitype.PermissionExpressionNotBuilder{
		PermissionBooleanExpressionUnaryBuilder: apitype.PermissionBooleanExpressionUnaryBuilder{Node: positive},
	}.Build(), nil
}
