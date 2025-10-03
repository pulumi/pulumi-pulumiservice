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

// Package pulumiapi provides clients for interacting with the Pulumi Service API.
package pulumiapi

import (
	"context"
	"fmt"
	"net/http"
	"path"
)

// ApprovalRuleClient provides methods for managing approval rules.
// In the future we might have different approval rule entities, for now it's only Environments.
// Because the API routes require envIdentifiers, I'm making these methods specific,
// but in the future we might want to generalize them.
type ApprovalRuleClient interface {
	CreateEnvironmentApprovalRule(
		ctx context.Context, orgName string, req CreateApprovalRuleRequest,
	) (*ApprovalRule, error)
	GetEnvironmentApprovalRule(ctx context.Context, orgName string, ruleID string) (*ApprovalRule, error)
	UpdateEnvironmentApprovalRule(
		ctx context.Context, orgName string, ruleID string, req UpdateApprovalRuleRequest,
	) (*ApprovalRule, error)
	DeleteEnvironmentApprovalRule(ctx context.Context, orgName string, ruleID string) error
}

// ApprovalRule represents an approval rule in Pulumi Service.
type ApprovalRule struct {
	ID        string                  `json:"id"`
	Name      string                  `json:"name"`
	Enabled   bool                    `json:"enabled"`
	CreatedAt string                  `json:"createdAt"`
	UpdatedAt string                  `json:"updatedAt"`
	Rule      ChangeGateRuleOutput    `json:"rule"`
	Target    *ChangeGateTargetOutput `json:"target,omitempty"`
}

// ApprovalRuleInput represents the input for creating or updating an approval rule.
type ApprovalRuleInput struct {
	NumApprovalsRequired      int                `json:"numApprovalsRequired"`
	AllowSelfApproval         bool               `json:"allowSelfApproval"`
	RequireReapprovalOnChange bool               `json:"requireReapprovalOnChange"`
	EligibleApprovers         []EligibleApprover `json:"eligibleApprovers"`
}

// ApprovalRuleEligibilityType defines the type of eligibility for an approver.
type ApprovalRuleEligibilityType string

const (
	// ApprovalRuleEligibilityTypeTeam indicates team member eligibility.
	ApprovalRuleEligibilityTypeTeam ApprovalRuleEligibilityType = "team_member"
	// ApprovalRuleEligibilityTypeUser indicates specific user eligibility.
	ApprovalRuleEligibilityTypeUser ApprovalRuleEligibilityType = "specific_user"
	// ApprovalRuleEligibilityTypePermission indicates permission-based eligibility.
	ApprovalRuleEligibilityTypePermission ApprovalRuleEligibilityType = "has_permission_on_target"
)

// EligibleApprover represents an entity eligible to approve.
type EligibleApprover struct {
	EligibilityType ApprovalRuleEligibilityType `json:"eligibilityType"`
	TeamName        string                      `json:"teamName,omitempty"`
	User            string                      `json:"userLogin,omitempty"`
	RbacPermission  string                      `json:"permission,omitempty"`
}

// EligibleApproverOutput represents an eligible approver in API responses.
type EligibleApproverOutput struct {
	EligibilityType ApprovalRuleEligibilityType `json:"eligibilityType"`
	TeamName        string                      `json:"name,omitempty"`
	User            UserInfo                    `json:"user,omitempty"`
	RbacPermission  string                      `json:"permission,omitempty"`
}

// UserInfo represents user information in approval contexts.
type UserInfo struct {
	Name        string `json:"name"`
	GithubLogin string `json:"githubLogin"`
}

func (out EligibleApproverOutput) toApprover() EligibleApprover {
	return EligibleApprover{
		EligibilityType: out.EligibilityType,
		TeamName:        out.TeamName,
		User:            out.User.GithubLogin,
		RbacPermission:  out.RbacPermission,
	}
}

// ToApprovers converts a slice of EligibleApproverOutput to EligibleApprover.
func ToApprovers(input []EligibleApproverOutput) []EligibleApprover {
	approvers := []EligibleApprover{}
	for _, approver := range input {
		approvers = append(approvers, approver.toApprover())
	}
	return approvers
}

// CreateApprovalRuleRequest represents a request to create an approval rule.
type CreateApprovalRuleRequest struct {
	Name    string                `json:"name"`
	Enabled bool                  `json:"enabled"`
	Rule    ChangeGateRuleInput   `json:"rule"`
	Target  ChangeGateTargetInput `json:"target"`
}

// UpdateApprovalRuleRequest represents a request to update an approval rule.
type UpdateApprovalRuleRequest struct {
	Name    string                `json:"name"`
	Enabled bool                  `json:"enabled"`
	Rule    ChangeGateRuleInput   `json:"rule"`
	Target  ChangeGateTargetInput `json:"target"`
}

// ChangeGateTargetInput represents the target for a change gate rule.
type ChangeGateTargetInput struct {
	ActionTypes   []string `json:"actionTypes"`
	EntityType    string   `json:"entityType"`
	QualifiedName string   `json:"qualifiedName"`
}

// ApprovalRuleType defines the type of approval rule.
type ApprovalRuleType string

const (
	// ChangeGateRuleTypeApproval indicates an approval-based change gate.
	ChangeGateRuleTypeApproval ApprovalRuleType = "approval_required"
)

// ChangeGateRuleInput represents input for a change gate rule.
type ChangeGateRuleInput struct {
	RuleType                  ApprovalRuleType   `json:"ruleType"`
	NumApprovalsRequired      int                `json:"numApprovalsRequired"`
	AllowSelfApproval         bool               `json:"allowSelfApproval"`
	RequireReapprovalOnChange bool               `json:"requireReapprovalOnChange"`
	EligibleApprovers         []EligibleApprover `json:"eligibleApprovers"`
}

// ChangeGateRuleOutput represents output from a change gate rule.
type ChangeGateRuleOutput struct {
	NumApprovalsRequired      int                      `json:"numApprovalsRequired"`
	AllowSelfApproval         bool                     `json:"allowSelfApproval"`
	RequireReapprovalOnChange bool                     `json:"requireReapprovalOnChange"`
	EligibleApproverOutputs   []EligibleApproverOutput `json:"eligibleApprovers"`
}

// ChangeGateTargetOutput represents the target output of a change gate.
type ChangeGateTargetOutput struct {
	ActionTypes   []string                    `json:"actionTypes"`
	QualifiedName string                      `json:"qualifiedName"`
	EntityType    string                      `json:"entityType"`
	EntityInfo    *ChangeGateTargetEntityInfo `json:"entityInfo,omitempty"`
}

// ChangeGateTargetEntityInfo represents entity information for a change gate target.
type ChangeGateTargetEntityInfo struct {
	Environment *EnvironmentEntity `json:"environment,omitempty"`
}

// EnvironmentEntity represents an environment entity in a change gate.
type EnvironmentEntity struct {
	Project string `json:"project"`
	Name    string `json:"name"`
}

// CreateEnvironmentApprovalRule creates an approval rule for an environment.
func (c *Client) CreateEnvironmentApprovalRule(
	ctx context.Context, orgName string, req CreateApprovalRuleRequest,
) (*ApprovalRule, error) {
	apiPath := path.Join("change-gates", orgName)

	var rule ApprovalRule
	_, err := c.do(ctx, http.MethodPost, apiPath, req, &rule)
	if err != nil {
		return nil, fmt.Errorf("failed to create approval rule: %w", err)
	}

	return &rule, nil
}

// GetEnvironmentApprovalRule retrieves an approval rule by ID.
func (c *Client) GetEnvironmentApprovalRule(ctx context.Context, orgName string, ruleID string) (*ApprovalRule, error) {
	apiPath := path.Join("change-gates", orgName, ruleID)

	var rule ApprovalRule
	_, err := c.do(ctx, http.MethodGet, apiPath, nil, &rule)
	if err != nil {
		return nil, fmt.Errorf("failed to get approval rule: %w", err)
	}

	return &rule, nil
}

// UpdateEnvironmentApprovalRule updates an existing approval rule.
func (c *Client) UpdateEnvironmentApprovalRule(
	ctx context.Context, orgName string, ruleID string, req UpdateApprovalRuleRequest,
) (*ApprovalRule, error) {
	apiPath := path.Join("change-gates", orgName, ruleID)

	var rule ApprovalRule
	_, err := c.do(ctx, http.MethodPut, apiPath, req, &rule)
	if err != nil {
		return nil, fmt.Errorf("failed to update approval rule: %w", err)
	}

	return &rule, nil
}

// DeleteEnvironmentApprovalRule deletes an approval rule.
func (c *Client) DeleteEnvironmentApprovalRule(ctx context.Context, orgName string, ruleID string) error {
	apiPath := path.Join("change-gates", orgName, ruleID)

	result, err := c.do(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		if result.StatusCode == 404 {
			return nil
		}
		return fmt.Errorf("failed to delete approval rule: %w", err)
	}

	return nil
}
