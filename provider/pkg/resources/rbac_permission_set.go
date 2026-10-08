// Copyright 2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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
	"net/http"
	"slices"
	"strings"
	"sync"

	_ "embed" // rbacExampleDocs

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

// rbacExampleDocs is the shared Example Usage section for RbacPermissionSet
// and RbacRole. Its YAML variant is examples/yaml-rbac-roles/Pulumi.yaml.
//
//go:embed docs/rbac_example.md
var rbacExampleDocs string

const (
	gcAdditionalPermissions = "additionalPermissions"
	gcResourceType          = "resourceType"
)

type RbacPermissionSet struct{}

var (
	_ infer.CustomCreate[RbacPermissionSetInput, RbacPermissionSetState] = &RbacPermissionSet{}
	_ infer.CustomCheck[RbacPermissionSetInput]                          = &RbacPermissionSet{}
	_ infer.CustomDelete[RbacPermissionSetState]                         = &RbacPermissionSet{}
	_ infer.CustomRead[RbacPermissionSetInput, RbacPermissionSetState]   = &RbacPermissionSet{}
	_ infer.CustomUpdate[RbacPermissionSetInput, RbacPermissionSetState] = &RbacPermissionSet{}
)

func (*RbacPermissionSet) Annotate(a infer.Annotator) {
	a.Describe(
		&RbacPermissionSet{},
		"A custom permission set: a named list of RBAC scopes that apply to one kind of entity "+
			"(the organization, stacks, environments, or Insights accounts). Permission sets are the "+
			"building blocks of an `RbacRole`, which grants them organization-wide or through entity "+
			"rules. This is the resource behind **Settings > Access management > Permission sets** in "+
			"the Pulumi Cloud console.\n\n"+
			"Built-in permission sets such as \"Stack Read\" or \"Environment Admin\" already exist in "+
			"every organization; look them up with `getRbacPermissionSet` instead of recreating them.\n\n"+
			"Requires the Custom Roles feature to be enabled on the organization.\n\n"+rbacExampleDocs,
	)
	a.SetToken("index", "RbacPermissionSet")
}

type RbacPermissionSetInput struct {
	OrganizationName      string           `pulumi:"organizationName"                provider:"replaceOnChanges"`
	Name                  string           `pulumi:"name"`
	Description           *string          `pulumi:"description,optional"`
	ResourceType          RbacResourceType `pulumi:"resourceType"                    provider:"replaceOnChanges"`
	Permissions           []RbacScope      `pulumi:"permissions"`
	AdditionalPermissions []string         `pulumi:"additionalPermissions,optional"`
}

func (i *RbacPermissionSetInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The Pulumi Cloud organization name.")
	a.Describe(&i.Name, "The permission set's display name. Must be unique within the organization.")
	a.Describe(&i.Description, "Human-readable description of what the permission set grants.")
	a.Describe(
		&i.ResourceType,
		"The kind of entity the permission set applies to. `global` sets grant organization-level "+
			"access; `stack`, `environment`, and `insights-account` sets are granted on entities "+
			"through a role's entity rules. Changing this replaces the permission set.",
	)
	a.Describe(
		&i.Permissions,
		"The scopes the permission set grants. Each scope must be valid for `resourceType`.",
	)
	a.Describe(
		&i.AdditionalPermissions,
		"Scopes to grant that are not yet in this SDK's `RbacScope` enum, for example ones Pulumi "+
			"Cloud added after this provider version was released. Values are passed through as-is. "+
			"Prefer `permissions` once the scope is available there.",
	)
}

type RbacPermissionSetState struct {
	RbacPermissionSetInput
	PermissionSetId string `pulumi:"permissionSetId"`
	Version         int    `pulumi:"version"`
}

func (s *RbacPermissionSetState) Annotate(a infer.Annotator) {
	a.Describe(&s.PermissionSetId, "The permission set's unique ID. Reference it from an `RbacRole`.")
	a.Describe(&s.Version, "The service-maintained version number that increments on every update.")
}

func (*RbacPermissionSet) Check(
	ctx context.Context,
	req infer.CheckRequest,
) (infer.CheckResponse[RbacPermissionSetInput], error) {
	in, failures, err := infer.DefaultCheck[RbacPermissionSetInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[RbacPermissionSetInput]{}, err
	}
	if !isUnknownInput(req.NewInputs, gcName) && in.Name == "" {
		failures = append(failures, p.CheckFailure{Property: gcName, Reason: gcNameNotEmpty})
	}

	permsKnown := !isUnknownInput(req.NewInputs, gcPermissions)
	extraKnown := !isUnknownInput(req.NewInputs, gcAdditionalPermissions)
	if permsKnown && extraKnown && len(in.Permissions) == 0 && len(in.AdditionalPermissions) == 0 {
		failures = append(failures, p.CheckFailure{
			Property: gcPermissions,
			Reason:   "a permission set must grant at least one scope",
		})
	}

	if permsKnown && !isUnknownInput(req.NewInputs, gcResourceType) {
		allowed := rbacScopesByResourceType[in.ResourceType]
		for i, s := range in.Permissions {
			if !slices.Contains(allowed, s) {
				failures = append(failures, p.CheckFailure{
					Property: fmt.Sprintf("%s[%d]", gcPermissions, i),
					Reason: fmt.Sprintf("scope %q cannot be granted by a %q permission set; it applies to: %s",
						s, in.ResourceType, strings.Join(scopeResourceTypes(s), ", ")),
				})
			}
		}
	}

	if extraKnown {
		log := p.GetLogger(ctx)
		for _, s := range in.AdditionalPermissions {
			if isKnownScope(s) {
				log.Warningf("scope %q is available in `permissions`; move it there for validation", s)
			}
		}
	}

	return infer.CheckResponse[RbacPermissionSetInput]{Inputs: in, Failures: failures}, nil
}

func (*RbacPermissionSet) Create(
	ctx context.Context,
	req infer.CreateRequest[RbacPermissionSetInput],
) (infer.CreateResponse[RbacPermissionSetState], error) {
	in := req.Inputs
	if req.DryRun {
		return infer.CreateResponse[RbacPermissionSetState]{
			ID:     fmt.Sprintf("%s/%s", in.OrganizationName, in.Name),
			Output: RbacPermissionSetState{RbacPermissionSetInput: in},
		}, nil
	}

	set, err := config.GetClient(ctx).CreateRole(ctx, in.OrganizationName, apitype.PermissionDescriptorBase{
		Name:         in.Name,
		Description:  util.OrZero(in.Description),
		ResourceType: string(in.ResourceType),
		UxPurpose:    apitype.PermissionDescriptorUXPurposeSet,
		Details:      permissionSetDetails(in),
	})
	if err != nil {
		return infer.CreateResponse[RbacPermissionSetState]{}, fmt.Errorf(
			"failed to create permission set %q: %w", in.Name, err)
	}
	return infer.CreateResponse[RbacPermissionSetState]{
		ID:     fmt.Sprintf("%s/%s", in.OrganizationName, set.ID),
		Output: RbacPermissionSetState{RbacPermissionSetInput: in, PermissionSetId: set.ID, Version: int(set.Version)},
	}, nil
}

func (*RbacPermissionSet) Update(
	ctx context.Context,
	req infer.UpdateRequest[RbacPermissionSetInput, RbacPermissionSetState],
) (infer.UpdateResponse[RbacPermissionSetState], error) {
	in := req.Inputs
	if req.DryRun {
		return infer.UpdateResponse[RbacPermissionSetState]{
			Output: RbacPermissionSetState{
				RbacPermissionSetInput: in,
				PermissionSetId:        req.State.PermissionSetId,
				Version:                req.State.Version,
			},
		}, nil
	}

	name := in.Name
	description := util.OrZero(in.Description)
	set, err := config.GetClient(ctx).UpdateRole(ctx, req.State.OrganizationName, req.State.PermissionSetId,
		apitype.UpdateRoleRequest{Name: &name, Description: &description, Details: permissionSetDetails(in)})
	if err != nil {
		return infer.UpdateResponse[RbacPermissionSetState]{}, fmt.Errorf(
			"failed to update permission set %q: %w", req.State.PermissionSetId, err)
	}
	return infer.UpdateResponse[RbacPermissionSetState]{
		Output: RbacPermissionSetState{
			RbacPermissionSetInput: in,
			PermissionSetId:        set.ID,
			Version:                int(set.Version),
		},
	}, nil
}

func (*RbacPermissionSet) Delete(
	ctx context.Context,
	req infer.DeleteRequest[RbacPermissionSetState],
) (infer.DeleteResponse, error) {
	err := deletePermissionDescriptor(ctx, req.State.OrganizationName, req.State.PermissionSetId)
	if err != nil && pulumiapi.GetErrorStatusCode(err) == http.StatusConflict {
		return infer.DeleteResponse{}, fmt.Errorf(
			"cannot delete permission set %q: Pulumi Cloud reports it is still in use by a role. "+
				"Remove it from every role that grants it first. Underlying error: %w",
			req.State.PermissionSetId, err)
	}
	return infer.DeleteResponse{}, err
}

func (*RbacPermissionSet) Read(
	ctx context.Context,
	req infer.ReadRequest[RbacPermissionSetInput, RbacPermissionSetState],
) (infer.ReadResponse[RbacPermissionSetInput, RbacPermissionSetState], error) {
	orgName, setID, err := splitOrgRoleID(req.ID)
	if err != nil {
		return infer.ReadResponse[RbacPermissionSetInput, RbacPermissionSetState]{}, err
	}
	set, err := config.GetClient(ctx).GetRole(ctx, orgName, setID)
	if err != nil {
		return infer.ReadResponse[RbacPermissionSetInput, RbacPermissionSetState]{}, fmt.Errorf(
			"failed to read permission set %q: %w", req.ID, err)
	}
	if set == nil {
		return infer.ReadResponse[RbacPermissionSetInput, RbacPermissionSetState]{}, nil
	}

	in, err := permissionSetInputFromAPI(orgName, req.Inputs, set)
	if err != nil {
		return infer.ReadResponse[RbacPermissionSetInput, RbacPermissionSetState]{}, err
	}
	return infer.ReadResponse[RbacPermissionSetInput, RbacPermissionSetState]{
		ID:     req.ID,
		Inputs: in,
		State:  RbacPermissionSetState{RbacPermissionSetInput: in, PermissionSetId: set.ID, Version: int(set.Version)},
	}, nil
}

// permissionSetDetails merges both scope lists into the Allow descriptor the
// service stores. Scopes this SDK doesn't know still marshal as-is.
func permissionSetDetails(in RbacPermissionSetInput) apitype.PermissionDescriptorAllow {
	scopes := make([]string, 0, len(in.Permissions)+len(in.AdditionalPermissions))
	for _, s := range in.Permissions {
		scopes = append(scopes, string(s))
	}
	var perms apitype.RbacPermissionSlice
	for _, s := range dedupe(append(scopes, in.AdditionalPermissions...)) {
		perms = append(perms, apitype.RbacPermission(s))
	}
	return apitype.PermissionDescriptorAllowBuilder{Permissions: perms}.Build()
}

// sdkCanRead reports whether the Cloud SDK can report scope s back on read.
//
// The SDK's RbacPermission enum is generated with fixup validation:
// RbacPermissionSlice.UnmarshalJSON silently drops any scope that isn't in
// the pinned SDK version, while marshalling sends every value verbatim (see
// TestCreateRoleSendsScopesUnknownToTheSDK in pulumiapi/roles_test.go). A
// scope Pulumi Cloud added after that SDK version can therefore be granted
// through `additionalPermissions` but never appears in a GetRole response, so
// the provider cannot tell whether it is still granted.
func sdkCanRead(s string) bool {
	return apitype.RbacPermission(s).IsValid()
}

func permissionSetInputFromAPI(
	orgName string,
	prior RbacPermissionSetInput,
	set *apitype.PermissionDescriptorRecord,
) (RbacPermissionSetInput, error) {
	if set.UxPurpose != apitype.PermissionDescriptorUXPurposeSet {
		return RbacPermissionSetInput{}, fmt.Errorf(
			"descriptor %q is not a permission set (uxPurpose=%q); use `RbacRole` for roles or "+
				"`pulumiservice:api:Role` for other descriptor kinds", set.ID, set.UxPurpose)
	}
	allow, ok := set.Details.(apitype.PermissionDescriptorAllow)
	if !ok {
		return RbacPermissionSetInput{}, fmt.Errorf(
			"permission set %q does not grant a flat list of scopes (%s) and cannot be managed by "+
				"`RbacPermissionSet`; use `pulumiservice:api:Role` instead", set.ID, typeName(set.Details))
	}

	in := RbacPermissionSetInput{
		OrganizationName: orgName,
		Name:             set.Name,
		Description:      prior.Description,
		ResourceType:     RbacResourceType(set.ResourceType),
	}
	if set.Description != "" || prior.Description != nil {
		in.Description = &set.Description
	}

	// Split the service's flat list back into the two inputs: scopes in this
	// provider's enum go to `permissions` unless the user listed them in
	// `additionalPermissions`.
	priorExtra := map[string]bool{}
	for _, s := range prior.AdditionalPermissions {
		priorExtra[s] = true
	}
	var perms []RbacScope
	var extra []string
	for _, p := range allow.Permissions() {
		s := string(p)
		if isKnownScope(s) && !priorExtra[s] {
			perms = append(perms, RbacScope(s))
		} else {
			extra = append(extra, s)
		}
	}
	// Carry forward scopes the SDK cannot report back, rather than showing
	// them as removed on every refresh. The trade-off: if such a scope is
	// removed outside Pulumi (in the console, or because Pulumi Cloud retires
	// it), refresh won't report the drift until the Cloud SDK is bumped to a
	// version that knows the scope. That is unlikely in practice, because
	// Pulumi Cloud rarely retires scopes and Renovate bumps the SDK often.
	for _, s := range prior.Permissions {
		if !sdkCanRead(string(s)) {
			perms = append(perms, s)
		}
	}
	for _, s := range prior.AdditionalPermissions {
		if !sdkCanRead(s) {
			extra = append(extra, s)
		}
	}
	in.Permissions = keepOrderIfSameSet(prior.Permissions, perms)
	in.AdditionalPermissions = keepOrderIfSameSet(prior.AdditionalPermissions, extra)
	return in, nil
}

// keepOrderIfSameSet returns prior when it holds the same elements as actual,
// so that a reordering by the service does not show up as drift.
func keepOrderIfSameSet[T comparable](prior, actual []T) []T {
	if len(prior) != len(actual) {
		return actual
	}
	seen := make(map[T]int, len(prior))
	for _, v := range prior {
		seen[v]++
	}
	for _, v := range actual {
		if seen[v] == 0 {
			return actual
		}
		seen[v]--
	}
	return prior
}

func isKnownScope(s string) bool {
	return knownScopes()[RbacScope(s)]
}

var knownScopes = sync.OnceValue(func() map[RbacScope]bool {
	known := map[RbacScope]bool{}
	for _, v := range RbacScope("").Values() {
		known[v.Value] = true
	}
	return known
})

func scopeResourceTypes(s RbacScope) []string {
	var out []string
	for _, v := range RbacResourceType("").Values() {
		if slices.Contains(rbacScopesByResourceType[v.Value], s) {
			out = append(out, string(v.Value))
		}
	}
	if len(out) == 0 {
		return []string{"(unknown scope)"}
	}
	return out
}

// deletePermissionDescriptor deletes a role, policy, or permission set. It
// tries the unprivileged delete first so tokens without force-delete scope can
// remove unreferenced descriptors, then escalates to force on 409 (still
// assigned to members, teams, or tokens).
func deletePermissionDescriptor(ctx context.Context, orgName, id string) error {
	client := config.GetClient(ctx)
	err := client.DeleteRole(ctx, orgName, id, false)
	if err != nil && pulumiapi.GetErrorStatusCode(err) == http.StatusConflict {
		err = client.DeleteRole(ctx, orgName, id, true)
	}
	return err
}
