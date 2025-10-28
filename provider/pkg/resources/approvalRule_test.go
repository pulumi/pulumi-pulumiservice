// Copyright 2016-2025, Pulumi Corporation.
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
	"context"
	"testing"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"
)

// Mock function types for ApprovalRule
type createApprovalRuleFunc func(ctx context.Context, orgName string, req pulumiapi.CreateApprovalRuleRequest) (*pulumiapi.ApprovalRule, error)
type getApprovalRuleFunc func(ctx context.Context, orgName, ruleID string) (*pulumiapi.ApprovalRule, error)
type updateApprovalRuleFunc func(ctx context.Context, orgName, ruleID string, req pulumiapi.UpdateApprovalRuleRequest) (*pulumiapi.ApprovalRule, error)
type deleteApprovalRuleFunc func(ctx context.Context, orgName, ruleID string) error

// ApprovalRuleClientMock mocks the pulumiapi.ApprovalRuleClient interface
type ApprovalRuleClientMock struct {
	createFunc createApprovalRuleFunc
	getFunc    getApprovalRuleFunc
	updateFunc updateApprovalRuleFunc
	deleteFunc deleteApprovalRuleFunc
}

func (c *ApprovalRuleClientMock) CreateEnvironmentApprovalRule(ctx context.Context, orgName string, req pulumiapi.CreateApprovalRuleRequest) (*pulumiapi.ApprovalRule, error) {
	if c.createFunc != nil {
		return c.createFunc(ctx, orgName, req)
	}
	return nil, nil
}

func (c *ApprovalRuleClientMock) GetEnvironmentApprovalRule(ctx context.Context, orgName, ruleID string) (*pulumiapi.ApprovalRule, error) {
	if c.getFunc != nil {
		return c.getFunc(ctx, orgName, ruleID)
	}
	return nil, nil
}

func (c *ApprovalRuleClientMock) UpdateEnvironmentApprovalRule(ctx context.Context, orgName, ruleID string, req pulumiapi.UpdateApprovalRuleRequest) (*pulumiapi.ApprovalRule, error) {
	if c.updateFunc != nil {
		return c.updateFunc(ctx, orgName, ruleID, req)
	}
	return nil, nil
}

func (c *ApprovalRuleClientMock) DeleteEnvironmentApprovalRule(ctx context.Context, orgName, ruleID string) error {
	if c.deleteFunc != nil {
		return c.deleteFunc(ctx, orgName, ruleID)
	}
	return nil
}

// Helper function to convert INPUT approvers to OUTPUT approvers for mock responses
func toEligibleApproverOutputs(inputApprovers []pulumiapi.EligibleApprover) []pulumiapi.EligibleApproverOutput {
	outputs := make([]pulumiapi.EligibleApproverOutput, len(inputApprovers))
	for i, approver := range inputApprovers {
		output := pulumiapi.EligibleApproverOutput{
			EligibilityType: approver.EligibilityType,
			TeamName:        approver.TeamName,
			RbacPermission:  approver.RbacPermission,
		}
		if approver.User != "" {
			output.User = pulumiapi.UserInfo{
				GithubLogin: approver.User,
				Name:        approver.User, // For tests, use same value
			}
		}
		outputs[i] = output
	}
	return outputs
}

// Helper function to convert INPUT rule to OUTPUT rule for mock responses
func toChangeGateRuleOutput(inputRule pulumiapi.ChangeGateRuleInput) pulumiapi.ChangeGateRuleOutput {
	return pulumiapi.ChangeGateRuleOutput{
		NumApprovalsRequired:      inputRule.NumApprovalsRequired,
		AllowSelfApproval:         inputRule.AllowSelfApproval,
		RequireReapprovalOnChange: inputRule.RequireReapprovalOnChange,
		EligibleApproverOutputs:   toEligibleApproverOutputs(inputRule.EligibleApprovers),
	}
}

// TestApprovalRule_Read_NotFound tests Read when rule not found
func TestApprovalRule_Read_NotFound(t *testing.T) {
	mockClient := &ApprovalRuleClientMock{
		getFunc: func(ctx context.Context, orgName, ruleID string) (*pulumiapi.ApprovalRule, error) {
			return nil, nil
		},
	}

	provider := PulumiServiceApprovalRuleResource{
		Client: mockClient,
	}

	req := &pulumirpc.ReadRequest{
		Id:  "environment/test-org/test-proj/test-env/rule-123",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
	}

	resp, err := provider.Read(req)

	assert.NoError(t, err)
	assert.Equal(t, "", resp.Id)
	assert.Nil(t, resp.Properties)
}

// TestApprovalRule_Read_Found tests Read when rule is found
func TestApprovalRule_Read_Found(t *testing.T) {
	mockClient := &ApprovalRuleClientMock{
		getFunc: func(ctx context.Context, orgName, ruleID string) (*pulumiapi.ApprovalRule, error) {
			assert.Equal(t, "test-org", orgName)
			assert.Equal(t, "rule-123", ruleID)
			return &pulumiapi.ApprovalRule{
				ID:      "rule-123",
				Name:    "my-rule",
				Enabled: true,
				Rule: pulumiapi.ChangeGateRuleOutput{
					NumApprovalsRequired: 1,
					AllowSelfApproval:    false,
					EligibleApproverOutputs: []pulumiapi.EligibleApproverOutput{
						{TeamName: "team1", EligibilityType: pulumiapi.ApprovalRuleEligibilityTypeTeam},
					},
				},
				Target: &pulumiapi.ChangeGateTargetOutput{
					QualifiedName: "test-proj/test-env",
					EntityType:    "environment",
					ActionTypes:   []string{"update"},
				},
			}, nil
		},
	}

	provider := PulumiServiceApprovalRuleResource{Client: mockClient}

	req := &pulumirpc.ReadRequest{
		Id:  "environment/test-org/test-proj/test-env/rule-123",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
	}

	resp, err := provider.Read(req)

	assert.NoError(t, err)
	assert.Equal(t, "environment/test-org/test-proj/test-env/rule-123", resp.Id)
	assert.NotNil(t, resp.Properties)
}

// TestApprovalRule_Read_MultipleApprovers tests Read with multiple eligible approvers
func TestApprovalRule_Read_MultipleApprovers(t *testing.T) {
	mockClient := &ApprovalRuleClientMock{
		getFunc: func(ctx context.Context, orgName, ruleID string) (*pulumiapi.ApprovalRule, error) {
			return &pulumiapi.ApprovalRule{
				ID:      "rule-123",
				Name:    "my-rule",
				Enabled: true,
				Rule: pulumiapi.ChangeGateRuleOutput{
					NumApprovalsRequired: 2,
					EligibleApproverOutputs: []pulumiapi.EligibleApproverOutput{
						{TeamName: "team1", EligibilityType: pulumiapi.ApprovalRuleEligibilityTypeTeam},
						{User: pulumiapi.UserInfo{GithubLogin: "user1", Name: "user1"}, EligibilityType: pulumiapi.ApprovalRuleEligibilityTypeUser},
						{RbacPermission: "admin", EligibilityType: pulumiapi.ApprovalRuleEligibilityTypePermission},
					},
				},
				Target: &pulumiapi.ChangeGateTargetOutput{QualifiedName: "test-proj/test-env"},
			}, nil
		},
	}

	provider := PulumiServiceApprovalRuleResource{Client: mockClient}

	req := &pulumirpc.ReadRequest{
		Id:  "environment/test-org/test-proj/test-env/rule-123",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
	}

	resp, err := provider.Read(req)

	assert.NoError(t, err)
	propMap, err := plugin.UnmarshalProperties(resp.Properties, plugin.MarshalOptions{SkipNulls: true})
	require.NoError(t, err)
	approvers := propMap["approvalRuleConfig"].ObjectValue()["eligibleApprovers"].ArrayValue()
	assert.Len(t, approvers, 3)
}

// TestApprovalRule_Read_InvalidID tests Read with malformed ID
func TestApprovalRule_Read_InvalidID(t *testing.T) {
	provider := PulumiServiceApprovalRuleResource{Client: &ApprovalRuleClientMock{}}

	req := &pulumirpc.ReadRequest{
		Id:  "invalid/id",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
	}

	resp, err := provider.Read(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestApprovalRule_Create_Success tests successful creation with team approver
func TestApprovalRule_Create_Success(t *testing.T) {
	mockClient := &ApprovalRuleClientMock{
		createFunc: func(ctx context.Context, orgName string, req pulumiapi.CreateApprovalRuleRequest) (*pulumiapi.ApprovalRule, error) {
			assert.Equal(t, "test-org", orgName)
			assert.Equal(t, "my-rule", req.Name)
			assert.True(t, req.Enabled)
			assert.Len(t, req.Rule.EligibleApprovers, 1)
			return &pulumiapi.ApprovalRule{
				ID:      "rule-123",
				Name:    req.Name,
				Enabled: req.Enabled,
				Rule:    toChangeGateRuleOutput(req.Rule),
				Target: &pulumiapi.ChangeGateTargetOutput{
					QualifiedName: "test-proj/test-env",
				},
			}, nil
		},
	}

	provider := PulumiServiceApprovalRuleResource{Client: mockClient}

	approverMap := resource.PropertyMap{
		"teamName": resource.NewStringProperty("team1"),
	}

	ruleConfigMap := resource.PropertyMap{
		"numApprovalsRequired":      resource.NewNumberProperty(1),
		"allowSelfApproval":         resource.NewBoolProperty(false),
		"requireReapprovalOnChange": resource.NewBoolProperty(false),
		"eligibleApprovers":         resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(approverMap)}),
	}

	envIdMap := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"name":         resource.NewStringProperty("test-env"),
	}

	inputs := resource.PropertyMap{
		"name":                  resource.NewStringProperty("my-rule"),
		"enabled":               resource.NewBoolProperty(true),
		"targetActionTypes":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(envIdMap),
		"approvalRuleConfig":    resource.NewObjectProperty(ruleConfigMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "environment/test-org/test-proj/test-env/rule-123", resp.Id)
}

// TestApprovalRule_Create_UserApprover tests creation with user approver
func TestApprovalRule_Create_UserApprover(t *testing.T) {
	mockClient := &ApprovalRuleClientMock{
		createFunc: func(ctx context.Context, orgName string, req pulumiapi.CreateApprovalRuleRequest) (*pulumiapi.ApprovalRule, error) {
			assert.Len(t, req.Rule.EligibleApprovers, 1)
			assert.Equal(t, "user1", req.Rule.EligibleApprovers[0].User)
			return &pulumiapi.ApprovalRule{
				ID:     "rule-123",
				Name:   req.Name,
				Rule:   toChangeGateRuleOutput(req.Rule),
				Target: &pulumiapi.ChangeGateTargetOutput{QualifiedName: "test-proj/test-env"},
			}, nil
		},
	}

	provider := PulumiServiceApprovalRuleResource{Client: mockClient}

	approverMap := resource.PropertyMap{
		"user": resource.NewStringProperty("user1"),
	}

	ruleConfigMap := resource.PropertyMap{
		"numApprovalsRequired":      resource.NewNumberProperty(1),
		"allowSelfApproval":         resource.NewBoolProperty(false),
		"requireReapprovalOnChange": resource.NewBoolProperty(false),
		"eligibleApprovers":         resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(approverMap)}),
	}

	envIdMap := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"name":         resource.NewStringProperty("test-env"),
	}

	inputs := resource.PropertyMap{
		"name":                  resource.NewStringProperty("my-rule"),
		"enabled":               resource.NewBoolProperty(true),
		"targetActionTypes":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(envIdMap),
		"approvalRuleConfig":    resource.NewObjectProperty(ruleConfigMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestApprovalRule_Create_RbacApprover tests creation with permission approver
func TestApprovalRule_Create_RbacApprover(t *testing.T) {
	mockClient := &ApprovalRuleClientMock{
		createFunc: func(ctx context.Context, orgName string, req pulumiapi.CreateApprovalRuleRequest) (*pulumiapi.ApprovalRule, error) {
			assert.Len(t, req.Rule.EligibleApprovers, 1)
			assert.Equal(t, "admin", req.Rule.EligibleApprovers[0].RbacPermission)
			return &pulumiapi.ApprovalRule{
				ID:     "rule-123",
				Name:   req.Name,
				Rule:   toChangeGateRuleOutput(req.Rule),
				Target: &pulumiapi.ChangeGateTargetOutput{QualifiedName: "test-proj/test-env"},
			}, nil
		},
	}

	provider := PulumiServiceApprovalRuleResource{Client: mockClient}

	approverMap := resource.PropertyMap{
		"rbacPermission": resource.NewStringProperty("admin"),
	}

	ruleConfigMap := resource.PropertyMap{
		"numApprovalsRequired":      resource.NewNumberProperty(1),
		"allowSelfApproval":         resource.NewBoolProperty(false),
		"requireReapprovalOnChange": resource.NewBoolProperty(false),
		"eligibleApprovers":         resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(approverMap)}),
	}

	envIdMap := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"name":         resource.NewStringProperty("test-env"),
	}

	inputs := resource.PropertyMap{
		"name":                  resource.NewStringProperty("my-rule"),
		"enabled":               resource.NewBoolProperty(true),
		"targetActionTypes":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(envIdMap),
		"approvalRuleConfig":    resource.NewObjectProperty(ruleConfigMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestApprovalRule_Create_MultipleApprovers tests creation with multiple eligible approvers
func TestApprovalRule_Create_MultipleApprovers(t *testing.T) {
	mockClient := &ApprovalRuleClientMock{
		createFunc: func(ctx context.Context, orgName string, req pulumiapi.CreateApprovalRuleRequest) (*pulumiapi.ApprovalRule, error) {
			assert.Len(t, req.Rule.EligibleApprovers, 3)
			return &pulumiapi.ApprovalRule{
				ID:     "rule-123",
				Name:   req.Name,
				Rule:   toChangeGateRuleOutput(req.Rule),
				Target: &pulumiapi.ChangeGateTargetOutput{QualifiedName: "test-proj/test-env"},
			}, nil
		},
	}

	provider := PulumiServiceApprovalRuleResource{Client: mockClient}

	ruleConfigMap := resource.PropertyMap{
		"numApprovalsRequired":      resource.NewNumberProperty(2),
		"allowSelfApproval":         resource.NewBoolProperty(false),
		"requireReapprovalOnChange": resource.NewBoolProperty(false),
		"eligibleApprovers": resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewObjectProperty(resource.PropertyMap{"teamName": resource.NewStringProperty("team1")}),
			resource.NewObjectProperty(resource.PropertyMap{"user": resource.NewStringProperty("user1")}),
			resource.NewObjectProperty(resource.PropertyMap{"rbacPermission": resource.NewStringProperty("admin")}),
		}),
	}

	envIdMap := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"name":         resource.NewStringProperty("test-env"),
	}

	inputs := resource.PropertyMap{
		"name":                  resource.NewStringProperty("my-rule"),
		"enabled":               resource.NewBoolProperty(true),
		"targetActionTypes":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(envIdMap),
		"approvalRuleConfig":    resource.NewObjectProperty(ruleConfigMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestApprovalRule_Create_AllowSelfApproval tests creation with allowSelfApproval=true
func TestApprovalRule_Create_AllowSelfApproval(t *testing.T) {
	mockClient := &ApprovalRuleClientMock{
		createFunc: func(ctx context.Context, orgName string, req pulumiapi.CreateApprovalRuleRequest) (*pulumiapi.ApprovalRule, error) {
			assert.True(t, req.Rule.AllowSelfApproval)
			return &pulumiapi.ApprovalRule{
				ID:     "rule-123",
				Name:   req.Name,
				Rule:   toChangeGateRuleOutput(req.Rule),
				Target: &pulumiapi.ChangeGateTargetOutput{QualifiedName: "test-proj/test-env"},
			}, nil
		},
	}

	provider := PulumiServiceApprovalRuleResource{Client: mockClient}

	approverMap := resource.PropertyMap{
		"teamName": resource.NewStringProperty("team1"),
	}

	ruleConfigMap := resource.PropertyMap{
		"numApprovalsRequired":      resource.NewNumberProperty(1),
		"allowSelfApproval":         resource.NewBoolProperty(true),
		"requireReapprovalOnChange": resource.NewBoolProperty(false),
		"eligibleApprovers":         resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(approverMap)}),
	}

	envIdMap := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"name":         resource.NewStringProperty("test-env"),
	}

	inputs := resource.PropertyMap{
		"name":                  resource.NewStringProperty("my-rule"),
		"enabled":               resource.NewBoolProperty(true),
		"targetActionTypes":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(envIdMap),
		"approvalRuleConfig":    resource.NewObjectProperty(ruleConfigMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CreateRequest{
		Urn:        "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		Properties: inputsStruct,
	}

	resp, err := provider.Create(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestApprovalRule_Update_Success tests successful update
func TestApprovalRule_Update_Success(t *testing.T) {
	mockClient := &ApprovalRuleClientMock{
		updateFunc: func(ctx context.Context, orgName, ruleID string, req pulumiapi.UpdateApprovalRuleRequest) (*pulumiapi.ApprovalRule, error) {
			assert.Equal(t, "test-org", orgName)
			assert.Equal(t, "rule-123", ruleID)
			return &pulumiapi.ApprovalRule{
				ID:     ruleID,
				Name:   req.Name,
				Rule:   toChangeGateRuleOutput(req.Rule),
				Target: &pulumiapi.ChangeGateTargetOutput{QualifiedName: "test-proj/test-env"},
			}, nil
		},
	}

	provider := PulumiServiceApprovalRuleResource{Client: mockClient}

	approverMap := resource.PropertyMap{
		"teamName": resource.NewStringProperty("team1"),
	}

	ruleConfigMap := resource.PropertyMap{
		"numApprovalsRequired":      resource.NewNumberProperty(1),
		"allowSelfApproval":         resource.NewBoolProperty(false),
		"requireReapprovalOnChange": resource.NewBoolProperty(false),
		"eligibleApprovers":         resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(approverMap)}),
	}

	envIdMap := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"name":         resource.NewStringProperty("test-env"),
	}

	inputs := resource.PropertyMap{
		"name":                  resource.NewStringProperty("updated-rule"),
		"enabled":               resource.NewBoolProperty(true),
		"targetActionTypes":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(envIdMap),
		"approvalRuleConfig":    resource.NewObjectProperty(ruleConfigMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.UpdateRequest{
		Id:   "environment/test-org/test-proj/test-env/rule-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		News: inputsStruct,
		Olds: inputsStruct,
	}

	resp, err := provider.Update(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestApprovalRule_Update_ChangeApprovers tests updating eligible approvers list
func TestApprovalRule_Update_ChangeApprovers(t *testing.T) {
	mockClient := &ApprovalRuleClientMock{
		updateFunc: func(ctx context.Context, orgName, ruleID string, req pulumiapi.UpdateApprovalRuleRequest) (*pulumiapi.ApprovalRule, error) {
			assert.Len(t, req.Rule.EligibleApprovers, 2)
			return &pulumiapi.ApprovalRule{
				ID:     ruleID,
				Name:   req.Name,
				Rule:   toChangeGateRuleOutput(req.Rule),
				Target: &pulumiapi.ChangeGateTargetOutput{QualifiedName: "test-proj/test-env"},
			}, nil
		},
	}

	provider := PulumiServiceApprovalRuleResource{Client: mockClient}

	ruleConfigMap := resource.PropertyMap{
		"numApprovalsRequired":      resource.NewNumberProperty(2),
		"allowSelfApproval":         resource.NewBoolProperty(false),
		"requireReapprovalOnChange": resource.NewBoolProperty(false),
		"eligibleApprovers": resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewObjectProperty(resource.PropertyMap{"teamName": resource.NewStringProperty("team1")}),
			resource.NewObjectProperty(resource.PropertyMap{"user": resource.NewStringProperty("user1")}),
		}),
	}

	envIdMap := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"name":         resource.NewStringProperty("test-env"),
	}

	inputs := resource.PropertyMap{
		"name":                  resource.NewStringProperty("my-rule"),
		"enabled":               resource.NewBoolProperty(true),
		"targetActionTypes":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(envIdMap),
		"approvalRuleConfig":    resource.NewObjectProperty(ruleConfigMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.UpdateRequest{
		Id:   "environment/test-org/test-proj/test-env/rule-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		News: inputsStruct,
		Olds: inputsStruct,
	}

	resp, err := provider.Update(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestApprovalRule_Update_DisableRule tests setting enabled=false
func TestApprovalRule_Update_DisableRule(t *testing.T) {
	mockClient := &ApprovalRuleClientMock{
		updateFunc: func(ctx context.Context, orgName, ruleID string, req pulumiapi.UpdateApprovalRuleRequest) (*pulumiapi.ApprovalRule, error) {
			assert.False(t, req.Enabled)
			return &pulumiapi.ApprovalRule{
				ID:      ruleID,
				Name:    req.Name,
				Enabled: false,
				Rule:    toChangeGateRuleOutput(req.Rule),
				Target:  &pulumiapi.ChangeGateTargetOutput{QualifiedName: "test-proj/test-env"},
			}, nil
		},
	}

	provider := PulumiServiceApprovalRuleResource{Client: mockClient}

	approverMap := resource.PropertyMap{
		"teamName": resource.NewStringProperty("team1"),
	}

	ruleConfigMap := resource.PropertyMap{
		"numApprovalsRequired":      resource.NewNumberProperty(1),
		"allowSelfApproval":         resource.NewBoolProperty(false),
		"requireReapprovalOnChange": resource.NewBoolProperty(false),
		"eligibleApprovers":         resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(approverMap)}),
	}

	envIdMap := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"name":         resource.NewStringProperty("test-env"),
	}

	inputs := resource.PropertyMap{
		"name":                  resource.NewStringProperty("my-rule"),
		"enabled":               resource.NewBoolProperty(false),
		"targetActionTypes":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(envIdMap),
		"approvalRuleConfig":    resource.NewObjectProperty(ruleConfigMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.UpdateRequest{
		Id:   "environment/test-org/test-proj/test-env/rule-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		News: inputsStruct,
		Olds: inputsStruct,
	}

	resp, err := provider.Update(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestApprovalRule_Update_ChangeTargetActions tests updating targetActionTypes
func TestApprovalRule_Update_ChangeTargetActions(t *testing.T) {
	mockClient := &ApprovalRuleClientMock{
		updateFunc: func(ctx context.Context, orgName, ruleID string, req pulumiapi.UpdateApprovalRuleRequest) (*pulumiapi.ApprovalRule, error) {
			assert.Contains(t, req.Target.ActionTypes, "update")
			assert.Contains(t, req.Target.ActionTypes, "destroy")
			return &pulumiapi.ApprovalRule{
				ID:     ruleID,
				Name:   req.Name,
				Rule:   toChangeGateRuleOutput(req.Rule),
				Target: &pulumiapi.ChangeGateTargetOutput{QualifiedName: req.Target.QualifiedName, ActionTypes: req.Target.ActionTypes},
			}, nil
		},
	}

	provider := PulumiServiceApprovalRuleResource{Client: mockClient}

	approverMap := resource.PropertyMap{
		"teamName": resource.NewStringProperty("team1"),
	}

	ruleConfigMap := resource.PropertyMap{
		"numApprovalsRequired":      resource.NewNumberProperty(1),
		"allowSelfApproval":         resource.NewBoolProperty(false),
		"requireReapprovalOnChange": resource.NewBoolProperty(false),
		"eligibleApprovers":         resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(approverMap)}),
	}

	envIdMap := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"name":         resource.NewStringProperty("test-env"),
	}

	inputs := resource.PropertyMap{
		"name":    resource.NewStringProperty("my-rule"),
		"enabled": resource.NewBoolProperty(true),
		"targetActionTypes": resource.NewArrayProperty([]resource.PropertyValue{
			resource.NewStringProperty("update"),
			resource.NewStringProperty("destroy"),
		}),
		"environmentIdentifier": resource.NewObjectProperty(envIdMap),
		"approvalRuleConfig":    resource.NewObjectProperty(ruleConfigMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.UpdateRequest{
		Id:   "environment/test-org/test-proj/test-env/rule-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		News: inputsStruct,
		Olds: inputsStruct,
	}

	resp, err := provider.Update(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestApprovalRule_Delete_Success tests successful deletion
func TestApprovalRule_Delete_Success(t *testing.T) {
	mockClient := &ApprovalRuleClientMock{
		deleteFunc: func(ctx context.Context, orgName, ruleID string) error {
			assert.Equal(t, "test-org", orgName)
			assert.Equal(t, "rule-123", ruleID)
			return nil
		},
	}

	provider := PulumiServiceApprovalRuleResource{Client: mockClient}

	req := &pulumirpc.DeleteRequest{
		Id:  "environment/test-org/test-proj/test-env/rule-123",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
	}

	resp, err := provider.Delete(req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

// TestApprovalRule_Delete_InvalidID tests deletion with malformed ID
func TestApprovalRule_Delete_InvalidID(t *testing.T) {
	provider := PulumiServiceApprovalRuleResource{Client: &ApprovalRuleClientMock{}}

	req := &pulumirpc.DeleteRequest{
		Id:  "invalid/id",
		Urn: "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
	}

	resp, err := provider.Delete(req)

	assert.Error(t, err)
	assert.Nil(t, resp)
}

// TestApprovalRule_Diff_EnvironmentChange tests that environment identifier change triggers replacement
// TODO(#586): This test currently fails because StandardDiff doesn't populate the Replaces array.
// The bug is in provider/pkg/util/diff.go - DetailedDiff is populated correctly but Replaces[] stays empty.
func TestApprovalRule_Diff_EnvironmentChange(t *testing.T) {
	t.Skip("TODO(#586): Skipping until StandardDiff populates Replaces array - see https://github.com/pulumi/pulumi-pulumiservice/issues/586")

	provider := PulumiServiceApprovalRuleResource{Client: &ApprovalRuleClientMock{}}

	oldInputs := resource.PropertyMap{
		"name":              resource.NewStringProperty("my-rule"),
		"enabled":           resource.NewBoolProperty(true),
		"targetActionTypes": resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(resource.PropertyMap{
			"organization": resource.NewStringProperty("old-org"),
			"project":      resource.NewStringProperty("test-proj"),
			"name":         resource.NewStringProperty("test-env"),
		}),
		"approvalRuleConfig": resource.NewObjectProperty(resource.PropertyMap{}),
	}

	oldState, err := structpb.NewStruct(oldInputs.Mappable())
	require.NoError(t, err)

	newInputs := resource.PropertyMap{
		"name":              resource.NewStringProperty("my-rule"),
		"enabled":           resource.NewBoolProperty(true),
		"targetActionTypes": resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(resource.PropertyMap{
			"organization": resource.NewStringProperty("new-org"),
			"project":      resource.NewStringProperty("test-proj"),
			"name":         resource.NewStringProperty("test-env"),
		}),
		"approvalRuleConfig": resource.NewObjectProperty(resource.PropertyMap{}),
	}

	newState, err := structpb.NewStruct(newInputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "environment/old-org/test-proj/test-env/rule-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		Olds: oldState,
		News: newState,
	}

	resp, err := provider.Diff(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Contains(t, resp.Replaces, "environmentIdentifier.organization")
}

// TestApprovalRule_Diff_NameChange tests that name change does not trigger replacement
func TestApprovalRule_Diff_NameChange(t *testing.T) {
	provider := PulumiServiceApprovalRuleResource{Client: &ApprovalRuleClientMock{}}

	oldInputs := resource.PropertyMap{
		"name":              resource.NewStringProperty("old-name"),
		"enabled":           resource.NewBoolProperty(true),
		"targetActionTypes": resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(resource.PropertyMap{
			"organization": resource.NewStringProperty("test-org"),
			"project":      resource.NewStringProperty("test-proj"),
			"name":         resource.NewStringProperty("test-env"),
		}),
		"approvalRuleConfig": resource.NewObjectProperty(resource.PropertyMap{}),
	}

	oldState, err := structpb.NewStruct(oldInputs.Mappable())
	require.NoError(t, err)

	newInputs := resource.PropertyMap{
		"name":              resource.NewStringProperty("new-name"),
		"enabled":           resource.NewBoolProperty(true),
		"targetActionTypes": resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(resource.PropertyMap{
			"organization": resource.NewStringProperty("test-org"),
			"project":      resource.NewStringProperty("test-proj"),
			"name":         resource.NewStringProperty("test-env"),
		}),
		"approvalRuleConfig": resource.NewObjectProperty(resource.PropertyMap{}),
	}

	newState, err := structpb.NewStruct(newInputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "environment/test-org/test-proj/test-env/rule-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		Olds: oldState,
		News: newState,
	}

	resp, err := provider.Diff(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotContains(t, resp.Replaces, "name")
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
}

// TestApprovalRule_Diff_NoChanges tests diff with no changes
// TODO(#587): This test currently fails because StandardDiff detects false changes.
// The bug is in provider/pkg/util/diff.go - it reports DIFF_SOME even when properties are identical.
func TestApprovalRule_Diff_NoChanges(t *testing.T) {
	t.Skip("TODO(#587): Skipping until StandardDiff false change detection is fixed - see https://github.com/pulumi/pulumi-pulumiservice/issues/587")

	provider := PulumiServiceApprovalRuleResource{Client: &ApprovalRuleClientMock{}}

	inputs := resource.PropertyMap{
		"name":              resource.NewStringProperty("my-rule"),
		"enabled":           resource.NewBoolProperty(true),
		"targetActionTypes": resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(resource.PropertyMap{
			"organization": resource.NewStringProperty("test-org"),
			"project":      resource.NewStringProperty("test-proj"),
			"name":         resource.NewStringProperty("test-env"),
		}),
		"approvalRuleConfig": resource.NewObjectProperty(resource.PropertyMap{}),
	}

	state, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.DiffRequest{
		Id:   "environment/test-org/test-proj/test-env/rule-123",
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		Olds: state,
		News: state,
	}

	resp, err := provider.Diff(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
}

// TestApprovalRule_Check_ValidTeamApprover tests Check accepts team approver only
func TestApprovalRule_Check_ValidTeamApprover(t *testing.T) {
	provider := PulumiServiceApprovalRuleResource{Client: &ApprovalRuleClientMock{}}

	approverMap := resource.PropertyMap{
		"teamName": resource.NewStringProperty("team1"),
	}

	ruleConfigMap := resource.PropertyMap{
		"numApprovalsRequired":      resource.NewNumberProperty(1),
		"allowSelfApproval":         resource.NewBoolProperty(false),
		"requireReapprovalOnChange": resource.NewBoolProperty(false),
		"eligibleApprovers":         resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(approverMap)}),
	}

	envIdMap := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"name":         resource.NewStringProperty("test-env"),
	}

	inputs := resource.PropertyMap{
		"name":                  resource.NewStringProperty("my-rule"),
		"enabled":               resource.NewBoolProperty(true),
		"targetActionTypes":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(envIdMap),
		"approvalRuleConfig":    resource.NewObjectProperty(ruleConfigMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Failures)
}

// TestApprovalRule_Check_ValidUserApprover tests Check accepts user approver only
func TestApprovalRule_Check_ValidUserApprover(t *testing.T) {
	provider := PulumiServiceApprovalRuleResource{Client: &ApprovalRuleClientMock{}}

	approverMap := resource.PropertyMap{
		"user": resource.NewStringProperty("user1"),
	}

	ruleConfigMap := resource.PropertyMap{
		"numApprovalsRequired":      resource.NewNumberProperty(1),
		"allowSelfApproval":         resource.NewBoolProperty(false),
		"requireReapprovalOnChange": resource.NewBoolProperty(false),
		"eligibleApprovers":         resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(approverMap)}),
	}

	envIdMap := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"name":         resource.NewStringProperty("test-env"),
	}

	inputs := resource.PropertyMap{
		"name":                  resource.NewStringProperty("my-rule"),
		"enabled":               resource.NewBoolProperty(true),
		"targetActionTypes":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(envIdMap),
		"approvalRuleConfig":    resource.NewObjectProperty(ruleConfigMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Failures)
}

// TestApprovalRule_Check_ValidRbacApprover tests Check accepts rbac approver only
func TestApprovalRule_Check_ValidRbacApprover(t *testing.T) {
	provider := PulumiServiceApprovalRuleResource{Client: &ApprovalRuleClientMock{}}

	approverMap := resource.PropertyMap{
		"rbacPermission": resource.NewStringProperty("admin"),
	}

	ruleConfigMap := resource.PropertyMap{
		"numApprovalsRequired":      resource.NewNumberProperty(1),
		"allowSelfApproval":         resource.NewBoolProperty(false),
		"requireReapprovalOnChange": resource.NewBoolProperty(false),
		"eligibleApprovers":         resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(approverMap)}),
	}

	envIdMap := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"name":         resource.NewStringProperty("test-env"),
	}

	inputs := resource.PropertyMap{
		"name":                  resource.NewStringProperty("my-rule"),
		"enabled":               resource.NewBoolProperty(true),
		"targetActionTypes":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(envIdMap),
		"approvalRuleConfig":    resource.NewObjectProperty(ruleConfigMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Failures)
}

// TestApprovalRule_Check_NoFieldsSet tests Check fails when no approver field set
func TestApprovalRule_Check_NoFieldsSet(t *testing.T) {
	provider := PulumiServiceApprovalRuleResource{Client: &ApprovalRuleClientMock{}}

	approverMap := resource.PropertyMap{
		// No fields set
	}

	ruleConfigMap := resource.PropertyMap{
		"numApprovalsRequired":      resource.NewNumberProperty(1),
		"allowSelfApproval":         resource.NewBoolProperty(false),
		"requireReapprovalOnChange": resource.NewBoolProperty(false),
		"eligibleApprovers":         resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(approverMap)}),
	}

	envIdMap := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"name":         resource.NewStringProperty("test-env"),
	}

	inputs := resource.PropertyMap{
		"name":                  resource.NewStringProperty("my-rule"),
		"enabled":               resource.NewBoolProperty(true),
		"targetActionTypes":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(envIdMap),
		"approvalRuleConfig":    resource.NewObjectProperty(ruleConfigMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Failures)
	assert.Contains(t, resp.Failures[0].Reason, "exactly one of")
}

// TestApprovalRule_Check_MultipleFieldsSet tests Check fails when teamName+user both set
func TestApprovalRule_Check_MultipleFieldsSet(t *testing.T) {
	provider := PulumiServiceApprovalRuleResource{Client: &ApprovalRuleClientMock{}}

	approverMap := resource.PropertyMap{
		"teamName": resource.NewStringProperty("team1"),
		"user":     resource.NewStringProperty("user1"),
	}

	ruleConfigMap := resource.PropertyMap{
		"numApprovalsRequired":      resource.NewNumberProperty(1),
		"allowSelfApproval":         resource.NewBoolProperty(false),
		"requireReapprovalOnChange": resource.NewBoolProperty(false),
		"eligibleApprovers":         resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(approverMap)}),
	}

	envIdMap := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"name":         resource.NewStringProperty("test-env"),
	}

	inputs := resource.PropertyMap{
		"name":                  resource.NewStringProperty("my-rule"),
		"enabled":               resource.NewBoolProperty(true),
		"targetActionTypes":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(envIdMap),
		"approvalRuleConfig":    resource.NewObjectProperty(ruleConfigMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Failures)
	assert.Contains(t, resp.Failures[0].Reason, "exactly one of")
}

// TestApprovalRule_Check_AllThreeFieldsSet tests Check fails when teamName+user+rbac all set
func TestApprovalRule_Check_AllThreeFieldsSet(t *testing.T) {
	provider := PulumiServiceApprovalRuleResource{Client: &ApprovalRuleClientMock{}}

	approverMap := resource.PropertyMap{
		"teamName":       resource.NewStringProperty("team1"),
		"user":           resource.NewStringProperty("user1"),
		"rbacPermission": resource.NewStringProperty("admin"),
	}

	ruleConfigMap := resource.PropertyMap{
		"numApprovalsRequired":      resource.NewNumberProperty(1),
		"allowSelfApproval":         resource.NewBoolProperty(false),
		"requireReapprovalOnChange": resource.NewBoolProperty(false),
		"eligibleApprovers":         resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(approverMap)}),
	}

	envIdMap := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"name":         resource.NewStringProperty("test-env"),
	}

	inputs := resource.PropertyMap{
		"name":                  resource.NewStringProperty("my-rule"),
		"enabled":               resource.NewBoolProperty(true),
		"targetActionTypes":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(envIdMap),
		"approvalRuleConfig":    resource.NewObjectProperty(ruleConfigMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Failures)
	assert.Contains(t, resp.Failures[0].Reason, "exactly one of")
}

// TestApprovalRule_Check_EmptyString tests that empty strings are treated as not set
func TestApprovalRule_Check_EmptyString(t *testing.T) {
	provider := PulumiServiceApprovalRuleResource{Client: &ApprovalRuleClientMock{}}

	approverMap := resource.PropertyMap{
		"teamName": resource.NewStringProperty("team1"),
		"user":     resource.NewStringProperty(""), // Empty string should be treated as not set
	}

	ruleConfigMap := resource.PropertyMap{
		"numApprovalsRequired":      resource.NewNumberProperty(1),
		"allowSelfApproval":         resource.NewBoolProperty(false),
		"requireReapprovalOnChange": resource.NewBoolProperty(false),
		"eligibleApprovers":         resource.NewArrayProperty([]resource.PropertyValue{resource.NewObjectProperty(approverMap)}),
	}

	envIdMap := resource.PropertyMap{
		"organization": resource.NewStringProperty("test-org"),
		"project":      resource.NewStringProperty("test-proj"),
		"name":         resource.NewStringProperty("test-env"),
	}

	inputs := resource.PropertyMap{
		"name":                  resource.NewStringProperty("my-rule"),
		"enabled":               resource.NewBoolProperty(true),
		"targetActionTypes":     resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("update")}),
		"environmentIdentifier": resource.NewObjectProperty(envIdMap),
		"approvalRuleConfig":    resource.NewObjectProperty(ruleConfigMap),
	}

	inputsStruct, err := structpb.NewStruct(inputs.Mappable())
	require.NoError(t, err)

	req := &pulumirpc.CheckRequest{
		Urn:  "urn:pulumi:dev::test::pulumiservice:index:ApprovalRule::testRule",
		News: inputsStruct,
	}

	resp, err := provider.Check(req)

	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Failures, "Empty strings should be treated as not set, so only teamName is set which is valid")
}
