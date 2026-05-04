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

// scopedAllowDescriptor builds an `OrganizationRole.permissions` descriptor
// that grants `permissions` only when the request's entity expression
// equals the supplied identity. The result is shaped to match the
// provider's translation contract: `discriminator` at the top (the SDK
// boundary the provider renames to `__type` for the wire) and `__type`
// at every nested level (the wire format Pulumi Cloud expects, passed
// through verbatim). The provider's `permissions` schema is `map[string]Any`
// so anything below the top is opaque — this helper provides the
// nested wire format because there is no provider machinery to do it
// for us.
//
// The (expressionDiscriminator, literalDiscriminator) pair is the per-
// entity-type PermissionExpression / PermissionLiteralExpression
// vocabulary from the Pulumi Cloud RBAC API:
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
func scopedAllowDescriptor(
	expressionDiscriminator, literalDiscriminator, identity string, permissions []string,
) map[string]interface{} {
	grants := make([]interface{}, len(permissions))
	for i, p := range permissions {
		grants[i] = p
	}
	return map[string]interface{}{
		// SDK boundary at the top — provider promotes this to __type.
		"discriminator": "PermissionDescriptorCondition",
		// Below the top: wire format, opaque to the provider, forwarded
		// verbatim to Pulumi Cloud.
		"condition": map[string]interface{}{
			"__type": "PermissionExpressionEqual",
			"left":   map[string]interface{}{"__type": expressionDiscriminator},
			"right":  map[string]interface{}{"__type": literalDiscriminator, "identity": identity},
		},
		"subNode": map[string]interface{}{
			"__type":      "PermissionDescriptorAllow",
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
// Global Allow helper
// ----------------------------------------------------------------------------

type BuildAllowPermissionsFunction struct{}

type BuildAllowPermissionsInput struct {
	Permissions []string `pulumi:"permissions"`
}

type BuildAllowPermissionsOutput struct {
	Permissions map[string]interface{} `pulumi:"permissions"`
}

func (BuildAllowPermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&BuildAllowPermissionsFunction{},
		"Builds an `OrganizationRole.permissions` descriptor that grants the "+
			"supplied scopes globally — i.e. on every entity of the matching "+
			"resource type. This is the simplest descriptor: a flat "+
			"`PermissionDescriptorAllow`. Use this helper instead of hand-"+
			"authoring `{discriminator: \"PermissionDescriptorAllow\", "+
			"permissions: [...]}` so the SDK boundary's discriminator field "+
			"name stays an implementation detail. For grants scoped to a "+
			"specific entity, see `buildEnvironmentScopedPermissions`, "+
			"`buildStackScopedPermissions`, or `buildInsightsAccountScopedPermissions`. "+
			scopedPermissionsHelpDoc,
	)
	a.SetToken("index", "buildAllowPermissions")
}

func (i *BuildAllowPermissionsInput) Annotate(a infer.Annotator) {
	a.Describe(
		&i.Permissions,
		"The set of scopes to grant globally (e.g. `stack:read`, `environment:open`, "+
			"`organization:billingManager`). Discover valid scope names via the "+
			"`getOrganizationRoleScopes` data source.",
	)
}

func (o *BuildAllowPermissionsOutput) Annotate(a infer.Annotator) {
	a.Describe(
		&o.Permissions,
		"A `PermissionDescriptorAllow` granting the supplied scopes on every "+
			"entity of the matching resource type, ready to assign to "+
			"`OrganizationRole.permissions`.",
	)
}

func (BuildAllowPermissionsFunction) Invoke(
	_ context.Context,
	req infer.FunctionRequest[BuildAllowPermissionsInput],
) (infer.FunctionResponse[BuildAllowPermissionsOutput], error) {
	if len(req.Input.Permissions) == 0 {
		return infer.FunctionResponse[BuildAllowPermissionsOutput]{},
			fmt.Errorf("`permissions` must not be empty")
	}
	grants := make([]interface{}, len(req.Input.Permissions))
	for i, p := range req.Input.Permissions {
		grants[i] = p
	}
	return infer.FunctionResponse[BuildAllowPermissionsOutput]{
		Output: BuildAllowPermissionsOutput{
			Permissions: map[string]interface{}{
				// Top-level uses the SDK boundary's `discriminator` key —
				// the provider's translator promotes it to the wire's
				// `__type` on Create/Update.
				"discriminator": "PermissionDescriptorAllow",
				"permissions":   grants,
			},
		},
	}, nil
}

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
			Permissions: scopedAllowDescriptor(
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
			Permissions: scopedAllowDescriptor(
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
			Permissions: scopedAllowDescriptor(
				"PermissionExpressionInsightsAccount",
				"PermissionLiteralExpressionInsightsAccount",
				req.Input.InsightsAccountID,
				req.Input.Permissions,
			),
		},
	}, nil
}
