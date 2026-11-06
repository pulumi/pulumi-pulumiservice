package resources

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const (
	gcABC     = "abc"
	gcMember1 = "member1"
	gcMember2 = "member2"
)

type TeamClientMock struct {
	config.Client
	getTeamFunc func(ctx context.Context, orgName string, teamName string) (*pulumiapi.Team, error)
	createdTeam *pulumiapi.Team
	addedUsers  []string
}

func (c *TeamClientMock) CreateTeam(
	_ context.Context,
	_, teamName, teamType, _, _ string,
	_ int64,
) (*pulumiapi.Team, error) {
	c.createdTeam = &pulumiapi.Team{Name: teamName, Type: teamType}
	return c.createdTeam, nil
}

func (c *TeamClientMock) AddMemberToTeam(_ context.Context, _, _, userName string) error {
	c.addedUsers = append(c.addedUsers, userName)
	return nil
}

func (c *TeamClientMock) GetTeam(ctx context.Context, orgName string, teamName string) (*pulumiapi.Team, error) {
	return c.getTeamFunc(ctx, orgName, teamName)
}

func TestTeam(t *testing.T) {
	t.Run("Read when the resource is not found", func(t *testing.T) {
		mockedClient := &TeamClientMock{
			getTeamFunc: func(_ context.Context, _ string, _ string) (*pulumiapi.Team, error) {
				return nil, nil
			},
		}

		ctx := config.WithMockClient(context.Background(), mockedClient)

		team := &Team{}
		req := infer.ReadRequest[TeamInput, TeamState]{
			ID: "abc/test",
			Inputs: TeamInput{
				TeamCore: TeamCore{
					OrganizationName: gcABC,
					Type:             teamTypePulumi,
					Name:             ref("test"),
				},
			},
			State: TeamState{
				TeamCore: TeamCore{
					OrganizationName: gcABC,
					Type:             teamTypePulumi,
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
			getTeamFunc: func(_ context.Context, _ string, _ string) (*pulumiapi.Team, error) {
				return &pulumiapi.Team{
					Type:        teamTypePulumi,
					Name:        "test",
					DisplayName: "test team",
					Description: "test team description",
					Members: []pulumiapi.TeamMember{
						{GithubLogin: gcMember1},
						{GithubLogin: gcMember2},
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
					OrganizationName: gcABC,
					Type:             teamTypePulumi,
					Name:             ref("test"),
				},
			},
			State: TeamState{
				TeamCore: TeamCore{
					OrganizationName: gcABC,
					Type:             teamTypePulumi,
					Name:             ref("test"),
				},
			},
		}

		resp, err := team.Read(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, "abc/test", resp.ID)
		assert.Equal(t, TeamInput{
			TeamCore: TeamCore{
				OrganizationName: gcABC,
				Type:             teamTypePulumi,
				Name:             ref("test"),
				DisplayName:      ref("test team"),
				Description:      ref("test team description"),
			},
			Members: []string{gcMember1, gcMember2},
		}, resp.Inputs)
		assert.Equal(t, TeamState{
			TeamCore: TeamCore{
				OrganizationName: gcABC,
				Type:             teamTypePulumi,
				Name:             ref("test"),
				DisplayName:      ref("test team"),
				Description:      ref("test team description"),
			},
			Members: []string{gcMember1, gcMember2},
		}, resp.State)
	})
}

func TestTeamCreateWarnsAboutUndeclaredMembers(t *testing.T) {
	const creator = "creator"
	for _, tc := range []struct {
		name     string
		declared []string
		wantWarn bool
	}{
		{"creator left out", []string{gcMember1}, true},
		{"no members declared", nil, true},
		{"creator declared", []string{creator, gcMember1}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var logs bytes.Buffer
			previous := slog.Default()
			slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
			t.Cleanup(func() { slog.SetDefault(previous) })

			client := &TeamClientMock{}
			client.getTeamFunc = func(_ context.Context, _, teamName string) (*pulumiapi.Team, error) {
				members := []pulumiapi.TeamMember{{GithubLogin: creator}}
				for _, u := range client.addedUsers {
					if u != creator {
						members = append(members, pulumiapi.TeamMember{GithubLogin: u})
					}
				}
				return &pulumiapi.Team{Name: teamName, Type: teamTypePulumi, Members: members}, nil
			}
			ctx := config.WithMockClient(context.Background(), client)

			_, err := (&Team{}).Create(ctx, infer.CreateRequest[TeamInput]{
				Inputs: TeamInput{
					TeamCore: TeamCore{OrganizationName: gcABC, Type: teamTypePulumi, Name: ref("test")},
					Members:  tc.declared,
				},
			})

			assert.NoError(t, err)
			if tc.wantWarn {
				assert.Contains(t, logs.String(), `Pulumi Cloud added \"creator\" to team \"test\"`,
					"expected a warning naming the undeclared creator")
			} else {
				assert.NotContains(t, logs.String(), "Pulumi Cloud added",
					"expected no warning when the creator is declared")
			}
		})
	}
}

func ref[T any](v T) *T { return &v }
