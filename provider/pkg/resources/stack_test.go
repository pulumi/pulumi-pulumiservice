// Copyright 2016-2025, Pulumi Corporation.

package resources

import (
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStack_Name(t *testing.T) {
	provider := PulumiServiceStackResource{}
	assert.Equal(t, "pulumiservice:index:Stack", provider.Name())
}

func TestPulumiServiceStack_ToPropertyMap(t *testing.T) {
	tests := []struct {
		name     string
		stack    PulumiServiceStack
		expected resource.PropertyMap
	}{
		{
			name: "basic stack without forceDestroy",
			stack: PulumiServiceStack{
				StackIdentifier: pulumiapi.StackIdentifier{
					OrgName:     "test-org",
					ProjectName: "test-project",
					StackName:   "dev",
				},
				ForceDestroy: false,
			},
			expected: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"projectName":      resource.NewPropertyValue("test-project"),
				"stackName":        resource.NewPropertyValue("dev"),
			},
		},
		{
			name: "stack with forceDestroy enabled",
			stack: PulumiServiceStack{
				StackIdentifier: pulumiapi.StackIdentifier{
					OrgName:     "prod-org",
					ProjectName: "infra",
					StackName:   "prod",
				},
				ForceDestroy: true,
			},
			expected: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("prod-org"),
				"projectName":      resource.NewPropertyValue("infra"),
				"stackName":        resource.NewPropertyValue("prod"),
				"forceDestroy":     resource.NewPropertyValue(true),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.stack.ToPropertyMap()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPulumiServiceStackResource_ToPulumiServiceStackTagInput(t *testing.T) {
	tests := []struct {
		name     string
		inputMap resource.PropertyMap
		expected *PulumiServiceStack
	}{
		{
			name: "valid input without forceDestroy",
			inputMap: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"projectName":      resource.NewPropertyValue("test-project"),
				"stackName":        resource.NewPropertyValue("dev"),
			},
			expected: &PulumiServiceStack{
				StackIdentifier: pulumiapi.StackIdentifier{
					OrgName:     "test-org",
					ProjectName: "test-project",
					StackName:   "dev",
				},
				ForceDestroy: false,
			},
		},
		{
			name: "valid input with forceDestroy false",
			inputMap: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"projectName":      resource.NewPropertyValue("test-project"),
				"stackName":        resource.NewPropertyValue("staging"),
				"forceDestroy":     resource.NewPropertyValue(false),
			},
			expected: &PulumiServiceStack{
				StackIdentifier: pulumiapi.StackIdentifier{
					OrgName:     "test-org",
					ProjectName: "test-project",
					StackName:   "staging",
				},
				ForceDestroy: false,
			},
		},
		{
			name: "valid input with forceDestroy true",
			inputMap: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"projectName":      resource.NewPropertyValue("test-project"),
				"stackName":        resource.NewPropertyValue("prod"),
				"forceDestroy":     resource.NewPropertyValue(true),
			},
			expected: &PulumiServiceStack{
				StackIdentifier: pulumiapi.StackIdentifier{
					OrgName:     "test-org",
					ProjectName: "test-project",
					StackName:   "prod",
				},
				ForceDestroy: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &PulumiServiceStackResource{}
			result, err := resource.ToPulumiServiceStackTagInput(tt.inputMap)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPulumiServiceStackResource_Diff(t *testing.T) {
	tests := []struct {
		name                string
		oldInputs           resource.PropertyMap
		newInputs           resource.PropertyMap
		expectedChanges     pulumirpc.DiffResponse_DiffChanges
		expectReplace       bool
		expectDetailedDiff  bool
		changedProperties   []string
	}{
		{
			name: "no changes",
			oldInputs: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"projectName":      resource.NewPropertyValue("test-project"),
				"stackName":        resource.NewPropertyValue("dev"),
			},
			newInputs: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"projectName":      resource.NewPropertyValue("test-project"),
				"stackName":        resource.NewPropertyValue("dev"),
			},
			expectedChanges:    pulumirpc.DiffResponse_DIFF_NONE,
			expectReplace:      false,
			expectDetailedDiff: false,
		},
		{
			name: "stackName changed",
			oldInputs: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"projectName":      resource.NewPropertyValue("test-project"),
				"stackName":        resource.NewPropertyValue("dev"),
			},
			newInputs: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"projectName":      resource.NewPropertyValue("test-project"),
				"stackName":        resource.NewPropertyValue("prod"),
			},
			expectedChanges:    pulumirpc.DiffResponse_DIFF_SOME,
			expectReplace:      true,
			expectDetailedDiff: true,
			changedProperties:  []string{"stackName"},
		},
		{
			name: "organizationName changed",
			oldInputs: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("old-org"),
				"projectName":      resource.NewPropertyValue("test-project"),
				"stackName":        resource.NewPropertyValue("dev"),
			},
			newInputs: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("new-org"),
				"projectName":      resource.NewPropertyValue("test-project"),
				"stackName":        resource.NewPropertyValue("dev"),
			},
			expectedChanges:    pulumirpc.DiffResponse_DIFF_SOME,
			expectReplace:      true,
			expectDetailedDiff: true,
			changedProperties:  []string{"organizationName"},
		},
		{
			name: "projectName changed",
			oldInputs: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"projectName":      resource.NewPropertyValue("old-project"),
				"stackName":        resource.NewPropertyValue("dev"),
			},
			newInputs: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"projectName":      resource.NewPropertyValue("new-project"),
				"stackName":        resource.NewPropertyValue("dev"),
			},
			expectedChanges:    pulumirpc.DiffResponse_DIFF_SOME,
			expectReplace:      true,
			expectDetailedDiff: true,
			changedProperties:  []string{"projectName"},
		},
		{
			name: "forceDestroy added",
			oldInputs: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"projectName":      resource.NewPropertyValue("test-project"),
				"stackName":        resource.NewPropertyValue("dev"),
			},
			newInputs: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"projectName":      resource.NewPropertyValue("test-project"),
				"stackName":        resource.NewPropertyValue("dev"),
				"forceDestroy":     resource.NewPropertyValue(true),
			},
			expectedChanges:    pulumirpc.DiffResponse_DIFF_SOME,
			expectReplace:      true,
			expectDetailedDiff: true,
			changedProperties:  []string{"forceDestroy"},
		},
		{
			name: "forceDestroy changed from false to true",
			oldInputs: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"projectName":      resource.NewPropertyValue("test-project"),
				"stackName":        resource.NewPropertyValue("dev"),
				"forceDestroy":     resource.NewPropertyValue(false),
			},
			newInputs: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"projectName":      resource.NewPropertyValue("test-project"),
				"stackName":        resource.NewPropertyValue("dev"),
				"forceDestroy":     resource.NewPropertyValue(true),
			},
			expectedChanges:    pulumirpc.DiffResponse_DIFF_SOME,
			expectReplace:      true,
			expectDetailedDiff: true,
			changedProperties:  []string{"forceDestroy"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := PulumiServiceStackResource{}

			olds, err := plugin.MarshalProperties(tt.oldInputs, plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
			require.NoError(t, err, "Failed to marshal old inputs")

			news, err := plugin.MarshalProperties(tt.newInputs, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
			require.NoError(t, err, "Failed to marshal new inputs")

			req := &pulumirpc.DiffRequest{
				Urn:       "urn:pulumi:test::test::pulumiservice:index:Stack::test",
				Id:        "test-org/test-project/dev",
				OldInputs: olds,
				News:      news,
			}

			resp, err := provider.Diff(req)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedChanges, resp.Changes)
			assert.Equal(t, tt.expectReplace, resp.DeleteBeforeReplace)
			assert.Equal(t, tt.expectDetailedDiff, resp.HasDetailedDiff)

			if tt.expectDetailedDiff {
				assert.NotEmpty(t, resp.DetailedDiff)
				// Verify that the expected properties are in the detailed diff
				for _, prop := range tt.changedProperties {
					assert.Contains(t, resp.DetailedDiff, prop, "Expected property %s to be in detailed diff", prop)
				}
			}
		})
	}
}

func TestPulumiServiceStackResource_Check(t *testing.T) {
	tests := []struct {
		name            string
		inputs          resource.PropertyMap
		expectFailures  bool
		failureMessages []string
	}{
		{
			name: "valid inputs",
			inputs: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"projectName":      resource.NewPropertyValue("test-project"),
				"stackName":        resource.NewPropertyValue("dev"),
			},
			expectFailures: false,
		},
		{
			name: "valid inputs with forceDestroy",
			inputs: resource.PropertyMap{
				"organizationName": resource.NewPropertyValue("test-org"),
				"projectName":      resource.NewPropertyValue("test-project"),
				"stackName":        resource.NewPropertyValue("dev"),
				"forceDestroy":     resource.NewPropertyValue(true),
			},
			expectFailures: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := PulumiServiceStackResource{}

			inputProps, err := plugin.MarshalProperties(tt.inputs, plugin.MarshalOptions{})
			require.NoError(t, err, "Failed to marshal input properties")

			req := &pulumirpc.CheckRequest{
				Urn:  "urn:pulumi:test::test::pulumiservice:index:Stack::test",
				News: inputProps,
			}

			resp, err := provider.Check(req)

			require.NoError(t, err)

			if tt.expectFailures {
				assert.NotNil(t, resp.Failures)
				assert.Greater(t, len(resp.Failures), 0)
				for _, expectedMsg := range tt.failureMessages {
					found := false
					for _, failure := range resp.Failures {
						if failure.Reason == expectedMsg {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected failure message not found: %s", expectedMsg)
				}
			} else {
				if resp.Failures != nil {
					assert.Empty(t, resp.Failures, "Expected no failures but got: %v", resp.Failures)
				}
			}

			// Check should return the inputs unchanged
			assert.Equal(t, inputProps, resp.Inputs)
		})
	}
}

func TestPulumiServiceStackResource_Update(t *testing.T) {
	provider := PulumiServiceStackResource{}

	req := &pulumirpc.UpdateRequest{
		Id:  "test-org/test-project/dev",
		Urn: "urn:pulumi:test::test::pulumiservice:index:Stack::test",
	}

	_, err := provider.Update(req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected call to update")
}
