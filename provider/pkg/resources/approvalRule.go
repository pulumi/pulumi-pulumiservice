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
	"context"
	"fmt"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type ApprovalRule struct{}

var (
	_ infer.CustomCheck[ApprovalRuleInput]                     = &ApprovalRule{}
	_ infer.CustomCreate[ApprovalRuleInput, ApprovalRuleState] = &ApprovalRule{}
	_ infer.CustomUpdate[ApprovalRuleInput, ApprovalRuleState] = &ApprovalRule{}
	_ infer.CustomDelete[ApprovalRuleState]                    = &ApprovalRule{}
	_ infer.CustomRead[ApprovalRuleInput, ApprovalRuleState]   = &ApprovalRule{}
)

func (*ApprovalRule) Annotate(a infer.Annotator) {
	a.Describe(&ApprovalRule{}, "An approval rule for environment deployments.")
	a.SetToken("index", "ApprovalRule")
}

// TargetActionType enumerates the action types an approval rule applies to.
type TargetActionType string

const (
	TargetActionTypeUpdate TargetActionType = "update"
)

const (
	gcEnvironment = "environment"
)

func (TargetActionType) Values() []infer.EnumValue[TargetActionType] {
	return []infer.EnumValue[TargetActionType]{
		{Name: "Update", Value: TargetActionTypeUpdate, Description: "Update action type for approval rules."},
	}
}

// RbacPermission enumerates the RBAC permissions that may be required for approval.
type RbacPermission string

const (
	RbacPermissionRead        RbacPermission = "environment:read"
	RbacPermissionReadDecrypt RbacPermission = "environment:read_decrypt"
	RbacPermissionOpen        RbacPermission = "environment:open"
	RbacPermissionWrite       RbacPermission = "environment:write"
	RbacPermissionDelete      RbacPermission = "environment:delete"
	RbacPermissionClone       RbacPermission = "environment:clone"
	RbacPermissionRotate      RbacPermission = "environment:rotate"
)

func (RbacPermission) Values() []infer.EnumValue[RbacPermission] {
	return []infer.EnumValue[RbacPermission]{
		{Name: "Read", Value: RbacPermissionRead, Description: "Read permission."},
		{Name: "ReadDecrypt", Value: RbacPermissionReadDecrypt, Description: "Read and decrypt permission."},
		{Name: "Open", Value: RbacPermissionOpen, Description: "Open permission."},
		{Name: "Write", Value: RbacPermissionWrite, Description: "Write permission."},
		{Name: "Delete", Value: RbacPermissionDelete, Description: "Delete permission."},
		{Name: "Clone", Value: RbacPermissionClone, Description: "Clone permission."},
		{Name: "Rotate", Value: RbacPermissionRotate, Description: "Rotate permission."},
	}
}

// EnvironmentIdentifier identifies the environment an approval rule applies to.
type EnvironmentIdentifier struct {
	Organization string `pulumi:"organization" provider:"replaceOnChanges"`
	Project      string `pulumi:"project"      provider:"replaceOnChanges"`
	Name         string `pulumi:"name"         provider:"replaceOnChanges"`
}

func (e *EnvironmentIdentifier) Annotate(a infer.Annotator) {
	a.Describe(&e.Organization, "The organization name.")
	a.Describe(&e.Project, "The project name.")
	a.Describe(&e.Name, "The environment name.")
}

// EligibleApprover describes a single entity that can approve a change gated by an
// approval rule. Exactly one of `teamName`, `user`, or `rbacPermission` must be set.
type EligibleApprover struct {
	TeamName       *string         `pulumi:"teamName,optional"`
	User           *string         `pulumi:"user,optional"`
	RbacPermission *RbacPermission `pulumi:"rbacPermission,optional"`
}

func (e *EligibleApprover) Annotate(a infer.Annotator) {
	a.Describe(&e.TeamName, "Name of the team that can approve.")
	a.Describe(&e.User, "Login of the user that can approve.")
	a.Describe(&e.RbacPermission, "RBAC permission that gives right to approve.")
}

// ApprovalRuleConfig captures the approval gate's enforcement configuration.
type ApprovalRuleConfig struct {
	NumApprovalsRequired      int                `pulumi:"numApprovalsRequired"`
	AllowSelfApproval         bool               `pulumi:"allowSelfApproval"`
	RequireReapprovalOnChange bool               `pulumi:"requireReapprovalOnChange"`
	EligibleApprovers         []EligibleApprover `pulumi:"eligibleApprovers"`
}

func (c *ApprovalRuleConfig) Annotate(a infer.Annotator) {
	a.Describe(&c.NumApprovalsRequired, "Number of approvals required.")
	a.Describe(&c.AllowSelfApproval, "Whether self-approval is allowed.")
	a.Describe(&c.RequireReapprovalOnChange, "Whether reapproval is required on changes.")
	a.Describe(&c.EligibleApprovers, "List of eligible approvers.")
}

type ApprovalRuleInput struct {
	Name                  string                `pulumi:"name"                  provider:"replaceOnChanges"`
	Enabled               bool                  `pulumi:"enabled"`
	TargetActionTypes     []TargetActionType    `pulumi:"targetActionTypes"`
	EnvironmentIdentifier EnvironmentIdentifier `pulumi:"environmentIdentifier"`
	ApprovalRuleConfig    ApprovalRuleConfig    `pulumi:"approvalRuleConfig"`
}

func (i *ApprovalRuleInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Name, "The name of the approval rule.")
	a.Describe(&i.Enabled, "Whether the approval rule is enabled.")
	a.Describe(&i.TargetActionTypes, "The type of action this rule applies to.")
	a.Describe(&i.EnvironmentIdentifier, "The environment this rule applies to.")
	a.Describe(&i.ApprovalRuleConfig, "The approval rule configuration.")
}

type ApprovalRuleState struct {
	ApprovalRuleInput
}

func (*ApprovalRule) Check(
	ctx context.Context, req infer.CheckRequest,
) (infer.CheckResponse[ApprovalRuleInput], error) {
	in, failures, err := infer.DefaultCheck[ApprovalRuleInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[ApprovalRuleInput]{}, err
	}
	for i, approver := range in.ApprovalRuleConfig.EligibleApprovers {
		set := 0
		if approver.TeamName != nil && *approver.TeamName != "" {
			set++
		}
		if approver.User != nil && *approver.User != "" {
			set++
		}
		if approver.RbacPermission != nil && *approver.RbacPermission != "" {
			set++
		}
		switch {
		case set == 0:
			failures = append(failures, p.CheckFailure{
				Property: fmt.Sprintf("approvalRuleConfig.eligibleApprovers[%d]", i),
				Reason:   "eligible approver must have exactly one of teamName, user, or rbacPermission set",
			})
		case set > 1:
			failures = append(failures, p.CheckFailure{
				Property: fmt.Sprintf("approvalRuleConfig.eligibleApprovers[%d]", i),
				Reason: "eligible approver must have exactly one of teamName, user, or rbacPermission set, " +
					"but multiple were provided",
			})
		}
	}
	return infer.CheckResponse[ApprovalRuleInput]{Inputs: in, Failures: failures}, nil
}

func (*ApprovalRule) Create(
	ctx context.Context, req infer.CreateRequest[ApprovalRuleInput],
) (infer.CreateResponse[ApprovalRuleState], error) {
	if req.DryRun {
		return infer.CreateResponse[ApprovalRuleState]{
			Output: ApprovalRuleState{ApprovalRuleInput: req.Inputs},
		}, nil
	}
	createReq := pulumiapi.CreateApprovalRuleRequest{
		Name:    req.Inputs.Name,
		Enabled: req.Inputs.Enabled,
		Rule:    approvalRuleConfigToAPI(req.Inputs.ApprovalRuleConfig),
		Target:  approvalRuleTargetToAPI(req.Inputs.EnvironmentIdentifier, req.Inputs.TargetActionTypes),
	}
	created, err := config.GetClient(ctx).CreateEnvironmentApprovalRule(
		ctx, req.Inputs.EnvironmentIdentifier.Organization, createReq,
	)
	if err != nil {
		return infer.CreateResponse[ApprovalRuleState]{}, fmt.Errorf(
			"creating approval rule %q: %w", req.Inputs.Name, err,
		)
	}
	return infer.CreateResponse[ApprovalRuleState]{
		ID:     buildApprovalRuleID(req.Inputs.EnvironmentIdentifier, created.ID),
		Output: ApprovalRuleState{ApprovalRuleInput: req.Inputs},
	}, nil
}

func (*ApprovalRule) Update(
	ctx context.Context, req infer.UpdateRequest[ApprovalRuleInput, ApprovalRuleState],
) (infer.UpdateResponse[ApprovalRuleState], error) {
	if req.DryRun {
		return infer.UpdateResponse[ApprovalRuleState]{
			Output: ApprovalRuleState{ApprovalRuleInput: req.Inputs},
		}, nil
	}
	_, ruleID, err := parseApprovalRuleID(req.ID)
	if err != nil {
		return infer.UpdateResponse[ApprovalRuleState]{}, err
	}
	updateReq := pulumiapi.UpdateApprovalRuleRequest{
		Name:    req.Inputs.Name,
		Enabled: req.Inputs.Enabled,
		Rule:    approvalRuleConfigToAPI(req.Inputs.ApprovalRuleConfig),
		Target:  approvalRuleTargetToAPI(req.Inputs.EnvironmentIdentifier, req.Inputs.TargetActionTypes),
	}
	if _, err = config.GetClient(ctx).UpdateEnvironmentApprovalRule(
		ctx, req.Inputs.EnvironmentIdentifier.Organization, ruleID, updateReq,
	); err != nil {
		return infer.UpdateResponse[ApprovalRuleState]{}, fmt.Errorf(
			"updating approval rule %q: %w", req.Inputs.Name, err,
		)
	}
	return infer.UpdateResponse[ApprovalRuleState]{
		Output: ApprovalRuleState{ApprovalRuleInput: req.Inputs},
	}, nil
}

func (*ApprovalRule) Delete(
	ctx context.Context, req infer.DeleteRequest[ApprovalRuleState],
) (infer.DeleteResponse, error) {
	envID, ruleID, err := parseApprovalRuleID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteEnvironmentApprovalRule(
		ctx, envID.Organization, ruleID,
	)
}

func (*ApprovalRule) Read(
	ctx context.Context, req infer.ReadRequest[ApprovalRuleInput, ApprovalRuleState],
) (infer.ReadResponse[ApprovalRuleInput, ApprovalRuleState], error) {
	envID, ruleID, err := parseApprovalRuleID(req.ID)
	if err != nil {
		return infer.ReadResponse[ApprovalRuleInput, ApprovalRuleState]{}, err
	}
	apiRule, err := config.GetClient(ctx).GetEnvironmentApprovalRule(ctx, envID.Organization, ruleID)
	if err != nil {
		return infer.ReadResponse[ApprovalRuleInput, ApprovalRuleState]{}, fmt.Errorf(
			"reading approval rule %q: %w", req.ID, err,
		)
	}
	if apiRule == nil {
		return infer.ReadResponse[ApprovalRuleInput, ApprovalRuleState]{}, nil
	}
	inputs := approvalRuleInputFromAPI(envID, *apiRule)
	return infer.ReadResponse[ApprovalRuleInput, ApprovalRuleState]{
		ID:     buildApprovalRuleID(envID, apiRule.ID),
		Inputs: inputs,
		State:  ApprovalRuleState{ApprovalRuleInput: inputs},
	}, nil
}

func approvalRuleConfigToAPI(c ApprovalRuleConfig) pulumiapi.ChangeGateRuleInput {
	return pulumiapi.ChangeGateRuleInput{
		RuleType:                  pulumiapi.ChangeGateRuleTypeApproval,
		NumApprovalsRequired:      c.NumApprovalsRequired,
		AllowSelfApproval:         c.AllowSelfApproval,
		RequireReapprovalOnChange: c.RequireReapprovalOnChange,
		EligibleApprovers:         eligibleApproversToAPI(c.EligibleApprovers),
	}
}

func approvalRuleTargetToAPI(
	env EnvironmentIdentifier, actionTypes []TargetActionType,
) pulumiapi.ChangeGateTargetInput {
	actions := make([]string, len(actionTypes))
	for i, t := range actionTypes {
		actions[i] = string(t)
	}
	return pulumiapi.ChangeGateTargetInput{
		ActionTypes:   actions,
		EntityType:    gcEnvironment,
		QualifiedName: fmt.Sprintf("%s/%s", env.Project, env.Name),
	}
}

func eligibleApproversToAPI(approvers []EligibleApprover) []pulumiapi.EligibleApprover {
	out := make([]pulumiapi.EligibleApprover, 0, len(approvers))
	for _, a := range approvers {
		api := pulumiapi.EligibleApprover{}
		switch {
		case a.TeamName != nil && *a.TeamName != "":
			api.EligibilityType = pulumiapi.ApprovalRuleEligibilityTypeTeam
			api.TeamName = *a.TeamName
		case a.User != nil && *a.User != "":
			api.EligibilityType = pulumiapi.ApprovalRuleEligibilityTypeUser
			api.User = *a.User
		case a.RbacPermission != nil && *a.RbacPermission != "":
			api.EligibilityType = pulumiapi.ApprovalRuleEligibilityTypePermission
			api.RbacPermission = string(*a.RbacPermission)
		}
		out = append(out, api)
	}
	return out
}

func approvalRuleInputFromAPI(env EnvironmentIdentifier, apiRule pulumiapi.ApprovalRule) ApprovalRuleInput {
	actionTypes := []TargetActionType{}
	if apiRule.Target != nil {
		actionTypes = make([]TargetActionType, len(apiRule.Target.ActionTypes))
		for i, t := range apiRule.Target.ActionTypes {
			actionTypes[i] = TargetActionType(t)
		}
	}
	approvers := make([]EligibleApprover, 0, len(apiRule.Rule.EligibleApproverOutputs))
	for _, a := range apiRule.Rule.EligibleApproverOutputs {
		approvers = append(approvers, eligibleApproverFromAPIOutput(a))
	}
	return ApprovalRuleInput{
		Name:                  apiRule.Name,
		Enabled:               apiRule.Enabled,
		TargetActionTypes:     actionTypes,
		EnvironmentIdentifier: env,
		ApprovalRuleConfig: ApprovalRuleConfig{
			NumApprovalsRequired:      apiRule.Rule.NumApprovalsRequired,
			AllowSelfApproval:         apiRule.Rule.AllowSelfApproval,
			RequireReapprovalOnChange: apiRule.Rule.RequireReapprovalOnChange,
			EligibleApprovers:         approvers,
		},
	}
}

func eligibleApproverFromAPIOutput(out pulumiapi.EligibleApproverOutput) EligibleApprover {
	approver := EligibleApprover{}
	switch out.EligibilityType {
	case pulumiapi.ApprovalRuleEligibilityTypeTeam:
		t := out.TeamName
		approver.TeamName = &t
	case pulumiapi.ApprovalRuleEligibilityTypeUser:
		u := out.User.GithubLogin
		approver.User = &u
	case pulumiapi.ApprovalRuleEligibilityTypePermission:
		perm := RbacPermission(out.RbacPermission)
		approver.RbacPermission = &perm
	}
	return approver
}

// buildApprovalRuleID encodes the composite resource ID using the legacy format so
// existing stacks continue to resolve. The format is:
//
//	environment/{org}/{project}/{env}/{ruleID}
func buildApprovalRuleID(env EnvironmentIdentifier, ruleID string) string {
	return fmt.Sprintf("environment/%s/%s/%s/%s", env.Organization, env.Project, env.Name, ruleID)
}

func parseApprovalRuleID(compositeID string) (EnvironmentIdentifier, string, error) {
	parts := strings.Split(compositeID, "/")
	if len(parts) != 5 || parts[0] != gcEnvironment {
		return EnvironmentIdentifier{}, "", fmt.Errorf(
			"invalid approval rule ID format: expected 'environment/{orgName}/{projectName}/{envName}/{ruleID}', got %q",
			compositeID,
		)
	}
	return EnvironmentIdentifier{
		Organization: parts[1],
		Project:      parts[2],
		Name:         parts[3],
	}, parts[4], nil
}
