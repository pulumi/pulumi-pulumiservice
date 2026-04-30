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

// scopedAllowWire builds an `OrganizationRole.permissions` descriptor that
// grants `permissions` only when the request's entity expression equals the
// supplied identity. Returns the wire-format Condition shape verbatim — the
// SDK boundary uses the same PascalCase `kind` values as Pulumi Cloud's REST
// API. No provider-side translation is needed; OrganizationRole's
// Create/Update wraps top-level Conditions in a single-entry Group only as a
// Cloud UI workaround, and Read collapses the wrap so refresh stays
// idempotent against the helper output.
//
// The (expressionKind, literalKind) pair is the per-entity-type
// PermissionExpression / PermissionLiteralExpression vocabulary from the
// Pulumi Cloud RBAC API:
//
//	environment      → PermissionExpressionEnvironment      / PermissionLiteralExpressionEnvironment
//	stack            → PermissionExpressionStack            / PermissionLiteralExpressionStack
//	insightsAccount  → PermissionExpressionInsightsAccount  / PermissionLiteralExpressionInsightsAccount
//
// Note: there is intentionally no "team" scoping helper. Roles are
// *associated with* teams via the TeamRoleAssignment resource, not gated
// on them via a permission descriptor; the wire grammar exposes
// PermissionExpressionTeam for advanced cases (e.g. roles imported from
// the Pulumi Cloud UI that mix team identity into a complex Compose),
// but the SDK does not advertise that as a recommended pattern.
func scopedAllowWire(expressionKind, literalKind, identity string, permissions []string) map[string]interface{} {
	grants := make([]interface{}, len(permissions))
	for i, p := range permissions {
		grants[i] = p
	}
	return map[string]interface{}{
		"kind": "PermissionDescriptorCondition",
		"condition": map[string]interface{}{
			"kind":  "PermissionExpressionEqual",
			"left":  map[string]interface{}{"kind": expressionKind},
			"right": map[string]interface{}{"kind": literalKind, "identity": identity},
		},
		"subNode": map[string]interface{}{
			"kind":        "PermissionDescriptorAllow",
			"permissions": grants,
		},
	}
}

// scopedPermissionsHelpDoc is the shared epilogue for the helpers'
// descriptions, kept identical so codegen documentation stays consistent.
const scopedPermissionsHelpDoc = "The result is directly assignable to " +
	"`OrganizationRole.permissions`. To grant scopes on more than one entity " +
	"in a single role, hand-roll a `PermissionDescriptorGroup` whose `entries` " +
	"list pulls the output of each helper."

// ----------------------------------------------------------------------------
// Environment-scoped helper
// ----------------------------------------------------------------------------

type BuildEnvironmentScopedPermissionsFunction struct{}

type BuildEnvironmentScopedPermissionsInput struct {
	EnvironmentID string   `pulumi:"environmentId"`
	Permissions   []string `pulumi:"permissions"`
}

type BuildEnvironmentScopedPermissionsOutput struct {
	Permissions map[string]interface{} `pulumi:"permissions"`
}

func (BuildEnvironmentScopedPermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&BuildEnvironmentScopedPermissionsFunction{},
		"Builds an `OrganizationRole.permissions` descriptor that grants the supplied scopes only on "+
			"the named environment. Pair with `Environment.environmentId` (or the `getEnvironment` data "+
			"source) to avoid hand-rolling the `PermissionDescriptorCondition` tree yourself. "+
			scopedPermissionsHelpDoc,
	)
	a.SetToken("index", "buildEnvironmentScopedPermissions")
}

func (i *BuildEnvironmentScopedPermissionsInput) Annotate(a infer.Annotator) {
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

func (o *BuildEnvironmentScopedPermissionsOutput) Annotate(a infer.Annotator) {
	a.Describe(
		&o.Permissions,
		"A `PermissionDescriptorCondition` tree gating a `PermissionDescriptorAllow` "+
			"on the named environment, ready to assign to `OrganizationRole.permissions`.",
	)
}

func (BuildEnvironmentScopedPermissionsFunction) Invoke(
	_ context.Context,
	req infer.FunctionRequest[BuildEnvironmentScopedPermissionsInput],
) (infer.FunctionResponse[BuildEnvironmentScopedPermissionsOutput], error) {
	if req.Input.EnvironmentID == "" {
		return infer.FunctionResponse[BuildEnvironmentScopedPermissionsOutput]{},
			fmt.Errorf("`environmentId` must not be empty")
	}
	if len(req.Input.Permissions) == 0 {
		return infer.FunctionResponse[BuildEnvironmentScopedPermissionsOutput]{},
			fmt.Errorf("`permissions` must not be empty")
	}
	return infer.FunctionResponse[BuildEnvironmentScopedPermissionsOutput]{
		Output: BuildEnvironmentScopedPermissionsOutput{
			Permissions: scopedAllowWire(
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

type BuildStackScopedPermissionsFunction struct{}

type BuildStackScopedPermissionsInput struct {
	StackID     string   `pulumi:"stackId"`
	Permissions []string `pulumi:"permissions"`
}

type BuildStackScopedPermissionsOutput struct {
	Permissions map[string]interface{} `pulumi:"permissions"`
}

func (BuildStackScopedPermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&BuildStackScopedPermissionsFunction{},
		"Builds an `OrganizationRole.permissions` descriptor that grants the supplied scopes only on "+
			"the named stack. The `stackId` is the stack's opaque Pulumi Cloud identifier — distinct "+
			"from the `organization/project/stack` triple. "+scopedPermissionsHelpDoc,
	)
	a.SetToken("index", "buildStackScopedPermissions")
}

func (i *BuildStackScopedPermissionsInput) Annotate(a infer.Annotator) {
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

func (o *BuildStackScopedPermissionsOutput) Annotate(a infer.Annotator) {
	a.Describe(
		&o.Permissions,
		"A `PermissionDescriptorCondition` tree gating a `PermissionDescriptorAllow` "+
			"on the named stack, ready to assign to `OrganizationRole.permissions`.",
	)
}

func (BuildStackScopedPermissionsFunction) Invoke(
	_ context.Context,
	req infer.FunctionRequest[BuildStackScopedPermissionsInput],
) (infer.FunctionResponse[BuildStackScopedPermissionsOutput], error) {
	if req.Input.StackID == "" {
		return infer.FunctionResponse[BuildStackScopedPermissionsOutput]{},
			fmt.Errorf("`stackId` must not be empty")
	}
	if len(req.Input.Permissions) == 0 {
		return infer.FunctionResponse[BuildStackScopedPermissionsOutput]{},
			fmt.Errorf("`permissions` must not be empty")
	}
	return infer.FunctionResponse[BuildStackScopedPermissionsOutput]{
		Output: BuildStackScopedPermissionsOutput{
			Permissions: scopedAllowWire(
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

type BuildInsightsAccountScopedPermissionsFunction struct{}

type BuildInsightsAccountScopedPermissionsInput struct {
	InsightsAccountID string   `pulumi:"insightsAccountId"`
	Permissions       []string `pulumi:"permissions"`
}

type BuildInsightsAccountScopedPermissionsOutput struct {
	Permissions map[string]interface{} `pulumi:"permissions"`
}

func (BuildInsightsAccountScopedPermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&BuildInsightsAccountScopedPermissionsFunction{},
		"Builds an `OrganizationRole.permissions` descriptor that grants the supplied scopes only on "+
			"the named insights account. Pair with `InsightsAccount.insightsAccountId` (or the "+
			"`getInsightsAccount` data source). "+scopedPermissionsHelpDoc,
	)
	a.SetToken("index", "buildInsightsAccountScopedPermissions")
}

func (i *BuildInsightsAccountScopedPermissionsInput) Annotate(a infer.Annotator) {
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

func (o *BuildInsightsAccountScopedPermissionsOutput) Annotate(a infer.Annotator) {
	a.Describe(
		&o.Permissions,
		"A `PermissionDescriptorCondition` tree gating a `PermissionDescriptorAllow` "+
			"on the named insights account, ready to assign to `OrganizationRole.permissions`.",
	)
}

func (BuildInsightsAccountScopedPermissionsFunction) Invoke(
	_ context.Context,
	req infer.FunctionRequest[BuildInsightsAccountScopedPermissionsInput],
) (infer.FunctionResponse[BuildInsightsAccountScopedPermissionsOutput], error) {
	if req.Input.InsightsAccountID == "" {
		return infer.FunctionResponse[BuildInsightsAccountScopedPermissionsOutput]{},
			fmt.Errorf("`insightsAccountId` must not be empty")
	}
	if len(req.Input.Permissions) == 0 {
		return infer.FunctionResponse[BuildInsightsAccountScopedPermissionsOutput]{},
			fmt.Errorf("`permissions` must not be empty")
	}
	return infer.FunctionResponse[BuildInsightsAccountScopedPermissionsOutput]{
		Output: BuildInsightsAccountScopedPermissionsOutput{
			Permissions: scopedAllowWire(
				"PermissionExpressionInsightsAccount",
				"PermissionLiteralExpressionInsightsAccount",
				req.Input.InsightsAccountID,
				req.Input.Permissions,
			),
		},
	}, nil
}
