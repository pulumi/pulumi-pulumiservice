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

package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

type OrganizationRole struct{}

var (
	_ infer.CustomCreate[OrganizationRoleInput, OrganizationRoleState] = &OrganizationRole{}
	_ infer.CustomCheck[OrganizationRoleInput]                         = &OrganizationRole{}
	_ infer.CustomDelete[OrganizationRoleState]                        = &OrganizationRole{}
	_ infer.CustomRead[OrganizationRoleInput, OrganizationRoleState]   = &OrganizationRole{}
	_ infer.CustomUpdate[OrganizationRoleInput, OrganizationRoleState] = &OrganizationRole{}
)

func (*OrganizationRole) Annotate(a infer.Annotator) {
	a.Describe(
		&OrganizationRole{},
		"A custom (fine-grained) role defined on a Pulumi Cloud organization. Custom roles allow precise "+
			"permission control beyond the built-in `admin` / `member` / `billing-manager` roles. Assign them "+
			"to members via the `OrganizationMember.roleId` field or to teams via `TeamRoleAssignment`.\n\n"+
			"Requires the Custom Roles feature to be enabled on the organization. See the "+
			"[Pulumi Cloud RBAC docs](https://www.pulumi.com/docs/pulumi-cloud/access-management/rbac/) for "+
			"the shape of the `permissions` descriptor.",
	)
}

type OrganizationRoleCore struct {
	OrganizationName string                 `pulumi:"organizationName" provider:"replaceOnChanges"`
	Name             string                 `pulumi:"name"`
	Description      *string                `pulumi:"description,optional"`
	ResourceType     *string                `pulumi:"resourceType,optional"`
	Permissions      map[string]interface{} `pulumi:"permissions"`
}

func (c *OrganizationRoleCore) Annotate(a infer.Annotator) {
	a.Describe(&c.OrganizationName, "The Pulumi Cloud organization name.")
	a.Describe(&c.Name, "The role's display name. Must be unique within the organization.")
	a.Describe(&c.Description, "Human-readable description of what the role grants.")
	a.Describe(
		&c.ResourceType,
		"The resource type the role's permissions apply to. Defaults to `global` (the org-wide role "+
			"that can be assigned to members and teams). Other valid values: `stack`, `environment`, "+
			"`insights-account`.",
	)
	a.Describe(
		&c.Permissions,
		"The role's permission descriptor tree — passed to the service verbatim. This is the `details` "+
			"field of a Pulumi Cloud PermissionDescriptor: an object with a `__type` discriminator (e.g. "+
			"`PermissionDescriptorAllow`, `PermissionDescriptorCompose`) describing which scopes are granted. "+
			"For per-entity scoping, prefer the `buildEnvironmentScopedPermissions`, `buildStackScopedPermissions`, "+
			"and `buildInsightsAccountScopedPermissions` helpers, which build the underlying "+
			"`PermissionDescriptorGroup` / `PermissionDescriptorCondition` tree for you.",
	)
}

type OrganizationRoleInput struct {
	OrganizationRoleCore
}

type OrganizationRoleState struct {
	OrganizationRoleCore
	RoleId  string `pulumi:"roleId"`
	Version int    `pulumi:"version"`
}

func (s *OrganizationRoleState) Annotate(a infer.Annotator) {
	a.Describe(&s.RoleId, "The unique identifier of the custom role.")
	a.Describe(&s.Version, "The service-maintained version number that increments on every update.")
}

func (*OrganizationRole) Check(
	ctx context.Context,
	req infer.CheckRequest,
) (infer.CheckResponse[OrganizationRoleInput], error) {
	in, failures, err := infer.DefaultCheck[OrganizationRoleInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[OrganizationRoleInput]{}, err
	}
	// Skip emptiness checks when the raw input arrived as unknown/computed
	// (e.g. `permissions: someResource.environmentId.apply(...)` or the
	// `_output` variant of `buildEnvironmentScopedPermissions`). At preview
	// the typed Go field decodes to its zero value, which would otherwise
	// trip the empty checks. Pulumi guarantees these inputs are concrete
	// by the time Create/Update runs, so the same checks belong there.
	if !isUnknownInput(req.NewInputs, "name") && in.Name == "" {
		failures = append(failures, p.CheckFailure{Property: "name", Reason: "name must not be empty"})
	}
	if !isUnknownInput(req.NewInputs, "permissions") && len(in.Permissions) == 0 {
		failures = append(failures, p.CheckFailure{
			Property: "permissions",
			Reason:   "permissions must not be empty — supply a PermissionDescriptor tree",
		})
	}
	return infer.CheckResponse[OrganizationRoleInput]{Inputs: in, Failures: failures}, nil
}

// isUnknownInput reports whether the value at key is present in the input map
// but not yet known. Pulumi's newer property.Value normalises Output-typed
// inputs that are still pending into the Computed form, so a single
// IsComputed check covers both `someResource.x.apply(...)` and the `_output`
// variant of a data-source helper.
func isUnknownInput(m property.Map, key string) bool {
	v, ok := m.GetOk(key)
	return ok && v.IsComputed()
}

func (*OrganizationRole) Create(
	ctx context.Context,
	req infer.CreateRequest[OrganizationRoleInput],
) (infer.CreateResponse[OrganizationRoleState], error) {
	if req.DryRun {
		return infer.CreateResponse[OrganizationRoleState]{
			ID:     fmt.Sprintf("%s/%s", req.Inputs.OrganizationName, req.Inputs.Name),
			Output: OrganizationRoleState{OrganizationRoleCore: req.Inputs.OrganizationRoleCore},
		}, nil
	}

	details, err := json.Marshal(req.Inputs.Permissions)
	if err != nil {
		return infer.CreateResponse[OrganizationRoleState]{}, fmt.Errorf(
			"failed to marshal permissions: %w",
			err,
		)
	}

	client := config.GetClient(ctx)
	role, err := client.CreateRole(ctx, req.Inputs.OrganizationName, pulumiapi.NewCreateRoleRequest(
		req.Inputs.Name,
		util.OrZero(req.Inputs.Description),
		util.OrZero(req.Inputs.ResourceType),
		details,
	))
	if err != nil {
		return infer.CreateResponse[OrganizationRoleState]{}, fmt.Errorf(
			"failed to create role %q: %w",
			req.Inputs.Name,
			err,
		)
	}

	return infer.CreateResponse[OrganizationRoleState]{
		ID:     fmt.Sprintf("%s/%s", req.Inputs.OrganizationName, role.ID),
		Output: orgRoleStateFromAPI(req.Inputs.OrganizationName, req.Inputs.OrganizationRoleCore, role),
	}, nil
}

func (*OrganizationRole) Update(
	ctx context.Context,
	req infer.UpdateRequest[OrganizationRoleInput, OrganizationRoleState],
) (infer.UpdateResponse[OrganizationRoleState], error) {
	core := req.Inputs.OrganizationRoleCore

	if req.DryRun {
		return infer.UpdateResponse[OrganizationRoleState]{
			Output: OrganizationRoleState{
				OrganizationRoleCore: core,
				RoleId:               req.State.RoleId,
				Version:              req.State.Version,
			},
		}, nil
	}

	details, err := json.Marshal(core.Permissions)
	if err != nil {
		return infer.UpdateResponse[OrganizationRoleState]{}, fmt.Errorf(
			"failed to marshal permissions: %w",
			err,
		)
	}

	client := config.GetClient(ctx)
	name := core.Name
	role, err := client.UpdateRole(
		ctx,
		req.State.OrganizationName,
		req.State.RoleId,
		&name,
		core.Description,
		details,
	)
	if err != nil {
		return infer.UpdateResponse[OrganizationRoleState]{}, fmt.Errorf(
			"failed to update role %q: %w",
			req.State.RoleId,
			err,
		)
	}
	return infer.UpdateResponse[OrganizationRoleState]{
		Output: orgRoleStateFromAPI(req.State.OrganizationName, core, role),
	}, nil
}

func (*OrganizationRole) Delete(
	ctx context.Context,
	req infer.DeleteRequest[OrganizationRoleState],
) (infer.DeleteResponse, error) {
	client := config.GetClient(ctx)
	// Force=true: Pulumi destroy should succeed even if the role is still
	// referenced; the alternative is telling users to manually unassign.
	return infer.DeleteResponse{}, client.DeleteRole(
		ctx,
		req.State.OrganizationName,
		req.State.RoleId,
		true,
	)
}

func (*OrganizationRole) Read(
	ctx context.Context,
	req infer.ReadRequest[OrganizationRoleInput, OrganizationRoleState],
) (infer.ReadResponse[OrganizationRoleInput, OrganizationRoleState], error) {
	orgName, roleID, err := splitOrgRoleID(req.ID)
	if err != nil {
		return infer.ReadResponse[OrganizationRoleInput, OrganizationRoleState]{}, err
	}

	client := config.GetClient(ctx)
	role, err := client.GetRole(ctx, orgName, roleID)
	if err != nil {
		return infer.ReadResponse[OrganizationRoleInput, OrganizationRoleState]{}, fmt.Errorf(
			"failed to read role %q: %w",
			req.ID,
			err,
		)
	}
	if role == nil {
		return infer.ReadResponse[OrganizationRoleInput, OrganizationRoleState]{}, nil
	}

	core, err := orgRoleCoreFromAPI(orgName, req.Inputs.OrganizationRoleCore, role)
	if err != nil {
		return infer.ReadResponse[OrganizationRoleInput, OrganizationRoleState]{}, err
	}
	state := orgRoleStateFromAPI(orgName, core, role)
	return infer.ReadResponse[OrganizationRoleInput, OrganizationRoleState]{
		ID:     req.ID,
		Inputs: OrganizationRoleInput{OrganizationRoleCore: core},
		State:  state,
	}, nil
}

func orgRoleCoreFromAPI(
	orgName string,
	prior OrganizationRoleCore,
	role *pulumiapi.RoleDescriptor,
) (OrganizationRoleCore, error) {
	core := OrganizationRoleCore{
		OrganizationName: orgName,
		Name:             role.Name,
		Description:      util.OrNil(role.Description),
		// Preserve user intent: if the user left resourceType unset, don't
		// let the service-computed default leak into state and cause refresh
		// drift.
		ResourceType: prior.ResourceType,
		Permissions:  prior.Permissions,
	}
	if core.ResourceType == nil && role.ResourceType != "" && role.ResourceType != "global" {
		core.ResourceType = util.OrNil(role.ResourceType)
	}
	if len(role.Details) > 0 {
		parsed := map[string]interface{}{}
		if err := json.Unmarshal(role.Details, &parsed); err != nil {
			return OrganizationRoleCore{}, fmt.Errorf("parsing role details for %q: %w", role.ID, err)
		}
		core.Permissions = parsed
	}
	return core, nil
}

func orgRoleStateFromAPI(
	orgName string,
	core OrganizationRoleCore,
	role *pulumiapi.RoleDescriptor,
) OrganizationRoleState {
	core.OrganizationName = orgName
	return OrganizationRoleState{
		OrganizationRoleCore: core,
		RoleId:               role.ID,
		Version:              role.Version,
	}
}

func splitOrgRoleID(id string) (string, string, error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("%q is invalid, must be in the format: organization/roleId", id)
	}
	return parts[0], parts[1], nil
}
