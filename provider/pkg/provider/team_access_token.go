package provider

import (
	"fmt"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

var (
	// Life-cycle participation
	_ infer.CustomResource[TeamAccessTokenInput, TeamAccessTokenState] = (*TeamAccessToken)(nil)
	_ infer.CustomDelete[TeamAccessTokenState]                         = (*TeamAccessToken)(nil)
	_ infer.CustomRead[TeamAccessTokenInput, TeamAccessTokenState]     = (*TeamAccessToken)(nil)

	// Schema documentation
	_ infer.Annotated = (*TeamAccessToken)(nil)
	_ infer.Annotated = (*TeamAccessTokenInput)(nil)
	_ infer.Annotated = (*TeamAccessTokenState)(nil)

	// Secret values
	_ infer.ExplicitDependencies[TeamAccessTokenInput, TeamAccessTokenState] = (*TeamAccessToken)(nil)
)

type TeamAccessToken struct{}

func (p *TeamAccessToken) Annotate(a infer.Annotator) {
	a.Describe(p, "The Pulumi Cloud allows users to create access tokens scoped to team. "+
		"Team access tokens is a resource to create them and assign them to a team")
}

func (p *TeamAccessToken) WireDependencies(
	f infer.FieldSelector, args *TeamAccessTokenInput, state *TeamAccessTokenState,
) {
	f.OutputField(&state.Value).AlwaysSecret()
}

type TeamAccessTokenInput struct {
	Name        string `pulumi:"name"`
	OrgName     string `pulumi:"organizationName"`
	TeamName    string `pulumi:"teamName"`
	Description string `pulumi:"description,optional"`
}

func (p *TeamAccessTokenInput) Annotate(a infer.Annotator) {
	// TODO: Fix typo in description
	a.Describe(&p.Name, "The name for the token. "+
		"This must be unique amongst all machine tokens within your organization.")
	a.Describe(&p.OrgName, "The organization's name.")
	a.Describe(&p.TeamName, "The team name.")
	a.Describe(&p.Description, "Optional. Description for the token.")
}

type TeamAccessTokenState struct {
	TeamAccessTokenInput

	Value string `pulumi:"value"`
}

func (p *TeamAccessTokenState) Annotate(a infer.Annotator) {
	a.Describe(&p.Value, "The token's value.")
}

func (*TeamAccessToken) Delete(ctx p.Context, id string, state TeamAccessTokenState) error {
	_, _, _, tokenId, err := splitTeamAccessTokenId(id)
	if err != nil {
		return err
	}
	return GetConfig(ctx).Client.DeleteTeamAccessToken(ctx, tokenId, state.OrgName, state.TeamName)
}

func (*TeamAccessToken) Create(
	ctx p.Context, name string, input TeamAccessTokenInput, preview bool,
) (string, TeamAccessTokenState, error) {
	if preview {
		return "", TeamAccessTokenState{TeamAccessTokenInput: input}, nil
	}

	accessToken, err := GetConfig(ctx).Client.CreateTeamAccessToken(ctx,
		input.Name, input.OrgName, input.TeamName, input.Description)
	if err != nil {
		return "", TeamAccessTokenState{}, err
	}

	id := fmt.Sprintf(input.OrgName + "/" + input.TeamName + "/" + input.Name + "/" + accessToken.ID)

	return id, TeamAccessTokenState{
		TeamAccessTokenInput: input,
		Value:                accessToken.TokenValue,
	}, nil

}

func (*TeamAccessToken) Read(
	ctx p.Context, id string, inputs TeamAccessTokenInput, state TeamAccessTokenState,
) (string, TeamAccessTokenInput, TeamAccessTokenState, error) {

	orgName, teamName, tokenName, tokenId, err := splitTeamAccessTokenId(id)
	if err != nil {
		return "", TeamAccessTokenInput{}, TeamAccessTokenState{}, err
	}

	// the team access token is immutable; if we get nil it got deleted
	accessToken, err := GetConfig(ctx).Client.GetTeamAccessToken(ctx, tokenId, orgName, teamName)
	if err != nil || accessToken == nil {
		return "", TeamAccessTokenInput{}, TeamAccessTokenState{}, err
	}

	input := TeamAccessTokenInput{
		Name:        tokenName,
		OrgName:     orgName,
		TeamName:    teamName,
		Description: accessToken.Description,
	}

	return id, input, TeamAccessTokenState{
		TeamAccessTokenInput: input,

		// TODO[BUG]: GetTeamAccessToken doesn't set the token value. Since the
		// value is immutable, this works for Refresh. This does not work for
		// import.
		//
		// TODO[DOCS]: That .Value is invalid after import should at least be
		// mentioned in the docs.
		Value: state.Value,
	}, nil
}

func splitTeamAccessTokenId(id string) (string, string, string, string, error) {
	// format: organization/teamName/tokenName/tokenId
	s := strings.Split(id, "/")
	if len(s) != 4 {
		return "", "", "", "", fmt.Errorf("%q is invalid, must contain a single slash ('/')", id)
	}
	return s[0], s[1], s[2], s[3], nil
}
