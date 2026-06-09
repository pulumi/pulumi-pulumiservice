// Copyright 2026, Pulumi Corporation.
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
	a.Describe(
		&PolicyGroup{},
		"A Policy Group allows you to apply policy packs to a set of stacks in your organization.",
	)
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

// PolicyGroupPolicyPackReferenceInput is the input shape for a policy pack
// applied to a policy group. The numeric `version` is intentionally omitted —
// it is server-derived from `versionTag` and only appears on output.
type PolicyGroupPolicyPackReferenceInput struct {
	Name        string                 `pulumi:"name"`
	DisplayName string                 `pulumi:"displayName,optional"`
	VersionTag  string                 `pulumi:"versionTag,optional"`
	Config      map[string]interface{} `pulumi:"config,optional"`
}

func (p *PolicyGroupPolicyPackReferenceInput) Annotate(a infer.Annotator) {
	a.Describe(p, "A reference to a policy pack within a policy group (input).")
	a.Describe(&p.Name, "The name of the policy pack.")
	a.Describe(&p.DisplayName, "The display name of the policy pack.")
	a.Describe(&p.VersionTag, "The version tag of the policy pack.")
	a.Describe(
		&p.Config,
		"Optional configuration for the policy pack. The special key `all` sets the default enforcement "+
			"level for every policy in the pack; per-policy entries override it.",
	)
}

// PolicyGroupPolicyPackReference is the output shape for a policy pack applied
// to a policy group; it includes the server-derived numeric `version`.
type PolicyGroupPolicyPackReference struct {
	Name        string                 `pulumi:"name"`
	DisplayName string                 `pulumi:"displayName,optional"`
	Version     int                    `pulumi:"version,optional"`
	VersionTag  string                 `pulumi:"versionTag,optional"`
	Config      map[string]interface{} `pulumi:"config,optional"`
}

func (p *PolicyGroupPolicyPackReference) Annotate(a infer.Annotator) {
	a.Describe(p, "A reference to a policy pack within a policy group.")
	a.Describe(&p.Name, "The name of the policy pack.")
	a.Describe(&p.DisplayName, "The display name of the policy pack.")
	a.Describe(
		&p.Version,
		"The server-derived numeric version of the policy pack. This is output-only; "+
			"use `versionTag` to pin a specific version.",
	)
	a.Describe(&p.VersionTag, "The version tag of the policy pack.")
	a.Describe(
		&p.Config,
		"Optional configuration for the policy pack. The special key `all` sets the default enforcement "+
			"level for every policy in the pack; per-policy entries override it.",
	)
}

type PolicyGroupInput struct {
	Name             string                                `pulumi:"name"             provider:"replaceOnChanges"`
	OrganizationName string                                `pulumi:"organizationName" provider:"replaceOnChanges"`
	EntityType       string                                `pulumi:"entityType,optional"       provider:"replaceOnChanges"`
	Mode             string                                `pulumi:"mode,optional"             provider:"replaceOnChanges"`
	Stacks           []PolicyGroupStackReference           `pulumi:"stacks,optional"`
	Accounts         []string                              `pulumi:"accounts,optional"`
	PolicyPacks      []PolicyGroupPolicyPackReferenceInput `pulumi:"policyPacks,optional"`
}

func (i *PolicyGroupInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Name, "The name of the policy group.")
	a.Describe(&i.OrganizationName, "The name of the Pulumi organization the policy group belongs to.")
	a.Describe(
		&i.EntityType,
		"The entity type for the policy group. Valid values are 'stacks' or 'accounts'. Defaults to 'stacks'.",
	)
	a.SetDefault(&i.EntityType, "stacks")
	a.Describe(
		&i.Mode,
		"The mode for the policy group. Valid values are 'audit' (reports violations) or "+
			"'preventative' (blocks operations). Defaults to 'audit'.",
	)
	a.SetDefault(&i.Mode, "audit")
	a.Describe(&i.Stacks, "List of stack references that belong to this policy group.")
	a.Describe(&i.Accounts, "List of accounts that belong to this policy group.")
	a.Describe(&i.PolicyPacks, "List of policy packs applied to this policy group.")
}

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
	a.Describe(
		&s.EntityType,
		"The entity type for the policy group. Valid values are 'stacks' or 'accounts'. Defaults to 'stacks'.",
	)
	a.Describe(
		&s.Mode,
		"The mode for the policy group. Valid values are 'audit' (reports violations) or "+
			"'preventative' (blocks operations). Defaults to 'audit'.",
	)
	a.Describe(&s.Stacks, "List of stack references that belong to this policy group.")
	a.Describe(&s.Accounts, "List of accounts that belong to this policy group.")
	a.Describe(&s.PolicyPacks, "List of policy packs applied to this policy group.")
}

// Check applies defaults, validates enum values, and strips the server-derived
// `version` field from policy pack inputs (so legacy programs that still set
// it upgrade cleanly).
func (*PolicyGroup) Check(
	ctx context.Context, req infer.CheckRequest,
) (infer.CheckResponse[PolicyGroupInput], error) {
	if packs, ok := req.NewInputs.GetOk("policyPacks"); ok && packs.IsArray() {
		items := packs.AsArray().AsSlice()
		stripped := make([]property.Value, 0, len(items))
		for _, item := range items {
			if item.IsMap() {
				m := item.AsMap()
				if _, has := m.GetOk("version"); has {
					m = m.Delete("version")
				}
				stripped = append(stripped, property.New(m))
				continue
			}
			stripped = append(stripped, item)
		}
		req.NewInputs = req.NewInputs.Set("policyPacks", property.New(property.NewArray(stripped)))
	}

	in, failures, err := infer.DefaultCheck[PolicyGroupInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[PolicyGroupInput]{}, err
	}
	if in.EntityType != "stacks" && in.EntityType != "accounts" {
		failures = append(failures, p.CheckFailure{
			Property: "entityType",
			Reason:   "entityType must be either 'stacks' or 'accounts'",
		})
	}
	if in.Mode != "audit" && in.Mode != "preventative" {
		failures = append(failures, p.CheckFailure{
			Property: "mode",
			Reason:   "mode must be either 'audit' or 'preventative'",
		})
	}
	return infer.CheckResponse[PolicyGroupInput]{Inputs: in, Failures: failures}, nil
}

// Diff compares inputs against state using order-independent comparison on the
// stacks/accounts/policyPacks slices, mirroring the legacy resource's
// behavior. Name, organizationName, entityType, and mode force replacement.
func (*PolicyGroup) Diff(
	_ context.Context,
	req infer.DiffRequest[PolicyGroupInput, PolicyGroupState],
) (infer.DiffResponse, error) {
	diff := map[string]p.PropertyDiff{}
	if req.State.Name != req.Inputs.Name {
		diff["name"] = p.PropertyDiff{Kind: p.UpdateReplace, InputDiff: true}
	}
	if req.State.OrganizationName != req.Inputs.OrganizationName {
		diff["organizationName"] = p.PropertyDiff{Kind: p.UpdateReplace, InputDiff: true}
	}
	if req.State.EntityType != req.Inputs.EntityType {
		diff["entityType"] = p.PropertyDiff{Kind: p.UpdateReplace, InputDiff: true}
	}
	if req.State.Mode != req.Inputs.Mode {
		diff["mode"] = p.PropertyDiff{Kind: p.UpdateReplace, InputDiff: true}
	}
	if !stackReferencesEqual(req.State.Stacks, req.Inputs.Stacks) {
		diff["stacks"] = p.PropertyDiff{Kind: p.Update, InputDiff: true}
	}
	if !util.ElementsEqual(req.State.Accounts, req.Inputs.Accounts) {
		diff["accounts"] = p.PropertyDiff{Kind: p.Update, InputDiff: true}
	}
	if !policyPackInputsEqualState(req.Inputs.PolicyPacks, req.State.PolicyPacks) {
		diff["policyPacks"] = p.PropertyDiff{Kind: p.Update, InputDiff: true}
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
	id := policyGroupResourceID(req.Inputs.OrganizationName, req.Inputs.Name)
	if req.DryRun {
		return infer.CreateResponse[PolicyGroupState]{
			ID:     id,
			Output: stateFromInputs(req.Inputs),
		}, nil
	}

	client := config.GetClient(ctx)

	err := client.CreatePolicyGroup(
		ctx, req.Inputs.OrganizationName, req.Inputs.Name, req.Inputs.EntityType, req.Inputs.Mode,
	)
	if err != nil {
		return infer.CreateResponse[PolicyGroupState]{}, fmt.Errorf(
			"error creating policy group '%s': %w", req.Inputs.Name, err,
		)
	}

	batchReqs := make([]pulumiapi.UpdatePolicyGroupRequest, 0,
		len(req.Inputs.Stacks)+len(req.Inputs.PolicyPacks)+len(req.Inputs.Accounts))
	for _, stack := range req.Inputs.Stacks {
		s := stackReferenceToAPI(stack)
		batchReqs = append(batchReqs, pulumiapi.UpdatePolicyGroupRequest{AddStack: &s})
	}
	for _, pp := range req.Inputs.PolicyPacks {
		api := policyPackInputToAPI(pp)
		batchReqs = append(batchReqs, pulumiapi.UpdatePolicyGroupRequest{AddPolicyPack: &api})
	}
	for _, account := range req.Inputs.Accounts {
		ref := pulumiapi.InsightsAccountReference{Name: account}
		batchReqs = append(batchReqs, pulumiapi.UpdatePolicyGroupRequest{AddInsightsAccount: &ref})
	}

	if len(batchReqs) > 0 {
		err = client.BatchUpdatePolicyGroup(ctx, req.Inputs.OrganizationName, req.Inputs.Name, batchReqs)
		if err != nil {
			return infer.CreateResponse[PolicyGroupState]{
					ID:     id,
					Output: stateFromInputs(req.Inputs),
				}, infer.ResourceInitFailedError{
					Reasons: []string{fmt.Sprintf("failed to add items to policy group: %s", err.Error())},
				}
		}
	}

	pg, err := client.GetPolicyGroup(ctx, req.Inputs.OrganizationName, req.Inputs.Name)
	if err != nil {
		return infer.CreateResponse[PolicyGroupState]{
				ID:     id,
				Output: stateFromInputs(req.Inputs),
			}, infer.ResourceInitFailedError{
				Reasons: []string{err.Error()},
			}
	}
	if pg == nil {
		return infer.CreateResponse[PolicyGroupState]{
				ID:     id,
				Output: stateFromInputs(req.Inputs),
			}, infer.ResourceInitFailedError{
				Reasons: []string{fmt.Sprintf("policy group '%s' not found after creation", req.Inputs.Name)},
			}
	}

	return infer.CreateResponse[PolicyGroupState]{
		ID:     id,
		Output: stateFromAPI(req.Inputs.OrganizationName, pg),
	}, nil
}

func (*PolicyGroup) Update(
	ctx context.Context, req infer.UpdateRequest[PolicyGroupInput, PolicyGroupState],
) (infer.UpdateResponse[PolicyGroupState], error) {
	if req.DryRun {
		return infer.UpdateResponse[PolicyGroupState]{
			Output: stateFromInputs(req.Inputs),
		}, nil
	}

	client := config.GetClient(ctx)
	batchReqs := buildUpdateBatch(req.State, req.Inputs)

	// The Cloud reorders/upserts policy packs by name when several mutations
	// share one batch, so send each op in its own request. On failure, earlier
	// ops have already mutated the Cloud; fetch the real state for the
	// checkpoint, detaching ctx so a cancellation that caused the failure does
	// not also kill the read.
	for i := range batchReqs {
		if err := client.BatchUpdatePolicyGroup(
			ctx, req.Inputs.OrganizationName, req.Inputs.Name, batchReqs[i:i+1],
		); err != nil {
			output := stateFromInputs(req.Inputs)
			readCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
			if pg, getErr := client.GetPolicyGroup(
				readCtx, req.Inputs.OrganizationName, req.Inputs.Name,
			); getErr == nil && pg != nil {
				output = stateFromAPI(req.Inputs.OrganizationName, pg)
			}
			cancel()
			return infer.UpdateResponse[PolicyGroupState]{Output: output},
				infer.ResourceInitFailedError{
					Reasons: []string{fmt.Sprintf("failed to update policy group: %s", err.Error())},
				}
		}
	}

	pg, err := client.GetPolicyGroup(ctx, req.Inputs.OrganizationName, req.Inputs.Name)
	if err != nil {
		return infer.UpdateResponse[PolicyGroupState]{}, fmt.Errorf(
			"failed to read policy group after update: %w", err,
		)
	}
	if pg == nil {
		return infer.UpdateResponse[PolicyGroupState]{}, fmt.Errorf(
			"policy group '%s' not found after update", req.Inputs.Name,
		)
	}
	return infer.UpdateResponse[PolicyGroupState]{
		Output: stateFromAPI(req.Inputs.OrganizationName, pg),
	}, nil
}

func (*PolicyGroup) Delete(
	ctx context.Context, req infer.DeleteRequest[PolicyGroupState],
) (infer.DeleteResponse, error) {
	err := config.GetClient(ctx).DeletePolicyGroup(ctx, req.State.OrganizationName, req.State.Name)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf(
			"failed to delete policy group %q/%q: %w",
			req.State.OrganizationName, req.State.Name, err,
		)
	}
	return infer.DeleteResponse{}, nil
}

// Read returns the current state of the policy group. For accounts, if the API
// returns the same set the previous state already had, we preserve the user's
// original input list (so auto-added child accounts don't pollute the program's
// declared input).
func (*PolicyGroup) Read(
	ctx context.Context, req infer.ReadRequest[PolicyGroupInput, PolicyGroupState],
) (infer.ReadResponse[PolicyGroupInput, PolicyGroupState], error) {
	orgName, policyGroupName, err := splitSingleSlashString(req.ID)
	if err != nil {
		return infer.ReadResponse[PolicyGroupInput, PolicyGroupState]{}, err
	}

	pg, err := config.GetClient(ctx).GetPolicyGroup(ctx, orgName, policyGroupName)
	if err != nil {
		return infer.ReadResponse[PolicyGroupInput, PolicyGroupState]{}, fmt.Errorf(
			"failed to read policy group (%q): %w", req.ID, err,
		)
	}
	if pg == nil {
		return infer.ReadResponse[PolicyGroupInput, PolicyGroupState]{}, nil
	}

	state := stateFromAPI(orgName, pg)

	// For accounts, if the API returned the same set the previous state had,
	// the user's input may be a subset (parent only) — preserve it.
	inputAccounts := state.Accounts
	if util.ElementsEqual(req.State.Accounts, state.Accounts) && req.Inputs.Accounts != nil {
		inputAccounts = req.Inputs.Accounts
	}

	inputs := PolicyGroupInput{
		Name:             state.Name,
		OrganizationName: state.OrganizationName,
		EntityType:       state.EntityType,
		Mode:             state.Mode,
		Stacks:           state.Stacks,
		Accounts:         inputAccounts,
		PolicyPacks:      policyPackStateToInputs(state.PolicyPacks),
	}

	return infer.ReadResponse[PolicyGroupInput, PolicyGroupState]{
		ID:     req.ID,
		Inputs: inputs,
		State:  state,
	}, nil
}

// --- helpers shared by Create/Update/Read ---

func policyGroupResourceID(orgName, policyGroupName string) string {
	return fmt.Sprintf("%s/%s", orgName, policyGroupName)
}

func stackReferenceToAPI(s PolicyGroupStackReference) pulumiapi.StackReference {
	return pulumiapi.StackReference{Name: s.Name, RoutingProject: s.RoutingProject}
}

func stackReferenceFromAPI(s pulumiapi.StackReference) PolicyGroupStackReference {
	return PolicyGroupStackReference{Name: s.Name, RoutingProject: s.RoutingProject}
}

func policyPackInputToAPI(p PolicyGroupPolicyPackReferenceInput) pulumiapi.PolicyPackMetadata {
	return pulumiapi.PolicyPackMetadata{
		Name:        p.Name,
		DisplayName: p.DisplayName,
		VersionTag:  p.VersionTag,
		Config:      p.Config,
	}
}

func policyPackFromAPI(p pulumiapi.PolicyPackMetadata) PolicyGroupPolicyPackReference {
	return PolicyGroupPolicyPackReference{
		Name:        p.Name,
		DisplayName: p.DisplayName,
		Version:     p.Version,
		VersionTag:  p.VersionTag,
		Config:      p.Config,
	}
}

func policyPackStateToInputs(packs []PolicyGroupPolicyPackReference) []PolicyGroupPolicyPackReferenceInput {
	if len(packs) == 0 {
		return nil
	}
	out := make([]PolicyGroupPolicyPackReferenceInput, len(packs))
	for i, p := range packs {
		out[i] = PolicyGroupPolicyPackReferenceInput{
			Name:        p.Name,
			DisplayName: p.DisplayName,
			VersionTag:  p.VersionTag,
			Config:      p.Config,
		}
	}
	return out
}

func stateFromAPI(orgName string, pg *pulumiapi.PolicyGroup) PolicyGroupState {
	state := PolicyGroupState{
		Name:             pg.Name,
		OrganizationName: orgName,
		EntityType:       pg.EntityType,
		Mode:             pg.Mode,
		Accounts:         pg.Accounts,
	}
	if len(pg.Stacks) > 0 {
		state.Stacks = make([]PolicyGroupStackReference, len(pg.Stacks))
		for i, s := range pg.Stacks {
			state.Stacks[i] = stackReferenceFromAPI(s)
		}
	}
	if len(pg.AppliedPolicyPacks) > 0 {
		state.PolicyPacks = make([]PolicyGroupPolicyPackReference, len(pg.AppliedPolicyPacks))
		for i, pp := range pg.AppliedPolicyPacks {
			state.PolicyPacks[i] = policyPackFromAPI(pp)
		}
	}
	return state
}

// stateFromInputs builds a preview-time state. The policy pack `version`
// field is server-derived; on preview we leave it zero.
func stateFromInputs(in PolicyGroupInput) PolicyGroupState {
	state := PolicyGroupState{
		Name:             in.Name,
		OrganizationName: in.OrganizationName,
		EntityType:       in.EntityType,
		Mode:             in.Mode,
		Stacks:           in.Stacks,
		Accounts:         in.Accounts,
	}
	if len(in.PolicyPacks) > 0 {
		state.PolicyPacks = make([]PolicyGroupPolicyPackReference, len(in.PolicyPacks))
		for i, p := range in.PolicyPacks {
			state.PolicyPacks[i] = PolicyGroupPolicyPackReference{
				Name:        p.Name,
				DisplayName: p.DisplayName,
				VersionTag:  p.VersionTag,
				Config:      p.Config,
			}
		}
	}
	return state
}

// --- diff & update helpers ---

func stackReferencesEqual(a, b []PolicyGroupStackReference) bool {
	return util.ElementsEqualFunc(a, b, compareNewStackRefs, newStackRefsEq)
}

func compareNewStackRefs(i, j PolicyGroupStackReference) int {
	if i.RoutingProject != j.RoutingProject {
		if i.RoutingProject < j.RoutingProject {
			return -1
		}
		return 1
	}
	if i.Name < j.Name {
		return -1
	}
	if i.Name > j.Name {
		return 1
	}
	return 0
}

func newStackRefsEq(a, b PolicyGroupStackReference) bool {
	return a.Name == b.Name && a.RoutingProject == b.RoutingProject
}

// policyPackInputsEqualState compares the user inputs (no `version`) against
// the stored state (with server-derived `version`). Equality is determined by
// (name, versionTag, config) and is order-independent. A nil config and an
// empty config map are treated as equal — both mean "no config".
func policyPackInputsEqualState(
	inputs []PolicyGroupPolicyPackReferenceInput, state []PolicyGroupPolicyPackReference,
) bool {
	if len(inputs) != len(state) {
		return false
	}
	matched := make([]bool, len(state))
	for _, in := range inputs {
		found := false
		for j, s := range state {
			if matched[j] {
				continue
			}
			if in.Name == s.Name && in.VersionTag == s.VersionTag && configEqual(in.Config, s.Config) {
				matched[j] = true
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// configEqual reports whether two policy-pack config maps are equivalent for
// diff purposes. A nil map and an empty map both mean "no config".
func configEqual(x, y map[string]interface{}) bool {
	if len(x) == 0 && len(y) == 0 {
		return true
	}
	return reflect.DeepEqual(x, y)
}

func buildUpdateBatch(state PolicyGroupState, inputs PolicyGroupInput) []pulumiapi.UpdatePolicyGroupRequest {
	var batch []pulumiapi.UpdatePolicyGroupRequest

	if !stackReferencesEqual(state.Stacks, inputs.Stacks) {
		for _, old := range state.Stacks {
			if !containsNewStackRef(inputs.Stacks, old) {
				ref := stackReferenceToAPI(old)
				batch = append(batch, pulumiapi.UpdatePolicyGroupRequest{RemoveStack: &ref})
			}
		}
		for _, n := range inputs.Stacks {
			if !containsNewStackRef(state.Stacks, n) {
				ref := stackReferenceToAPI(n)
				batch = append(batch, pulumiapi.UpdatePolicyGroupRequest{AddStack: &ref})
			}
		}
	}

	if !policyPackInputsEqualState(inputs.PolicyPacks, state.PolicyPacks) {
		for _, old := range state.PolicyPacks {
			if !containsPolicyPackInputByNameTag(inputs.PolicyPacks, old.Name, old.VersionTag) {
				api := pulumiapi.PolicyPackMetadata{
					Name:        old.Name,
					DisplayName: old.DisplayName,
					Version:     old.Version,
					VersionTag:  old.VersionTag,
					Config:      old.Config,
				}
				batch = append(batch, pulumiapi.UpdatePolicyGroupRequest{RemovePolicyPack: &api})
			}
		}
		for _, n := range inputs.PolicyPacks {
			if !containsPolicyPackStateByNameTag(state.PolicyPacks, n.Name, n.VersionTag) {
				api := policyPackInputToAPI(n)
				batch = append(batch, pulumiapi.UpdatePolicyGroupRequest{AddPolicyPack: &api})
			}
		}
	}

	// Skip removal of child accounts whose parent is still present — child
	// accounts are auto-managed when their parent is added.
	if !util.ElementsEqual(state.Accounts, inputs.Accounts) {
		for _, old := range state.Accounts {
			if !slices.Contains(inputs.Accounts, old) && !hasParentAccount(old, inputs.Accounts) {
				ref := pulumiapi.InsightsAccountReference{Name: old}
				batch = append(batch, pulumiapi.UpdatePolicyGroupRequest{RemoveInsightsAccount: &ref})
			}
		}
		for _, n := range inputs.Accounts {
			if !slices.Contains(state.Accounts, n) {
				ref := pulumiapi.InsightsAccountReference{Name: n}
				batch = append(batch, pulumiapi.UpdatePolicyGroupRequest{AddInsightsAccount: &ref})
			}
		}
	}

	return batch
}

func containsNewStackRef(stacks []PolicyGroupStackReference, target PolicyGroupStackReference) bool {
	for _, s := range stacks {
		if s.Name == target.Name && s.RoutingProject == target.RoutingProject {
			return true
		}
	}
	return false
}

func containsPolicyPackInputByNameTag(packs []PolicyGroupPolicyPackReferenceInput, name, tag string) bool {
	for _, p := range packs {
		if p.Name == name && p.VersionTag == tag {
			return true
		}
	}
	return false
}

func containsPolicyPackStateByNameTag(packs []PolicyGroupPolicyPackReference, name, tag string) bool {
	for _, p := range packs {
		if p.Name == name && p.VersionTag == tag {
			return true
		}
	}
	return false
}

// hasParentAccount checks if the given account has a parent account in the
// list. Account names use "/" as a separator, so "parent/child" has parent
// "parent".
func hasParentAccount(account string, accounts []string) bool {
	for _, acc := range accounts {
		if strings.HasPrefix(account, acc+"/") {
			return true
		}
	}
	return false
}
