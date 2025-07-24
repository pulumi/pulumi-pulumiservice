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
package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"path"
)

// In the future we might have different approval rule entities, for now it's only Environments
// Because the API routes require envIdentifiers, I'm making these methods specific, but in the future
// we might want to generalize them
type ApprovalRuleClient interface {
	CreateEnvironmentApprovalRule(ctx context.Context, envId EnvironmentIdentifier, req CreateApprovalRuleRequest) (*ApprovalRule, error)
	GetEnvironmentApprovalRule(ctx context.Context, envId EnvironmentIdentifier, ruleId string) (*ApprovalRule, error)
	UpdateEnvironmentApprovalRule(ctx context.Context, envId EnvironmentIdentifier, ruleId string, req UpdateApprovalRuleRequest) (*ApprovalRule, error)
	DeleteEnvironmentApprovalRule(ctx context.Context, envId EnvironmentIdentifier, ruleId string) error
}

type ApprovalRule struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Enabled     bool                    `json:"enabled"`
	CreatedAt   string                  `json:"createdAt"`
	UpdatedAt   string                  `json:"updatedAt"`
	Rule        ChangeGateRuleOutput    `json:"rule"`
	Target      *ChangeGateTargetOutput `json:"target,omitempty"`
}

type ApprovalRuleInput struct {
	NumApprovalsRequired      int                `json:"numApprovalsRequired"`
	AllowSelfApproval         bool               `json:"allowSelfApproval"`
	RequireReapprovalOnChange bool               `json:"requireReapprovalOnChange"`
	EligibleApprovers         []EligibleApprover `json:"eligibleApprovers"`
}

type ApprovalRuleEligibilityType string

const (
	ApprovalRuleEligibilityTypeTeam       ApprovalRuleEligibilityType = "team_member"
	ApprovalRuleEligibilityTypeUser       ApprovalRuleEligibilityType = "specific_user"
	ApprovalRuleEligibilityTypePermission ApprovalRuleEligibilityType = "has_permission_on_target"
)

type EligibleApprover struct {
	EligibilityType ApprovalRuleEligibilityType `json:"eligibilityType"`
	TeamName        string                      `json:"teamName,omitempty"`
	User            string                      `json:"userLogin,omitempty"`
	RbacPermission  string                      `json:"permission,omitempty"`
}

type EligibleApproverOutput struct {
	EligibilityType ApprovalRuleEligibilityType `json:"eligibilityType"`
	TeamName        string                      `json:"teamName,omitempty"`
	GithubLogin     string                      `json:"githubLogin,omitempty"`
	RbacPermission  string                      `json:"permission,omitempty"`
}

func (out EligibleApproverOutput) toApprover() EligibleApprover {
	return EligibleApprover{
		EligibilityType: out.EligibilityType,
		TeamName:        out.TeamName,
		User:            out.GithubLogin,
		RbacPermission:  out.RbacPermission,
	}
}

func ToApprovers(input []EligibleApproverOutput) []EligibleApprover {
	approvers := []EligibleApprover{}
	for _, approver := range input {
		approvers = append(approvers, approver.toApprover())
	}
	return approvers
}

type CreateApprovalRuleRequest struct {
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Enabled     bool                  `json:"enabled"`
	Rule        ChangeGateRuleInput   `json:"rule"`
	Target      ChangeGateTargetInput `json:"target"`
}

type UpdateApprovalRuleRequest struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Enabled     bool                `json:"enabled"`
	Rule        ChangeGateRuleInput `json:"rule"`
}

type ChangeGateTargetInput struct {
	ActionType string `json:"actionType"`
}

type ApprovalRuleType string

const (
	ChangeGateRuleTypeApproval ApprovalRuleType = "approval_required"
)

type ChangeGateRuleInput struct {
	RuleType                  ApprovalRuleType   `json:"ruleType"`
	NumApprovalsRequired      int                `json:"numApprovalsRequired"`
	AllowSelfApproval         bool               `json:"allowSelfApproval"`
	RequireReapprovalOnChange bool               `json:"requireReapprovalOnChange"`
	EligibleApprovers         []EligibleApprover `json:"eligibleApprovers"`
}

type ChangeGateRuleOutput struct {
	NumApprovalsRequired      int                      `json:"numApprovalsRequired"`
	AllowSelfApproval         bool                     `json:"allowSelfApproval"`
	RequireReapprovalOnChange bool                     `json:"requireReapprovalOnChange"`
	EligibleApproverOutputs   []EligibleApproverOutput `json:"eligibleApprovers"`
}

type ChangeGateTargetOutput struct {
	ActionType string                      `json:"actionType"`
	EntityInfo *ChangeGateTargetEntityInfo `json:"entityInfo,omitempty"`
}

type ChangeGateTargetEntityInfo struct {
	Environment *EnvironmentEntity `json:"environment,omitempty"`
}

type EnvironmentEntity struct {
	Project string `json:"project"`
	Name    string `json:"name"`
}

func (c *Client) CreateEnvironmentApprovalRule(ctx context.Context, envId EnvironmentIdentifier, req CreateApprovalRuleRequest) (*ApprovalRule, error) {
	apiPath := path.Join("preview", "esc", "environments",
		envId.OrgName,
		envId.ProjectName,
		envId.EnvName,
		"change-gates")

	var rule ApprovalRule
	_, err := c.do(ctx, http.MethodPost, apiPath, req, &rule)
	if err != nil {
		return nil, fmt.Errorf("failed to create approval rule: %w", err)
	}

	return &rule, nil
}

func (c *Client) GetEnvironmentApprovalRule(ctx context.Context, envId EnvironmentIdentifier, ruleId string) (*ApprovalRule, error) {
	apiPath := path.Join("preview", "esc", "environments",
		envId.OrgName,
		envId.ProjectName,
		envId.EnvName,
		"change-gates",
		ruleId)

	var rule ApprovalRule
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &rule)
	if err != nil {
		return nil, fmt.Errorf("failed to get approval rule: %w", err)
	}

	return &rule, nil
}

func (c *Client) UpdateEnvironmentApprovalRule(ctx context.Context, envId EnvironmentIdentifier, ruleId string, req UpdateApprovalRuleRequest) (*ApprovalRule, error) {
	apiPath := path.Join("preview", "esc", "environments",
		envId.OrgName,
		envId.ProjectName,
		envId.EnvName,
		"change-gates",
		ruleId)

	var rule ApprovalRule
	_, err := c.do(ctx, http.MethodPut, apiPath, req, &rule)
	if err != nil {
		return nil, fmt.Errorf("failed to update approval rule: %w", err)
	}

	return &rule, nil
}

func (c *Client) DeleteEnvironmentApprovalRule(ctx context.Context, envId EnvironmentIdentifier, ruleId string) error {
	apiPath := path.Join("preview", "esc", "environments",
		envId.OrgName,
		envId.ProjectName,
		envId.EnvName,
		"change-gates",
		ruleId)

	result, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		if result.StatusCode == 404 {
			return nil
		}
		return fmt.Errorf("failed to delete approval rule: %w", err)
	}

	return nil
}
