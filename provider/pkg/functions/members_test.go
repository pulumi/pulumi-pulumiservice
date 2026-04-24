// Copyright 2016-2025, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package functions

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

// ---------------------------------------------------------------------------
// mocks — embed config.Client so only the methods under test need overriding
// ---------------------------------------------------------------------------

type membersClientMock struct {
	config.Client // embed for all other interface methods
	listOrgMembersFunc func(ctx context.Context, orgName string) (*pulumiapi.Members, error)
}

func (m *membersClientMock) ListOrgMembers(ctx context.Context, orgName string) (*pulumiapi.Members, error) {
	if m.listOrgMembersFunc != nil {
		return m.listOrgMembersFunc(ctx, orgName)
	}
	return &pulumiapi.Members{}, nil
}

type teamMembersClientMock struct {
	config.Client // embed for all other interface methods
	getTeamFunc func(ctx context.Context, orgName, teamName string) (*pulumiapi.Team, error)
}

func (m *teamMembersClientMock) GetTeam(ctx context.Context, orgName, teamName string) (*pulumiapi.Team, error) {
	if m.getTeamFunc != nil {
		return m.getTeamFunc(ctx, orgName, teamName)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// getOrgMembers tests
// ---------------------------------------------------------------------------

func TestGetOrgMembers_HappyPath(t *testing.T) {
	t.Parallel()
	mock := &membersClientMock{
		listOrgMembersFunc: func(_ context.Context, orgName string) (*pulumiapi.Members, error) {
			assert.Equal(t, "myorg", orgName)
			return &pulumiapi.Members{
				Members: []pulumiapi.Member{
					{
						Role: "admin",
						User: pulumiapi.User{
							Name:        "Alice Example",
							GithubLogin: "alice",
							AvatarURL:   "https://example.com/alice.png",
							Email:       "alice@example.com",
						},
					},
					{
						Role: "member",
						User: pulumiapi.User{
							Name:        "Bob Example",
							GithubLogin: "bob",
							AvatarURL:   "https://example.com/bob.png",
							Email:       "bob@example.com",
						},
					},
				},
			}, nil
		},
	}

	ctx := config.WithMockClient(t.Context(), mock)
	fn := GetOrgMembersFunction{}
	resp, err := fn.Invoke(ctx, infer.FunctionRequest[GetOrgMembersInput]{
		Input: GetOrgMembersInput{OrganizationName: "myorg"},
	})

	require.NoError(t, err)
	require.Len(t, resp.Output.Members, 2)

	assert.Equal(t, OrgMember{
		Role:        "admin",
		GithubLogin: "alice",
		Name:        "Alice Example",
		AvatarURL:   "https://example.com/alice.png",
		Email:       "alice@example.com",
	}, resp.Output.Members[0])

	assert.Equal(t, OrgMember{
		Role:        "member",
		GithubLogin: "bob",
		Name:        "Bob Example",
		AvatarURL:   "https://example.com/bob.png",
		Email:       "bob@example.com",
	}, resp.Output.Members[1])
}

func TestGetOrgMembers_EmptyOrganizationName(t *testing.T) {
	t.Parallel()
	ctx := config.WithMockClient(t.Context(), &membersClientMock{})
	fn := GetOrgMembersFunction{}
	_, err := fn.Invoke(ctx, infer.FunctionRequest[GetOrgMembersInput]{
		Input: GetOrgMembersInput{OrganizationName: ""},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "organizationName must not be empty")
}

func TestGetOrgMembers_APIError(t *testing.T) {
	t.Parallel()
	mock := &membersClientMock{
		listOrgMembersFunc: func(_ context.Context, _ string) (*pulumiapi.Members, error) {
			return nil, errors.New("network error")
		},
	}
	ctx := config.WithMockClient(t.Context(), mock)
	fn := GetOrgMembersFunction{}
	_, err := fn.Invoke(ctx, infer.FunctionRequest[GetOrgMembersInput]{
		Input: GetOrgMembersInput{OrganizationName: "myorg"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to list org members")
	assert.Contains(t, err.Error(), "network error")
}

func TestGetOrgMembers_EmptyResult(t *testing.T) {
	t.Parallel()
	mock := &membersClientMock{
		listOrgMembersFunc: func(_ context.Context, _ string) (*pulumiapi.Members, error) {
			return &pulumiapi.Members{Members: nil}, nil
		},
	}
	ctx := config.WithMockClient(t.Context(), mock)
	fn := GetOrgMembersFunction{}
	resp, err := fn.Invoke(ctx, infer.FunctionRequest[GetOrgMembersInput]{
		Input: GetOrgMembersInput{OrganizationName: "myorg"},
	})
	require.NoError(t, err)
	assert.Empty(t, resp.Output.Members)
}

// ---------------------------------------------------------------------------
// getTeamMembers tests
// ---------------------------------------------------------------------------

func TestGetTeamMembers_HappyPath(t *testing.T) {
	t.Parallel()
	mock := &teamMembersClientMock{
		getTeamFunc: func(_ context.Context, orgName, teamName string) (*pulumiapi.Team, error) {
			assert.Equal(t, "myorg", orgName)
			assert.Equal(t, "backend", teamName)
			return &pulumiapi.Team{
				Name: "backend",
				Members: []pulumiapi.TeamMember{
					{GithubLogin: "alice", Name: "Alice Example", AvatarURL: "https://example.com/alice.png", Role: "admin"},
					{GithubLogin: "charlie", Name: "Charlie Example", AvatarURL: "https://example.com/charlie.png", Role: "member"},
				},
			}, nil
		},
	}

	ctx := config.WithMockClient(t.Context(), mock)
	fn := GetTeamMembersFunction{}
	resp, err := fn.Invoke(ctx, infer.FunctionRequest[GetTeamMembersInput]{
		Input: GetTeamMembersInput{OrganizationName: "myorg", TeamName: "backend"},
	})

	require.NoError(t, err)
	require.Len(t, resp.Output.Members, 2)

	assert.Equal(t, TeamMemberEntry{
		GithubLogin: "alice",
		Name:        "Alice Example",
		AvatarURL:   "https://example.com/alice.png",
		Role:        "admin",
	}, resp.Output.Members[0])

	assert.Equal(t, TeamMemberEntry{
		GithubLogin: "charlie",
		Name:        "Charlie Example",
		AvatarURL:   "https://example.com/charlie.png",
		Role:        "member",
	}, resp.Output.Members[1])
}

func TestGetTeamMembers_EmptyOrganizationName(t *testing.T) {
	t.Parallel()
	ctx := config.WithMockClient(t.Context(), &teamMembersClientMock{})
	fn := GetTeamMembersFunction{}
	_, err := fn.Invoke(ctx, infer.FunctionRequest[GetTeamMembersInput]{
		Input: GetTeamMembersInput{OrganizationName: "", TeamName: "backend"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "organizationName must not be empty")
}

func TestGetTeamMembers_EmptyTeamName(t *testing.T) {
	t.Parallel()
	ctx := config.WithMockClient(t.Context(), &teamMembersClientMock{})
	fn := GetTeamMembersFunction{}
	_, err := fn.Invoke(ctx, infer.FunctionRequest[GetTeamMembersInput]{
		Input: GetTeamMembersInput{OrganizationName: "myorg", TeamName: ""},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "teamName must not be empty")
}

func TestGetTeamMembers_NotFound(t *testing.T) {
	t.Parallel()
	mock := &teamMembersClientMock{
		getTeamFunc: func(_ context.Context, _, _ string) (*pulumiapi.Team, error) {
			return nil, nil // nil → not found
		},
	}
	ctx := config.WithMockClient(t.Context(), mock)
	fn := GetTeamMembersFunction{}
	_, err := fn.Invoke(ctx, infer.FunctionRequest[GetTeamMembersInput]{
		Input: GetTeamMembersInput{OrganizationName: "myorg", TeamName: "nonexistent"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetTeamMembers_APIError(t *testing.T) {
	t.Parallel()
	mock := &teamMembersClientMock{
		getTeamFunc: func(_ context.Context, _, _ string) (*pulumiapi.Team, error) {
			return nil, errors.New("service unavailable")
		},
	}
	ctx := config.WithMockClient(t.Context(), mock)
	fn := GetTeamMembersFunction{}
	_, err := fn.Invoke(ctx, infer.FunctionRequest[GetTeamMembersInput]{
		Input: GetTeamMembersInput{OrganizationName: "myorg", TeamName: "backend"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get team")
	assert.Contains(t, err.Error(), "service unavailable")
}

func TestGetTeamMembers_EmptyTeam(t *testing.T) {
	t.Parallel()
	mock := &teamMembersClientMock{
		getTeamFunc: func(_ context.Context, _, _ string) (*pulumiapi.Team, error) {
			return &pulumiapi.Team{Name: "empty-team", Members: nil}, nil
		},
	}
	ctx := config.WithMockClient(t.Context(), mock)
	fn := GetTeamMembersFunction{}
	resp, err := fn.Invoke(ctx, infer.FunctionRequest[GetTeamMembersInput]{
		Input: GetTeamMembersInput{OrganizationName: "myorg", TeamName: "empty-team"},
	})
	require.NoError(t, err)
	assert.Empty(t, resp.Output.Members)
}
