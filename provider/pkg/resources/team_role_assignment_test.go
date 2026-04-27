package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type teamRoleClientMock struct {
	config.Client
	assign func(ctx context.Context, org, team, roleID string) error
	remove func(ctx context.Context, org, team, roleID string) error
	get    func(ctx context.Context, org, team, roleID string) (*pulumiapi.TeamRoleRef, error)
}

func (c *teamRoleClientMock) AssignRoleToTeam(ctx context.Context, org, team, roleID string) error {
	return c.assign(ctx, org, team, roleID)
}

func (c *teamRoleClientMock) RemoveRoleFromTeam(ctx context.Context, org, team, roleID string) error {
	return c.remove(ctx, org, team, roleID)
}

func (c *teamRoleClientMock) GetTeamRole(
	ctx context.Context, org, team, roleID string,
) (*pulumiapi.TeamRoleRef, error) {
	return c.get(ctx, org, team, roleID)
}

func TestTeamRoleAssignmentCreate(t *testing.T) {
	assigned := false
	mock := &teamRoleClientMock{
		assign: func(_ context.Context, org, team, roleID string) error {
			assigned = true
			assert.Equal(t, "acme", org)
			assert.Equal(t, "devops", team)
			assert.Equal(t, "role-123", roleID)
			return nil
		},
		get: func(_ context.Context, _, _, _ string) (*pulumiapi.TeamRoleRef, error) {
			return &pulumiapi.TeamRoleRef{ID: "role-123", Name: "devops-role"}, nil
		},
	}
	ctx := config.WithMockClient(context.Background(), mock)

	r := &TeamRoleAssignment{}
	resp, err := r.Create(ctx, infer.CreateRequest[TeamRoleAssignmentInput]{
		Inputs: TeamRoleAssignmentInput{
			OrganizationName: "acme",
			TeamName:         "devops",
			RoleId:           "role-123",
		},
	})
	assert.NoError(t, err)
	assert.True(t, assigned)
	assert.Equal(t, "acme/devops/role-123", resp.ID)
	assert.Equal(t, "devops-role", resp.Output.RoleName)
}

func TestTeamRoleAssignmentRead(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		mock := &teamRoleClientMock{
			get: func(_ context.Context, _, _, _ string) (*pulumiapi.TeamRoleRef, error) {
				return &pulumiapi.TeamRoleRef{ID: "role-123", Name: "devops"}, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)
		r := &TeamRoleAssignment{}
		resp, err := r.Read(ctx, infer.ReadRequest[TeamRoleAssignmentInput, TeamRoleAssignmentState]{
			ID: "acme/devops/role-123",
		})
		assert.NoError(t, err)
		assert.Equal(t, "acme/devops/role-123", resp.ID)
		assert.Equal(t, "role-123", resp.State.RoleId)
	})

	t.Run("not found empties", func(t *testing.T) {
		mock := &teamRoleClientMock{
			get: func(_ context.Context, _, _, _ string) (*pulumiapi.TeamRoleRef, error) {
				return nil, nil
			},
		}
		ctx := config.WithMockClient(context.Background(), mock)
		r := &TeamRoleAssignment{}
		resp, err := r.Read(ctx, infer.ReadRequest[TeamRoleAssignmentInput, TeamRoleAssignmentState]{
			ID: "acme/devops/role-123",
		})
		assert.NoError(t, err)
		assert.Equal(t, "", resp.ID)
	})

	t.Run("bad id", func(t *testing.T) {
		r := &TeamRoleAssignment{}
		_, err := r.Read(context.Background(), infer.ReadRequest[TeamRoleAssignmentInput, TeamRoleAssignmentState]{
			ID: "bad",
		})
		assert.ErrorContains(t, err, "organization/team/roleId")
	})
}

func TestTeamRoleAssignmentDelete(t *testing.T) {
	removed := false
	mock := &teamRoleClientMock{
		remove: func(_ context.Context, org, team, roleID string) error {
			removed = true
			assert.Equal(t, "acme", org)
			assert.Equal(t, "devops", team)
			assert.Equal(t, "role-123", roleID)
			return nil
		},
	}
	ctx := config.WithMockClient(context.Background(), mock)
	r := &TeamRoleAssignment{}
	_, err := r.Delete(ctx, infer.DeleteRequest[TeamRoleAssignmentState]{
		State: TeamRoleAssignmentState{
			TeamRoleAssignmentInput: TeamRoleAssignmentInput{
				OrganizationName: "acme",
				TeamName:         "devops",
				RoleId:           "role-123",
			},
		},
	})
	assert.NoError(t, err)
	assert.True(t, removed)
}
