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
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

const (
	gcName             = "name"
	gcOrganizationName = "organizationName"
	gcEntityType       = "entityType"
	gcMode             = "mode"
	gcStacks           = "stacks"
	gcAccounts         = "accounts"
	gcPolicyPacks      = "policyPacks"
	gcDisplayName      = "displayName"
	gcVersionTag       = "versionTag"
	gcAudit            = "audit"
	gcPreventative     = "preventative"
)

const gcPolicyPackVersion = "version"

// PolicyGroup manages a Pulumi Cloud policy group: a set of stacks or
// Insights accounts that a set of policy packs applies to.
type PolicyGroup struct{}

var (
	_ infer.CustomCheck[PolicyGroupInput]                    = &PolicyGroup{}
	_ infer.CustomDiff[PolicyGroupInput, PolicyGroupState]   = &PolicyGroup{}
	_ infer.CustomCreate[PolicyGroupInput, PolicyGroupState] = &PolicyGroup{}
	_ infer.CustomUpdate[PolicyGroupInput, PolicyGroupState] = &PolicyGroup{}
	_ infer.CustomDelete[PolicyGroupState]                   = &PolicyGroup{}
	_ infer.CustomRead[PolicyGroupInput, PolicyGroupState]   = &PolicyGroup{}
)

func (*PolicyGroup) Annotate(a infer.Annotator) {
	a.Describe(&PolicyGroup{},
		"A Policy Group allows you to apply policy packs to a set of stacks in your organization."+
			util.ImportDocs("pulumiservice:index:PolicyGroup",
				"{organization}/{policyGroupName}", "my-org/my-policy-group"))
	a.SetToken("index", "PolicyGroup")
}

// PolicyGroupStackReference is a reference to a stack within a policy group.
type PolicyGroupStackReference struct {
	Name           string `pulumi:"name"`
	RoutingProject string `pulumi:"routingProject"`
}

func (s *PolicyGroupStackReference) Annotate(a infer.Annotator) {
	a.Describe(s, "A reference to a stack within a policy group.")
	a.Describe(&s.Name, "The name of the stack.")
	a.Describe(&s.RoutingProject, "The routing project name (also known as project name).")
}

const policyPackConfigDescription = "Optional configuration for the policy pack. The special key `all` sets " +
	"the default enforcement level for every policy in the pack; per-policy entries override it."

// PolicyGroupPolicyPackReferenceInput is the input shape of a policy pack
// applied to a policy group. It has no `version`: that's derived by the
// server from `versionTag` and only appears in outputs.
type PolicyGroupPolicyPackReferenceInput struct {
	Name        string         `pulumi:"name"`
	DisplayName string         `pulumi:"displayName,optional"`
	VersionTag  string         `pulumi:"versionTag,optional"`
	Config      map[string]any `pulumi:"config,optional"`
}

func (r *PolicyGroupPolicyPackReferenceInput) Annotate(a infer.Annotator) {
	a.Describe(r, "A reference to a policy pack within a policy group (input).")
	a.Describe(&r.Name, "The name of the policy pack.")
	a.Describe(&r.DisplayName, "The display name of the policy pack.")
	a.Describe(&r.VersionTag, "The version tag of the policy pack.")
	a.Describe(&r.Config, policyPackConfigDescription)
}

// PolicyGroupPolicyPackReference is the output shape of a policy pack applied
// to a policy group.
//
// Version is a float64 so the schema keeps the `number` type that the
// pre-infer resource published. Changing it to `integer` would change the
// generated SDK types.
type PolicyGroupPolicyPackReference struct {
	Name        string         `pulumi:"name"`
	DisplayName string         `pulumi:"displayName,optional"`
	Version     float64        `pulumi:"version,optional"`
	VersionTag  string         `pulumi:"versionTag,optional"`
	Config      map[string]any `pulumi:"config,optional"`
}

func (r *PolicyGroupPolicyPackReference) Annotate(a infer.Annotator) {
	a.Describe(r, "A reference to a policy pack within a policy group.")
	a.Describe(&r.Name, "The name of the policy pack.")
	a.Describe(&r.DisplayName, "The display name of the policy pack.")
	a.Describe(&r.Version, "The server-derived numeric version of the policy pack. This is output-only; "+
		"use `versionTag` to pin a specific version.")
	a.Describe(&r.VersionTag, "The version tag of the policy pack.")
	a.Describe(&r.Config, policyPackConfigDescription)
}

const (
	policyGroupEntityTypeDescription = "The entity type for the policy group. Valid values are 'stacks' or " +
		"'accounts'. Defaults to 'stacks'."
	policyGroupModeDescription = "The mode for the policy group. Valid values are 'audit' (reports violations) " +
		"or 'preventative' (blocks operations). Defaults to 'audit'."
)

type PolicyGroupInput struct {
	Name             string                                `pulumi:"name"                provider:"replaceOnChanges"`
	OrganizationName string                                `pulumi:"organizationName"    provider:"replaceOnChanges"`
	EntityType       string                                `pulumi:"entityType,optional" provider:"replaceOnChanges"`
	Mode             string                                `pulumi:"mode,optional"       provider:"replaceOnChanges"`
	Stacks           []PolicyGroupStackReference           `pulumi:"stacks,optional"`
	Accounts         []string                              `pulumi:"accounts,optional"`
	PolicyPacks      []PolicyGroupPolicyPackReferenceInput `pulumi:"policyPacks,optional"`
}

func (i *PolicyGroupInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Name, "The name of the policy group.")
	a.Describe(&i.OrganizationName, "The name of the Pulumi organization the policy group belongs to.")
	a.Describe(&i.EntityType, policyGroupEntityTypeDescription)
	a.SetDefault(&i.EntityType, gcStacks)
	a.Describe(&i.Mode, policyGroupModeDescription)
	a.SetDefault(&i.Mode, gcAudit)
	a.Describe(&i.Stacks, "List of stack references that belong to this policy group.")
	a.Describe(&i.Accounts, "List of accounts that belong to this policy group.")
	a.Describe(&i.PolicyPacks, "List of policy packs applied to this policy group.")
}

// PolicyGroupState doesn't embed PolicyGroupInput: its policy packs carry the
// server-derived `version`.
type PolicyGroupState struct {
	Name             string                           `pulumi:"name"`
	OrganizationName string                           `pulumi:"organizationName"`
	EntityType       string                           `pulumi:"entityType"`
	Mode             string                           `pulumi:"mode"`
	Stacks           []PolicyGroupStackReference      `pulumi:"stacks,optional"`
	Accounts         []string                         `pulumi:"accounts,optional"`
	PolicyPacks      []PolicyGroupPolicyPackReference `pulumi:"policyPacks,optional"`
}

func (s *PolicyGroupState) Annotate(a infer.Annotator) {
	a.Describe(&s.Name, "The name of the policy group.")
	a.Describe(&s.OrganizationName, "The name of the Pulumi organization the policy group belongs to.")
	a.Describe(&s.EntityType, policyGroupEntityTypeDescription)
	a.Describe(&s.Mode, policyGroupModeDescription)
	a.Describe(&s.Stacks, "List of stack references that belong to this policy group.")
	a.Describe(&s.Accounts, "List of accounts that belong to this policy group.")
	a.Describe(&s.PolicyPacks, "List of policy packs applied to this policy group.")
}

// Check applies defaults and validates `entityType` and `mode`. It also strips
// `version` from each policy pack: that field is derived by the server and is
// output-only, but older programs may still set it.
func (*PolicyGroup) Check(
	ctx context.Context, req infer.CheckRequest,
) (infer.CheckResponse[PolicyGroupInput], error) {
	news := stripPolicyPackVersion(req.NewInputs)
	inputs, failures, err := infer.DefaultCheck[PolicyGroupInput](ctx, news)
	if err != nil || len(failures) > 0 {
		return infer.CheckResponse[PolicyGroupInput]{Inputs: inputs, Failures: failures}, err
	}

	// Only validate values that are known. An unknown value decodes as "".
	known := func(key string) bool {
		v, ok := news.GetOk(key)
		return !ok || !v.IsComputed()
	}
	if known(gcEntityType) && inputs.EntityType != gcStacks && inputs.EntityType != gcAccounts {
		failures = append(failures, p.CheckFailure{
			Property: gcEntityType,
			Reason:   "entityType must be either 'stacks' or 'accounts'",
		})
	}
	if known(gcMode) && inputs.Mode != gcAudit && inputs.Mode != gcPreventative {
		failures = append(failures, p.CheckFailure{
			Property: gcMode,
			Reason:   "mode must be either 'audit' or 'preventative'",
		})
	}
	return infer.CheckResponse[PolicyGroupInput]{Inputs: inputs, Failures: failures}, nil
}

// stripPolicyPackVersion removes `version` from each policy pack in inputs.
func stripPolicyPackVersion(inputs property.Map) property.Map {
	packs, ok := inputs.GetOk(gcPolicyPacks)
	if !ok || !packs.IsArray() {
		return inputs
	}
	items := packs.AsArray().AsSlice()
	stripped := make([]property.Value, len(items))
	for i, item := range items {
		if item.IsMap() {
			item = property.New(item.AsMap().Delete(gcPolicyPackVersion)).
				WithSecret(item.Secret()).WithDependencies(item.Dependencies())
		}
		stripped[i] = item
	}
	return inputs.Set(gcPolicyPacks, property.New(property.NewArray(stripped)).
		WithSecret(packs.Secret()).WithDependencies(packs.Dependencies()))
}

// Diff compares the inputs against the state without regard to the order of
// stacks, accounts and policy packs. A list differs exactly when Update would
// send an operation for it; see [policyGroupOps].
func (*PolicyGroup) Diff(
	_ context.Context, req infer.DiffRequest[PolicyGroupInput, PolicyGroupState],
) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	for _, f := range []struct {
		name       string
		state, new string
	}{
		{gcName, req.State.Name, req.Inputs.Name},
		{gcOrganizationName, req.State.OrganizationName, req.Inputs.OrganizationName},
		{gcEntityType, req.State.EntityType, req.Inputs.EntityType},
		// The API can't change the mode of a policy group.
		{gcMode, req.State.Mode, req.Inputs.Mode},
	} {
		if f.state != f.new {
			diff[f.name] = p.PropertyDiff{Kind: p.UpdateReplace, InputDiff: true}
		}
	}

	ops := policyGroupOps(req.State, req.Inputs)
	for name, ops := range map[string][]pulumiapi.UpdatePolicyGroupRequest{
		gcStacks:      ops.stacks,
		gcPolicyPacks: ops.policyPacks,
		gcAccounts:    ops.accounts,
	} {
		if len(ops) > 0 {
			diff[name] = p.PropertyDiff{Kind: p.Update, InputDiff: true}
		}
	}

	return infer.DiffResponse{
		HasChanges:          len(diff) > 0,
		DetailedDiff:        diff,
		DeleteBeforeReplace: true,
	}, nil
}

func (*PolicyGroup) Create(
	ctx context.Context, req infer.CreateRequest[PolicyGroupInput],
) (infer.CreateResponse[PolicyGroupState], error) {
	in := req.Inputs
	id := policyGroupID(in.OrganizationName, in.Name)
	if req.DryRun {
		return infer.CreateResponse[PolicyGroupState]{ID: id, Output: policyGroupStateFromInputs(in)}, nil
	}

	client := config.GetClient(ctx)
	if err := client.CreatePolicyGroup(ctx, in.OrganizationName, in.Name, in.EntityType, in.Mode); err != nil {
		return infer.CreateResponse[PolicyGroupState]{}, fmt.Errorf(
			"error creating policy group '%s': %w", in.Name, err)
	}

	// The policy group exists from here on, so failures return
	// ResourceInitFailedError to keep it in state.
	initFailed := func(state PolicyGroupState, reason string) (infer.CreateResponse[PolicyGroupState], error) {
		return infer.CreateResponse[PolicyGroupState]{ID: id, Output: state},
			infer.ResourceInitFailedError{Reasons: []string{reason}}
	}

	// Create only adds, so the Cloud's reorder/upsert-by-name behavior that
	// Update works around doesn't apply. Send everything in one request.
	ops := policyGroupOps(PolicyGroupState{}, in)
	batch := slices.Concat(ops.stacks, ops.policyPacks, ops.accounts)
	if len(batch) > 0 {
		if err := client.BatchUpdatePolicyGroup(ctx, in.OrganizationName, in.Name, batch); err != nil {
			return initFailed(policyGroupStateFromInputs(in), fmt.Sprintf("failed to add items to policy group: %s", err))
		}
	}

	pg, err := client.GetPolicyGroup(ctx, in.OrganizationName, in.Name)
	if err != nil {
		return initFailed(policyGroupStateFromInputs(in), err.Error())
	}
	if pg == nil {
		return initFailed(policyGroupStateFromInputs(in),
			fmt.Sprintf("policy group %q not found after creation", id))
	}
	return infer.CreateResponse[PolicyGroupState]{ID: id, Output: policyGroupStateFromAPI(in.OrganizationName, pg)}, nil
}

// Update sends each operation in its own request: when one batch holds
// several policy pack changes, the Cloud reorders them and upserts by name.
// If an operation fails, earlier ones have already changed the policy group,
// so Update reads back the real state and returns it with
// ResourceInitFailedError.
func (*PolicyGroup) Update(
	ctx context.Context, req infer.UpdateRequest[PolicyGroupInput, PolicyGroupState],
) (infer.UpdateResponse[PolicyGroupState], error) {
	in := req.Inputs
	if req.DryRun {
		return infer.UpdateResponse[PolicyGroupState]{Output: policyGroupStateFromInputs(in)}, nil
	}

	client := config.GetClient(ctx)
	ops := policyGroupOps(req.State, in)
	for _, op := range slices.Concat(ops.stacks, ops.policyPacks, ops.accounts) {
		err := client.BatchUpdatePolicyGroup(ctx, in.OrganizationName, in.Name,
			[]pulumiapi.UpdatePolicyGroupRequest{op})
		if err == nil {
			continue
		}
		// Detach ctx so a cancellation that caused the failure doesn't also
		// stop the read.
		readCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
		defer cancel()
		state := policyGroupStateFromInputs(in)
		if pg, getErr := client.GetPolicyGroup(readCtx, in.OrganizationName, in.Name); getErr == nil && pg != nil {
			state = policyGroupStateFromAPI(in.OrganizationName, pg)
		}
		return infer.UpdateResponse[PolicyGroupState]{Output: state}, infer.ResourceInitFailedError{
			Reasons: []string{fmt.Sprintf("failed to update policy group: %s", err)},
		}
	}

	// Read back the full state: adding an account also adds its child accounts.
	pg, err := client.GetPolicyGroup(ctx, in.OrganizationName, in.Name)
	if err != nil {
		return infer.UpdateResponse[PolicyGroupState]{}, fmt.Errorf("failed to read policy group after update: %w", err)
	}
	if pg == nil {
		return infer.UpdateResponse[PolicyGroupState]{}, fmt.Errorf(
			"policy group %q not found after update", policyGroupID(in.OrganizationName, in.Name))
	}
	return infer.UpdateResponse[PolicyGroupState]{Output: policyGroupStateFromAPI(in.OrganizationName, pg)}, nil
}

func (*PolicyGroup) Delete(
	ctx context.Context, req infer.DeleteRequest[PolicyGroupState],
) (infer.DeleteResponse, error) {
	orgName, name, err := splitSingleSlashString(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	if err := config.GetClient(ctx).DeletePolicyGroup(ctx, orgName, name); err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete policy group %q: %w", req.ID, err)
	}
	return infer.DeleteResponse{}, nil
}

// Read returns the policy group's current state. Adding an account also adds
// its child accounts, so when the accounts haven't changed since the last
// operation, Read keeps the accounts from the previous inputs. Otherwise the
// child accounts would show up in the program's inputs.
func (*PolicyGroup) Read(
	ctx context.Context, req infer.ReadRequest[PolicyGroupInput, PolicyGroupState],
) (infer.ReadResponse[PolicyGroupInput, PolicyGroupState], error) {
	orgName, name, err := splitSingleSlashString(req.ID)
	if err != nil {
		return infer.ReadResponse[PolicyGroupInput, PolicyGroupState]{}, err
	}

	pg, err := config.GetClient(ctx).GetPolicyGroup(ctx, orgName, name)
	if err != nil {
		return infer.ReadResponse[PolicyGroupInput, PolicyGroupState]{}, fmt.Errorf(
			"failed to read policy group (%q): %w", req.ID, err)
	}
	if pg == nil {
		return infer.ReadResponse[PolicyGroupInput, PolicyGroupState]{}, nil
	}

	state := policyGroupStateFromAPI(orgName, pg)
	inputAccounts := state.Accounts
	if sameStrings(req.State.Accounts, state.Accounts) {
		inputAccounts = req.Inputs.Accounts
	}

	inputs := PolicyGroupInput{
		Name:             state.Name,
		OrganizationName: state.OrganizationName,
		EntityType:       state.EntityType,
		Mode:             state.Mode,
		Stacks:           state.Stacks,
		Accounts:         inputAccounts,
	}
	for _, pp := range state.PolicyPacks {
		inputs.PolicyPacks = append(inputs.PolicyPacks, PolicyGroupPolicyPackReferenceInput{
			Name:        pp.Name,
			DisplayName: pp.DisplayName,
			VersionTag:  pp.VersionTag,
			Config:      pp.Config,
		})
	}

	return infer.ReadResponse[PolicyGroupInput, PolicyGroupState]{ID: req.ID, Inputs: inputs, State: state}, nil
}

func policyGroupID(orgName, name string) string {
	return fmt.Sprintf("%s/%s", orgName, name)
}

func policyGroupStateFromAPI(orgName string, pg *pulumiapi.PolicyGroup) PolicyGroupState {
	state := PolicyGroupState{
		Name:             pg.Name,
		OrganizationName: orgName,
		EntityType:       pg.EntityType,
		Mode:             pg.Mode,
		Accounts:         pg.Accounts,
	}
	for _, s := range pg.Stacks {
		state.Stacks = append(state.Stacks, PolicyGroupStackReference(s))
	}
	for _, pp := range pg.AppliedPolicyPacks {
		state.PolicyPacks = append(state.PolicyPacks, PolicyGroupPolicyPackReference{
			Name:        pp.Name,
			DisplayName: pp.DisplayName,
			Version:     float64(pp.Version),
			VersionTag:  pp.VersionTag,
			Config:      pp.Config,
		})
	}
	return state
}

// policyGroupStateFromInputs builds the state to report when the API can't
// say what it is: on preview, and when an operation fails before the policy
// group can be read back. It has no policy pack `version`.
func policyGroupStateFromInputs(in PolicyGroupInput) PolicyGroupState {
	state := PolicyGroupState{
		Name:             in.Name,
		OrganizationName: in.OrganizationName,
		EntityType:       in.EntityType,
		Mode:             in.Mode,
		Stacks:           in.Stacks,
		Accounts:         in.Accounts,
	}
	for _, pp := range in.PolicyPacks {
		state.PolicyPacks = append(state.PolicyPacks, PolicyGroupPolicyPackReference{
			Name:        pp.Name,
			DisplayName: pp.DisplayName,
			VersionTag:  pp.VersionTag,
			Config:      pp.Config,
		})
	}
	return state
}

// policyGroupUpdateOps holds the operations that bring a policy group from
// its state to its inputs, by list. Removals come before additions.
type policyGroupUpdateOps struct {
	stacks, policyPacks, accounts []pulumiapi.UpdatePolicyGroupRequest
}

// policyGroupOps computes the operations that bring a policy group from state
// to inputs. Order within each list doesn't matter.
//
//   - Stacks match on name and routing project.
//   - Policy packs match on name and version tag. A pack without a version tag
//     in the inputs matches a pack of that name at any version. A matched pack
//     whose config changed is removed and added again. A nil config and an
//     empty one are equal.
//   - An account that's no longer in the inputs isn't removed while its parent
//     account still is: the Cloud manages child accounts along with their
//     parent.
func policyGroupOps(state PolicyGroupState, in PolicyGroupInput) policyGroupUpdateOps {
	var ops policyGroupUpdateOps

	// Stacks
	for _, s := range state.Stacks {
		if !slices.Contains(in.Stacks, s) {
			ref := pulumiapi.StackReference(s)
			ops.stacks = append(ops.stacks, pulumiapi.UpdatePolicyGroupRequest{RemoveStack: &ref})
		}
	}
	for _, s := range in.Stacks {
		if !slices.Contains(state.Stacks, s) {
			ref := pulumiapi.StackReference(s)
			ops.stacks = append(ops.stacks, pulumiapi.UpdatePolicyGroupRequest{AddStack: &ref})
		}
	}

	// Policy packs
	matched := make([]bool, len(state.PolicyPacks))
	var removes, adds []pulumiapi.UpdatePolicyGroupRequest
	match := func(want PolicyGroupPolicyPackReferenceInput, anyVersion bool) int {
		for j, have := range state.PolicyPacks {
			if !matched[j] && have.Name == want.Name &&
				(have.VersionTag == want.VersionTag || anyVersion && want.VersionTag == "") {
				return j
			}
		}
		return -1
	}
	// Match exact version tags first, so a pack without a version tag can't
	// claim a state entry that another input names exactly.
	inputMatch := make([]int, len(in.PolicyPacks))
	for i, want := range in.PolicyPacks {
		if inputMatch[i] = match(want, false); inputMatch[i] >= 0 {
			matched[inputMatch[i]] = true
		}
	}
	for i, want := range in.PolicyPacks {
		if inputMatch[i] < 0 {
			if inputMatch[i] = match(want, true); inputMatch[i] >= 0 {
				matched[inputMatch[i]] = true
			}
		}
	}
	for i, want := range in.PolicyPacks {
		add := pulumiapi.PolicyPackMetadata{
			Name:        want.Name,
			DisplayName: want.DisplayName,
			VersionTag:  want.VersionTag,
			Config:      want.Config,
		}
		if j := inputMatch[i]; j >= 0 {
			if policyPackConfigEqual(state.PolicyPacks[j].Config, want.Config) {
				continue
			}
			matched[j] = false // Remove it below, then add it back.
		}
		adds = append(adds, pulumiapi.UpdatePolicyGroupRequest{AddPolicyPack: &add})
	}
	for j, have := range state.PolicyPacks {
		if !matched[j] {
			remove := pulumiapi.PolicyPackMetadata{
				Name:        have.Name,
				DisplayName: have.DisplayName,
				Version:     int(have.Version),
				VersionTag:  have.VersionTag,
				Config:      have.Config,
			}
			removes = append(removes, pulumiapi.UpdatePolicyGroupRequest{RemovePolicyPack: &remove})
		}
	}
	ops.policyPacks = append(removes, adds...)

	// Accounts
	for _, a := range state.Accounts {
		if !slices.Contains(in.Accounts, a) && !hasParentAccount(a, in.Accounts) {
			ref := pulumiapi.InsightsAccountReference{Name: a}
			ops.accounts = append(ops.accounts, pulumiapi.UpdatePolicyGroupRequest{RemoveInsightsAccount: &ref})
		}
	}
	for _, a := range in.Accounts {
		if !slices.Contains(state.Accounts, a) {
			ref := pulumiapi.InsightsAccountReference{Name: a}
			ops.accounts = append(ops.accounts, pulumiapi.UpdatePolicyGroupRequest{AddInsightsAccount: &ref})
		}
	}

	return ops
}

// policyPackConfigEqual reports whether two policy pack configs are equal. A
// nil config and an empty one both mean "no config".
func policyPackConfigEqual(x, y map[string]any) bool {
	if len(x) == 0 && len(y) == 0 {
		return true
	}
	return reflect.DeepEqual(x, y)
}

// sameStrings reports whether a and b hold the same strings, ignoring order.
func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	a, b = slices.Clone(a), slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}

// hasParentAccount checks if the given account has a parent account in the list.
// Account names use "/" as a separator, so "parent/child" has parent "parent".
func hasParentAccount(account string, accounts []string) bool {
	for _, acc := range accounts {
		if strings.HasPrefix(account, acc+"/") {
			return true
		}
	}
	return false
}
