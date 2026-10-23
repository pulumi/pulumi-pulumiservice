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
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

const (
	gcTeamName = "teamName"
)

type TeamAccessToken struct{}

var (
	_ infer.CustomCheck[TeamAccessTokenInput]                        = &TeamAccessToken{}
	_ infer.CustomCreate[TeamAccessTokenInput, TeamAccessTokenState] = &TeamAccessToken{}
	_ infer.CustomDelete[TeamAccessTokenState]                       = &TeamAccessToken{}
	_ infer.CustomRead[TeamAccessTokenInput, TeamAccessTokenState]   = &TeamAccessToken{}
	_ infer.CustomStateMigrations[TeamAccessTokenState]              = &TeamAccessToken{}
)

func (*TeamAccessToken) Annotate(a infer.Annotator) {
	a.Describe(
		&TeamAccessToken{},
		"Creates a team access token in Pulumi Cloud. The token isn't tied to a user, and it has the "+
			"permissions granted to its team.\n\n"+
			"Unlike `OrgAccessToken`, a provider authenticated with an organization or team access token "+
			"can create this resource."+
			util.ImportDocs("pulumiservice:index:TeamAccessToken", "{organization}/{team}/{name}/{tokenId}",
				"my-org/my-team/my-token/5c7e0a92-1b3d-4f6e-8a24-9d1b7e3c0f58",
				util.APIIDHint("token ID", "/api/orgs/my-org/teams/my-team/tokens", ".tokens[] | {name, id}"),
				"The token's `value` can't be read back from Pulumi Cloud, so it's empty after import."),
	)
}

type TeamAccessTokenInput struct {
	Name             string  `pulumi:"name,optional"        provider:"replaceOnChanges"`
	OrganizationName string  `pulumi:"organizationName"     provider:"replaceOnChanges"`
	TeamName         string  `pulumi:"teamName"             provider:"replaceOnChanges"`
	Description      *string `pulumi:"description,optional" provider:"replaceOnChanges"`
}

func (i *TeamAccessTokenInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Name, tokenNameDescription)
	a.Describe(&i.OrganizationName, "The organization's name.")
	a.Describe(&i.TeamName, "The team name.")
	a.Describe(&i.Description, "A description of the token.")
}

type TeamAccessTokenState struct {
	TeamAccessTokenInput
	// name is optional to supply but always set once Check has run, so it's a
	// required output. Shadowing the embedded optional input keeps the output
	// type non-optional in the SDKs.
	Name  string `pulumi:"name"  provider:"replaceOnChanges"`
	Value string `pulumi:"value" provider:"secret"`
}

func (s *TeamAccessTokenState) Annotate(a infer.Annotator) {
	a.Describe(&s.Name, tokenNameDescription)
	a.Describe(&s.Value, "The token's value.")
}

// newTeamAccessTokenState builds the state from a set of inputs, populating
// the required name output that shadows the optional input.
func newTeamAccessTokenState(input TeamAccessTokenInput, value string) TeamAccessTokenState {
	return TeamAccessTokenState{TeamAccessTokenInput: input, Name: input.Name, Value: value}
}

func (*TeamAccessToken) Create(
	ctx context.Context,
	req infer.CreateRequest[TeamAccessTokenInput],
) (infer.CreateResponse[TeamAccessTokenState], error) {
	if req.DryRun {
		return infer.CreateResponse[TeamAccessTokenState]{
			Output: newTeamAccessTokenState(req.Inputs, ""),
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
		ID:     teamAccessTokenID(req.Inputs.OrganizationName, req.Inputs.TeamName, req.Inputs.Name, token.ID),
		Output: newTeamAccessTokenState(req.Inputs, token.TokenValue),
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
		// Token values aren't retrievable from the API after creation; carry
		// the existing secret from state so refresh does not erase it.
		State: newTeamAccessTokenState(inputs, req.State.Value),
	}, nil
}

// Check generates the token's name when the program omits it.
func (*TeamAccessToken) Check(
	ctx context.Context,
	req infer.CheckRequest,
) (infer.CheckResponse[TeamAccessTokenInput], error) {
	in, failures, err := infer.DefaultCheck[TeamAccessTokenInput](ctx, util.AutoName(ctx, req, gcName, tokenNameMaxLen))
	if err != nil {
		return infer.CheckResponse[TeamAccessTokenInput]{}, err
	}
	return infer.CheckResponse[TeamAccessTokenInput]{Inputs: in, Failures: failures}, nil
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
	if _, ok := old.GetOk(gcInputs); !ok {
		return infer.MigrationResult[TeamAccessTokenState]{}, nil
	}
	state := TeamAccessTokenState{}
	if v, ok := old.GetOk(gcName); ok && v.IsString() {
		state.Name = v.AsString()
		state.TeamAccessTokenInput.Name = v.AsString()
	}
	if v, ok := old.GetOk(gcOrganizationName); ok && v.IsString() {
		state.OrganizationName = v.AsString()
	}
	if v, ok := old.GetOk(gcTeamName); ok && v.IsString() {
		state.TeamName = v.AsString()
	}
	if v, ok := old.GetOk(gcDescription); ok && v.IsString() {
		s := v.AsString()
		state.Description = &s
	}
	if v, ok := old.GetOk(gcValue); ok && v.IsString() {
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
