package resources

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/stretchr/testify/assert"
)

type TeamClientMock struct {
	config.Client
	getTeamFunc func(ctx context.Context, orgName string, teamName string) (*pulumiapi.Team, error)
}

func (c *TeamClientMock) GetTeam(ctx context.Context, orgName string, teamName string) (*pulumiapi.Team, error) {
	return c.getTeamFunc(ctx, orgName, teamName)
}

func TestTeam(t *testing.T) {
	t.Run("Read when the resource is not found", func(t *testing.T) {
		mockedClient := &TeamClientMock{
			getTeamFunc: func(ctx context.Context, orgName string, teamName string) (*pulumiapi.Team, error) {
				return nil, nil
			},
		}

		ctx := config.WithMockClient(context.Background(), mockedClient)

		team := &Team{}
		req := infer.ReadRequest[TeamInput, TeamState]{
			ID: "abc/test",
			Inputs: TeamInput{
				TeamCore: TeamCore{
					OrganizationName: "abc",
					Type:             "pulumi",
					Name:             ref("test"),
				},
			},
			State: TeamState{
				TeamCore: TeamCore{
					OrganizationName: "abc",
					Type:             "pulumi",
					Name:             ref("test"),
				},
			},
		}

		resp, err := team.Read(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "", resp.ID)
		assert.Equal(t, TeamInput{}, resp.Inputs)
		assert.Equal(t, TeamState{}, resp.State)
	})

	t.Run("Read when the resource is found", func(t *testing.T) {
		mockedClient := &TeamClientMock{
			getTeamFunc: func(ctx context.Context, orgName string, teamName string) (*pulumiapi.Team, error) {
				return &pulumiapi.Team{
					Type:        "pulumi",
					Name:        "test",
					DisplayName: "test team",
					Description: "test team description",
					Members: []pulumiapi.TeamMember{
						{GithubLogin: "member1"},
						{GithubLogin: "member2"},
					},
				}, nil
			},
		}

		ctx := config.WithMockClient(context.Background(), mockedClient)

		team := &Team{}
		req := infer.ReadRequest[TeamInput, TeamState]{
			ID: "abc/test",
			Inputs: TeamInput{
				TeamCore: TeamCore{
					OrganizationName: "abc",
					Type:             "pulumi",
					Name:             ref("test"),
				},
			},
			State: TeamState{
				TeamCore: TeamCore{
					OrganizationName: "abc",
					Type:             "pulumi",
					Name:             ref("test"),
				},
			},
		}

		resp, err := team.Read(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "abc/test", resp.ID)
		assert.Equal(t, TeamInput{
			TeamCore: TeamCore{
				OrganizationName: "abc",
				Type:             "pulumi",
				Name:             ref("test"),
				DisplayName:      ref("test team"),
				Description:      ref("test team description"),
			},
			Members: []string{"member1", "member2"},
		}, resp.Inputs)
		assert.Equal(t, TeamState{
			TeamCore: TeamCore{
				OrganizationName: "abc",
				Type:             "pulumi",
				Name:             ref("test"),
				DisplayName:      ref("test team"),
				Description:      ref("test team description"),
			},
			Members: []string{"member1", "member2"},
		}, resp.State)
	})
}

func ref[T any](v T) *T { return &v }
