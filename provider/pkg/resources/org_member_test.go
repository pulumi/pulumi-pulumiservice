// Copyright 2016-2025, Pulumi Corporation.
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
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
)

func TestOrgMemberResource_Name(t *testing.T) {
	resource := &PulumiServiceOrgMemberResource{}
	assert.Equal(t, "pulumiservice:index:OrgMember", resource.Name())
}

func TestOrgMemberResource_Check(t *testing.T) {
	res := &PulumiServiceOrgMemberResource{}

	tests := []struct {
		name          string
		role          string
		expectFailure bool
		failureReason string
	}{
		{
			name:          "Valid role: admin",
			role:          "admin",
			expectFailure: false,
		},
		{
			name:          "Valid role: member",
			role:          "member",
			expectFailure: false,
		},
		{
			name:          "Invalid role: owner",
			role:          "owner",
			expectFailure: true,
			failureReason: "role must be either 'admin' or 'member'",
		},
		{
			name:          "Invalid role: readonly",
			role:          "readonly",
			expectFailure: true,
			failureReason: "role must be either 'admin' or 'member'",
		},
		{
			name:          "Invalid role: empty",
			role:          "",
			expectFailure: true,
			failureReason: "role is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputMap := resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"userName":         resource.NewPropertyValue("test-user"),
				"role":             resource.NewPropertyValue(tt.role),
			}

			properties, err := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})
			assert.NoError(t, err)

			req := &pulumirpc.CheckRequest{
				News: properties,
			}

			resp, err := res.Check(req)
			assert.NoError(t, err)

			if tt.expectFailure {
				assert.NotNil(t, resp.Failures)
				assert.Len(t, resp.Failures, 1)
				assert.Equal(t, "role", resp.Failures[0].Property)
				assert.Equal(t, tt.failureReason, resp.Failures[0].Reason)
			} else {
				assert.Nil(t, resp.Failures)
			}
		})
	}
}

func TestToPulumiServiceOrgMemberInput(t *testing.T) {
	res := &PulumiServiceOrgMemberResource{}
	inputMap := resource.PropertyMap{
		"organizationName": resource.NewPropertyValue("test-org"),
		"userName":         resource.NewPropertyValue("test-user"),
		"role":             resource.NewPropertyValue("admin"),
	}

	input := res.ToPulumiServiceOrgMemberInput(inputMap)

	assert.Equal(t, "test-org", input.OrganizationName)
	assert.Equal(t, "test-user", input.UserName)
	assert.Equal(t, "admin", input.Role)
}

func TestPulumiServiceOrgMemberInput_ToPropertyMap(t *testing.T) {
	input := PulumiServiceOrgMemberInput{
		OrganizationName: "my-org",
		UserName:         "bob",
		Role:             "member",
	}

	pm := input.ToPropertyMap()

	assert.Equal(t, "my-org", pm["organizationName"].StringValue())
	assert.Equal(t, "bob", pm["userName"].StringValue())
	assert.Equal(t, "member", pm["role"].StringValue())
}

func TestOrgMemberResource_Diff(t *testing.T) {
	res := &PulumiServiceOrgMemberResource{}

	tests := []struct {
		name               string
		oldOrg             string
		newOrg             string
		oldUser            string
		newUser            string
		oldRole            string
		newRole            string
		expectChanges      pulumirpc.DiffResponse_DiffChanges
		expectReplaces     []string
	}{
		{
			name:          "No changes",
			oldOrg:        "test-org",
			newOrg:        "test-org",
			oldUser:       "alice",
			newUser:       "alice",
			oldRole:       "admin",
			newRole:       "admin",
			expectChanges: pulumirpc.DiffResponse_DIFF_NONE,
			expectReplaces: nil,
		},
		{
			name:          "Role change only",
			oldOrg:        "test-org",
			newOrg:        "test-org",
			oldUser:       "alice",
			newUser:       "alice",
			oldRole:       "member",
			newRole:       "admin",
			expectChanges: pulumirpc.DiffResponse_DIFF_SOME,
			expectReplaces: nil,
		},
		{
			name:          "Organization change (replace)",
			oldOrg:        "old-org",
			newOrg:        "new-org",
			oldUser:       "alice",
			newUser:       "alice",
			oldRole:       "admin",
			newRole:       "admin",
			expectChanges: pulumirpc.DiffResponse_DIFF_SOME,
			expectReplaces: []string{"organizationName"},
		},
		{
			name:          "User change (replace)",
			oldOrg:        "test-org",
			newOrg:        "test-org",
			oldUser:       "alice",
			newUser:       "bob",
			oldRole:       "admin",
			newRole:       "admin",
			expectChanges: pulumirpc.DiffResponse_DIFF_SOME,
			expectReplaces: []string{"userName"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldInputMap := resource.PropertyMap{
				"organizationName": resource.NewPropertyValue(tt.oldOrg),
				"userName":         resource.NewPropertyValue(tt.oldUser),
				"role":             resource.NewPropertyValue(tt.oldRole),
			}

			newInputMap := resource.PropertyMap{
				"organizationName": resource.NewPropertyValue(tt.newOrg),
				"userName":         resource.NewPropertyValue(tt.newUser),
				"role":             resource.NewPropertyValue(tt.newRole),
			}

			oldProps, err := plugin.MarshalProperties(oldInputMap, plugin.MarshalOptions{})
			assert.NoError(t, err)

			newProps, err := plugin.MarshalProperties(newInputMap, plugin.MarshalOptions{})
			assert.NoError(t, err)

			req := &pulumirpc.DiffRequest{
				Olds: oldProps,
				News: newProps,
			}

			resp, err := res.Diff(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectChanges, resp.Changes)
			assert.Equal(t, tt.expectReplaces, resp.Replaces)
		})
	}
}
