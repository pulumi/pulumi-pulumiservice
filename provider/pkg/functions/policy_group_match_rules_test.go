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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-go-provider/infer"
)

const (
	tagEnv  = "env"
	tagProd = "prod"
)

func operator(o PolicyGroupTagOperator) *PolicyGroupTagOperator { return &o }

func value(v string) *string { return &v }

// The Pulumi Cloud console's encoding of the rules in TestBuildStackMatchRule.
const consoleStackMatchRuleJSON = `{
  "__type": "PermissionExpressionOr",
  "left": {
    "__type": "PermissionExpressionAnd",
    "left": {
      "__type": "PermissionExpressionEqual",
      "left": {
        "__type": "PermissionExpressionTag",
        "context": {"__type": "PermissionExpressionStack"},
        "key": "env"
      },
      "right": {"__type": "PermissionLiteralExpressionString", "value": "prod"}
    },
    "right": {
      "__type": "PermissionExpressionNot",
      "node": {
        "__type": "PermissionExpressionEqual",
        "left": {
          "__type": "PermissionExpressionTag",
          "context": {"__type": "PermissionExpressionStack"},
          "key": "team"
        },
        "right": {"__type": "PermissionLiteralExpressionString", "value": "legacy"}
      }
    }
  },
  "right": {
    "__type": "PermissionExpressionAnd",
    "left": {
      "__type": "PermissionExpressionHasTag",
      "context": {"__type": "PermissionExpressionStack"},
      "key": "cost-center"
    },
    "right": {
      "__type": "PermissionExpressionNot",
      "node": {
        "__type": "PermissionExpressionHasTag",
        "context": {"__type": "PermissionExpressionStack"},
        "key": "deprecated"
      }
    }
  }
}`

func TestBuildStackMatchRule(t *testing.T) {
	t.Parallel()
	resp, err := BuildStackMatchRuleFunction{}.Invoke(context.Background(), infer.FunctionRequest[BuildMatchRuleInput]{
		Input: BuildMatchRuleInput{Rules: []PolicyGroupMatchRule{
			{Conditions: []PolicyGroupTagCondition{
				{Key: tagEnv, Value: value(tagProd)},
				{Key: "team", Operator: operator(PolicyGroupTagOperatorNotEquals), Value: value("legacy")},
			}},
			{Conditions: []PolicyGroupTagCondition{
				{Key: "cost-center", Operator: operator(PolicyGroupTagOperatorExists)},
				{Key: "deprecated", Operator: operator(PolicyGroupTagOperatorNotExists)},
			}},
		}},
	})
	require.NoError(t, err)
	got, err := json.Marshal(resp.Output.MatchRule)
	require.NoError(t, err)
	assert.JSONEq(t, consoleStackMatchRuleJSON, string(got))
}

func TestBuildAccountMatchRule(t *testing.T) {
	t.Parallel()
	resp, err := BuildAccountMatchRuleFunction{}.Invoke(context.Background(), infer.FunctionRequest[BuildMatchRuleInput]{
		Input: BuildMatchRuleInput{Rules: []PolicyGroupMatchRule{
			{Conditions: []PolicyGroupTagCondition{{Key: tagEnv, Operator: operator(PolicyGroupTagOperatorExists)}}},
		}},
	})
	require.NoError(t, err)
	got, err := json.Marshal(resp.Output.MatchRule)
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"__type": "PermissionExpressionHasTag",
		"context": {"__type": "PermissionExpressionInsightsAccount"},
		"key": "env"
	}`, string(got))
}

func TestBuildMatchRule_Errors(t *testing.T) {
	t.Parallel()
	for name, tc := range map[string]struct {
		rules []PolicyGroupMatchRule
		err   string
	}{
		"no rules": {nil, "`rules` must not be empty"},
		"no conditions": {
			[]PolicyGroupMatchRule{{}},
			"rules[0]: `conditions` must not be empty",
		},
		"empty key": {
			[]PolicyGroupMatchRule{{Conditions: []PolicyGroupTagCondition{{Key: " ", Value: value(tagProd)}}}},
			"rules[0].conditions[0]: `key` must not be empty",
		},
		"team-access key": {
			[]PolicyGroupMatchRule{{Conditions: []PolicyGroupTagCondition{
				{Key: "pulumi:team-access:ops", Operator: operator(PolicyGroupTagOperatorExists)},
			}}},
			"can't use team-access tags",
		},
		"equals without a value": {
			[]PolicyGroupMatchRule{{Conditions: []PolicyGroupTagCondition{{Key: tagEnv}}}},
			"`value` is required with the \"equals\" operator",
		},
		"exists with a value": {
			[]PolicyGroupMatchRule{{Conditions: []PolicyGroupTagCondition{
				{Key: tagEnv, Operator: operator(PolicyGroupTagOperatorExists), Value: value(tagProd)},
			}}},
			"`value` can't be set with the \"exists\" operator",
		},
		"unknown operator": {
			[]PolicyGroupMatchRule{{Conditions: []PolicyGroupTagCondition{
				{Key: tagEnv, Operator: operator("contains"), Value: value(tagProd)},
			}}},
			"unknown operator \"contains\"",
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := BuildStackMatchRuleFunction{}.Invoke(context.Background(), infer.FunctionRequest[BuildMatchRuleInput]{
				Input: BuildMatchRuleInput{Rules: tc.rules},
			})
			assert.ErrorContains(t, err, tc.err)
		})
	}
}
