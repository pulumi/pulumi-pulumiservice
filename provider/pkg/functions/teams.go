// Copyright 2016-2026, Pulumi Corporation.
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
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type TeamInfo struct {
	Name        string `pulumi:"name"`
	DisplayName string `pulumi:"displayName"`
	Description string `pulumi:"description"`
	Type        string `pulumi:"type"`
	Everyone    bool   `pulumi:"everyone"`
}

func (t *TeamInfo) Annotate(a infer.Annotator) {
	a.Describe(&t.Name, "The team's name, used to grant to it.")
	a.Describe(&t.DisplayName, "The team's display name.")
	a.Describe(&t.Description, "The team's description.")
	a.Describe(&t.Type, "The type of team: `pulumi` or `github`.")
	a.Describe(
		&t.Everyone,
		"Whether this is the Everyone team: every admin and member of the organization is in it, never "+
			"billing managers. Pulumi manages it, so it can't be managed with the `Team` resource.",
	)
}

func teamInfoFrom(t pulumiapi.Team) TeamInfo {
	return TeamInfo{
		Name:        t.Name,
		DisplayName: t.DisplayName,
		Description: t.Description,
		Type:        t.Type,
		Everyone:    t.Everyone,
	}
}

type GetTeamsFunction struct{}

type GetTeamsInput struct {
	OrganizationName string `pulumi:"organizationName"`
}

type GetTeamsOutput struct {
	Teams []TeamInfo `pulumi:"teams"`
}

func (GetTeamsFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetTeamsFunction{},
		"Lists the teams of a Pulumi Cloud organization, including the Everyone team.",
	)
	a.SetToken("index", "getTeams")
}

func (i *GetTeamsInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The name of the Pulumi organization.")
}

func (GetTeamsFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetTeamsInput],
) (infer.FunctionResponse[GetTeamsOutput], error) {
	teams, err := config.GetClient(ctx).ListTeams(ctx, req.Input.OrganizationName)
	if err != nil {
		return infer.FunctionResponse[GetTeamsOutput]{}, err
	}
	out := make([]TeamInfo, 0, len(teams))
	for _, t := range teams {
		out = append(out, teamInfoFrom(t))
	}
	return infer.FunctionResponse[GetTeamsOutput]{Output: GetTeamsOutput{Teams: out}}, nil
}

type GetTeamFunction struct{}

type GetTeamInput struct {
	OrganizationName string `pulumi:"organizationName"`
	Name             string `pulumi:"name"`
}

type GetTeamOutput struct {
	TeamInfo
	Members []string `pulumi:"members"`
}

func (GetTeamFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetTeamFunction{},
		"Looks up a team of a Pulumi Cloud organization by name, for example the Everyone team, `everyone`. "+
			"Returns an error when the team is not found.",
	)
	a.SetToken("index", "getTeam")
}

func (i *GetTeamInput) Annotate(a infer.Annotator) {
	a.Describe(&i.OrganizationName, "The name of the Pulumi organization.")
	a.Describe(&i.Name, "The team's name.")
}

func (o *GetTeamOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.Members, "Pulumi Cloud usernames of the team's members.")
}

func (GetTeamFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetTeamInput],
) (infer.FunctionResponse[GetTeamOutput], error) {
	name := strings.TrimSpace(req.Input.Name)
	if name == "" {
		return infer.FunctionResponse[GetTeamOutput]{}, fmt.Errorf("`name` must not be empty")
	}
	team, err := config.GetClient(ctx).GetTeam(ctx, req.Input.OrganizationName, name)
	if err != nil {
		return infer.FunctionResponse[GetTeamOutput]{}, err
	}
	if team == nil {
		return infer.FunctionResponse[GetTeamOutput]{}, fmt.Errorf(
			"organization %q has no team named %q", req.Input.OrganizationName, name)
	}
	members := make([]string, 0, len(team.Members))
	for _, m := range team.Members {
		members = append(members, m.GithubLogin)
	}
	return infer.FunctionResponse[GetTeamOutput]{
		Output: GetTeamOutput{TeamInfo: teamInfoFrom(*team), Members: members},
	}, nil
}
