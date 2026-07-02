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

func ptr[T any](v T) *T { return &v }

func TestGenerateWebhookID(t *testing.T) {
	t.Run("organization scope", func(t *testing.T) {
		id := generateWebhookID(WebhookInput{OrganizationName: gcMyOrg}, "hook-1")
		assert.Equal(t, "my-org/hook-1", id)
	})

	t.Run("stack scope", func(t *testing.T) {
		id := generateWebhookID(WebhookInput{
			OrganizationName: gcMyOrg,
			ProjectName:      ptr(gcMyProject),
			StackName:        ptr("my-stack"),
		}, "hook-2")
		assert.Equal(t, "my-org/my-project/my-stack/hook-2", id)
	})

	t.Run("environment scope", func(t *testing.T) {
		id := generateWebhookID(WebhookInput{
			OrganizationName: gcMyOrg,
			ProjectName:      ptr(gcMyProject),
			EnvironmentName:  ptr("dev"),
		}, "hook-3")
		assert.Equal(t, "my-org/environment/my-project/dev/hook-3", id)
	})
}

func TestSplitWebhookID(t *testing.T) {
	t.Run("organization scope", func(t *testing.T) {
		got, err := splitWebhookID("my-org/hook-1")
		require.NoError(t, err)
		assert.Equal(t, &webhookID{
			organizationName: gcMyOrg,
			webhookName:      "hook-1",
		}, got)
	})

	t.Run("stack scope", func(t *testing.T) {
		got, err := splitWebhookID("my-org/my-project/my-stack/hook-2")
		require.NoError(t, err)
		assert.Equal(t, &webhookID{
			organizationName: gcMyOrg,
			projectName:      ptr(gcMyProject),
			stackName:        ptr("my-stack"),
			webhookName:      "hook-2",
		}, got)
	})

	t.Run("environment scope", func(t *testing.T) {
		got, err := splitWebhookID("my-org/environment/my-project/dev/hook-3")
		require.NoError(t, err)
		assert.Equal(t, &webhookID{
			organizationName: gcMyOrg,
			projectName:      ptr(gcMyProject),
			environmentName:  ptr("dev"),
			webhookName:      "hook-3",
		}, got)
	})

	t.Run("malformed", func(t *testing.T) {
		_, err := splitWebhookID("only-one")
		require.Error(t, err)
		_, err = splitWebhookID("a/b/c")
		require.Error(t, err)
		_, err = splitWebhookID("a/b/c/d/e/f")
		require.Error(t, err)
	})

	t.Run("round-trip", func(t *testing.T) {
		// All round-trip cases must produce the same scope-shape on parse.
		cases := []WebhookInput{
			{OrganizationName: gcOrg},
			{OrganizationName: gcOrg, ProjectName: ptr("proj"), StackName: ptr("stk")},
			{OrganizationName: gcOrg, ProjectName: ptr("proj"), EnvironmentName: ptr(gcEnv)},
		}
		for _, in := range cases {
			id := generateWebhookID(in, "hook")
			parsed, err := splitWebhookID(id)
			require.NoError(t, err)
			assert.Equal(t, "hook", parsed.webhookName)
			assert.Equal(t, in.OrganizationName, parsed.organizationName)
			assert.Equal(t, in.ProjectName, parsed.projectName)
			assert.Equal(t, in.StackName, parsed.stackName)
			assert.Equal(t, in.EnvironmentName, parsed.environmentName)
		}
	})
}
