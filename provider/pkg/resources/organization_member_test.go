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

const testAdminRole = "admin"

type orgMemberClientMock struct {
	config.Client
	addFunc       func(ctx context.Context, userName, orgName, role string) error
	updateFunc    func(ctx context.Context, orgName, userName, role string, fgaRoleID *string) error
	deleteFunc    func(ctx context.Context, orgName, userName string) error
	getFunc       func(ctx context.Context, orgName, userName string) (*pulumiapi.Member, error)
	listRolesFunc func(ctx context.Context, orgName, uxPurpose string) ([]pulumiapi.RoleDescriptor, error)
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

func (c *orgMemberClientMock) ListOrgRoles(
	ctx context.Context, orgName, uxPurpose string,
) ([]pulumiapi.RoleDescriptor, error) {
	if c.listRolesFunc == nil {
		return nil, nil
	}
	return c.listRolesFunc(ctx, orgName, uxPurpose)
}

// builtinRoles is the role-catalogue fixture used by tests that don't care
// about the exact role layout, only that ListOrgRoles returns the standard
// built-in slugs so applyMemberRoleToState can round-trip an FGARole back
// to the matching built-in Role string.
var builtinRoles = []pulumiapi.RoleDescriptor{
	{ID: "admin-fga", Name: "Admin", DefaultIdentifier: "admin"},
	{ID: "member-fga", Name: "Member", DefaultIdentifier: "member"},
	{ID: "billing-fga", Name: "Billing Manager", DefaultIdentifier: "billing-manager"},
}

func stubBuiltinRoles(_ context.Context, _, _ string) ([]pulumiapi.RoleDescriptor, error) {
	return builtinRoles, nil
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
		// Regression: the Pulumi Cloud member response returns the role's
		// display name ("Admin"), not its slug ("admin"). Earlier versions
		// of applyMemberRoleToState compared the display name directly
		// against the built-in slug list, so Read populated roleId instead
		// of role and every refresh drifted +roleId-role.
		mock := &orgMemberClientMock{
			getFunc: func(_ context.Context, _, _ string) (*pulumiapi.Member, error) {
				return &pulumiapi.Member{
					Role: "admin",
					User: pulumiapi.User{
						Name:        "Alice",
						GithubLogin: "alice",
						Email:       "alice@example.com",
					},
					FGARole: &pulumiapi.FGARole{ID: "admin-fga", Name: "Admin"},
				}, nil
			},
			listRolesFunc: stubBuiltinRoles,
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
		assert.Equal(t, "alice", resp.State.Username)
	})

	t.Run("custom role surfaces the catalogue ID, not the FGA-side ID", func(t *testing.T) {
		// Regression: Pulumi Cloud hands out two separate identifiers for
		// the same role — the role-catalogue ID (`RoleDescriptor.ID`, what
		// users supply as `roleId`) and the FGA-system ID returned on the
		// member list endpoint (`member.fgaRole.id`). Surfacing the FGA id
		// made every refresh of a custom-role assignment drift `~roleId`
		// against the user's original input. The lookup has to bridge the
		// two by matching the shared display Name.
		const catalogueID = "role-catalogue-xyz"
		const fgaID = "role-fga-xyz"
		mock := &orgMemberClientMock{
			getFunc: func(_ context.Context, _, _ string) (*pulumiapi.Member, error) {
				return &pulumiapi.Member{
					Role:    "member",
					User:    pulumiapi.User{Name: "Bob", GithubLogin: "bob"},
					FGARole: &pulumiapi.FGARole{ID: fgaID, Name: "read-only-devops"},
				}, nil
			},
			listRolesFunc: func(_ context.Context, _, _ string) ([]pulumiapi.RoleDescriptor, error) {
				return append([]pulumiapi.RoleDescriptor{
					{ID: catalogueID, Name: "read-only-devops"},
				}, builtinRoles...), nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)

		r := &OrganizationMember{}
		resp, err := r.Read(ctx, infer.ReadRequest[OrganizationMemberInput, OrganizationMemberState]{
			ID: "acme/bob",
		})
		assert.NoError(t, err)
		if assert.NotNil(t, resp.State.RoleId) {
			assert.Equal(t, catalogueID, *resp.State.RoleId,
				"RoleId must be the role-catalogue ID, not the FGA-side ID")
		}
		assert.Nil(t, resp.State.Role)
		assert.Equal(t, "read-only-devops", resp.State.RoleName)
	})

	t.Run("legacy member without FGARole clears prior roleId", func(t *testing.T) {
		// Regression: when the service returns a member without an FGARole
		// (legacy / uninitialized state), the built-in branch must clear any
		// roleId carried over from prior state so the resource preserves the
		// data-source contract: built-in → role set, custom → roleId set,
		// never both.
		mock := &orgMemberClientMock{
			getFunc: func(_ context.Context, _, _ string) (*pulumiapi.Member, error) {
				return &pulumiapi.Member{
					Role:    "admin",
					User:    pulumiapi.User{GithubLogin: "alice"},
					FGARole: nil,
				}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)
		priorRoleID := "stale-role-id"

		r := &OrganizationMember{}
		resp, err := r.Read(ctx, infer.ReadRequest[OrganizationMemberInput, OrganizationMemberState]{
			ID:     "acme/alice",
			Inputs: OrganizationMemberInput{OrganizationMemberCore: OrganizationMemberCore{RoleId: &priorRoleID}},
		})
		assert.NoError(t, err)
		if assert.NotNil(t, resp.State.Role) {
			assert.Equal(t, "admin", *resp.State.Role)
		}
		assert.Nil(t, resp.State.RoleId, "RoleId must be cleared in built-in branch")
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
					FGARole: &pulumiapi.FGARole{ID: "admin-fga", Name: "Admin"},
				}, nil
			},
			listRolesFunc: stubBuiltinRoles,
		}
		ctx := config.WithMockClient(context.Background(), mock)
		role := testAdminRole

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
			listRolesFunc: stubBuiltinRoles,
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

func TestOrganizationMemberDelete(t *testing.T) {
	t.Run("adopted: un-assigns role but keeps user", func(t *testing.T) {
		deleted := false
		updated := false
		mock := &orgMemberClientMock{
			deleteFunc: func(_ context.Context, _, _ string) error {
				deleted = true
				return nil
			},
			updateFunc: func(_ context.Context, _, _, role string, fgaRoleID *string) error {
				updated = true
				assert.Equal(t, "member", role)
				// Nil fgaRoleID: server falls through to the legacy path and
				// applies the built-in role as the new FGA role, clearing
				// any prior custom-role assignment.
				assert.Nil(t, fgaRoleID)
				return nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)
		r := &OrganizationMember{}
		_, err := r.Delete(ctx, infer.DeleteRequest[OrganizationMemberState]{
			State: OrganizationMemberState{
				OrganizationMemberCore: OrganizationMemberCore{
					OrganizationName: "acme",
					Username:         "alice",
				},
				Adopted: true,
			},
		})
		assert.NoError(t, err)
		assert.True(t, updated, "expected UpdateOrgMemberRole to un-assign the custom role")
		assert.False(t, deleted, "must not remove an adopted member from the org")
	})

	t.Run("owned: removes the user", func(t *testing.T) {
		deleted := false
		mock := &orgMemberClientMock{
			deleteFunc: func(_ context.Context, _, _ string) error {
				deleted = true
				return nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)
		r := &OrganizationMember{}
		_, err := r.Delete(ctx, infer.DeleteRequest[OrganizationMemberState]{
			State: OrganizationMemberState{
				OrganizationMemberCore: OrganizationMemberCore{
					OrganizationName: "acme",
					Username:         "alice",
				},
				Adopted: false,
			},
		})
		assert.NoError(t, err)
		assert.True(t, deleted)
	})
}

func TestOrganizationMemberUpdateDryRunPreservesAdopted(t *testing.T) {
	// Regression: Update's DryRun branch returned OrganizationMemberState{}
	// with Adopted defaulting to false. `pulumi preview` on an adopted
	// member then showed a spurious `adopted: true → false` diff.
	ctx := config.WithMockClient(context.Background(), &orgMemberClientMock{})
	role := testAdminRole

	r := &OrganizationMember{}
	resp, err := r.Update(ctx, infer.UpdateRequest[OrganizationMemberInput, OrganizationMemberState]{
		DryRun: true,
		Inputs: OrganizationMemberInput{
			OrganizationMemberCore: OrganizationMemberCore{
				OrganizationName: "acme",
				Username:         "alice",
				Role:             &role,
			},
		},
		State: OrganizationMemberState{
			OrganizationMemberCore: OrganizationMemberCore{
				OrganizationName: "acme",
				Username:         "alice",
			},
			Adopted: true,
		},
	})
	assert.NoError(t, err)
	assert.True(t, resp.Output.Adopted, "Update DryRun must preserve Adopted from prior state")
}

func TestOrganizationMemberUpdatePreservesAdopted(t *testing.T) {
	// Regression: Update called readOrgMemberState (which returns a fresh
	// state with Adopted=false) without carrying Adopted over from the prior
	// state. That caused a subsequent Delete to take the non-adopted branch
	// and remove the user from the org — see the ts-rbac manual test run
	// that deleted pulumiux-test.
	mock := &orgMemberClientMock{
		updateFunc: func(_ context.Context, _, _, _ string, _ *string) error { return nil },
		getFunc: func(_ context.Context, _, _ string) (*pulumiapi.Member, error) {
			return &pulumiapi.Member{
				Role: "admin",
				User: pulumiapi.User{GithubLogin: "alice"},
			}, nil
		},
	}
	ctx := config.WithMockClient(context.Background(), mock)
	role := testAdminRole

	r := &OrganizationMember{}
	resp, err := r.Update(ctx, infer.UpdateRequest[OrganizationMemberInput, OrganizationMemberState]{
		Inputs: OrganizationMemberInput{
			OrganizationMemberCore: OrganizationMemberCore{
				OrganizationName: "acme",
				Username:         "alice",
				Role:             &role,
			},
		},
		State: OrganizationMemberState{
			OrganizationMemberCore: OrganizationMemberCore{
				OrganizationName: "acme",
				Username:         "alice",
			},
			Adopted: true,
		},
	})
	assert.NoError(t, err)
	assert.True(t, resp.Output.Adopted, "Update must preserve Adopted from prior state")
}

func TestOrganizationMemberReadPreservesAdopted(t *testing.T) {
	// Regression: Read rebuilt state from the server without carrying the
	// Adopted flag; refresh would silently drop it and cause a subsequent
	// Delete to remove the user from the org.
	mock := &orgMemberClientMock{
		getFunc: func(_ context.Context, _, _ string) (*pulumiapi.Member, error) {
			return &pulumiapi.Member{
				Role: "member",
				User: pulumiapi.User{GithubLogin: "alice"},
			}, nil
		},
	}
	ctx := config.WithMockClient(context.Background(), mock)

	r := &OrganizationMember{}
	resp, err := r.Read(ctx, infer.ReadRequest[OrganizationMemberInput, OrganizationMemberState]{
		ID: "acme/alice",
		State: OrganizationMemberState{
			OrganizationMemberCore: OrganizationMemberCore{
				OrganizationName: "acme",
				Username:         "alice",
			},
			Adopted: true,
		},
	})
	assert.NoError(t, err)
	assert.True(t, resp.State.Adopted, "Read must preserve Adopted from prior state")
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
