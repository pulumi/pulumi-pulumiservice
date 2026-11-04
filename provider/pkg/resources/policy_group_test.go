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
	"errors"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/urn"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const (
	gcParentChild     = "parent/child"
	gcParent          = "parent"
	gcTestPolicyGroup = "test-policy-group"
	gcOther           = "other"
	gcProject1        = "project-1"
	gcStack1          = "stack-1"
	gcPolicyPack1     = "policy-pack-1"
	gcPolicyPack2     = "policy-pack-2"
	gcVersion200      = "2.0.0"
	gcStack2          = "stack-2"
	gcProject2        = "project-2"
	gcAllPolicies     = "all"
)

const testPolicyGroupID = gcTestOrg + "/" + gcTestPolicyGroup

var policyGroupURN = urn.New("dev", "test", "", "pulumiservice:index:PolicyGroup", "testPolicyGroup")

// PolicyGroupClientMock records every batch it's sent.
type PolicyGroupClientMock struct {
	config.Client
	getPolicyGroupFunc         func() (*pulumiapi.PolicyGroup, error)
	createPolicyGroupFunc      func(ctx context.Context, orgName, policyGroupName, entityType, mode string) error
	batchUpdatePolicyGroupFunc func(
		ctx context.Context, orgName, policyGroupName string, reqs []pulumiapi.UpdatePolicyGroupRequest,
	) error
	deletePolicyGroupFunc func(ctx context.Context, orgName, policyGroupName string) error

	batches [][]pulumiapi.UpdatePolicyGroupRequest
}

func (c *PolicyGroupClientMock) GetPolicyGroup(context.Context, string, string) (*pulumiapi.PolicyGroup, error) {
	if c.getPolicyGroupFunc != nil {
		return c.getPolicyGroupFunc()
	}
	return nil, nil
}

func (c *PolicyGroupClientMock) CreatePolicyGroup(
	ctx context.Context, orgName, policyGroupName, entityType, mode string,
) error {
	if c.createPolicyGroupFunc != nil {
		return c.createPolicyGroupFunc(ctx, orgName, policyGroupName, entityType, mode)
	}
	return nil
}

func (c *PolicyGroupClientMock) BatchUpdatePolicyGroup(
	ctx context.Context, orgName, policyGroupName string, reqs []pulumiapi.UpdatePolicyGroupRequest,
) error {
	c.batches = append(c.batches, reqs)
	if c.batchUpdatePolicyGroupFunc != nil {
		return c.batchUpdatePolicyGroupFunc(ctx, orgName, policyGroupName, reqs)
	}
	return nil
}

func (c *PolicyGroupClientMock) DeletePolicyGroup(ctx context.Context, orgName, policyGroupName string) error {
	if c.deletePolicyGroupFunc != nil {
		return c.deletePolicyGroupFunc(ctx, orgName, policyGroupName)
	}
	return nil
}

// ops flattens the recorded batches.
func (c *PolicyGroupClientMock) ops() []pulumiapi.UpdatePolicyGroupRequest {
	var all []pulumiapi.UpdatePolicyGroupRequest
	for _, b := range c.batches {
		all = append(all, b...)
	}
	return all
}

func returnPolicyGroup(pg *pulumiapi.PolicyGroup) func() (*pulumiapi.PolicyGroup, error) {
	return func() (*pulumiapi.PolicyGroup, error) { return pg, nil }
}

func mockPolicyGroupCtx(client *PolicyGroupClientMock) context.Context {
	return config.WithMockClient(context.Background(), client)
}

func newPolicyGroupServer(t *testing.T, client *PolicyGroupClientMock) integration.Server {
	t.Helper()
	prov, err := infer.NewProviderBuilder().
		WithResources(infer.Resource(&PolicyGroup{})).
		Build()
	require.NoError(t, err)
	server, err := integration.NewServer(config.WithMockClient(t.Context(), client),
		"pulumiservice", semver.MustParse("0.0.0"), integration.WithProvider(prov))
	require.NoError(t, err)
	return server
}

// Builders

type stackRef struct {
	name, project string
}

type policyPackRef struct {
	name       string
	version    int
	versionTag string
	config     map[string]any
}

func (s stackRef) ref() PolicyGroupStackReference {
	return PolicyGroupStackReference{Name: s.name, RoutingProject: s.project}
}

func (pp policyPackRef) input() PolicyGroupPolicyPackReferenceInput {
	return PolicyGroupPolicyPackReferenceInput{Name: pp.name, VersionTag: pp.versionTag, Config: pp.config}
}

func (pp policyPackRef) state() PolicyGroupPolicyPackReference {
	return PolicyGroupPolicyPackReference{
		Name: pp.name, Version: float64(pp.version), VersionTag: pp.versionTag, Config: pp.config,
	}
}

func (pp policyPackRef) api() pulumiapi.PolicyPackMetadata {
	return pulumiapi.PolicyPackMetadata{
		Name: pp.name, Version: pp.version, VersionTag: pp.versionTag, Config: pp.config,
	}
}

type policyGroupBuilder struct {
	entityType  string
	stacks      []stackRef
	policyPacks []policyPackRef
	accounts    []string
}

func newPolicyGroup() *policyGroupBuilder {
	return &policyGroupBuilder{entityType: gcStacks}
}

func (b *policyGroupBuilder) withStacks(stacks ...stackRef) *policyGroupBuilder {
	b.stacks = stacks
	return b
}

func (b *policyGroupBuilder) withPolicyPacks(packs ...policyPackRef) *policyGroupBuilder {
	b.policyPacks = packs
	return b
}

func (b *policyGroupBuilder) withAccounts(accounts ...string) *policyGroupBuilder {
	b.entityType = gcAccounts
	b.accounts = accounts
	return b
}

func (b *policyGroupBuilder) input() PolicyGroupInput {
	in := PolicyGroupInput{
		Name: gcTestPolicyGroup, OrganizationName: gcTestOrg, EntityType: b.entityType, Mode: gcAudit,
		Accounts: b.accounts,
	}
	for _, s := range b.stacks {
		in.Stacks = append(in.Stacks, s.ref())
	}
	for _, pp := range b.policyPacks {
		in.PolicyPacks = append(in.PolicyPacks, pp.input())
	}
	return in
}

func (b *policyGroupBuilder) state() PolicyGroupState {
	st := PolicyGroupState{
		Name: gcTestPolicyGroup, OrganizationName: gcTestOrg, EntityType: b.entityType, Mode: gcAudit,
		Accounts: b.accounts,
	}
	for _, s := range b.stacks {
		st.Stacks = append(st.Stacks, s.ref())
	}
	for _, pp := range b.policyPacks {
		st.PolicyPacks = append(st.PolicyPacks, pp.state())
	}
	return st
}

func (b *policyGroupBuilder) api() *pulumiapi.PolicyGroup {
	pg := &pulumiapi.PolicyGroup{
		Name: gcTestPolicyGroup, EntityType: b.entityType, Mode: gcAudit, Accounts: b.accounts,
	}
	for _, s := range b.stacks {
		pg.Stacks = append(pg.Stacks, pulumiapi.StackReference{Name: s.name, RoutingProject: s.project})
	}
	for _, pp := range b.policyPacks {
		pg.AppliedPolicyPacks = append(pg.AppliedPolicyPacks, pp.api())
	}
	return pg
}

// policyGroupProps builds a property map as the SDKs (inputs) or the
// pre-infer resource (state) wrote it.
func policyGroupProps(extra map[string]property.Value) property.Map {
	m := map[string]property.Value{
		gcName:             property.New(gcTestPolicyGroup),
		gcOrganizationName: property.New(gcTestOrg),
	}
	for k, v := range extra {
		m[k] = v
	}
	return property.NewMap(m)
}

func TestPolicyGroupCheck(t *testing.T) {
	t.Parallel()
	server := newPolicyGroupServer(t, &PolicyGroupClientMock{})
	check := func(t *testing.T, inputs property.Map) p.CheckResponse {
		t.Helper()
		resp, err := server.Check(p.CheckRequest{Urn: policyGroupURN, Inputs: inputs})
		require.NoError(t, err)
		return resp
	}

	t.Run("applies defaults", func(t *testing.T) {
		resp := check(t, policyGroupProps(nil))
		require.Empty(t, resp.Failures)
		assert.Equal(t, gcStacks, resp.Inputs.Get(gcEntityType).AsString())
		assert.Equal(t, gcAudit, resp.Inputs.Get(gcMode).AsString())
	})

	t.Run("keeps explicit values", func(t *testing.T) {
		resp := check(t, policyGroupProps(map[string]property.Value{
			gcEntityType: property.New(gcAccounts),
			gcMode:       property.New(gcPreventative),
			gcAccounts:   property.New([]property.Value{property.New("account-1")}),
		}))
		require.Empty(t, resp.Failures)
		assert.Equal(t, gcAccounts, resp.Inputs.Get(gcEntityType).AsString())
		assert.Equal(t, gcPreventative, resp.Inputs.Get(gcMode).AsString())
		assert.Equal(t, "account-1", resp.Inputs.Get(gcAccounts).AsArray().Get(0).AsString())
	})

	t.Run("strips policy pack version", func(t *testing.T) {
		resp := check(t, policyGroupProps(map[string]property.Value{
			gcPolicyPacks: property.New([]property.Value{property.New(map[string]property.Value{
				gcName:              property.New(gcPolicyPack1),
				gcVersionTag:        property.New(gcVersion100),
				gcPolicyPackVersion: property.New(3.0),
			})}),
		}))
		require.Empty(t, resp.Failures)
		pack := resp.Inputs.Get(gcPolicyPacks).AsArray().Get(0).AsMap()
		_, hasVersion := pack.GetOk(gcPolicyPackVersion)
		assert.False(t, hasVersion, "version should be stripped from inputs")
		assert.Equal(t, gcPolicyPack1, pack.Get(gcName).AsString())
		assert.Equal(t, gcVersion100, pack.Get(gcVersionTag).AsString())
	})

	t.Run("rejects invalid entityType", func(t *testing.T) {
		resp := check(t, policyGroupProps(map[string]property.Value{gcEntityType: property.New("invalid")}))
		require.Len(t, resp.Failures, 1)
		assert.Equal(t, gcEntityType, resp.Failures[0].Property)
		assert.Contains(t, resp.Failures[0].Reason, "entityType must be either 'stacks' or 'accounts'")
	})

	t.Run("rejects invalid mode", func(t *testing.T) {
		resp := check(t, policyGroupProps(map[string]property.Value{gcMode: property.New("invalid")}))
		require.Len(t, resp.Failures, 1)
		assert.Equal(t, gcMode, resp.Failures[0].Property)
		assert.Contains(t, resp.Failures[0].Reason, "mode must be either 'audit' or 'preventative'")
	})

	t.Run("doesn't validate unknown values", func(t *testing.T) {
		resp := check(t, policyGroupProps(map[string]property.Value{
			gcEntityType: property.New(property.Computed),
			gcMode:       property.New(property.Computed),
		}))
		assert.Empty(t, resp.Failures)
	})

	t.Run("requires name and organizationName", func(t *testing.T) {
		resp := check(t, property.NewMap(nil))
		props := make([]string, 0, len(resp.Failures))
		for _, f := range resp.Failures {
			props = append(props, f.Property)
		}
		assert.ElementsMatch(t, []string{gcName, gcOrganizationName}, props)
	})
}

func TestPolicyGroupDiff(t *testing.T) {
	t.Parallel()
	stack1 := stackRef{name: gcStack1, project: gcProject1}
	stack2 := stackRef{name: gcStack2, project: gcProject2}
	pp1 := policyPackRef{name: gcPolicyPack1, version: 1, versionTag: gcVersion100}
	pp2 := policyPackRef{name: gcPolicyPack2, version: 2, versionTag: gcVersion200}

	diff := func(t *testing.T, state PolicyGroupState, inputs PolicyGroupInput) p.DiffResponse {
		t.Helper()
		resp, err := (&PolicyGroup{}).Diff(context.Background(),
			infer.DiffRequest[PolicyGroupInput, PolicyGroupState]{ID: testPolicyGroupID, State: state, Inputs: inputs})
		require.NoError(t, err)
		return resp
	}

	t.Run("no diff when unchanged", func(t *testing.T) {
		b := newPolicyGroup().withStacks(stack1, stack2).withPolicyPacks(pp1, pp2)
		resp := diff(t, b.state(), b.input())
		assert.False(t, resp.HasChanges)
		assert.Empty(t, resp.DetailedDiff)
	})

	t.Run("lists are compared without regard to order", func(t *testing.T) {
		resp := diff(t,
			newPolicyGroup().withStacks(stack1, stack2).withPolicyPacks(pp1, pp2).state(),
			newPolicyGroup().withStacks(stack2, stack1).withPolicyPacks(pp2, pp1).input())
		assert.False(t, resp.HasChanges)

		resp = diff(t,
			newPolicyGroup().withAccounts("account-b", "account-a").state(),
			newPolicyGroup().withAccounts("account-a", "account-b").input())
		assert.False(t, resp.HasChanges)
	})

	t.Run("nil and empty lists are equal", func(t *testing.T) {
		st := newPolicyGroup().state()
		st.Stacks, st.Accounts, st.PolicyPacks = []PolicyGroupStackReference{}, []string{}, []PolicyGroupPolicyPackReference{}
		resp := diff(t, st, newPolicyGroup().input())
		assert.False(t, resp.HasChanges)
	})

	for _, field := range []string{gcName, gcOrganizationName, gcEntityType, gcMode} {
		t.Run(field+" change replaces", func(t *testing.T) {
			in := newPolicyGroup().input()
			switch field {
			case gcName:
				in.Name = gcOther
			case gcOrganizationName:
				in.OrganizationName = gcOther
			case gcEntityType:
				in.EntityType = gcAccounts
			case gcMode:
				in.Mode = gcPreventative
			}
			resp := diff(t, newPolicyGroup().state(), in)
			assert.True(t, resp.HasChanges)
			assert.Equal(t, map[string]p.PropertyDiff{field: {Kind: p.UpdateReplace, InputDiff: true}}, resp.DetailedDiff)
			assert.True(t, resp.DeleteBeforeReplace)
		})
	}

	t.Run("stack change updates in place", func(t *testing.T) {
		resp := diff(t, newPolicyGroup().withStacks(stack1).state(), newPolicyGroup().withStacks(stack2).input())
		assert.Equal(t, map[string]p.PropertyDiff{gcStacks: {Kind: p.Update, InputDiff: true}}, resp.DetailedDiff)
	})

	t.Run("policy pack version change updates in place", func(t *testing.T) {
		newer := pp1
		newer.versionTag = gcVersion200
		resp := diff(t, newPolicyGroup().withPolicyPacks(pp1).state(), newPolicyGroup().withPolicyPacks(newer).input())
		assert.Equal(t, map[string]p.PropertyDiff{gcPolicyPacks: {Kind: p.Update, InputDiff: true}}, resp.DetailedDiff)
	})

	t.Run("policy pack config change updates in place", func(t *testing.T) {
		configured := pp1
		configured.config = map[string]any{gcAllPolicies: "mandatory"}
		resp := diff(t, newPolicyGroup().withPolicyPacks(pp1).state(),
			newPolicyGroup().withPolicyPacks(configured).input())
		assert.Equal(t, map[string]p.PropertyDiff{gcPolicyPacks: {Kind: p.Update, InputDiff: true}}, resp.DetailedDiff)
	})

	t.Run("nil and empty policy pack config are equal", func(t *testing.T) {
		empty := pp1
		empty.config = map[string]any{}
		resp := diff(t, newPolicyGroup().withPolicyPacks(pp1).state(), newPolicyGroup().withPolicyPacks(empty).input())
		assert.False(t, resp.HasChanges)
		resp = diff(t, newPolicyGroup().withPolicyPacks(empty).state(), newPolicyGroup().withPolicyPacks(pp1).input())
		assert.False(t, resp.HasChanges)
	})

	t.Run("policy pack without versionTag matches any version", func(t *testing.T) {
		untagged := pp1
		untagged.versionTag = ""
		resp := diff(t, newPolicyGroup().withPolicyPacks(pp1).state(), newPolicyGroup().withPolicyPacks(untagged).input())
		assert.False(t, resp.HasChanges)
	})

	t.Run("child account isn't a diff while its parent is in the inputs", func(t *testing.T) {
		resp := diff(t, newPolicyGroup().withAccounts(gcParent, gcParentChild).state(),
			newPolicyGroup().withAccounts(gcParent).input())
		assert.False(t, resp.HasChanges)
	})

	t.Run("account change updates in place", func(t *testing.T) {
		resp := diff(t, newPolicyGroup().withAccounts(gcParent, gcParentChild).state(),
			newPolicyGroup().withAccounts(gcOther).input())
		assert.Equal(t, map[string]p.PropertyDiff{gcAccounts: {Kind: p.Update, InputDiff: true}}, resp.DetailedDiff)
	})
}

// TestPolicyGroupDiffLegacyState checks that state written by the pre-infer
// resource, with old inputs that still set `version`, upgrades without a diff.
func TestPolicyGroupDiffLegacyState(t *testing.T) {
	t.Parallel()
	server := newPolicyGroupServer(t, &PolicyGroupClientMock{})

	pack := func(withVersion bool) property.Value {
		m := map[string]property.Value{
			gcName:        property.New("hitrust-aws"),
			gcDisplayName: property.New(""),
			gcVersionTag:  property.New("2.4.1"),
			"config": property.New(map[string]property.Value{
				gcAllPolicies: property.New(map[string]property.Value{"enforcementLevel": property.New("mandatory")}),
			}),
		}
		if withVersion {
			m[gcPolicyPackVersion] = property.New(7.0)
		}
		return property.New(m)
	}
	stacks := property.New([]property.Value{property.New(map[string]property.Value{
		gcName: property.New("dev"), "routingProject": property.New("proj"),
	})})
	state := policyGroupProps(map[string]property.Value{
		gcEntityType:  property.New(gcStacks),
		gcMode:        property.New(gcAudit),
		gcStacks:      stacks,
		gcAccounts:    property.New([]property.Value{}),
		gcPolicyPacks: property.New([]property.Value{pack(true)}),
	})

	checked, err := server.Check(p.CheckRequest{Urn: policyGroupURN, Inputs: policyGroupProps(
		map[string]property.Value{
			gcEntityType:  property.New(gcStacks),
			gcMode:        property.New(gcAudit),
			gcStacks:      stacks,
			gcPolicyPacks: property.New([]property.Value{pack(true)}),
		})})
	require.NoError(t, err)
	require.Empty(t, checked.Failures)

	resp, err := server.Diff(p.DiffRequest{
		ID: testPolicyGroupID, Urn: policyGroupURN, State: state, Inputs: checked.Inputs,
	})
	require.NoError(t, err)
	assert.False(t, resp.HasChanges)
	assert.Empty(t, resp.DetailedDiff)
}

func TestPolicyGroupOps(t *testing.T) {
	t.Parallel()
	pp1 := policyPackRef{name: gcPolicyPack1, version: 1, versionTag: gcVersion100}

	t.Run("version change removes then adds", func(t *testing.T) {
		newer := pp1
		newer.versionTag = gcVersion200
		ops := policyGroupOps(newPolicyGroup().withPolicyPacks(pp1).state(), newPolicyGroup().withPolicyPacks(newer).input())
		require.Len(t, ops.policyPacks, 2)
		require.NotNil(t, ops.policyPacks[0].RemovePolicyPack)
		assert.Equal(t, gcVersion100, ops.policyPacks[0].RemovePolicyPack.VersionTag)
		assert.Equal(t, 1, ops.policyPacks[0].RemovePolicyPack.Version)
		require.NotNil(t, ops.policyPacks[1].AddPolicyPack)
		assert.Equal(t, gcVersion200, ops.policyPacks[1].AddPolicyPack.VersionTag)
	})

	t.Run("config change removes then adds", func(t *testing.T) {
		configured := pp1
		configured.config = map[string]any{gcAllPolicies: "mandatory"}
		ops := policyGroupOps(newPolicyGroup().withPolicyPacks(pp1).state(),
			newPolicyGroup().withPolicyPacks(configured).input())
		require.Len(t, ops.policyPacks, 2)
		require.NotNil(t, ops.policyPacks[0].RemovePolicyPack)
		require.NotNil(t, ops.policyPacks[1].AddPolicyPack)
		assert.Equal(t, configured.config, ops.policyPacks[1].AddPolicyPack.Config)
	})

	t.Run("exact versionTag wins over an untagged input", func(t *testing.T) {
		newer := pp1
		newer.versionTag = gcVersion200
		untagged := pp1
		untagged.versionTag = ""
		// State has 1.0.0 and 2.0.0; inputs ask for "any" and 1.0.0. The
		// untagged input must claim 2.0.0, not 1.0.0.
		ops := policyGroupOps(newPolicyGroup().withPolicyPacks(pp1, newer).state(),
			newPolicyGroup().withPolicyPacks(untagged, pp1).input())
		assert.Empty(t, ops.policyPacks)
	})
}

func TestHasParentAccount(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		account  string
		accounts []string
		expected bool
	}{
		{"child account with parent in list", gcParentChild, []string{gcParent, gcOther}, true},
		{"child account without parent in list", gcParentChild, []string{gcOther, "another"}, false},
		{"parent account (no parent exists)", gcParent, []string{gcOther, "another"}, false},
		{"empty accounts list", gcParentChild, []string{}, false},
		{"similar prefix but not parent", "parent-other/child", []string{gcParent, gcOther}, false},
		{"exact match is not a parent", gcParent, []string{gcParent}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, hasParentAccount(tt.account, tt.accounts))
		})
	}
}

func TestSameStrings(t *testing.T) {
	t.Parallel()
	assert.True(t, sameStrings(nil, []string{}))
	assert.True(t, sameStrings([]string{"a", "b"}, []string{"b", "a"}))
	assert.False(t, sameStrings([]string{"a", "b"}, []string{"a"}))
	assert.False(t, sameStrings([]string{"a", "a"}, []string{"a", "b"}))
}

func TestPolicyGroupRead(t *testing.T) {
	t.Parallel()
	parentAccount := "parent-account"
	childAccount := "parent-account/child"

	read := func(
		t *testing.T, client *PolicyGroupClientMock, inputs PolicyGroupInput, state PolicyGroupState,
	) infer.ReadResponse[PolicyGroupInput, PolicyGroupState] {
		t.Helper()
		resp, err := (&PolicyGroup{}).Read(mockPolicyGroupCtx(client),
			infer.ReadRequest[PolicyGroupInput, PolicyGroupState]{ID: testPolicyGroupID, Inputs: inputs, State: state})
		require.NoError(t, err)
		return resp
	}

	t.Run("preserves input accounts when API state matches previous state", func(t *testing.T) {
		// The user set the parent, and the Cloud added its child.
		client := &PolicyGroupClientMock{
			getPolicyGroupFunc: returnPolicyGroup(newPolicyGroup().withAccounts(parentAccount, childAccount).api()),
		}
		resp := read(t, client, newPolicyGroup().withAccounts(parentAccount).input(),
			newPolicyGroup().withAccounts(childAccount, parentAccount).state())
		assert.Equal(t, testPolicyGroupID, resp.ID)
		assert.Equal(t, []string{parentAccount, childAccount}, resp.State.Accounts)
		assert.Equal(t, []string{parentAccount}, resp.Inputs.Accounts)
	})

	t.Run("updates input accounts when API state changed externally", func(t *testing.T) {
		client := &PolicyGroupClientMock{
			getPolicyGroupFunc: returnPolicyGroup(newPolicyGroup().withAccounts("a", "b", "c").api()),
		}
		resp := read(t, client, newPolicyGroup().withAccounts("a", "b").input(),
			newPolicyGroup().withAccounts("a", "b").state())
		assert.Equal(t, []string{"a", "b", "c"}, resp.Inputs.Accounts)
	})

	t.Run("import reads everything from the API", func(t *testing.T) {
		pp := policyPackRef{name: gcPolicyPack1, version: 4, versionTag: gcVersion100, config: map[string]any{"k": "v"}}
		stack := stackRef{name: gcStack1, project: gcProject1}
		client := &PolicyGroupClientMock{
			getPolicyGroupFunc: returnPolicyGroup(newPolicyGroup().withStacks(stack).withPolicyPacks(pp).api()),
		}
		resp := read(t, client, PolicyGroupInput{}, PolicyGroupState{})
		want := newPolicyGroup().withStacks(stack).withPolicyPacks(pp)
		assert.Equal(t, want.state(), resp.State)
		assert.Equal(t, want.input(), resp.Inputs)
		assert.InDelta(t, 4.0, resp.State.PolicyPacks[0].Version, 0)
	})

	t.Run("returns empty response when not found", func(t *testing.T) {
		resp := read(t, &PolicyGroupClientMock{}, PolicyGroupInput{}, PolicyGroupState{})
		assert.Empty(t, resp.ID)
	})

	t.Run("returns API errors", func(t *testing.T) {
		client := &PolicyGroupClientMock{getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
			return nil, errors.New("boom")
		}}
		_, err := (&PolicyGroup{}).Read(mockPolicyGroupCtx(client),
			infer.ReadRequest[PolicyGroupInput, PolicyGroupState]{ID: testPolicyGroupID})
		assert.ErrorContains(t, err, "boom")
	})

	t.Run("rejects a malformed ID", func(t *testing.T) {
		_, err := (&PolicyGroup{}).Read(mockPolicyGroupCtx(&PolicyGroupClientMock{}),
			infer.ReadRequest[PolicyGroupInput, PolicyGroupState]{ID: "no-slash"})
		assert.Error(t, err)
	})
}

func TestPolicyGroupCreate(t *testing.T) {
	t.Parallel()
	stack1 := stackRef{name: gcStack1, project: gcProject1}
	stack2 := stackRef{name: gcStack2, project: gcProject2}
	pp1 := policyPackRef{name: gcPolicyPack1, version: 1, versionTag: gcVersion100}

	create := func(
		client *PolicyGroupClientMock, in PolicyGroupInput, dryRun bool,
	) (infer.CreateResponse[PolicyGroupState], error) {
		return (&PolicyGroup{}).Create(mockPolicyGroupCtx(client),
			infer.CreateRequest[PolicyGroupInput]{Name: "testPolicyGroup", Inputs: in, DryRun: dryRun})
	}

	t.Run("creates empty policy group", func(t *testing.T) {
		var created bool
		client := &PolicyGroupClientMock{
			createPolicyGroupFunc: func(_ context.Context, orgName, name, entityType, mode string) error {
				created = true
				assert.Equal(t, gcTestOrg, orgName)
				assert.Equal(t, gcTestPolicyGroup, name)
				assert.Equal(t, gcStacks, entityType)
				assert.Equal(t, gcAudit, mode)
				return nil
			},
			getPolicyGroupFunc: returnPolicyGroup(newPolicyGroup().api()),
		}
		resp, err := create(client, newPolicyGroup().input(), false)
		require.NoError(t, err)
		assert.True(t, created)
		assert.Empty(t, client.batches, "no batch for an empty policy group")
		assert.Equal(t, testPolicyGroupID, resp.ID)
	})

	t.Run("adds stacks, policy packs and accounts in one batch", func(t *testing.T) {
		b := newPolicyGroup().withStacks(stack1, stack2).withPolicyPacks(pp1)
		in := b.input()
		in.Accounts = []string{"account-1"}
		client := &PolicyGroupClientMock{getPolicyGroupFunc: returnPolicyGroup(b.api())}
		_, err := create(client, in, false)
		require.NoError(t, err)
		require.Len(t, client.batches, 1)
		ops := client.batches[0]
		require.Len(t, ops, 4)
		assert.Equal(t, gcStack1, ops[0].AddStack.Name)
		assert.Equal(t, gcStack2, ops[1].AddStack.Name)
		assert.Equal(t, gcPolicyPack1, ops[2].AddPolicyPack.Name)
		assert.Equal(t, gcVersion100, ops[2].AddPolicyPack.VersionTag)
		assert.Equal(t, "account-1", ops[3].AddInsightsAccount.Name)
	})

	t.Run("returns state from API response", func(t *testing.T) {
		// The API adds a child account and resolves the pack version.
		api := newPolicyGroup().withAccounts(gcParent, gcParentChild).withPolicyPacks(pp1).api()
		client := &PolicyGroupClientMock{getPolicyGroupFunc: returnPolicyGroup(api)}
		resp, err := create(client, newPolicyGroup().withAccounts(gcParent).input(), false)
		require.NoError(t, err)
		assert.Equal(t, []string{gcParent, gcParentChild}, resp.Output.Accounts)
		require.Len(t, resp.Output.PolicyPacks, 1)
		assert.InDelta(t, 1.0, resp.Output.PolicyPacks[0].Version, 0)
	})

	t.Run("doesn't call the API on preview", func(t *testing.T) {
		client := &PolicyGroupClientMock{createPolicyGroupFunc: func(context.Context, string, string, string, string) error {
			t.Fatal("CreatePolicyGroup called on preview")
			return nil
		}}
		resp, err := create(client, newPolicyGroup().withStacks(stack1).input(), true)
		require.NoError(t, err)
		assert.Equal(t, testPolicyGroupID, resp.ID)
		assert.Equal(t, newPolicyGroup().withStacks(stack1).state(), resp.Output)
	})

	t.Run("returns error from CreatePolicyGroup", func(t *testing.T) {
		client := &PolicyGroupClientMock{createPolicyGroupFunc: func(context.Context, string, string, string, string) error {
			return errors.New("create failed")
		}}
		_, err := create(client, newPolicyGroup().input(), false)
		assert.ErrorContains(t, err, "create failed")
		assert.False(t, errors.As(err, &infer.ResourceInitFailedError{}))
	})

	t.Run("keeps the policy group in state when BatchUpdatePolicyGroup fails", func(t *testing.T) {
		client := &PolicyGroupClientMock{
			batchUpdatePolicyGroupFunc: func(context.Context, string, string, []pulumiapi.UpdatePolicyGroupRequest) error {
				return errors.New("batch failed")
			},
		}
		resp, err := create(client, newPolicyGroup().withStacks(stack1).input(), false)
		var initErr infer.ResourceInitFailedError
		require.ErrorAs(t, err, &initErr)
		assert.Contains(t, initErr.Reasons[0], "batch failed")
		assert.Equal(t, testPolicyGroupID, resp.ID)
		assert.Equal(t, newPolicyGroup().withStacks(stack1).state(), resp.Output)
	})

	t.Run("keeps the policy group in state when the read back fails", func(t *testing.T) {
		client := &PolicyGroupClientMock{getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
			return nil, errors.New("get failed")
		}}
		resp, err := create(client, newPolicyGroup().input(), false)
		var initErr infer.ResourceInitFailedError
		require.ErrorAs(t, err, &initErr)
		assert.Contains(t, initErr.Reasons[0], "get failed")
		assert.Equal(t, testPolicyGroupID, resp.ID)
	})

	t.Run("keeps the policy group in state when the read back finds nothing", func(t *testing.T) {
		_, err := create(&PolicyGroupClientMock{}, newPolicyGroup().input(), false)
		require.ErrorAs(t, err, &infer.ResourceInitFailedError{})
	})
}

func TestPolicyGroupUpdate(t *testing.T) {
	t.Parallel()
	stack1 := stackRef{name: gcStack1, project: gcProject1}
	stack2 := stackRef{name: gcStack2, project: gcProject2}
	pp1 := policyPackRef{name: gcPolicyPack1, version: 1, versionTag: gcVersion100}
	pp2 := policyPackRef{name: gcPolicyPack2, version: 1, versionTag: gcVersion100}
	parentAccount := "parent-account"
	childAccount := "parent-account/child"

	update := func(
		client *PolicyGroupClientMock, state PolicyGroupState, in PolicyGroupInput,
	) (infer.UpdateResponse[PolicyGroupState], error) {
		return (&PolicyGroup{}).Update(mockPolicyGroupCtx(client),
			infer.UpdateRequest[PolicyGroupInput, PolicyGroupState]{ID: testPolicyGroupID, State: state, Inputs: in})
	}

	t.Run("adds new stacks", func(t *testing.T) {
		client := &PolicyGroupClientMock{getPolicyGroupFunc: returnPolicyGroup(newPolicyGroup().api())}
		_, err := update(client, newPolicyGroup().withStacks(stack1).state(),
			newPolicyGroup().withStacks(stack1, stack2).input())
		require.NoError(t, err)
		ops := client.ops()
		require.Len(t, ops, 1)
		require.NotNil(t, ops[0].AddStack)
		assert.Equal(t, gcStack2, ops[0].AddStack.Name)
	})

	t.Run("removes stacks", func(t *testing.T) {
		client := &PolicyGroupClientMock{getPolicyGroupFunc: returnPolicyGroup(newPolicyGroup().api())}
		_, err := update(client, newPolicyGroup().withStacks(stack1, stack2).state(),
			newPolicyGroup().withStacks(stack1).input())
		require.NoError(t, err)
		ops := client.ops()
		require.Len(t, ops, 1)
		require.NotNil(t, ops[0].RemoveStack)
		assert.Equal(t, gcStack2, ops[0].RemoveStack.Name)
	})

	t.Run("adds and removes policy packs, one op per request", func(t *testing.T) {
		client := &PolicyGroupClientMock{getPolicyGroupFunc: returnPolicyGroup(newPolicyGroup().api())}
		_, err := update(client, newPolicyGroup().withPolicyPacks(pp1).state(),
			newPolicyGroup().withPolicyPacks(pp2).input())
		require.NoError(t, err)
		require.Len(t, client.batches, 2)
		for _, b := range client.batches {
			assert.Len(t, b, 1)
		}
		require.NotNil(t, client.batches[0][0].RemovePolicyPack)
		assert.Equal(t, gcPolicyPack1, client.batches[0][0].RemovePolicyPack.Name)
		require.NotNil(t, client.batches[1][0].AddPolicyPack)
		assert.Equal(t, gcPolicyPack2, client.batches[1][0].AddPolicyPack.Name)
	})

	t.Run("adds accounts", func(t *testing.T) {
		client := &PolicyGroupClientMock{getPolicyGroupFunc: returnPolicyGroup(newPolicyGroup().api())}
		_, err := update(client, newPolicyGroup().withAccounts("account-1").state(),
			newPolicyGroup().withAccounts("account-1", "account-2").input())
		require.NoError(t, err)
		ops := client.ops()
		require.Len(t, ops, 1)
		require.NotNil(t, ops[0].AddInsightsAccount)
		assert.Equal(t, "account-2", ops[0].AddInsightsAccount.Name)
	})

	t.Run("skips child accounts while their parent is still present", func(t *testing.T) {
		client := &PolicyGroupClientMock{
			getPolicyGroupFunc: returnPolicyGroup(newPolicyGroup().withAccounts(parentAccount, childAccount).api()),
		}
		_, err := update(client, newPolicyGroup().withAccounts(parentAccount, childAccount).state(),
			newPolicyGroup().withAccounts(parentAccount).input())
		require.NoError(t, err)
		assert.Empty(t, client.batches)
	})

	t.Run("removes child account when parent is also removed", func(t *testing.T) {
		client := &PolicyGroupClientMock{getPolicyGroupFunc: returnPolicyGroup(newPolicyGroup().withAccounts().api())}
		_, err := update(client, newPolicyGroup().withAccounts(parentAccount, childAccount).state(),
			newPolicyGroup().withAccounts().input())
		require.NoError(t, err)
		var removed []string
		for _, op := range client.ops() {
			require.NotNil(t, op.RemoveInsightsAccount)
			removed = append(removed, op.RemoveInsightsAccount.Name)
		}
		assert.ElementsMatch(t, []string{parentAccount, childAccount}, removed)
	})

	t.Run("no changes when inputs match state", func(t *testing.T) {
		b := newPolicyGroup().withStacks(stack1).withPolicyPacks(pp1)
		client := &PolicyGroupClientMock{getPolicyGroupFunc: returnPolicyGroup(b.api())}
		_, err := update(client, b.state(), b.input())
		require.NoError(t, err)
		assert.Empty(t, client.batches)
	})

	t.Run("doesn't call the API on preview", func(t *testing.T) {
		client := &PolicyGroupClientMock{}
		resp, err := (&PolicyGroup{}).Update(mockPolicyGroupCtx(client),
			infer.UpdateRequest[PolicyGroupInput, PolicyGroupState]{
				ID: testPolicyGroupID, State: newPolicyGroup().state(),
				Inputs: newPolicyGroup().withStacks(stack1).input(), DryRun: true,
			})
		require.NoError(t, err)
		assert.Empty(t, client.batches)
		assert.Equal(t, newPolicyGroup().withStacks(stack1).state(), resp.Output)
	})

	t.Run("returns error from GetPolicyGroup after update", func(t *testing.T) {
		client := &PolicyGroupClientMock{getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) {
			return nil, errors.New("get failed")
		}}
		_, err := update(client, newPolicyGroup().state(), newPolicyGroup().withStacks(stack1).input())
		assert.ErrorContains(t, err, "get failed")
	})

	t.Run("returns updated state from API response", func(t *testing.T) {
		api := newPolicyGroup().withAccounts(parentAccount, childAccount).api()
		client := &PolicyGroupClientMock{getPolicyGroupFunc: returnPolicyGroup(api)}
		resp, err := update(client, newPolicyGroup().withAccounts().state(),
			newPolicyGroup().withAccounts(parentAccount).input())
		require.NoError(t, err)
		assert.Equal(t, []string{parentAccount, childAccount}, resp.Output.Accounts)
		assert.Equal(t, gcTestOrg, resp.Output.OrganizationName)
	})

	t.Run("partial failure returns the actual state", func(t *testing.T) {
		// Removing pp1 succeeds, adding pp2 fails.
		var calls int
		actual := newPolicyGroup().withStacks(stack1).api()
		client := &PolicyGroupClientMock{
			batchUpdatePolicyGroupFunc: func(context.Context, string, string, []pulumiapi.UpdatePolicyGroupRequest) error {
				calls++
				if calls == 1 {
					return nil
				}
				return errors.New("backend reorder/upsert bug fired")
			},
			getPolicyGroupFunc: returnPolicyGroup(actual),
		}
		resp, err := update(client, newPolicyGroup().withStacks(stack1).withPolicyPacks(pp1).state(),
			newPolicyGroup().withStacks(stack1).withPolicyPacks(pp2).input())
		var initErr infer.ResourceInitFailedError
		require.ErrorAs(t, err, &initErr)
		assert.Contains(t, initErr.Reasons[0], "failed to update policy group")
		assert.Equal(t, 2, calls, "stops at the failed op")
		assert.Equal(t, newPolicyGroup().withStacks(stack1).state(), resp.Output)
	})

	t.Run("partial failure falls back to inputs when the read fails", func(t *testing.T) {
		client := &PolicyGroupClientMock{
			batchUpdatePolicyGroupFunc: func(context.Context, string, string, []pulumiapi.UpdatePolicyGroupRequest) error {
				return errors.New("batch failed")
			},
			getPolicyGroupFunc: func() (*pulumiapi.PolicyGroup, error) { return nil, errors.New("get failed") },
		}
		in := newPolicyGroup().withStacks(stack1).input()
		resp, err := update(client, newPolicyGroup().state(), in)
		require.ErrorAs(t, err, &infer.ResourceInitFailedError{})
		assert.Equal(t, newPolicyGroup().withStacks(stack1).state(), resp.Output)
	})
}

func TestPolicyGroupDelete(t *testing.T) {
	t.Parallel()
	t.Run("deletes by ID", func(t *testing.T) {
		var gotOrg, gotName string
		client := &PolicyGroupClientMock{deletePolicyGroupFunc: func(_ context.Context, org, name string) error {
			gotOrg, gotName = org, name
			return nil
		}}
		_, err := (&PolicyGroup{}).Delete(mockPolicyGroupCtx(client),
			infer.DeleteRequest[PolicyGroupState]{ID: testPolicyGroupID, State: newPolicyGroup().state()})
		require.NoError(t, err)
		assert.Equal(t, gcTestOrg, gotOrg)
		assert.Equal(t, gcTestPolicyGroup, gotName)
	})

	t.Run("returns API errors", func(t *testing.T) {
		client := &PolicyGroupClientMock{deletePolicyGroupFunc: func(context.Context, string, string) error {
			return errors.New("delete failed")
		}}
		_, err := (&PolicyGroup{}).Delete(mockPolicyGroupCtx(client),
			infer.DeleteRequest[PolicyGroupState]{ID: testPolicyGroupID})
		assert.ErrorContains(t, err, "delete failed")
	})
}

// TestPolicyGroupConfigRoundTrip checks that a nested policy pack config
// survives the round trip through the server's encoding.
func TestPolicyGroupConfigRoundTrip(t *testing.T) {
	t.Parallel()
	cfg := map[string]any{
		"approvedAmiIds": []any{"ami-123", "ami-456"},
		"maxInstances":   float64(10),
		"nestedConfig":   map[string]any{"regions": []any{"us-east-1", "us-west-2"}, "enabled": true},
	}
	pp := policyPackRef{name: "aws-compliance", version: 1, versionTag: "v1.0.0", config: cfg}
	b := newPolicyGroup().withStacks(stackRef{name: "stack1", project: "project1"}).withPolicyPacks(pp)
	client := &PolicyGroupClientMock{getPolicyGroupFunc: returnPolicyGroup(b.api())}
	server := newPolicyGroupServer(t, client)

	resp, err := server.Read(p.ReadRequest{ID: testPolicyGroupID, Urn: policyGroupURN})
	require.NoError(t, err)

	// Diffing the read state against the read inputs shows no changes.
	diff, err := server.Diff(p.DiffRequest{
		ID: testPolicyGroupID, Urn: policyGroupURN, State: resp.Properties, Inputs: resp.Inputs,
	})
	require.NoError(t, err)
	assert.False(t, diff.HasChanges)

	packs := resp.Properties.Get(gcPolicyPacks).AsArray()
	require.Equal(t, 1, packs.Len())
	got := packs.Get(0).AsMap()
	assert.InDelta(t, 1.0, got.Get(gcPolicyPackVersion).AsNumber(), 0)
	assert.Equal(t, "ami-456",
		got.Get("config").AsMap().Get("approvedAmiIds").AsArray().Get(1).AsString())
}
