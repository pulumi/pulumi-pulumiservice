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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const (
	testRbacRoleID   = "role-1"
	testRbacPolicyID = "policy-1"
	testRoleName     = "Platform"
	gcAll            = "all"
	gcPermissionSets = "permissionSetIds"
)

type rbacRoleClientMock struct {
	config.Client
	descriptors map[string]*apitype.PermissionDescriptorRecord
	sets        []pulumiapi.PermissionSet

	created *apitype.PermissionDescriptorBase
	updates map[string]apitype.UpdateRoleRequest
	deleted []string
}

func newRbacRoleClientMock(t *testing.T, policyWire string) *rbacRoleClientMock {
	t.Helper()
	return &rbacRoleClientMock{
		descriptors: map[string]*apitype.PermissionDescriptorRecord{
			testRbacRoleID: {
				PermissionDescriptorBase: apitype.PermissionDescriptorBase{
					Name:      testRoleName,
					UxPurpose: apitype.PermissionDescriptorUXPurposeRole,
					Details: mustParseDescriptor(t,
						`{"__type":"PermissionDescriptorCompose","permissionDescriptors":["policy-1"]}`),
				},
				ID: testRbacRoleID,
			},
			testRbacPolicyID: {
				PermissionDescriptorBase: apitype.PermissionDescriptorBase{
					Name:      testRoleName,
					UxPurpose: apitype.PermissionDescriptorUXPurposePolicy,
					Details:   mustParseDescriptor(t, policyWire),
				},
				ID:      testRbacPolicyID,
				Version: 4,
			},
		},
		sets: []pulumiapi.PermissionSet{
			{ID: testOrgSetID, ResourceType: "global"},
			{ID: testStackSetID, ResourceType: string(RbacResourceTypeStack)},
			{ID: testEnvSetID, ResourceType: string(RbacResourceTypeEnvironment)},
		},
		updates: map[string]apitype.UpdateRoleRequest{},
	}
}

func (m *rbacRoleClientMock) CreateRoleWithPolicy(
	_ context.Context, _ string, policy apitype.PermissionDescriptorBase,
) (*apitype.PermissionDescriptorRecord, error) {
	m.created = &policy
	return m.descriptors[testRbacRoleID], nil
}

func (m *rbacRoleClientMock) GetRole(_ context.Context, _, id string) (*apitype.PermissionDescriptorRecord, error) {
	return m.descriptors[id], nil
}

func (m *rbacRoleClientMock) UpdateRole(
	_ context.Context, _, id string, req apitype.UpdateRoleRequest,
) (*apitype.PermissionDescriptorRecord, error) {
	m.updates[id] = req
	rec := *m.descriptors[id]
	rec.Version++
	return &rec, nil
}

func (m *rbacRoleClientMock) DeleteRole(_ context.Context, _, id string, _ bool) error {
	m.deleted = append(m.deleted, id)
	return nil
}

func (m *rbacRoleClientMock) ListPermissionSets(context.Context, string) ([]pulumiapi.PermissionSet, error) {
	return m.sets, nil
}

func (m *rbacRoleClientMock) GetPermissionSet(context.Context, string, string) (*pulumiapi.PermissionSet, error) {
	return nil, nil
}

func testRoleCore() RbacRoleCore {
	core := consoleRoleModel()
	core.OrganizationName = gcAcme
	core.Name = testRoleName
	return core
}

func TestRbacRoleCreate(t *testing.T) {
	t.Parallel()
	mock := newRbacRoleClientMock(t, consoleRole)
	ctx := config.WithMockClient(context.Background(), mock)

	resp, err := (&RbacRole{}).Create(ctx, infer.CreateRequest[RbacRoleInput]{
		Inputs: RbacRoleInput{RbacRoleCore: testRoleCore()},
	})
	require.NoError(t, err)
	assert.Equal(t, "acme/role-1", resp.ID)
	assert.Equal(t, testRbacRoleID, resp.Output.RoleId)
	assert.Equal(t, testRbacPolicyID, resp.Output.PolicyId)
	assert.Equal(t, 4, resp.Output.Version)

	require.NotNil(t, mock.created)
	assert.Equal(t, apitype.PermissionDescriptorUXPurposeRole, mock.created.UxPurpose)
	assert.Empty(t, mock.created.ResourceType, "console roles carry no resource type")
	got, err := json.Marshal(mock.created.Details)
	require.NoError(t, err)
	assert.JSONEq(t, consoleRole, string(got))
}

func TestRbacRoleUpdate(t *testing.T) {
	t.Parallel()
	prior := testRoleCore()

	t.Run("permissions only touch the policy", func(t *testing.T) {
		mock := newRbacRoleClientMock(t, consoleRole)
		next := testRoleCore()
		next.EntityRules = next.EntityRules[:1]
		resp, err := (&RbacRole{}).Update(config.WithMockClient(context.Background(), mock),
			infer.UpdateRequest[RbacRoleInput, RbacRoleState]{
				Inputs: RbacRoleInput{RbacRoleCore: next},
				State:  RbacRoleState{RbacRoleCore: prior, RoleId: testRbacRoleID, PolicyId: testRbacPolicyID},
			})
		require.NoError(t, err)
		assert.Contains(t, mock.updates, testRbacPolicyID)
		assert.NotContains(t, mock.updates, testRbacRoleID)
		assert.Equal(t, 5, resp.Output.Version)
	})

	t.Run("rename updates both descriptors", func(t *testing.T) {
		mock := newRbacRoleClientMock(t, consoleRole)
		next := testRoleCore()
		next.Name = "Platform Engineers"
		_, err := (&RbacRole{}).Update(config.WithMockClient(context.Background(), mock),
			infer.UpdateRequest[RbacRoleInput, RbacRoleState]{
				Inputs: RbacRoleInput{RbacRoleCore: next},
				State:  RbacRoleState{RbacRoleCore: prior, RoleId: testRbacRoleID, PolicyId: testRbacPolicyID},
			})
		require.NoError(t, err)
		require.Contains(t, mock.updates, testRbacRoleID)
		assert.Equal(t, "Platform Engineers", *mock.updates[testRbacRoleID].Name)
		assert.Nil(t, mock.updates[testRbacRoleID].Details, "the role's own details stay Compose[policy]")
		assert.Equal(t, "Platform Engineers", *mock.updates[testRbacPolicyID].Name)
	})
}

func TestRbacRoleDeleteOnlyDeletesRole(t *testing.T) {
	t.Parallel()
	mock := newRbacRoleClientMock(t, consoleRole)
	_, err := (&RbacRole{}).Delete(config.WithMockClient(context.Background(), mock),
		infer.DeleteRequest[RbacRoleState]{State: RbacRoleState{
			RbacRoleCore: testRoleCore(), RoleId: testRbacRoleID, PolicyId: testRbacPolicyID,
		}})
	require.NoError(t, err)
	assert.Equal(t, []string{testRbacRoleID}, mock.deleted)
}

func TestRbacRoleRead(t *testing.T) {
	t.Parallel()
	read := func(_ *testing.T, mock *rbacRoleClientMock, prior RbacRoleCore) (
		infer.ReadResponse[RbacRoleInput, RbacRoleState], error,
	) {
		return (&RbacRole{}).Read(config.WithMockClient(context.Background(), mock),
			infer.ReadRequest[RbacRoleInput, RbacRoleState]{ID: "acme/role-1", Inputs: RbacRoleInput{RbacRoleCore: prior}})
	}

	t.Run("keeps the user's layout when nothing changed", func(t *testing.T) {
		prior := testRoleCore()
		slicesReverse(prior.EntityRules)
		resp, err := read(t, newRbacRoleClientMock(t, consoleRole), prior)
		require.NoError(t, err)
		assert.Equal(t, prior.EntityRules, resp.Inputs.EntityRules)
		assert.Equal(t, testRbacPolicyID, resp.State.PolicyId)
		assert.Equal(t, 4, resp.State.Version)
	})

	t.Run("reports console edits", func(t *testing.T) {
		edited := `{"__type":"PermissionDescriptorGroup","entries":[
		  {"__type":"PermissionDescriptorCompose","permissionDescriptors":["org-set","env-set"]}]}`
		resp, err := read(t, newRbacRoleClientMock(t, edited), testRoleCore())
		require.NoError(t, err)
		assert.Equal(t, []string{testOrgSetID}, resp.Inputs.OrganizationPermissionSetIds)
		require.Len(t, resp.Inputs.EntityRules, 1)
		assert.Equal(t, pointerTo(true), resp.Inputs.EntityRules[0].Environment.All)
	})

	t.Run("import", func(t *testing.T) {
		resp, err := read(t, newRbacRoleClientMock(t, consoleRole), RbacRoleCore{})
		require.NoError(t, err)
		assert.Equal(t, gcAcme, resp.Inputs.OrganizationName)
		assert.Equal(t, testRoleName, resp.Inputs.Name)
		assert.Len(t, resp.Inputs.EntityRules, 4)
	})

	t.Run("unrepresentable role points at alternatives", func(t *testing.T) {
		mock := newRbacRoleClientMock(t, consoleRole)
		mock.descriptors[testRbacRoleID].Details = mustParseDescriptor(t,
			`{"__type":"PermissionDescriptorAllow","permissions":["stack:read"]}`)
		_, err := read(t, mock, RbacRoleCore{})
		assert.ErrorContains(t, err, "OrganizationRole")
	})
}

func TestRbacRoleCheck(t *testing.T) {
	t.Parallel()
	check := func(t *testing.T, rules property.Value) map[string]string {
		t.Helper()
		resp, err := (&RbacRole{}).Check(context.Background(), infer.CheckRequest{
			NewInputs: property.NewMap(map[string]property.Value{
				gcOrganizationName: property.New(gcAcme),
				gcName:             property.New(testRoleName),
				gcEntityRules:      rules,
			}),
		})
		require.NoError(t, err)
		failures := map[string]string{}
		for _, f := range resp.Failures {
			failures[f.Property] = f.Reason
		}
		return failures
	}
	rule := func(fields map[string]property.Value) property.Value {
		return property.New([]property.Value{property.New(fields)})
	}
	sets := property.New([]property.Value{property.New(testStackSetID)})

	t.Run("valid tag rule", func(t *testing.T) {
		assert.Empty(t, check(t, rule(map[string]property.Value{
			gcPermissionSets: sets,
			gcStack: property.New(map[string]property.Value{
				"tags": property.New([]property.Value{property.New(map[string]property.Value{
					"key": property.New("env"), "value": property.New("prod"), "operator": property.New("notEquals"),
				})}),
			}),
		})))
	})

	t.Run("two entity kinds", func(t *testing.T) {
		failures := check(t, rule(map[string]property.Value{
			gcPermissionSets: sets,
			"stack":          property.New(map[string]property.Value{gcAll: property.New(true)}),
			"environment":    property.New(map[string]property.Value{gcAll: property.New(true)}),
		}))
		assert.Contains(t, failures, "entityRules[0]")
	})

	t.Run("id and all together", func(t *testing.T) {
		failures := check(t, rule(map[string]property.Value{
			gcPermissionSets: sets,
			gcStack: property.New(map[string]property.Value{
				gcAll: property.New(true), "id": property.New(testStackUUID),
			}),
		}))
		assert.Contains(t, failures, "entityRules[0].stack")
	})

	t.Run("grants nothing", func(t *testing.T) {
		failures := check(t, property.New([]property.Value{}))
		assert.Contains(t, failures, gcOrganizationPermissionSetIds)
	})

	t.Run("unknown stack id at preview", func(t *testing.T) {
		assert.Empty(t, check(t, rule(map[string]property.Value{
			gcPermissionSets: property.New([]property.Value{property.New(property.Computed)}),
			"stack":          property.New(map[string]property.Value{"id": property.New(property.Computed)}),
		})))
	})
}
