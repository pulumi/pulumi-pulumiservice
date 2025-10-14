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

func (c *TeamClientMock) GetTeam(
	_ context.Context, _ /* orgName */ string, _ /* teamName */ string,
) (*pulumiapi.Team, error) {
	return c.getTeamFunc()
}

func (c *TeamClientMock) ListTeams(_ context.Context, _ /* orgName */ string) ([]pulumiapi.Team, error) {
	return nil, nil
}

func (c *TeamClientMock) CreateTeam(
	_ context.Context,
	_ /* orgName */, _ /* teamName */, _ /* teamType */, _ /* displayName */, _ /* description */ string,
	_ /* teamID */ int64,
) (*pulumiapi.Team, error) {
	return nil, nil
}

func (c *TeamClientMock) UpdateTeam(
	_ context.Context, _ /* orgName */, _ /* teamName */, _ /* displayName */, _ /* description */ string,
) error {
	return nil
}

func (c *TeamClientMock) DeleteTeam(_ context.Context, _ /* orgName */, _ /* teamName */ string) error {
	return nil
}

func (c *TeamClientMock) AddMemberToTeam(
	_ context.Context, _ /* orgName */, _ /* teamName */, _ /* userName */ string,
) error {
	return nil
}

func (c *TeamClientMock) DeleteMemberFromTeam(
	_ context.Context, _ /* orgName */, _ /* teamName */, _ /* userName */ string,
) error {
	return nil
}

func (c *TeamClientMock) AddStackPermission(
	_ context.Context,
	_ /* stack */ pulumiapi.StackIdentifier,
	_ /* teamName */ string,
	_ /* permission */ int,
) error {
	return nil
}

func (c *TeamClientMock) RemoveStackPermission(
	_ context.Context,
	_ /* stack */ pulumiapi.StackIdentifier,
	_ /* teamName */ string,
) error {
	return nil
}

func (c *TeamClientMock) GetTeamStackPermission(
	_ context.Context,
	_ /* stack */ pulumiapi.StackIdentifier,
	_ /* teamName */ string,
) (*int, error) {
	return nil, nil
}

func (c *TeamClientMock) AddEnvironmentSettings(
	_ context.Context,
	_ /* req */ pulumiapi.CreateTeamEnvironmentSettingsRequest,
) error {
	return nil
}

func (c *TeamClientMock) RemoveEnvironmentSettings(
	_ context.Context,
	_ /* req */ pulumiapi.TeamEnvironmentSettingsRequest,
) error {
	return nil
}

func (c *TeamClientMock) GetTeamEnvironmentSettings(
	_ context.Context,
	_ /* req */ pulumiapi.TeamEnvironmentSettingsRequest,
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
