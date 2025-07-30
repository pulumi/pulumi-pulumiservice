package resources

import (
	"context"
	"fmt"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceApprovalRuleResource struct {
	Client pulumiapi.ApprovalRuleClient
}

type PulumiServiceApprovalRule struct {
	Name                  string                          `json:"name"`
	Enabled               bool                            `json:"enabled"`
	TargetActionType      string                          `json:"targetActionType"`
	EnvironmentIdentifier pulumiapi.EnvironmentIdentifier `json:"environmentIdentifier"`
	ApprovalRuleConfig    pulumiapi.ApprovalRuleInput     `json:"approvalRuleConfig"`
}

func (i *PulumiServiceApprovalRule) ToPropertyMap() resource.PropertyMap {
	pm := resource.PropertyMap{}
	pm["name"] = resource.NewPropertyValue(i.Name)
	pm["enabled"] = resource.NewPropertyValue(i.Enabled)
	pm["targetActionType"] = resource.NewPropertyValue(i.TargetActionType)

	// Environment identifier
	envMap := resource.PropertyMap{}
	envMap["organization"] = resource.NewPropertyValue(i.EnvironmentIdentifier.OrgName)
	envMap["project"] = resource.NewPropertyValue(i.EnvironmentIdentifier.ProjectName)
	envMap["name"] = resource.NewPropertyValue(i.EnvironmentIdentifier.EnvName)
	pm["environmentIdentifier"] = resource.NewPropertyValue(envMap)

	// Approval rule config
	ruleMap := resource.PropertyMap{}
	ruleMap["numApprovalsRequired"] = resource.NewPropertyValue(i.ApprovalRuleConfig.NumApprovalsRequired)
	ruleMap["allowSelfApproval"] = resource.NewPropertyValue(i.ApprovalRuleConfig.AllowSelfApproval)
	ruleMap["requireReapprovalOnChange"] = resource.NewPropertyValue(i.ApprovalRuleConfig.RequireReapprovalOnChange)

	// Eligible approvers
	approvers := []resource.PropertyValue{}
	for _, approver := range i.ApprovalRuleConfig.EligibleApprovers {
		approverMap := resource.PropertyMap{}
		if approver.TeamName != "" {
			approverMap["teamName"] = resource.NewPropertyValue(approver.TeamName)
		}
		if approver.User != "" {
			approverMap["user"] = resource.NewPropertyValue(approver.User)
		}
		if approver.RbacPermission != "" {
			approverMap["rbacPermission"] = resource.NewPropertyValue(approver.RbacPermission)
		}
		approvers = append(approvers, resource.NewPropertyValue(approverMap))
	}
	ruleMap["eligibleApprovers"] = resource.NewPropertyValue(approvers)
	pm["approvalRuleConfig"] = resource.NewPropertyValue(ruleMap)

	return pm
}

func (s *PulumiServiceApprovalRuleResource) ToPulumiServiceApprovalRuleInput(inputMap resource.PropertyMap) (*PulumiServiceApprovalRule, error) {
	rule := PulumiServiceApprovalRule{}

	rule.Name = inputMap["name"].StringValue()
	rule.Enabled = inputMap["enabled"].BoolValue()
	rule.TargetActionType = inputMap["targetActionType"].StringValue()

	// Parse environment identifier
	if inputMap["environmentIdentifier"].HasValue() && inputMap["environmentIdentifier"].IsObject() {
		envMap := inputMap["environmentIdentifier"].ObjectValue()
		rule.EnvironmentIdentifier = pulumiapi.EnvironmentIdentifier{
			OrgName:     envMap["organization"].StringValue(),
			ProjectName: envMap["project"].StringValue(),
			EnvName:     envMap["name"].StringValue(),
		}
	}

	// Parse approval rule config
	if inputMap["approvalRuleConfig"].HasValue() && inputMap["approvalRuleConfig"].IsObject() {
		ruleInputMap := inputMap["approvalRuleConfig"].ObjectValue()
		rule.ApprovalRuleConfig = pulumiapi.ApprovalRuleInput{
			NumApprovalsRequired:      int(ruleInputMap["numApprovalsRequired"].NumberValue()),
			AllowSelfApproval:         ruleInputMap["allowSelfApproval"].BoolValue(),
			RequireReapprovalOnChange: ruleInputMap["requireReapprovalOnChange"].BoolValue(),
		}

		// Parse eligible approvers
		if ruleInputMap["eligibleApprovers"].HasValue() && ruleInputMap["eligibleApprovers"].IsArray() {
			for _, approverValue := range ruleInputMap["eligibleApprovers"].ArrayValue() {
				if approverValue.IsObject() {
					approverMap := approverValue.ObjectValue()
					approver := pulumiapi.EligibleApprover{}

					if approverMap["teamName"].HasValue() && approverMap["teamName"].IsString() {
						approver.TeamName = approverMap["teamName"].StringValue()
						approver.EligibilityType = pulumiapi.ApprovalRuleEligibilityTypeTeam
					}
					if approverMap["user"].HasValue() && approverMap["user"].IsString() {
						approver.User = approverMap["user"].StringValue()
						approver.EligibilityType = pulumiapi.ApprovalRuleEligibilityTypeUser
					}
					if approverMap["rbacPermission"].HasValue() && approverMap["rbacPermission"].IsString() {
						approver.RbacPermission = approverMap["rbacPermission"].StringValue()
						approver.EligibilityType = pulumiapi.ApprovalRuleEligibilityTypePermission
					}

					rule.ApprovalRuleConfig.EligibleApprovers = append(rule.ApprovalRuleConfig.EligibleApprovers, approver)
				}
			}
		}
	}

	return &rule, nil
}

func (s *PulumiServiceApprovalRuleResource) Name() string {
	return "pulumiservice:index:ApprovalRule"
}

func buildApprovalRuleID(envID pulumiapi.EnvironmentIdentifier, ruleID string) string {
	return fmt.Sprintf("environment/%s/%s/%s/%s", envID.OrgName, envID.ProjectName, envID.EnvName, ruleID)
}

func parseApprovalRuleID(compositeID string) (pulumiapi.EnvironmentIdentifier, string, error) {
	parts := strings.Split(compositeID, "/")
	if len(parts) != 5 || parts[0] != "environment" {
		// For now, this is the only type, but we expect to have more types of approvals later on
		return pulumiapi.EnvironmentIdentifier{}, "", fmt.Errorf("invalid approval rule ID format: expected 'environment/{orgName}/{projectName}/{envName}/{ruleID}', got %q", compositeID)
	}

	envID := pulumiapi.EnvironmentIdentifier{
		OrgName:     parts[1],
		ProjectName: parts[2],
		EnvName:     parts[3],
	}
	ruleID := parts[4]

	return envID, ruleID, nil
}

func (s *PulumiServiceApprovalRuleResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	replaceProperties := map[string]bool{
		"environmentIdentifier.organization": true,
		"environmentIdentifier.project":      true,
		"environmentIdentifier.name":         true,
	}
	return util.StandardDiff(req, replaceProperties)
}

func (s *PulumiServiceApprovalRuleResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()

	envID, ruleID, err := parseApprovalRuleID(req.GetId())
	if err != nil {
		return nil, err
	}

	err = s.Client.DeleteEnvironmentApprovalRule(ctx, envID, ruleID)
	if err != nil {
		return nil, err
	}

	return &pbempty.Empty{}, nil
}

func (s *PulumiServiceApprovalRuleResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputs, err := plugin.UnmarshalProperties(req.GetProperties(), util.StandardUnmarshal)
	if err != nil {
		return nil, err
	}

	rule, err := s.ToPulumiServiceApprovalRuleInput(inputs)
	if err != nil {
		return nil, err
	}

	createReq := pulumiapi.CreateApprovalRuleRequest{
		Name:    rule.Name,
		Enabled: rule.Enabled,
		Rule: pulumiapi.ChangeGateRuleInput{
			NumApprovalsRequired:      rule.ApprovalRuleConfig.NumApprovalsRequired,
			AllowSelfApproval:         rule.ApprovalRuleConfig.AllowSelfApproval,
			RequireReapprovalOnChange: rule.ApprovalRuleConfig.RequireReapprovalOnChange,
			EligibleApprovers:         rule.ApprovalRuleConfig.EligibleApprovers,
			RuleType:                  pulumiapi.ChangeGateRuleTypeApproval,
		},
		Target: pulumiapi.ChangeGateTargetInput{
			ActionType: rule.TargetActionType,
		},
	}

	createdRule, err := s.Client.CreateEnvironmentApprovalRule(ctx, rule.EnvironmentIdentifier, createReq)
	if err != nil {
		return nil, err
	}

	outputProperties, err := plugin.MarshalProperties(
		rule.ToPropertyMap(),
		util.StandardMarshal,
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         buildApprovalRuleID(rule.EnvironmentIdentifier, createdRule.ID),
		Properties: outputProperties,
	}, nil
}

func (s *PulumiServiceApprovalRuleResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	inputs, err := plugin.UnmarshalProperties(req.GetNews(), util.StandardUnmarshal)
	if err != nil {
		return nil, err
	}

	var failures []*pulumirpc.CheckFailure

	// Validate eligible approvers if approvalRuleConfig is present
	if inputs["approvalRuleConfig"].HasValue() && inputs["approvalRuleConfig"].IsObject() {
		ruleInputMap := inputs["approvalRuleConfig"].ObjectValue()
		if ruleInputMap["eligibleApprovers"].HasValue() && ruleInputMap["eligibleApprovers"].IsArray() {
			approvers := ruleInputMap["eligibleApprovers"].ArrayValue()
			for i, approverValue := range approvers {
				if approverValue.IsObject() {
					approverMap := approverValue.ObjectValue()

					// Count how many fields are set
					fieldsSet := 0
					if approverMap["teamName"].HasValue() && approverMap["teamName"].IsString() && approverMap["teamName"].StringValue() != "" {
						fieldsSet++
					}
					if approverMap["user"].HasValue() && approverMap["user"].IsString() && approverMap["user"].StringValue() != "" {
						fieldsSet++
					}
					if approverMap["rbacPermission"].HasValue() && approverMap["rbacPermission"].IsString() && approverMap["rbacPermission"].StringValue() != "" {
						fieldsSet++
					}

					// Validate exactly one field is set
					if fieldsSet == 0 {
						failures = append(failures, &pulumirpc.CheckFailure{
							Property: fmt.Sprintf("approvalRuleConfig.eligibleApprovers[%d]", i),
							Reason:   "eligible approver must have exactly one of teamName, user, or rbacPermission set",
						})
					} else if fieldsSet > 1 {
						failures = append(failures, &pulumirpc.CheckFailure{
							Property: fmt.Sprintf("approvalRuleConfig.eligibleApprovers[%d]", i),
							Reason:   "eligible approver must have exactly one of teamName, user, or rbacPermission set, but multiple were provided",
						})
					}
				}
			}
		}
	}

	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: failures}, nil
}

func (s *PulumiServiceApprovalRuleResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	ctx := context.Background()
	inputs, err := plugin.UnmarshalProperties(req.GetNews(), util.StandardUnmarshal)
	if err != nil {
		return nil, err
	}

	rule, err := s.ToPulumiServiceApprovalRuleInput(inputs)
	if err != nil {
		return nil, err
	}

	envID, ruleID, err := parseApprovalRuleID(req.GetId())
	if err != nil {
		return nil, err
	}

	updateReq := pulumiapi.UpdateApprovalRuleRequest{
		Name:    rule.Name,
		Enabled: rule.Enabled,
		Rule: pulumiapi.ChangeGateRuleInput{
			NumApprovalsRequired:      rule.ApprovalRuleConfig.NumApprovalsRequired,
			AllowSelfApproval:         rule.ApprovalRuleConfig.AllowSelfApproval,
			RequireReapprovalOnChange: rule.ApprovalRuleConfig.RequireReapprovalOnChange,
			EligibleApprovers:         rule.ApprovalRuleConfig.EligibleApprovers,
			RuleType:                  pulumiapi.ChangeGateRuleTypeApproval,
		},
	}

	_, err = s.Client.UpdateEnvironmentApprovalRule(ctx, envID, ruleID, updateReq)
	if err != nil {
		return nil, err
	}

	outputProperties, err := plugin.MarshalProperties(
		rule.ToPropertyMap(),
		util.StandardMarshal,
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.UpdateResponse{
		Properties: outputProperties,
	}, nil
}

func (s *PulumiServiceApprovalRuleResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()

	envID, ruleID, err := parseApprovalRuleID(req.GetId())
	if err != nil {
		return nil, err
	}

	apiRule, err := s.Client.GetEnvironmentApprovalRule(ctx, envID, ruleID)
	if err != nil {
		return nil, fmt.Errorf("failure while reading approval rule %q: %w", req.Id, err)
	}
	if apiRule == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	// Convert API response back to our input format
	readRule := PulumiServiceApprovalRule{
		Name:                  apiRule.Name,
		Enabled:               apiRule.Enabled,
		TargetActionType:      apiRule.Target.ActionType,
		EnvironmentIdentifier: envID,
		ApprovalRuleConfig: pulumiapi.ApprovalRuleInput{
			NumApprovalsRequired:      apiRule.Rule.NumApprovalsRequired,
			AllowSelfApproval:         apiRule.Rule.AllowSelfApproval,
			RequireReapprovalOnChange: apiRule.Rule.RequireReapprovalOnChange,
			EligibleApprovers:         pulumiapi.ToApprovers(apiRule.Rule.EligibleApproverOutputs),
		},
	}

	outputs, err := plugin.MarshalProperties(
		readRule.ToPropertyMap(),
		util.StandardMarshal,
	)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         buildApprovalRuleID(envID, apiRule.ID),
		Properties: outputs,
		Inputs:     outputs,
	}, nil
}
