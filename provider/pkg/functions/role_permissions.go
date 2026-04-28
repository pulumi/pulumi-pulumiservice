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

// scopedAllow builds a kind-shaped Allow descriptor scoped to the supplied
// entity. The result is directly assignable to OrganizationRole.permissions.
// The provider's translator (in resources/permission_descriptor.go) expands
// the on: modifier into the wire-format Condition wrapper at Create time.
func scopedAllow(entityType, identity string, permissions []string) map[string]interface{} {
	grants := make([]interface{}, len(permissions))
	for i, p := range permissions {
		grants[i] = p
	}
	return map[string]interface{}{
		"kind":        "allow",
		"on":          map[string]interface{}{entityType: identity},
		"permissions": grants,
	}
}

// scopedPermissionsHelpDoc is the shared epilogue for the three helpers'
// descriptions, kept identical so codegen documentation stays consistent.
const scopedPermissionsHelpDoc = "The result is directly assignable to " +
	"`OrganizationRole.permissions`. To grant scopes on more than one entity " +
	"in a single role, hand-roll a `group` whose `entries` list pulls the " +
	"output of each helper."

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
			"source) to avoid hand-rolling the `on:` modifier yourself. "+scopedPermissionsHelpDoc,
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
		"A `kind: allow` descriptor with an `on: { environment: <uuid> }` modifier, "+
			"ready to assign to `OrganizationRole.permissions`.",
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
			Permissions: scopedAllow("environment", req.Input.EnvironmentID, req.Input.Permissions),
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
		"A `kind: allow` descriptor with an `on: { stack: <id> }` modifier, "+
			"ready to assign to `OrganizationRole.permissions`.",
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
			Permissions: scopedAllow("stack", req.Input.StackID, req.Input.Permissions),
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
		"A `kind: allow` descriptor with an `on: { insightsAccount: <id> }` modifier, "+
			"ready to assign to `OrganizationRole.permissions`.",
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
			Permissions: scopedAllow("insightsAccount", req.Input.InsightsAccountID, req.Input.Permissions),
		},
	}, nil
}
