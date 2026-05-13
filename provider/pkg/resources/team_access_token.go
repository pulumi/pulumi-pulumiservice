// Copyright 2026, Pulumi Corporation.
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

package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

type TeamAccessToken struct{}

var (
	_ infer.CustomCreate[TeamAccessTokenInput, TeamAccessTokenState] = &TeamAccessToken{}
	_ infer.CustomDelete[TeamAccessTokenState]                       = &TeamAccessToken{}
	_ infer.CustomRead[TeamAccessTokenInput, TeamAccessTokenState]   = &TeamAccessToken{}
	_ infer.CustomStateMigrations[TeamAccessTokenState]              = &TeamAccessToken{}
)

func (*TeamAccessToken) Annotate(a infer.Annotator) {
	a.Describe(
		&TeamAccessToken{},
		"The Pulumi Cloud allows users to create access tokens scoped to team. "+
			"Team access tokens is a resource to create them and assign them to a team",
	)
}

type TeamAccessTokenInput struct {
	Name             string  `pulumi:"name"             provider:"replaceOnChanges"`
	OrganizationName string  `pulumi:"organizationName" provider:"replaceOnChanges"`
	TeamName         string  `pulumi:"teamName"         provider:"replaceOnChanges"`
	Description      *string `pulumi:"description,optional" provider:"replaceOnChanges"`
}

func (i *TeamAccessTokenInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Name, "The name for the token. This must be unique amongst all machine tokens within your organization.")
	a.Describe(&i.OrganizationName, "The organization's name.")
	a.Describe(&i.TeamName, "The team name.")
	a.Describe(&i.Description, "Optional. Description for the token.")
}

type TeamAccessTokenState struct {
	TeamAccessTokenInput
	Value string `pulumi:"value" provider:"secret"`
}

func (s *TeamAccessTokenState) Annotate(a infer.Annotator) {
	a.Describe(&s.Value, "The token's value.")
}

func (*TeamAccessToken) Create(
	ctx context.Context,
	req infer.CreateRequest[TeamAccessTokenInput],
) (infer.CreateResponse[TeamAccessTokenState], error) {
	if req.DryRun {
		return infer.CreateResponse[TeamAccessTokenState]{
			Output: TeamAccessTokenState{TeamAccessTokenInput: req.Inputs},
		}, nil
	}
	desc := ""
	if req.Inputs.Description != nil {
		desc = *req.Inputs.Description
	}
	token, err := config.GetClient(ctx).CreateTeamAccessToken(
		ctx, req.Inputs.Name, req.Inputs.OrganizationName, req.Inputs.TeamName, desc,
	)
	if err != nil {
		return infer.CreateResponse[TeamAccessTokenState]{}, fmt.Errorf(
			"error creating team access token %q: %w", req.Inputs.Name, err,
		)
	}
	return infer.CreateResponse[TeamAccessTokenState]{
		ID: teamAccessTokenID(req.Inputs.OrganizationName, req.Inputs.TeamName, req.Inputs.Name, token.ID),
		Output: TeamAccessTokenState{
			TeamAccessTokenInput: req.Inputs,
			Value:                token.TokenValue,
		},
	}, nil
}

func (*TeamAccessToken) Delete(
	ctx context.Context,
	req infer.DeleteRequest[TeamAccessTokenState],
) (infer.DeleteResponse, error) {
	orgName, teamName, _, tokenID, err := splitTeamAccessTokenID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteTeamAccessToken(ctx, tokenID, orgName, teamName)
}

func (*TeamAccessToken) Read(
	ctx context.Context,
	req infer.ReadRequest[TeamAccessTokenInput, TeamAccessTokenState],
) (infer.ReadResponse[TeamAccessTokenInput, TeamAccessTokenState], error) {
	orgName, teamName, tokenName, tokenID, err := splitTeamAccessTokenID(req.ID)
	if err != nil {
		return infer.ReadResponse[TeamAccessTokenInput, TeamAccessTokenState]{}, err
	}

	token, err := config.GetClient(ctx).GetTeamAccessToken(ctx, tokenID, orgName, teamName)
	if err != nil {
		return infer.ReadResponse[TeamAccessTokenInput, TeamAccessTokenState]{}, err
	}
	if token == nil {
		return infer.ReadResponse[TeamAccessTokenInput, TeamAccessTokenState]{}, nil
	}

	inputs := TeamAccessTokenInput{
		Name:             tokenName,
		OrganizationName: orgName,
		TeamName:         teamName,
		Description:      stringPtrIfNonEmpty(token.Description),
	}
	return infer.ReadResponse[TeamAccessTokenInput, TeamAccessTokenState]{
		ID:     req.ID,
		Inputs: inputs,
		State: TeamAccessTokenState{
			TeamAccessTokenInput: inputs,
			// Token values aren't retrievable from the API after creation; carry
			// the existing secret from state so refresh does not erase it.
			Value: req.State.Value,
		},
	}, nil
}

// StateMigrations strips the legacy "__inputs" outputs property that the
// pre-infer TeamAccessToken resource embedded in state. infer rejects unknown
// fields when decoding state, so without this migration a refresh against an
// existing stack errors with "Unrecognized field '__inputs'".
func (*TeamAccessToken) StateMigrations(context.Context) []infer.StateMigrationFunc[TeamAccessTokenState] {
	return []infer.StateMigrationFunc[TeamAccessTokenState]{
		infer.StateMigration(migrateTeamAccessTokenLegacyInputs),
	}
}

func migrateTeamAccessTokenLegacyInputs(
	_ context.Context, old property.Map,
) (infer.MigrationResult[TeamAccessTokenState], error) {
	if _, ok := old.GetOk("__inputs"); !ok {
		return infer.MigrationResult[TeamAccessTokenState]{}, nil
	}
	state := TeamAccessTokenState{}
	if v, ok := old.GetOk("name"); ok && v.IsString() {
		state.Name = v.AsString()
	}
	if v, ok := old.GetOk("organizationName"); ok && v.IsString() {
		state.OrganizationName = v.AsString()
	}
	if v, ok := old.GetOk("teamName"); ok && v.IsString() {
		state.TeamName = v.AsString()
	}
	if v, ok := old.GetOk("description"); ok && v.IsString() {
		s := v.AsString()
		state.Description = &s
	}
	if v, ok := old.GetOk("value"); ok && v.IsString() {
		state.Value = v.AsString()
	}
	return infer.MigrationResult[TeamAccessTokenState]{Result: &state}, nil
}

func teamAccessTokenID(org, team, name, tokenID string) string {
	return fmt.Sprintf("%s/%s/%s/%s", org, team, name, tokenID)
}

func splitTeamAccessTokenID(id string) (string, string, string, string, error) {
	// format: organization/teamName/tokenName/tokenID
	s := strings.Split(id, "/")
	if len(s) != 4 {
		return "", "", "", "", fmt.Errorf("%q is invalid, must be of the form organization/team/name/tokenId", id)
	}
	return s[0], s[1], s[2], s[3], nil
}

func stringPtrIfNonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
