// Copyright 2016-2026, Pulumi Corporation.
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
	"fmt"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// scopedPermissionsDescriptor builds a PermissionDescriptor tree that grants
// the supplied scopes only when the request targets the entity identified by
// (expressionType, identity).
//
// Wire shape: a single-entry PermissionDescriptorGroup whose entry is a
// PermissionDescriptorCondition wrapping a PermissionDescriptorAllow. The
// condition compares the request's entity expression to a literal identity.
// All three supported entities (environment, stack, insights-account) share
// this flat {__type, identity} literal shape on the service side.
func scopedPermissionsDescriptor(
	expressionType, literalType, identity string,
	permissions []string,
) map[string]interface{} {
	grants := make([]interface{}, len(permissions))
	for i, p := range permissions {
		grants[i] = p
	}
	return map[string]interface{}{
		"__type": "PermissionDescriptorGroup",
		"entries": []interface{}{
			map[string]interface{}{
				"__type": "PermissionDescriptorCondition",
				"condition": map[string]interface{}{
					"__type": "PermissionExpressionEqual",
					"left":   map[string]interface{}{"__type": expressionType},
					"right": map[string]interface{}{
						"__type":   literalType,
						"identity": identity,
					},
				},
				"subNode": map[string]interface{}{
					"__type":      "PermissionDescriptorAllow",
					"permissions": grants,
				},
			},
		},
	}
}

// scopedPermissionsHelpDoc is the shared epilogue for the three helpers'
// descriptions, kept identical so codegen documentation stays consistent.
const scopedPermissionsHelpDoc = "The result is directly assignable to " +
	"`OrganizationRole.permissions`. To grant scopes on more than one entity " +
	"in a single role, hand-roll a `PermissionDescriptorGroup` whose `entries` " +
	"list pulls a `PermissionDescriptorCondition` from each helper output."

// ----------------------------------------------------------------------------
// Environment-scoped helper
// ----------------------------------------------------------------------------

// GetEnvironmentScopedPermissionsFunction builds an OrganizationRole.permissions
// descriptor that grants the supplied scopes only on the named environment.
type GetEnvironmentScopedPermissionsFunction struct{}

type GetEnvironmentScopedPermissionsInput struct {
	EnvironmentID string   `pulumi:"environmentId"`
	Permissions   []string `pulumi:"permissions"`
}

type GetEnvironmentScopedPermissionsOutput struct {
	Permissions map[string]interface{} `pulumi:"permissions"`
}

func (GetEnvironmentScopedPermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetEnvironmentScopedPermissionsFunction{},
		"Builds an `OrganizationRole.permissions` descriptor that grants the supplied scopes only on "+
			"the named environment. Pair with `Environment.environmentId` (or the `getEnvironment` data "+
			"source) to avoid hand-rolling the underlying `PermissionDescriptorGroup` / "+
			"`PermissionDescriptorCondition` / `PermissionLiteralExpressionEnvironment` JSON. "+
			scopedPermissionsHelpDoc,
	)
	a.SetToken("index", "getEnvironmentScopedPermissions")
}

func (i *GetEnvironmentScopedPermissionsInput) Annotate(a infer.Annotator) {
	a.Describe(
		&i.EnvironmentID,
		"The target environment's UUID. Use the `environmentId` output of an `Environment` resource "+
			"or the `getEnvironment` data source.",
	)
	a.Describe(
		&i.Permissions,
		"The set of `environment:*` scopes to grant on the target environment "+
			"(e.g. `environment:read`, `environment:open`, `environment:update`). "+
			"Discover valid scope names via the `getOrganizationRoleScopes` data source.",
	)
}

func (o *GetEnvironmentScopedPermissionsOutput) Annotate(a infer.Annotator) {
	a.Describe(
		&o.Permissions,
		"A `PermissionDescriptor` tree ready to assign to `OrganizationRole.permissions`.",
	)
}

func (GetEnvironmentScopedPermissionsFunction) Invoke(
	_ context.Context,
	req infer.FunctionRequest[GetEnvironmentScopedPermissionsInput],
) (infer.FunctionResponse[GetEnvironmentScopedPermissionsOutput], error) {
	if req.Input.EnvironmentID == "" {
		return infer.FunctionResponse[GetEnvironmentScopedPermissionsOutput]{},
			fmt.Errorf("`environmentId` must not be empty")
	}
	if len(req.Input.Permissions) == 0 {
		return infer.FunctionResponse[GetEnvironmentScopedPermissionsOutput]{},
			fmt.Errorf("`permissions` must not be empty")
	}
	return infer.FunctionResponse[GetEnvironmentScopedPermissionsOutput]{
		Output: GetEnvironmentScopedPermissionsOutput{
			Permissions: scopedPermissionsDescriptor(
				"PermissionExpressionEnvironment",
				"PermissionLiteralExpressionEnvironment",
				req.Input.EnvironmentID,
				req.Input.Permissions,
			),
		},
	}, nil
}

// ----------------------------------------------------------------------------
// Stack-scoped helper
// ----------------------------------------------------------------------------

// GetStackScopedPermissionsFunction builds an OrganizationRole.permissions
// descriptor that grants the supplied scopes only on the named stack.
type GetStackScopedPermissionsFunction struct{}

type GetStackScopedPermissionsInput struct {
	StackID     string   `pulumi:"stackId"`
	Permissions []string `pulumi:"permissions"`
}

type GetStackScopedPermissionsOutput struct {
	Permissions map[string]interface{} `pulumi:"permissions"`
}

func (GetStackScopedPermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetStackScopedPermissionsFunction{},
		"Builds an `OrganizationRole.permissions` descriptor that grants the supplied scopes only on "+
			"the named stack. The `stackId` is the stack's opaque Pulumi Cloud identifier — distinct "+
			"from the `organization/project/stack` triple — and is what `PermissionLiteralExpressionStack` "+
			"expects. "+scopedPermissionsHelpDoc,
	)
	a.SetToken("index", "getStackScopedPermissions")
}

func (i *GetStackScopedPermissionsInput) Annotate(a infer.Annotator) {
	a.Describe(
		&i.StackID,
		"The target stack's opaque Pulumi Cloud identifier (not the `organization/project/stack` triple).",
	)
	a.Describe(
		&i.Permissions,
		"The set of `stack:*` scopes to grant on the target stack "+
			"(e.g. `stack:read`, `stack:edit`, `stack:admin`). "+
			"Discover valid scope names via the `getOrganizationRoleScopes` data source.",
	)
}

func (o *GetStackScopedPermissionsOutput) Annotate(a infer.Annotator) {
	a.Describe(
		&o.Permissions,
		"A `PermissionDescriptor` tree ready to assign to `OrganizationRole.permissions`.",
	)
}

func (GetStackScopedPermissionsFunction) Invoke(
	_ context.Context,
	req infer.FunctionRequest[GetStackScopedPermissionsInput],
) (infer.FunctionResponse[GetStackScopedPermissionsOutput], error) {
	if req.Input.StackID == "" {
		return infer.FunctionResponse[GetStackScopedPermissionsOutput]{},
			fmt.Errorf("`stackId` must not be empty")
	}
	if len(req.Input.Permissions) == 0 {
		return infer.FunctionResponse[GetStackScopedPermissionsOutput]{},
			fmt.Errorf("`permissions` must not be empty")
	}
	return infer.FunctionResponse[GetStackScopedPermissionsOutput]{
		Output: GetStackScopedPermissionsOutput{
			Permissions: scopedPermissionsDescriptor(
				"PermissionExpressionStack",
				"PermissionLiteralExpressionStack",
				req.Input.StackID,
				req.Input.Permissions,
			),
		},
	}, nil
}

// ----------------------------------------------------------------------------
// Insights-account-scoped helper
// ----------------------------------------------------------------------------

// GetInsightsAccountScopedPermissionsFunction builds an
// OrganizationRole.permissions descriptor that grants the supplied scopes only
// on the named insights account.
type GetInsightsAccountScopedPermissionsFunction struct{}

type GetInsightsAccountScopedPermissionsInput struct {
	InsightsAccountID string   `pulumi:"insightsAccountId"`
	Permissions       []string `pulumi:"permissions"`
}

type GetInsightsAccountScopedPermissionsOutput struct {
	Permissions map[string]interface{} `pulumi:"permissions"`
}

func (GetInsightsAccountScopedPermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetInsightsAccountScopedPermissionsFunction{},
		"Builds an `OrganizationRole.permissions` descriptor that grants the supplied scopes only on "+
			"the named insights account. Pair with `InsightsAccount.insightsAccountId` (or the "+
			"`getInsightsAccount` data source) to avoid hand-rolling the underlying "+
			"`PermissionLiteralExpressionInsightsAccount` JSON. "+scopedPermissionsHelpDoc,
	)
	a.SetToken("index", "getInsightsAccountScopedPermissions")
}

func (i *GetInsightsAccountScopedPermissionsInput) Annotate(a infer.Annotator) {
	a.Describe(
		&i.InsightsAccountID,
		"The target insights account's identifier. Use the `insightsAccountId` output of an "+
			"`InsightsAccount` resource or the `getInsightsAccount` data source.",
	)
	a.Describe(
		&i.Permissions,
		"The set of `insights-account:*` scopes to grant on the target account. "+
			"Discover valid scope names via the `getOrganizationRoleScopes` data source.",
	)
}

func (o *GetInsightsAccountScopedPermissionsOutput) Annotate(a infer.Annotator) {
	a.Describe(
		&o.Permissions,
		"A `PermissionDescriptor` tree ready to assign to `OrganizationRole.permissions`.",
	)
}

func (GetInsightsAccountScopedPermissionsFunction) Invoke(
	_ context.Context,
	req infer.FunctionRequest[GetInsightsAccountScopedPermissionsInput],
) (infer.FunctionResponse[GetInsightsAccountScopedPermissionsOutput], error) {
	if req.Input.InsightsAccountID == "" {
		return infer.FunctionResponse[GetInsightsAccountScopedPermissionsOutput]{},
			fmt.Errorf("`insightsAccountId` must not be empty")
	}
	if len(req.Input.Permissions) == 0 {
		return infer.FunctionResponse[GetInsightsAccountScopedPermissionsOutput]{},
			fmt.Errorf("`permissions` must not be empty")
	}
	return infer.FunctionResponse[GetInsightsAccountScopedPermissionsOutput]{
		Output: GetInsightsAccountScopedPermissionsOutput{
			Permissions: scopedPermissionsDescriptor(
				"PermissionExpressionInsightsAccount",
				"PermissionLiteralExpressionInsightsAccount",
				req.Input.InsightsAccountID,
				req.Input.Permissions,
			),
		},
	}, nil
}
