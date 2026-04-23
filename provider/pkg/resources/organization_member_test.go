package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type orgMemberClientMock struct {
	config.Client
	addFunc    func(ctx context.Context, userName, orgName, role string) error
	updateFunc func(ctx context.Context, orgName, userName, role string, fgaRoleID *string) error
	deleteFunc func(ctx context.Context, orgName, userName string) error
	getFunc    func(ctx context.Context, orgName, userName string) (*pulumiapi.Member, error)
}

func (c *orgMemberClientMock) AddMemberToOrg(ctx context.Context, userName, orgName, role string) error {
	return c.addFunc(ctx, userName, orgName, role)
}

func (c *orgMemberClientMock) UpdateOrgMemberRole(
	ctx context.Context, orgName, userName, role string, fgaRoleID *string,
) error {
	return c.updateFunc(ctx, orgName, userName, role, fgaRoleID)
}

func (c *orgMemberClientMock) DeleteMemberFromOrg(ctx context.Context, orgName, userName string) error {
	return c.deleteFunc(ctx, orgName, userName)
}

func (c *orgMemberClientMock) GetOrgMember(
	ctx context.Context, orgName, userName string,
) (*pulumiapi.Member, error) {
	return c.getFunc(ctx, orgName, userName)
}

func TestOrganizationMemberRead(t *testing.T) {
	t.Run("not found returns empty", func(t *testing.T) {
		mock := &orgMemberClientMock{
			getFunc: func(_ context.Context, _, _ string) (*pulumiapi.Member, error) { return nil, nil },
		}
		ctx := config.WithMockClient(context.Background(), mock)

		r := &OrganizationMember{}
		resp, err := r.Read(ctx, infer.ReadRequest[OrganizationMemberInput, OrganizationMemberState]{
			ID: "acme/alice",
		})
		assert.NoError(t, err)
		assert.Equal(t, "", resp.ID)
	})

	t.Run("built-in role surfaces as role, not roleId", func(t *testing.T) {
		mock := &orgMemberClientMock{
			getFunc: func(_ context.Context, _, _ string) (*pulumiapi.Member, error) {
				return &pulumiapi.Member{
					Role: "admin",
					User: pulumiapi.User{
						Name:        "Alice",
						GithubLogin: "alice",
						Email:       "alice@example.com",
					},
					KnownToPulumi: true,
					FGARole:       &pulumiapi.FGARole{ID: "admin", Name: "admin"},
				}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)

		r := &OrganizationMember{}
		resp, err := r.Read(ctx, infer.ReadRequest[OrganizationMemberInput, OrganizationMemberState]{
			ID: "acme/alice",
		})
		assert.NoError(t, err)
		assert.Equal(t, "acme/alice", resp.ID)
		if assert.NotNil(t, resp.State.Role) {
			assert.Equal(t, "admin", *resp.State.Role)
		}
		assert.Nil(t, resp.State.RoleId)
		assert.Equal(t, "alice", resp.State.GithubLogin)
	})

	t.Run("custom role surfaces as roleId", func(t *testing.T) {
		mock := &orgMemberClientMock{
			getFunc: func(_ context.Context, _, _ string) (*pulumiapi.Member, error) {
				return &pulumiapi.Member{
					Role:          "member",
					User:          pulumiapi.User{Name: "Bob", GithubLogin: "bob"},
					KnownToPulumi: true,
					FGARole:       &pulumiapi.FGARole{ID: "role-xyz", Name: "read-only-devops"},
				}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)

		r := &OrganizationMember{}
		resp, err := r.Read(ctx, infer.ReadRequest[OrganizationMemberInput, OrganizationMemberState]{
			ID: "acme/bob",
		})
		assert.NoError(t, err)
		if assert.NotNil(t, resp.State.RoleId) {
			assert.Equal(t, "role-xyz", *resp.State.RoleId)
		}
		assert.Nil(t, resp.State.Role)
		assert.Equal(t, "read-only-devops", resp.State.RoleName)
	})

	t.Run("bad id", func(t *testing.T) {
		r := &OrganizationMember{}
		_, err := r.Read(context.Background(), infer.ReadRequest[OrganizationMemberInput, OrganizationMemberState]{
			ID: "no-slash",
		})
		assert.ErrorContains(t, err, "must be in the format")
	})
}

func TestOrganizationMemberCreate(t *testing.T) {
	t.Run("built-in role", func(t *testing.T) {
		added, updated := false, false
		mock := &orgMemberClientMock{
			addFunc: func(_ context.Context, user, org, role string) error {
				added = true
				assert.Equal(t, "alice", user)
				assert.Equal(t, "acme", org)
				assert.Equal(t, "admin", role)
				return nil
			},
			updateFunc: func(_ context.Context, _, _, _ string, _ *string) error {
				updated = true
				return nil
			},
			getFunc: func(_ context.Context, _, _ string) (*pulumiapi.Member, error) {
				return &pulumiapi.Member{
					Role:    "admin",
					User:    pulumiapi.User{GithubLogin: "alice"},
					FGARole: &pulumiapi.FGARole{ID: "admin", Name: "admin"},
				}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)
		role := "admin"

		r := &OrganizationMember{}
		resp, err := r.Create(ctx, infer.CreateRequest[OrganizationMemberInput]{
			Inputs: OrganizationMemberInput{
				OrganizationMemberCore: OrganizationMemberCore{
					OrganizationName: "acme",
					Username:         "alice",
					Role:             &role,
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, "acme/alice", resp.ID)
		assert.True(t, added)
		assert.False(t, updated, "no UpdateOrgMemberRole expected when only built-in role is set")
	})

	t.Run("custom role triggers follow-up update", func(t *testing.T) {
		updateRoleID := ""
		mock := &orgMemberClientMock{
			addFunc: func(_ context.Context, _, _, role string) error {
				assert.Equal(t, "member", role, "defaults to member when roleId is set")
				return nil
			},
			updateFunc: func(_ context.Context, _, _, role string, fgaRoleID *string) error {
				assert.Equal(t, "", role, "role argument must be empty when promoting to custom role")
				if assert.NotNil(t, fgaRoleID) {
					updateRoleID = *fgaRoleID
				}
				return nil
			},
			getFunc: func(_ context.Context, _, _ string) (*pulumiapi.Member, error) {
				return &pulumiapi.Member{
					Role:    "member",
					User:    pulumiapi.User{GithubLogin: "bob"},
					FGARole: &pulumiapi.FGARole{ID: "role-xyz", Name: "custom"},
				}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)
		roleID := "role-xyz"

		r := &OrganizationMember{}
		_, err := r.Create(ctx, infer.CreateRequest[OrganizationMemberInput]{
			Inputs: OrganizationMemberInput{
				OrganizationMemberCore: OrganizationMemberCore{
					OrganizationName: "acme",
					Username:         "bob",
					RoleId:           &roleID,
				},
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, "role-xyz", updateRoleID)
	})
}

func TestOrganizationMemberCheck(t *testing.T) {
	r := &OrganizationMember{}
	bad := "owner"

	resp, err := r.Check(context.Background(), infer.CheckRequest{
		NewInputs: property.NewMap(map[string]property.Value{
			"organizationName": property.New("acme"),
			"username":         property.New("alice"),
			"role":             property.New(bad),
		}),
	})
	assert.NoError(t, err)
	if assert.Len(t, resp.Failures, 1) {
		assert.Equal(t, "role", resp.Failures[0].Property)
	}
}
