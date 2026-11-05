// Copyright 2016-2026, Pulumi Corporation.

package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

type defaultRoleClientMock struct {
	config.Client
	roles []apitype.PermissionDescriptorRecord
	set   []string
}

func (c *defaultRoleClientMock) ListOrgRoles(
	context.Context, string, apitype.PermissionDescriptorUXPurpose,
) ([]apitype.PermissionDescriptorRecord, error) {
	return c.roles, nil
}

func (c *defaultRoleClientMock) SetOrgDefaultRole(_ context.Context, _, roleID string) error {
	c.set = append(c.set, roleID)
	return nil
}

func defaultRoleCatalogue(defaultID string) []apitype.PermissionDescriptorRecord {
	record := func(id, name, identifier string) apitype.PermissionDescriptorRecord {
		return apitype.PermissionDescriptorRecord{
			PermissionDescriptorBase: apitype.PermissionDescriptorBase{Name: name},
			ID:                       id,
			DefaultIdentifier:        identifier,
			IsOrgDefault:             id == defaultID,
		}
	}
	return []apitype.PermissionDescriptorRecord{
		record("reader-id", "Reader", "reader"),
		record("builder-id", "Builder", gcBuilder),
		record("custom-id", "release-managers", ""),
	}
}

func strPtr(s string) *string { return &s }

func TestOrganizationDefaultRoleCheck(t *testing.T) {
	check := func(props map[string]property.Value) []string {
		resp, err := (&OrganizationDefaultRole{}).Check(t.Context(), infer.CheckRequest{
			NewInputs: property.NewMap(props),
		})
		require.NoError(t, err)
		reasons := []string{}
		for _, f := range resp.Failures {
			reasons = append(reasons, f.Reason)
		}
		return reasons
	}
	org := property.New(gcAcme)

	assert.Empty(t, check(map[string]property.Value{gcOrganizationName: org, roleProperty: property.New(gcBuilder)}))
	assert.Empty(t, check(map[string]property.Value{gcOrganizationName: org, "roleId": property.New("custom-id")}))
	assert.NotEmpty(t, check(map[string]property.Value{gcOrganizationName: org}), "one of role and roleId is required")
	assert.NotEmpty(t, check(map[string]property.Value{
		gcOrganizationName: org, roleProperty: property.New(gcBuilder), "roleId": property.New("custom-id"),
	}), "role and roleId are mutually exclusive")
	assert.NotEmpty(t, check(map[string]property.Value{gcOrganizationName: org, roleProperty: property.New("admin")}),
		"admin cannot be the default role")
	assert.NotEmpty(t, check(map[string]property.Value{gcOrganizationName: org, roleProperty: property.New("member")}),
		"member is not listed once an organization stores a default role")
}

func TestOrganizationDefaultRoleCreate(t *testing.T) {
	t.Run("built-in role is resolved to its ID", func(t *testing.T) {
		mock := &defaultRoleClientMock{roles: defaultRoleCatalogue("reader-id")}
		ctx := config.WithMockClient(t.Context(), mock)

		resp, err := (&OrganizationDefaultRole{}).Create(ctx, infer.CreateRequest[OrganizationDefaultRoleInput]{
			Inputs: OrganizationDefaultRoleInput{OrganizationName: gcAcme, Role: strPtr(gcBuilder)},
		})
		require.NoError(t, err)
		assert.Equal(t, gcAcme, resp.ID)
		assert.Equal(t, []string{"builder-id"}, mock.set)
		assert.Equal(t, "Builder", resp.Output.RoleName)
	})

	t.Run("custom role is set by ID", func(t *testing.T) {
		mock := &defaultRoleClientMock{roles: defaultRoleCatalogue("reader-id")}
		ctx := config.WithMockClient(t.Context(), mock)

		resp, err := (&OrganizationDefaultRole{}).Create(ctx, infer.CreateRequest[OrganizationDefaultRoleInput]{
			Inputs: OrganizationDefaultRoleInput{OrganizationName: gcAcme, RoleId: strPtr("custom-id")},
		})
		require.NoError(t, err)
		assert.Equal(t, []string{"custom-id"}, mock.set)
		assert.Equal(t, "release-managers", resp.Output.RoleName)
	})

	t.Run("unknown role is an error and sets nothing", func(t *testing.T) {
		mock := &defaultRoleClientMock{roles: defaultRoleCatalogue("reader-id")}
		ctx := config.WithMockClient(t.Context(), mock)

		_, err := (&OrganizationDefaultRole{}).Create(ctx, infer.CreateRequest[OrganizationDefaultRoleInput]{
			Inputs: OrganizationDefaultRoleInput{OrganizationName: gcAcme, Role: strPtr("maintainer")},
		})
		assert.ErrorContains(t, err, `built-in role "maintainer" not found`)
		assert.Empty(t, mock.set)
	})

	t.Run("preview does not call the service", func(t *testing.T) {
		mock := &defaultRoleClientMock{}
		ctx := config.WithMockClient(t.Context(), mock)

		_, err := (&OrganizationDefaultRole{}).Create(ctx, infer.CreateRequest[OrganizationDefaultRoleInput]{
			Inputs: OrganizationDefaultRoleInput{OrganizationName: gcAcme, Role: strPtr(gcBuilder)},
			DryRun: true,
		})
		require.NoError(t, err)
		assert.Empty(t, mock.set)
	})
}

func TestOrganizationDefaultRoleRead(t *testing.T) {
	read := func(t *testing.T, defaultID string, in OrganizationDefaultRoleInput) OrganizationDefaultRoleInput {
		t.Helper()
		ctx := config.WithMockClient(t.Context(), &defaultRoleClientMock{roles: defaultRoleCatalogue(defaultID)})
		resp, err := (&OrganizationDefaultRole{}).Read(ctx,
			infer.ReadRequest[OrganizationDefaultRoleInput, OrganizationDefaultRoleState]{ID: gcAcme, Inputs: in})
		require.NoError(t, err)
		return resp.Inputs
	}

	t.Run("built-in default reads as role", func(t *testing.T) {
		got := read(t, "builder-id", OrganizationDefaultRoleInput{})
		assert.Equal(t, strPtr(gcBuilder), got.Role)
		assert.Nil(t, got.RoleId)
	})

	t.Run("custom default reads as roleId", func(t *testing.T) {
		got := read(t, "custom-id", OrganizationDefaultRoleInput{})
		assert.Nil(t, got.Role)
		assert.Equal(t, strPtr("custom-id"), got.RoleId)
	})

	t.Run("a roleId that points at a built-in role stays a roleId", func(t *testing.T) {
		got := read(t, "builder-id", OrganizationDefaultRoleInput{RoleId: strPtr("builder-id")})
		assert.Nil(t, got.Role)
		assert.Equal(t, strPtr("builder-id"), got.RoleId)
	})

	t.Run("drift to another role is reported", func(t *testing.T) {
		got := read(t, "reader-id", OrganizationDefaultRoleInput{Role: strPtr(gcBuilder)})
		assert.Equal(t, strPtr("reader"), got.Role)
	})
}

func TestOrganizationDefaultRoleDeleteLeavesDefaultInPlace(t *testing.T) {
	mock := &defaultRoleClientMock{roles: defaultRoleCatalogue("builder-id")}
	ctx := config.WithMockClient(t.Context(), mock)

	_, err := (&OrganizationDefaultRole{}).Delete(ctx, infer.DeleteRequest[OrganizationDefaultRoleState]{})
	require.NoError(t, err)
	assert.Empty(t, mock.set)
}
