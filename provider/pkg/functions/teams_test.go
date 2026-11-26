// Copyright 2016-2026, Pulumi Corporation.

package functions

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const (
	everyoneTeamName = "everyone"
	platformTeamName = "platform"
	teamMemberAlice  = "alice"
	teamMemberBob    = "bob"
)

func newTeamsTestServer(t *testing.T) integration.Server {
	t.Helper()
	everyone := pulumiapi.Team{
		Type:        "pulumi",
		Name:        everyoneTeamName,
		DisplayName: "Everyone",
		Everyone:    true,
		Members:     []pulumiapi.TeamMember{{GithubLogin: teamMemberAlice}, {GithubLogin: teamMemberBob}},
	}
	platform := pulumiapi.Team{Type: "pulumi", Name: platformTeamName}
	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/orgs/anOrg/teams"):
			require.NoError(t, json.NewEncoder(w).Encode(pulumiapi.Teams{Teams: []pulumiapi.Team{everyone, platform}}))
		case strings.HasSuffix(r.URL.Path, "/orgs/anOrg/teams/everyone"):
			require.NoError(t, json.NewEncoder(w).Encode(everyone))
		default:
			w.WriteHeader(http.StatusNotFound)
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"code": 404, "message": "not found"}))
		}
	}))
	t.Cleanup(httpServer.Close)

	client, err := pulumiapi.NewClient(&http.Client{}, "tok", httpServer.URL)
	require.NoError(t, err)
	prov, err := infer.NewProviderBuilder().
		WithFunctions(infer.Function(&GetTeamFunction{}), infer.Function(&GetTeamsFunction{})).
		Build()
	require.NoError(t, err)
	server, err := integration.NewServer(config.WithMockClient(t.Context(), client),
		"pulumiservice", semver.MustParse("0.0.0"), integration.WithProvider(prov))
	require.NoError(t, err)
	return server
}

func TestGetTeamsMarksTheEveryoneTeam(t *testing.T) {
	resp, err := newTeamsTestServer(t).Invoke(p.InvokeRequest{
		Token: "pulumiservice:index:getTeams",
		Args:  property.NewMap(map[string]property.Value{keyOrganizationName: property.New("anOrg")}),
	})
	require.NoError(t, err)
	require.Empty(t, resp.Failures)
	teams := resp.Return.Get("teams").AsArray().AsSlice()
	require.Len(t, teams, 2)
	assert.Equal(t, everyoneTeamName, teams[0].AsMap().Get(ppKeyName).AsString())
	assert.True(t, teams[0].AsMap().Get("everyone").AsBool())
	assert.Equal(t, platformTeamName, teams[1].AsMap().Get(ppKeyName).AsString())
	assert.False(t, teams[1].AsMap().Get("everyone").AsBool())
}

func TestGetTeamReturnsTheEveryoneTeamsMembers(t *testing.T) {
	resp, err := newTeamsTestServer(t).Invoke(p.InvokeRequest{
		Token: "pulumiservice:index:getTeam",
		Args: property.NewMap(map[string]property.Value{
			keyOrganizationName: property.New("anOrg"),
			ppKeyName:           property.New(everyoneTeamName),
		}),
	})
	require.NoError(t, err)
	require.Empty(t, resp.Failures)
	assert.True(t, resp.Return.Get("everyone").AsBool())
	assert.Equal(t, "Everyone", resp.Return.Get("displayName").AsString())
	members := resp.Return.Get("members").AsArray().AsSlice()
	require.Len(t, members, 2)
	assert.Equal(t, teamMemberAlice, members[0].AsString())
	assert.Equal(t, teamMemberBob, members[1].AsString())
}

func TestGetTeamFailsForAMissingTeam(t *testing.T) {
	_, err := newTeamsTestServer(t).Invoke(p.InvokeRequest{
		Token: "pulumiservice:index:getTeam",
		Args: property.NewMap(map[string]property.Value{
			keyOrganizationName: property.New("anOrg"),
			ppKeyName:           property.New("nobody"),
		}),
	})
	require.ErrorContains(t, err, `organization "anOrg" has no team named "nobody"`)
}
