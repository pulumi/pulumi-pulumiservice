// Copyright 2026, Pulumi Corporation.

package functions

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

const testFunctionRolesOrgName = "test-org"

type orgRolesClientMock struct {
	config.Client
	listOrgRolesFunc func(
		ctx context.Context,
		orgName string,
		uxPurpose apitype.PermissionDescriptorUXPurpose,
	) ([]apitype.PermissionDescriptorRecord, error)
}

func (c *orgRolesClientMock) ListOrgRoles(
	ctx context.Context,
	orgName string,
	uxPurpose apitype.PermissionDescriptorUXPurpose,
) ([]apitype.PermissionDescriptorRecord, error) {
	if c.listOrgRolesFunc == nil {
		return nil, nil
	}
	return c.listOrgRolesFunc(ctx, orgName, uxPurpose)
}

func testRole(id, name, defaultIdentifier string) apitype.PermissionDescriptorRecord {
	return apitype.PermissionDescriptorRecord{
		PermissionDescriptorBase: apitype.PermissionDescriptorBase{
			Name:         name,
			Description:  name + " description",
			ResourceType: "global",
			UxPurpose:    apitype.PermissionDescriptorUXPurposeRole,
			Details: apitype.PermissionDescriptorAllowBuilder{
				Permissions: apitype.RbacPermissionSlice{apitype.RbacPermission("stack:read")},
			}.Build(),
		},
		ID:                id,
		DefaultIdentifier: defaultIdentifier,
	}
}

// notRole returns a permission descriptor record of another kind, such as a
// permission set or policy, which the role functions must never return.
func notRole(id, name string, uxPurpose apitype.PermissionDescriptorUXPurpose) apitype.PermissionDescriptorRecord {
	r := testRole(id, name, "")
	r.UxPurpose = uxPurpose
	return r
}

// stubOrgRoles returns a fixed role list and checks that the function asks
// for the console's role catalogue.
func stubOrgRoles(t *testing.T, roles ...apitype.PermissionDescriptorRecord) *orgRolesClientMock {
	return &orgRolesClientMock{
		listOrgRolesFunc: func(
			_ context.Context,
			orgName string,
			uxPurpose apitype.PermissionDescriptorUXPurpose,
		) ([]apitype.PermissionDescriptorRecord, error) {
			assert.Equal(t, testFunctionRolesOrgName, orgName)
			assert.Equal(t, apitype.PermissionDescriptorUXPurposeRole, uxPurpose)
			return roles, nil
		},
	}
}

func invokeGetOrganizationRole(
	t *testing.T,
	client config.Client,
	name string,
) (infer.FunctionResponse[GetOrganizationRoleOutput], error) {
	ctx := config.WithMockClient(t.Context(), client)
	return GetOrganizationRoleFunction{}.Invoke(ctx, infer.FunctionRequest[GetOrganizationRoleInput]{
		Input: GetOrganizationRoleInput{OrganizationName: testFunctionRolesOrgName, Name: name},
	})
}

func TestGetOrganizationRoleFunction(t *testing.T) {
	t.Parallel()

	t.Run("finds a custom role by name", func(t *testing.T) {
		t.Parallel()
		client := stubOrgRoles(t,
			testRole("admin-id", "Admin", "admin"),
			testRole("custom-id", "CI Deployer", ""),
		)

		resp, err := invokeGetOrganizationRole(t, client, "CI Deployer")
		require.NoError(t, err)

		out := resp.Output
		assert.Equal(t, "custom-id", out.RoleId)
		assert.Equal(t, "CI Deployer", out.Name)
		require.NotNil(t, out.Description)
		assert.Equal(t, "CI Deployer description", *out.Description)
		assert.False(t, out.BuiltIn)
	})

	t.Run("finds a built-in role by name", func(t *testing.T) {
		t.Parallel()
		client := stubOrgRoles(t,
			testRole("admin-id", "Admin", "admin"),
			testRole("custom-id", "CI Deployer", ""),
		)

		resp, err := invokeGetOrganizationRole(t, client, "Admin")
		require.NoError(t, err)
		assert.Equal(t, "admin-id", resp.Output.RoleId)
		assert.True(t, resp.Output.BuiltIn)
	})

	t.Run("trims surrounding whitespace from the name", func(t *testing.T) {
		t.Parallel()
		client := stubOrgRoles(t, testRole("custom-id", "CI Deployer", ""))

		resp, err := invokeGetOrganizationRole(t, client, "  CI Deployer ")
		require.NoError(t, err)
		assert.Equal(t, "custom-id", resp.Output.RoleId)
	})

	t.Run("match is case-insensitive", func(t *testing.T) {
		t.Parallel()
		client := stubOrgRoles(t, testRole("custom-id", "CI Deployer", ""))

		resp, err := invokeGetOrganizationRole(t, client, "ci deployer")
		require.NoError(t, err)
		assert.Equal(t, "custom-id", resp.Output.RoleId)
	})

	t.Run("does not find a permission set by name", func(t *testing.T) {
		t.Parallel()
		client := stubOrgRoles(t, notRole("set-id", "Stack Write", apitype.PermissionDescriptorUXPurposeSet))

		_, err := invokeGetOrganizationRole(t, client, "Stack Write")
		assert.ErrorContains(t, err, `organization "test-org" has no role named "Stack Write"`)
	})

	t.Run("errors when no role matches", func(t *testing.T) {
		t.Parallel()
		client := stubOrgRoles(t, testRole("admin-id", "Admin", "admin"))

		_, err := invokeGetOrganizationRole(t, client, "Missing")
		assert.ErrorContains(t, err, `organization "test-org" has no role named "Missing"`)
	})

	t.Run("errors on an empty name without calling the API", func(t *testing.T) {
		t.Parallel()
		client := &orgRolesClientMock{
			listOrgRolesFunc: func(
				context.Context, string, apitype.PermissionDescriptorUXPurpose,
			) ([]apitype.PermissionDescriptorRecord, error) {
				t.Fatal("ListOrgRoles should not be called")
				return nil, nil
			},
		}

		_, err := invokeGetOrganizationRole(t, client, "  ")
		assert.ErrorContains(t, err, "`name` must not be empty")
	})

	t.Run("propagates API errors", func(t *testing.T) {
		t.Parallel()
		client := &orgRolesClientMock{
			listOrgRolesFunc: func(
				context.Context, string, apitype.PermissionDescriptorUXPurpose,
			) ([]apitype.PermissionDescriptorRecord, error) {
				return nil, errors.New("boom")
			},
		}

		_, err := invokeGetOrganizationRole(t, client, "Admin")
		assert.ErrorContains(t, err, `failed to list roles for organization "test-org": boom`)
	})
}

func TestGetOrganizationRolesFunction(t *testing.T) {
	t.Parallel()

	t.Run("returns all roles sorted by name", func(t *testing.T) {
		t.Parallel()
		client := stubOrgRoles(t,
			testRole("member-id", "Member", "member"),
			testRole("custom-id", "CI Deployer", ""),
			testRole("admin-id", "Admin", "admin"),
		)
		ctx := config.WithMockClient(t.Context(), client)

		resp, err := GetOrganizationRolesFunction{}.Invoke(ctx, infer.FunctionRequest[GetOrganizationRolesInput]{
			Input: GetOrganizationRolesInput{OrganizationName: testFunctionRolesOrgName},
		})
		require.NoError(t, err)

		names := make([]string, 0, len(resp.Output.Roles))
		for _, r := range resp.Output.Roles {
			names = append(names, r.Name)
		}
		assert.Equal(t, []string{"Admin", "CI Deployer", "Member"}, names)
	})

	t.Run("leaves out permission sets and policies", func(t *testing.T) {
		t.Parallel()
		client := stubOrgRoles(t,
			testRole("admin-id", "Admin", "admin"),
			notRole("set-id", "Stack Write", apitype.PermissionDescriptorUXPurposeSet),
			notRole("policy-id", "Deploy Policy", apitype.PermissionDescriptorUXPurposePolicy),
		)
		ctx := config.WithMockClient(t.Context(), client)

		resp, err := GetOrganizationRolesFunction{}.Invoke(ctx, infer.FunctionRequest[GetOrganizationRolesInput]{
			Input: GetOrganizationRolesInput{OrganizationName: testFunctionRolesOrgName},
		})
		require.NoError(t, err)
		require.Len(t, resp.Output.Roles, 1)
		assert.Equal(t, "admin-id", resp.Output.Roles[0].RoleId)
	})

	t.Run("propagates API errors", func(t *testing.T) {
		t.Parallel()
		client := &orgRolesClientMock{
			listOrgRolesFunc: func(
				context.Context, string, apitype.PermissionDescriptorUXPurpose,
			) ([]apitype.PermissionDescriptorRecord, error) {
				return nil, errors.New("boom")
			},
		}
		ctx := config.WithMockClient(t.Context(), client)

		_, err := GetOrganizationRolesFunction{}.Invoke(ctx, infer.FunctionRequest[GetOrganizationRolesInput]{
			Input: GetOrganizationRolesInput{OrganizationName: testFunctionRolesOrgName},
		})
		assert.ErrorContains(t, err, `failed to list roles for organization "test-org": boom`)
	})
}
