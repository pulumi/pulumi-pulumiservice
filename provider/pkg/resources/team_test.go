package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type getTeamFunc func() (*pulumiapi.Team, error)

type TeamClientMock struct {
	getTeamFunc getTeamFunc
}

func (c *TeamClientMock) GetTeam(ctx context.Context, orgName string, teamName string) (*pulumiapi.Team, error) {
	return c.getTeamFunc()
}

func (c *TeamClientMock) ListTeams(ctx context.Context, orgName string) ([]pulumiapi.Team, error) {
	return nil, nil
}

func (c *TeamClientMock) CreateTeam(
	ctx context.Context,
	orgName, teamName, teamType, displayName, description string,
	teamID int64,
) (*pulumiapi.Team, error) {
	return nil, nil
}
func (c *TeamClientMock) UpdateTeam(ctx context.Context, orgName, teamName, displayName, description string) error {
	return nil
}
func (c *TeamClientMock) DeleteTeam(ctx context.Context, orgName, teamName string) error { return nil }
func (c *TeamClientMock) AddMemberToTeam(ctx context.Context, orgName, teamName, userName string) error {
	return nil
}
func (c *TeamClientMock) DeleteMemberFromTeam(ctx context.Context, orgName, teamName, userName string) error {
	return nil
}

func (c *TeamClientMock) AddStackPermission(
	ctx context.Context,
	stack pulumiapi.StackIdentifier,
	teamName string,
	permission int,
) error {
	return nil
}

func (c *TeamClientMock) RemoveStackPermission(
	ctx context.Context,
	stack pulumiapi.StackIdentifier,
	teamName string,
) error {
	return nil
}

func (c *TeamClientMock) GetTeamStackPermission(
	ctx context.Context,
	stack pulumiapi.StackIdentifier,
	teamName string,
) (*int, error) {
	return nil, nil
}

func (c *TeamClientMock) AddEnvironmentSettings(
	ctx context.Context,
	req pulumiapi.CreateTeamEnvironmentSettingsRequest,
) error {
	return nil
}

func (c *TeamClientMock) RemoveEnvironmentSettings(
	ctx context.Context,
	req pulumiapi.TeamEnvironmentSettingsRequest,
) error {
	return nil
}

func (c *TeamClientMock) GetTeamEnvironmentSettings(
	ctx context.Context,
	req pulumiapi.TeamEnvironmentSettingsRequest,
) (*string, *pulumiapi.Duration, error) {
	return nil, nil, nil
}

func buildTeamClientMock(getTeamFunc getTeamFunc) *TeamClientMock {
	return &TeamClientMock{
		getTeamFunc,
	}
}

func TestTeam(t *testing.T) {
	t.Run("Read when the resource is not found", func(t *testing.T) {
		mockedClient := buildTeamClientMock(
			func() (*pulumiapi.Team, error) { return nil, nil },
		)

		provider := PulumiServiceTeamResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "abc/123",
			Urn: "urn:123",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "")
		assert.Nil(t, resp.Properties)
	})

	t.Run("Read when the resource is found", func(t *testing.T) {
		mockedClient := buildTeamClientMock(
			func() (*pulumiapi.Team, error) {
				return &pulumiapi.Team{
					Type:        "pulumi",
					Name:        "test",
					DisplayName: "test team",
					Description: "test team description",
					Members: []pulumiapi.TeamMember{
						{Name: "member1"},
						{Name: "member2"},
					},
				}, nil
			},
		)

		provider := PulumiServiceTeamResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "abc/123",
			Urn: "urn:123",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "abc/123")
	})
}
