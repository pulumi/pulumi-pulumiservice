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
	"fmt"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

// ---------------------------------------------------------------------------
// getOrgMembers
// ---------------------------------------------------------------------------

// GetOrgMembersFunction is an invoke function that lists all members of a
// Pulumi Cloud organization, together with their role and basic user profile.
type GetOrgMembersFunction struct{}

// GetOrgMembersInput is the input to getOrgMembers.
type GetOrgMembersInput struct {
	// OrganizationName is the name of the Pulumi Cloud organization.
	OrganizationName string `pulumi:"organizationName"`
}

// OrgMember holds the membership details for a single user within an org.
type OrgMember struct {
	// Role is the member's org-level role ("admin" or "member").
	Role string `pulumi:"role"`
	// GithubLogin is the user's GitHub username (also used as the Pulumi login).
	GithubLogin string `pulumi:"githubLogin"`
	// Name is the user's display name.
	Name string `pulumi:"name"`
	// AvatarURL is the URL of the user's avatar image.
	AvatarURL string `pulumi:"avatarUrl"`
	// Email is the user's email address.
	Email string `pulumi:"email"`
}

// GetOrgMembersOutput is the output of getOrgMembers.
type GetOrgMembersOutput struct {
	// Members is the list of all organization members.
	Members []OrgMember `pulumi:"members"`
}

func (GetOrgMembersFunction) Annotate(a infer.Annotator) {
	a.Describe(&GetOrgMembersFunction{},
		"List all members of a Pulumi Cloud organization.\n\n"+
			"Returns each member's GitHub login, display name, email, avatar URL, and "+
			"org-level role ('admin' or 'member'). This is the primary way to resolve a "+
			"user's login name for use in team membership, stack permissions, and other "+
			"resources that reference users by their GitHub login.",
	)
	a.SetToken("index", "getOrgMembers")
}

func (GetOrgMembersFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetOrgMembersInput],
) (infer.FunctionResponse[GetOrgMembersOutput], error) {
	client := config.GetClient(ctx)

	orgName := req.Input.OrganizationName
	if orgName == "" {
		return infer.FunctionResponse[GetOrgMembersOutput]{},
			fmt.Errorf("organizationName must not be empty")
	}

	members, err := client.ListOrgMembers(ctx, orgName)
	if err != nil {
		return infer.FunctionResponse[GetOrgMembersOutput]{},
			fmt.Errorf("failed to list org members for %q: %w", orgName, err)
	}

	output := make([]OrgMember, 0, len(members.Members))
	for _, m := range members.Members {
		output = append(output, OrgMember{
			Role:        m.Role,
			GithubLogin: m.User.GithubLogin,
			Name:        m.User.Name,
			AvatarURL:   m.User.AvatarURL,
			Email:       m.User.Email,
		})
	}

	return infer.FunctionResponse[GetOrgMembersOutput]{
		Output: GetOrgMembersOutput{Members: output},
	}, nil
}

// ---------------------------------------------------------------------------
// getTeamMembers
// ---------------------------------------------------------------------------

// GetTeamMembersFunction is an invoke function that returns the members of a
// specific team within a Pulumi Cloud organization.
type GetTeamMembersFunction struct{}

// GetTeamMembersInput is the input to getTeamMembers.
type GetTeamMembersInput struct {
	// OrganizationName is the name of the Pulumi Cloud organization.
	OrganizationName string `pulumi:"organizationName"`
	// TeamName is the name of the team within the organization.
	TeamName string `pulumi:"teamName"`
}

// TeamMemberEntry holds the profile for a single team member.
type TeamMemberEntry struct {
	// GithubLogin is the user's GitHub username (also used as the Pulumi login).
	GithubLogin string `pulumi:"githubLogin"`
	// Name is the user's display name.
	Name string `pulumi:"name"`
	// AvatarURL is the URL of the user's avatar image.
	AvatarURL string `pulumi:"avatarUrl"`
	// Role is the member's role within the team.
	Role string `pulumi:"role"`
}

// GetTeamMembersOutput is the output of getTeamMembers.
type GetTeamMembersOutput struct {
	// Members is the list of team members.
	Members []TeamMemberEntry `pulumi:"members"`
}

func (GetTeamMembersFunction) Annotate(a infer.Annotator) {
	a.Describe(&GetTeamMembersFunction{},
		"List all members of a team in a Pulumi Cloud organization.\n\n"+
			"Returns each member's GitHub login, display name, avatar URL, and team role. "+
			"Useful for resolving user logins when configuring stack or environment "+
			"permissions for individual team members.",
	)
	a.SetToken("index", "getTeamMembers")
}

func (GetTeamMembersFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetTeamMembersInput],
) (infer.FunctionResponse[GetTeamMembersOutput], error) {
	client := config.GetClient(ctx)

	orgName := req.Input.OrganizationName
	teamName := req.Input.TeamName

	if orgName == "" {
		return infer.FunctionResponse[GetTeamMembersOutput]{},
			fmt.Errorf("organizationName must not be empty")
	}
	if teamName == "" {
		return infer.FunctionResponse[GetTeamMembersOutput]{},
			fmt.Errorf("teamName must not be empty")
	}

	team, err := client.GetTeam(ctx, orgName, teamName)
	if err != nil {
		return infer.FunctionResponse[GetTeamMembersOutput]{},
			fmt.Errorf("failed to get team %q in org %q: %w", teamName, orgName, err)
	}
	if team == nil {
		return infer.FunctionResponse[GetTeamMembersOutput]{},
			fmt.Errorf("team %q not found in org %q", teamName, orgName)
	}

	output := make([]TeamMemberEntry, 0, len(team.Members))
	for _, m := range team.Members {
		output = append(output, TeamMemberEntry{
			GithubLogin: m.GithubLogin,
			Name:        m.Name,
			AvatarURL:   m.AvatarURL,
			Role:        m.Role,
		})
	}

	return infer.FunctionResponse[GetTeamMembersOutput]{
		Output: GetTeamMembersOutput{Members: output},
	}, nil
}
