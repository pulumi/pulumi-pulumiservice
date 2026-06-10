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

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildApprovalRuleID(t *testing.T) {
	got := buildApprovalRuleID(EnvironmentIdentifier{
		Organization: "my-org",
		Project:      "my-project",
		Name:         "my-env",
	}, "rule-abc")
	assert.Equal(t, "environment/my-org/my-project/my-env/rule-abc", got)
}

func TestParseApprovalRuleID(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		env, ruleID, err := parseApprovalRuleID("environment/my-org/my-project/my-env/rule-abc")
		require.NoError(t, err)
		assert.Equal(t, EnvironmentIdentifier{
			Organization: "my-org",
			Project:      "my-project",
			Name:         "my-env",
		}, env)
		assert.Equal(t, "rule-abc", ruleID)
	})

	t.Run("wrong prefix", func(t *testing.T) {
		_, _, err := parseApprovalRuleID("stack/my-org/my-project/my-env/rule-abc")
		require.Error(t, err)
	})

	t.Run("too few parts", func(t *testing.T) {
		_, _, err := parseApprovalRuleID("environment/my-org/my-project/my-env")
		require.Error(t, err)
	})

	t.Run("too many parts", func(t *testing.T) {
		_, _, err := parseApprovalRuleID("environment/my-org/my-project/my-env/rule-abc/extra")
		require.Error(t, err)
	})
}
