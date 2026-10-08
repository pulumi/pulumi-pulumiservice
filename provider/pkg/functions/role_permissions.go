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
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// scopedPermissionsHelpDoc is the shared epilogue for the helpers'
// descriptions, kept identical so codegen documentation stays consistent.
const scopedPermissionsHelpDoc = "The result is directly assignable to " +
	"`OrganizationRole.permissions` or `api.Role.details`. To combine several " +
	"grants in one descriptor, pass the output of each helper to " +
	"`buildGroupPermissions`."

// grantHelpDoc describes the mutually exclusive `permissions` / `setIds`
// inputs shared by the conditional helpers.
const grantHelpDoc = "Set exactly one of `permissions` or `setIds`. `permissions` " +
	"grants the scopes inline. `setIds` grants the referenced permission sets " +
	"(`api.Role` with `uxPurpose: set`), which is the shape Pulumi Cloud uses for " +
	"policies (`api.Role` with `uxPurpose: policy`)."

// descriptorToSDKMap marshals a typed apitype.PermissionDescriptor to the
// SDK-boundary map shape the provider expects on
// `OrganizationRole.permissions`. The typed Marshaler emits the wire format
// directly (`__type` discriminator at every level), which is exactly what
// the SDK boundary now uses — the Python SDK preserves `__`-prefixed keys
// across resource inputs as of pulumi/pulumi#22834 (3.235.0+, pinned via
// the Python SDK's runtime requirement), so no rename is needed.
func descriptorToSDKMap(descriptor apitype.PermissionDescriptor) (map[string]any, error) {
	raw, err := json.Marshal(descriptor)
	if err != nil {
		return nil, fmt.Errorf("marshalling typed descriptor: %w", err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decoding descriptor JSON: %w", err)
	}
	return out, nil
}

// rbacPermissionSlice converts a []string of scope names to the typed
// apitype.RbacPermissionSlice the apitype builders consume, validating
// each scope against the generated apitype.RbacPermission enum as it
// goes. The IsValid() method is generated from the same OpenAPI spec
// the API uses, so the catalogue stays in sync without the provider
// having to maintain its own list. An invalid scope here surfaces as
// a clear preview-time error rather than a 400 at apply.
func rbacPermissionSlice(scopes []string) (apitype.RbacPermissionSlice, error) {
	out := make(apitype.RbacPermissionSlice, len(scopes))
	for i, s := range scopes {
		p := apitype.RbacPermission(s)
		if !p.IsValid() {
			return nil, fmt.Errorf(
				"%q is not a valid permission scope; discover valid scope names "+
					"via the `getOrganizationRoleScopes` data source",
				s,
			)
		}
		out[i] = p
	}
	return out, nil
}

// grantDescriptor returns the node a conditional helper gates: an inline
// Allow of scopes, or a Compose that references permission sets.
func grantDescriptor(permissions, setIDs []string) (apitype.PermissionDescriptor, error) {
	if (len(permissions) == 0) == (len(setIDs) == 0) {
		return nil, fmt.Errorf("set exactly one of `permissions` or `setIds`")
	}
	if len(setIDs) > 0 {
		if err := requireIDs("setIds", setIDs); err != nil {
			return nil, err
		}
		return apitype.PermissionDescriptorComposeBuilder{PermissionDescriptors: setIDs}.Build(), nil
	}
	scopes, err := rbacPermissionSlice(permissions)
	if err != nil {
		return nil, err
	}
	return apitype.PermissionDescriptorAllowBuilder{Permissions: scopes}.Build(), nil
}

func requireIDs(field string, ids []string) error {
	if len(ids) == 0 {
		return fmt.Errorf("`%s` must not be empty", field)
	}
	for i, id := range ids {
		if id == "" {
			return fmt.Errorf("`%s[%d]` must not be empty", field, i)
		}
	}
	return nil
}

// conditionalDescriptor builds the SDK-shape map for a Condition that gates
// the grant described by permissions / setIDs.
//
// Note: there is intentionally no "team" scoping helper. Roles are
// *associated with* teams via the TeamRoleAssignment resource, not gated
// on them via a permission descriptor; the wire grammar exposes
// `PermissionExpressionTeam` for advanced cases (e.g. roles imported from
// the Pulumi Cloud UI that mix team identity into a complex Compose),
// but the SDK does not advertise that as a recommended pattern.
func conditionalDescriptor(
	condition apitype.PermissionBooleanExpression,
	permissions []string,
	setIDs []string,
) (map[string]any, error) {
	grant, err := grantDescriptor(permissions, setIDs)
	if err != nil {
		return nil, err
	}
	descriptor := apitype.PermissionDescriptorConditionBuilder{
		Condition: condition,
		SubNode:   grant,
	}.Build()
	return descriptorToSDKMap(descriptor)
}

// entityEqual is the `entity == literal` condition the entity-scoped helpers gate on.
func entityEqual(expression, literal apitype.PermissionExpression) apitype.PermissionBooleanExpression {
	return apitype.PermissionExpressionEqualBuilder{Left: expression, Right: literal}.Build()
}

// ----------------------------------------------------------------------------
// Global Allow helper
// ----------------------------------------------------------------------------

type BuildAllowPermissionsFunction struct{}

type BuildAllowPermissionsInput struct {
	Permissions []string `pulumi:"permissions"`
}

type BuildAllowPermissionsOutput struct {
	Permissions map[string]any `pulumi:"permissions"`
}

func (BuildAllowPermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&BuildAllowPermissionsFunction{},
		"Builds an `OrganizationRole.permissions` descriptor that grants the "+
			"supplied scopes globally — i.e. on every entity of the matching "+
			"resource type. This is the simplest descriptor: a flat "+
			"`PermissionDescriptorAllow`. Use this helper instead of hand-"+
			"authoring the descriptor literal so the wire-format `__type` "+
			"discriminator stays an implementation detail. For grants scoped "+
			"to a specific entity, see `buildEnvironmentScopedPermissions`, "+
			"`buildStackScopedPermissions`, or "+
			"`buildInsightsAccountScopedPermissions`. "+
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
	scopes, err := rbacPermissionSlice(req.Input.Permissions)
	if err != nil {
		return infer.FunctionResponse[BuildAllowPermissionsOutput]{}, err
	}
	descriptor := apitype.PermissionDescriptorAllowBuilder{
		Permissions: scopes,
	}.Build()
	out, err := descriptorToSDKMap(descriptor)
	if err != nil {
		return infer.FunctionResponse[BuildAllowPermissionsOutput]{}, err
	}
	return infer.FunctionResponse[BuildAllowPermissionsOutput]{
		Output: BuildAllowPermissionsOutput{Permissions: out},
	}, nil
}

// ----------------------------------------------------------------------------
// Environment-scoped helper
// ----------------------------------------------------------------------------

type BuildEnvironmentScopedPermissionsFunction struct{}

type BuildEnvironmentScopedPermissionsInput struct {
	EnvironmentID string   `pulumi:"environmentId"`
	Permissions   []string `pulumi:"permissions,optional"`
	SetIDs        []string `pulumi:"setIds,optional"`
}

type BuildEnvironmentScopedPermissionsOutput struct {
	Permissions map[string]any `pulumi:"permissions"`
}

func (BuildEnvironmentScopedPermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&BuildEnvironmentScopedPermissionsFunction{},
		"Builds a permission descriptor that grants the supplied scopes or permission sets only on "+
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
	a.Describe(
		&i.SetIDs,
		"The IDs of the permission sets to grant on the target environment. "+grantHelpDoc,
	)
}

func (o *BuildEnvironmentScopedPermissionsOutput) Annotate(a infer.Annotator) {
	a.Describe(
		&o.Permissions,
		"A `PermissionDescriptorCondition` tree gating the grant "+
			"on the named environment.",
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
	out, err := conditionalDescriptor(
		entityEqual(
			apitype.PermissionExpressionEnvironmentBuilder{}.Build(),
			apitype.PermissionLiteralExpressionEnvironmentBuilder{Identity: req.Input.EnvironmentID}.Build(),
		),
		req.Input.Permissions,
		req.Input.SetIDs,
	)
	if err != nil {
		return infer.FunctionResponse[BuildEnvironmentScopedPermissionsOutput]{}, err
	}
	return infer.FunctionResponse[BuildEnvironmentScopedPermissionsOutput]{
		Output: BuildEnvironmentScopedPermissionsOutput{Permissions: out},
	}, nil
}

// ----------------------------------------------------------------------------
// Stack-scoped helper
// ----------------------------------------------------------------------------

type BuildStackScopedPermissionsFunction struct{}

type BuildStackScopedPermissionsInput struct {
	StackID     string   `pulumi:"stackId"`
	Permissions []string `pulumi:"permissions,optional"`
	SetIDs      []string `pulumi:"setIds,optional"`
}

type BuildStackScopedPermissionsOutput struct {
	Permissions map[string]any `pulumi:"permissions"`
}

func (BuildStackScopedPermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&BuildStackScopedPermissionsFunction{},
		"Builds a permission descriptor that grants the supplied scopes or permission sets only on "+
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
	a.Describe(
		&i.SetIDs,
		"The IDs of the permission sets to grant on the target stack. "+grantHelpDoc,
	)
}

func (o *BuildStackScopedPermissionsOutput) Annotate(a infer.Annotator) {
	a.Describe(
		&o.Permissions,
		"A `PermissionDescriptorCondition` tree gating the grant "+
			"on the named stack.",
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
	out, err := conditionalDescriptor(
		entityEqual(
			apitype.PermissionExpressionStackBuilder{}.Build(),
			apitype.PermissionLiteralExpressionStackBuilder{Identity: req.Input.StackID}.Build(),
		),
		req.Input.Permissions,
		req.Input.SetIDs,
	)
	if err != nil {
		return infer.FunctionResponse[BuildStackScopedPermissionsOutput]{}, err
	}
	return infer.FunctionResponse[BuildStackScopedPermissionsOutput]{
		Output: BuildStackScopedPermissionsOutput{Permissions: out},
	}, nil
}

// ----------------------------------------------------------------------------
// Insights-account-scoped helper
// ----------------------------------------------------------------------------

type BuildInsightsAccountScopedPermissionsFunction struct{}

type BuildInsightsAccountScopedPermissionsInput struct {
	InsightsAccountID string   `pulumi:"insightsAccountId"`
	Permissions       []string `pulumi:"permissions,optional"`
	SetIDs            []string `pulumi:"setIds,optional"`
}

type BuildInsightsAccountScopedPermissionsOutput struct {
	Permissions map[string]any `pulumi:"permissions"`
}

func (BuildInsightsAccountScopedPermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&BuildInsightsAccountScopedPermissionsFunction{},
		"Builds a permission descriptor that grants the supplied scopes or permission sets only on "+
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
	a.Describe(
		&i.SetIDs,
		"The IDs of the permission sets to grant on the target insights account. "+grantHelpDoc,
	)
}

func (o *BuildInsightsAccountScopedPermissionsOutput) Annotate(a infer.Annotator) {
	a.Describe(
		&o.Permissions,
		"A `PermissionDescriptorCondition` tree gating the grant "+
			"on the named insights account.",
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
	out, err := conditionalDescriptor(
		entityEqual(
			apitype.PermissionExpressionInsightsAccountBuilder{}.Build(),
			apitype.PermissionLiteralExpressionInsightsAccountBuilder{Identity: req.Input.InsightsAccountID}.Build(),
		),
		req.Input.Permissions,
		req.Input.SetIDs,
	)
	if err != nil {
		return infer.FunctionResponse[BuildInsightsAccountScopedPermissionsOutput]{}, err
	}
	return infer.FunctionResponse[BuildInsightsAccountScopedPermissionsOutput]{
		Output: BuildInsightsAccountScopedPermissionsOutput{Permissions: out},
	}, nil
}

// ----------------------------------------------------------------------------
// Tag-conditional helper
// ----------------------------------------------------------------------------

type BuildTagConditionalPermissionsFunction struct{}

type BuildTagConditionalPermissionsInput struct {
	EntityType  RbacEntityType `pulumi:"entityType"`
	TagKey      string         `pulumi:"tagKey"`
	TagValue    string         `pulumi:"tagValue,optional"`
	Permissions []RbacScope    `pulumi:"permissions,optional"`
	SetIDs      []string       `pulumi:"setIds,optional"`
}

type BuildTagConditionalPermissionsOutput struct {
	Permissions map[string]any `pulumi:"permissions"`
}

func (BuildTagConditionalPermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&BuildTagConditionalPermissionsFunction{},
		"Builds a permission descriptor that grants the supplied scopes or permission sets only on "+
			"entities that carry a tag. With `tagValue`, the tag must have that exact value; without it, "+
			"the tag key must exist. Tags are evaluated on existing entities, so grant create rights "+
			"unconditionally. "+scopedPermissionsHelpDoc,
	)
	a.SetToken("index", "buildTagConditionalPermissions")
}

func (i *BuildTagConditionalPermissionsInput) Annotate(a infer.Annotator) {
	a.Describe(
		&i.EntityType,
		"The kind of entity whose tags are evaluated: `stack`, `environment`, or `insights-account`.",
	)
	a.Describe(&i.TagKey, "The tag key to match.")
	a.Describe(
		&i.TagValue,
		"The tag value to match. Omit it to match every entity that has `tagKey`, whatever its value.",
	)
	a.Describe(
		&i.Permissions,
		"The scopes to grant on matching entities. Discover valid scope names via the "+
			"`getOrganizationRoleScopes` data source.",
	)
	a.Describe(
		&i.SetIDs,
		"The IDs of the permission sets to grant on matching entities. "+grantHelpDoc,
	)
}

func (o *BuildTagConditionalPermissionsOutput) Annotate(a infer.Annotator) {
	a.Describe(
		&o.Permissions,
		"A `PermissionDescriptorCondition` tree gating the grant on the tag.",
	)
}

func tagContext(entityType RbacEntityType) (apitype.PermissionContextExpression, error) {
	switch entityType {
	case RbacEntityTypeStack:
		return apitype.PermissionExpressionStackBuilder{}.Build(), nil
	case RbacEntityTypeEnvironment:
		return apitype.PermissionExpressionEnvironmentBuilder{}.Build(), nil
	case RbacEntityTypeInsightsAccount:
		return apitype.PermissionExpressionInsightsAccountBuilder{}.Build(), nil
	default:
		return nil, fmt.Errorf(
			"`entityType` must be one of `stack`, `environment`, or `insights-account`; got %q", entityType)
	}
}

func (BuildTagConditionalPermissionsFunction) Invoke(
	_ context.Context,
	req infer.FunctionRequest[BuildTagConditionalPermissionsInput],
) (infer.FunctionResponse[BuildTagConditionalPermissionsOutput], error) {
	in := req.Input
	entity, err := tagContext(in.EntityType)
	if err != nil {
		return infer.FunctionResponse[BuildTagConditionalPermissionsOutput]{}, err
	}
	if in.TagKey == "" {
		return infer.FunctionResponse[BuildTagConditionalPermissionsOutput]{},
			fmt.Errorf("`tagKey` must not be empty")
	}
	var condition apitype.PermissionBooleanExpression
	if in.TagValue == "" {
		condition = apitype.PermissionExpressionHasTagBuilder{Context: entity, Key: in.TagKey}.Build()
	} else {
		condition = apitype.PermissionExpressionEqualBuilder{
			Left:  apitype.PermissionExpressionTagBuilder{Context: entity, Key: in.TagKey}.Build(),
			Right: apitype.PermissionLiteralExpressionStringBuilder{Value: in.TagValue}.Build(),
		}.Build()
	}
	out, err := conditionalDescriptor(condition, scopeStrings(in.Permissions), in.SetIDs)
	if err != nil {
		return infer.FunctionResponse[BuildTagConditionalPermissionsOutput]{}, err
	}
	return infer.FunctionResponse[BuildTagConditionalPermissionsOutput]{
		Output: BuildTagConditionalPermissionsOutput{Permissions: out},
	}, nil
}

// ----------------------------------------------------------------------------
// Compose helper
// ----------------------------------------------------------------------------

type BuildComposePermissionsFunction struct{}

type BuildComposePermissionsInput struct {
	PermissionDescriptorIDs []string `pulumi:"permissionDescriptorIds"`
}

type BuildComposePermissionsOutput struct {
	Permissions map[string]any `pulumi:"permissions"`
}

func (BuildComposePermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&BuildComposePermissionsFunction{},
		"Builds a `PermissionDescriptorCompose` that grants the union of the referenced permission "+
			"descriptors. Pulumi Cloud models a role in three layers: a role composes policies, a "+
			"policy composes permission sets, and a set grants scopes. Use this helper for an "+
			"`api.Role` with `uxPurpose: role` (pass policy IDs) and for the unconditional entry of a "+
			"policy (pass set IDs).",
	)
	a.SetToken("index", "buildComposePermissions")
}

func (i *BuildComposePermissionsInput) Annotate(a infer.Annotator) {
	a.Describe(
		&i.PermissionDescriptorIDs,
		"The IDs of the descriptors to compose. A role may reference only policies, and a policy "+
			"may reference only sets.",
	)
}

func (o *BuildComposePermissionsOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.Permissions, "A `PermissionDescriptorCompose` referencing the supplied IDs.")
}

func (BuildComposePermissionsFunction) Invoke(
	_ context.Context,
	req infer.FunctionRequest[BuildComposePermissionsInput],
) (infer.FunctionResponse[BuildComposePermissionsOutput], error) {
	ids := req.Input.PermissionDescriptorIDs
	if err := requireIDs("permissionDescriptorIds", ids); err != nil {
		return infer.FunctionResponse[BuildComposePermissionsOutput]{}, err
	}
	out, err := descriptorToSDKMap(apitype.PermissionDescriptorComposeBuilder{PermissionDescriptors: ids}.Build())
	if err != nil {
		return infer.FunctionResponse[BuildComposePermissionsOutput]{}, err
	}
	return infer.FunctionResponse[BuildComposePermissionsOutput]{
		Output: BuildComposePermissionsOutput{Permissions: out},
	}, nil
}

// ----------------------------------------------------------------------------
// Group helper
// ----------------------------------------------------------------------------

type BuildGroupPermissionsFunction struct{}

type BuildGroupPermissionsInput struct {
	Entries []map[string]any `pulumi:"entries"`
}

type BuildGroupPermissionsOutput struct {
	Permissions map[string]any `pulumi:"permissions"`
}

func (BuildGroupPermissionsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&BuildGroupPermissionsFunction{},
		"Builds a `PermissionDescriptorGroup` that grants the union of its entries. Use it for the "+
			"`details` of an `api.Role` with `uxPurpose: policy`: one `buildComposePermissions` entry "+
			"for the sets granted everywhere, plus one conditional entry (for example from "+
			"`buildTagConditionalPermissions` or `buildStackScopedPermissions` with `setIds`) per "+
			"targeted grant.",
	)
	a.SetToken("index", "buildGroupPermissions")
}

func (i *BuildGroupPermissionsInput) Annotate(a infer.Annotator) {
	a.Describe(
		&i.Entries,
		"The descriptors to group, typically the `permissions` outputs of the other `build*` helpers.",
	)
}

func (o *BuildGroupPermissionsOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.Permissions, "A `PermissionDescriptorGroup` containing the supplied entries.")
}

func (BuildGroupPermissionsFunction) Invoke(
	_ context.Context,
	req infer.FunctionRequest[BuildGroupPermissionsInput],
) (infer.FunctionResponse[BuildGroupPermissionsOutput], error) {
	entries := req.Input.Entries
	if len(entries) == 0 {
		return infer.FunctionResponse[BuildGroupPermissionsOutput]{},
			fmt.Errorf("`entries` must not be empty")
	}
	for i, entry := range entries {
		if kind, _ := entry["__type"].(string); kind == "" {
			return infer.FunctionResponse[BuildGroupPermissionsOutput]{},
				fmt.Errorf("`entries[%d]` is not a permission descriptor: it has no `__type`", i)
		}
	}
	// Entries pass through verbatim: a typed round-trip would silently drop
	// scopes the generated enum does not know.
	return infer.FunctionResponse[BuildGroupPermissionsOutput]{
		Output: BuildGroupPermissionsOutput{Permissions: map[string]any{
			"__type":  "PermissionDescriptorGroup",
			"entries": entries,
		}},
	}, nil
}
