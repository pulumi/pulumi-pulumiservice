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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const (
	testServicePrincipalLogin = "service-principal:d333a711-4aa0-402f-be6d-72af9665fc37"
	testServicePrincipalName  = "ci-deployer"
)

var testServicePrincipalID = servicePrincipalID(gcMyOrg, testServicePrincipalLogin)

type servicePrincipalClientMock struct {
	config.Client
	create      func(orgName, name string, roleIDs []string) (*pulumiapi.ServicePrincipal, error)
	get         func(orgName, login string) (*pulumiapi.ServicePrincipal, error)
	rename      func(orgName, login, name string) error
	updateRoles func(orgName, login string, roleIDs []string) ([]string, error)
	del         func(orgName, login string) error
}

func (c *servicePrincipalClientMock) CreateServicePrincipal(
	_ context.Context, orgName, name string, roleIDs []string,
) (*pulumiapi.ServicePrincipal, error) {
	return c.create(orgName, name, roleIDs)
}

func (c *servicePrincipalClientMock) GetServicePrincipal(
	_ context.Context, orgName, login string,
) (*pulumiapi.ServicePrincipal, error) {
	return c.get(orgName, login)
}

func (c *servicePrincipalClientMock) RenameServicePrincipal(_ context.Context, orgName, login, name string) error {
	return c.rename(orgName, login, name)
}

func (c *servicePrincipalClientMock) UpdateServicePrincipalRoles(
	_ context.Context, orgName, login string, roleIDs []string,
) ([]string, error) {
	return c.updateRoles(orgName, login, roleIDs)
}

func (c *servicePrincipalClientMock) DeleteServicePrincipal(_ context.Context, orgName, login string) error {
	return c.del(orgName, login)
}

func TestServicePrincipalCreate(t *testing.T) {
	t.Run("creates with roles and outputs the login", func(t *testing.T) {
		mock := &servicePrincipalClientMock{
			create: func(orgName, name string, roleIDs []string) (*pulumiapi.ServicePrincipal, error) {
				assert.Equal(t, gcMyOrg, orgName)
				assert.Equal(t, testServicePrincipalName, name)
				assert.Equal(t, []string{"role-1"}, roleIDs)
				return &pulumiapi.ServicePrincipal{Login: testServicePrincipalLogin, Name: name, RoleIDs: roleIDs}, nil
			},
		}
		ctx := config.WithMockClient(t.Context(), mock)
		in := ServicePrincipalInput{
			OrganizationName: gcMyOrg,
			Name:             testServicePrincipalName,
			RoleIDs:          []string{"role-1"},
		}
		resp, err := (&ServicePrincipal{}).Create(ctx, infer.CreateRequest[ServicePrincipalInput]{Inputs: in})
		require.NoError(t, err)
		assert.Equal(t, testServicePrincipalID, resp.ID)
		assert.Equal(t, ServicePrincipalState{ServicePrincipalInput: in, Login: testServicePrincipalLogin}, resp.Output)
	})

	t.Run("surfaces the service's error", func(t *testing.T) {
		mock := &servicePrincipalClientMock{
			create: func(string, string, []string) (*pulumiapi.ServicePrincipal, error) {
				return nil, errors.New("400 API error: a service principal holds at most one role")
			},
		}
		ctx := config.WithMockClient(t.Context(), mock)
		_, err := (&ServicePrincipal{}).Create(ctx, infer.CreateRequest[ServicePrincipalInput]{
			Inputs: ServicePrincipalInput{OrganizationName: gcMyOrg, Name: testServicePrincipalName},
		})
		assert.ErrorContains(t, err, "at most one role")
	})

	t.Run("preview makes no call", func(t *testing.T) {
		ctx := config.WithMockClient(t.Context(), &servicePrincipalClientMock{})
		resp, err := (&ServicePrincipal{}).Create(ctx, infer.CreateRequest[ServicePrincipalInput]{
			Inputs: ServicePrincipalInput{OrganizationName: gcMyOrg, Name: testServicePrincipalName},
			DryRun: true,
		})
		require.NoError(t, err)
		assert.Empty(t, resp.Output.Login)
	})
}

func TestServicePrincipalUpdate(t *testing.T) {
	state := ServicePrincipalState{
		ServicePrincipalInput: ServicePrincipalInput{
			OrganizationName: gcMyOrg,
			Name:             testServicePrincipalName,
			RoleIDs:          []string{"role-1", "role-2"},
		},
		Login: testServicePrincipalLogin,
	}

	run := func(t *testing.T, inputs ServicePrincipalInput) (renamed *string, roles *[]string) {
		t.Helper()
		mock := &servicePrincipalClientMock{
			rename: func(orgName, login, name string) error {
				assert.Equal(t, gcMyOrg, orgName)
				assert.Equal(t, testServicePrincipalLogin, login)
				renamed = &name
				return nil
			},
			updateRoles: func(orgName, login string, roleIDs []string) ([]string, error) {
				assert.Equal(t, gcMyOrg, orgName)
				assert.Equal(t, testServicePrincipalLogin, login)
				roles = &roleIDs
				return roleIDs, nil
			},
		}
		ctx := config.WithMockClient(t.Context(), mock)
		resp, err := (&ServicePrincipal{}).Update(ctx, infer.UpdateRequest[ServicePrincipalInput, ServicePrincipalState]{
			ID:     testServicePrincipalID,
			Inputs: inputs,
			State:  state,
		})
		require.NoError(t, err)
		assert.Equal(t, ServicePrincipalState{ServicePrincipalInput: inputs, Login: testServicePrincipalLogin}, resp.Output)
		return renamed, roles
	}

	t.Run("rename only", func(t *testing.T) {
		in := state.ServicePrincipalInput
		in.Name = "new-name"
		renamed, roles := run(t, in)
		require.NotNil(t, renamed)
		assert.Equal(t, "new-name", *renamed)
		assert.Nil(t, roles, "roles must not be rewritten")
	})

	t.Run("roles only", func(t *testing.T) {
		in := state.ServicePrincipalInput
		in.RoleIDs = []string{"role-3"}
		renamed, roles := run(t, in)
		assert.Nil(t, renamed, "name must not be rewritten")
		require.NotNil(t, roles)
		assert.Equal(t, []string{"role-3"}, *roles)
	})

	t.Run("removing every role", func(t *testing.T) {
		in := state.ServicePrincipalInput
		in.RoleIDs = nil
		_, roles := run(t, in)
		require.NotNil(t, roles)
		assert.Empty(t, *roles)
	})

	t.Run("reordered roles make no call", func(t *testing.T) {
		in := state.ServicePrincipalInput
		in.RoleIDs = []string{"role-2", "role-1"}
		renamed, roles := run(t, in)
		assert.Nil(t, renamed)
		assert.Nil(t, roles)
	})

	t.Run("preview makes no call", func(t *testing.T) {
		ctx := config.WithMockClient(t.Context(), &servicePrincipalClientMock{})
		in := state.ServicePrincipalInput
		in.Name = "new-name"
		resp, err := (&ServicePrincipal{}).Update(ctx, infer.UpdateRequest[ServicePrincipalInput, ServicePrincipalState]{
			ID:     testServicePrincipalID,
			Inputs: in,
			State:  state,
			DryRun: true,
		})
		require.NoError(t, err)
		assert.Equal(t, testServicePrincipalLogin, resp.Output.Login)
	})
}

func TestServicePrincipalRead(t *testing.T) {
	read := func(
		t *testing.T, sp *pulumiapi.ServicePrincipal, state ServicePrincipalState,
	) infer.ReadResponse[ServicePrincipalInput, ServicePrincipalState] {
		t.Helper()
		mock := &servicePrincipalClientMock{
			get: func(orgName, login string) (*pulumiapi.ServicePrincipal, error) {
				assert.Equal(t, gcMyOrg, orgName)
				assert.Equal(t, testServicePrincipalLogin, login)
				return sp, nil
			},
		}
		ctx := config.WithMockClient(t.Context(), mock)
		resp, err := (&ServicePrincipal{}).Read(ctx, infer.ReadRequest[ServicePrincipalInput, ServicePrincipalState]{
			ID:    testServicePrincipalID,
			State: state,
		})
		require.NoError(t, err)
		return resp
	}

	t.Run("import", func(t *testing.T) {
		resp := read(t, &pulumiapi.ServicePrincipal{
			Login:   testServicePrincipalLogin,
			Name:    testServicePrincipalName,
			RoleIDs: []string{"role-1"},
		}, ServicePrincipalState{})
		want := ServicePrincipalInput{
			OrganizationName: gcMyOrg,
			Name:             testServicePrincipalName,
			RoleIDs:          []string{"role-1"},
		}
		assert.Equal(t, testServicePrincipalID, resp.ID)
		assert.Equal(t, want, resp.Inputs)
		assert.Equal(t, ServicePrincipalState{ServicePrincipalInput: want, Login: testServicePrincipalLogin}, resp.State)
	})

	t.Run("keeps the prior order of the same roles", func(t *testing.T) {
		resp := read(t, &pulumiapi.ServicePrincipal{
			Login:   testServicePrincipalLogin,
			Name:    testServicePrincipalName,
			RoleIDs: []string{"role-1", "role-2"},
		}, ServicePrincipalState{ServicePrincipalInput: ServicePrincipalInput{RoleIDs: []string{"role-2", "role-1"}}})
		assert.Equal(t, []string{"role-2", "role-1"}, resp.Inputs.RoleIDs)
	})

	t.Run("reports drift in roles and name", func(t *testing.T) {
		resp := read(t, &pulumiapi.ServicePrincipal{
			Login:   testServicePrincipalLogin,
			Name:    "renamed-in-console",
			RoleIDs: []string{"role-3"},
		}, ServicePrincipalState{ServicePrincipalInput: ServicePrincipalInput{
			Name:    testServicePrincipalName,
			RoleIDs: []string{"role-1"},
		}})
		assert.Equal(t, "renamed-in-console", resp.Inputs.Name)
		assert.Equal(t, []string{"role-3"}, resp.Inputs.RoleIDs)
	})

	t.Run("no roles", func(t *testing.T) {
		resp := read(t, &pulumiapi.ServicePrincipal{
			Login:   testServicePrincipalLogin,
			Name:    testServicePrincipalName,
			RoleIDs: []string{},
		}, ServicePrincipalState{})
		assert.Nil(t, resp.Inputs.RoleIDs)
	})

	t.Run("deleted", func(t *testing.T) {
		resp := read(t, nil, ServicePrincipalState{})
		assert.Empty(t, resp.ID)
	})
}

func TestServicePrincipalDelete(t *testing.T) {
	var deleted string
	mock := &servicePrincipalClientMock{
		del: func(orgName, login string) error {
			assert.Equal(t, gcMyOrg, orgName)
			deleted = login
			return nil
		},
	}
	ctx := config.WithMockClient(t.Context(), mock)
	_, err := (&ServicePrincipal{}).Delete(ctx, infer.DeleteRequest[ServicePrincipalState]{ID: testServicePrincipalID})
	require.NoError(t, err)
	assert.Equal(t, testServicePrincipalLogin, deleted)
}

func TestSplitServicePrincipalID(t *testing.T) {
	org, login, err := splitServicePrincipalID(testServicePrincipalID)
	require.NoError(t, err)
	assert.Equal(t, gcMyOrg, org)
	assert.Equal(t, testServicePrincipalLogin, login)

	for _, id := range []string{"", "my-org", "my-org/", "/login", "my-org/a/b"} {
		_, _, err := splitServicePrincipalID(id)
		assert.Error(t, err, id)
	}
}
