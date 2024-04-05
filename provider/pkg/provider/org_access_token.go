package provider

import (
	"fmt"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

var (
	// Life-cycle participation
	_ infer.CustomResource[OrgAccessTokenInput, OrgAccessTokenState] = (*OrgAccessToken)(nil)
	_ infer.CustomDelete[OrgAccessTokenState]                        = (*OrgAccessToken)(nil)
	_ infer.CustomRead[OrgAccessTokenInput, OrgAccessTokenState]     = (*OrgAccessToken)(nil)

	// Schema documentation
	_ infer.Annotated = (*OrgAccessToken)(nil)
	_ infer.Annotated = (*OrgAccessTokenInput)(nil)
	_ infer.Annotated = (*OrgAccessTokenState)(nil)
)

type OrgAccessToken struct{}

func (p *OrgAccessToken) Annotate(a infer.Annotator) {
	a.Describe(p, "The Pulumi Cloud allows users to create access tokens scoped to orgs. "+
		"Org access tokens is a resource to create them and assign them to an org")
}

type OrgAccessTokenInput struct {
	OrgName     string `pulumi:"organizationName"`
	Description string `pulumi:"description,optional"`
	Name        string `pulumi:"name"`
	Admin       bool   `pulumi:"admin,optional"`
}

func (p *OrgAccessTokenInput) Annotate(a infer.Annotator) {
	a.Describe(&p.OrgName, "The organization's name.")
	a.Describe(&p.Description, "Optional. Description for the token.")
	a.Describe(&p.Name, "The name for the token.")
	a.Describe(&p.Admin, "Optional. True if this is an admin token.")
}

type OrgAccessTokenState struct {
	OrgAccessTokenInput
	Value string `pulumi:"value" provider:"secret"`
}

func (p *OrgAccessTokenState) Annotate(a infer.Annotator) {
	a.Describe(&p.Value, "The token's value.")
}

func (*OrgAccessToken) Delete(ctx p.Context, id string, props OrgAccessTokenState) error {
	orgName, _, tokenId, err := splitOrgAccessTokenId(id)
	if err != nil {
		return err
	}
	return GetConfig(ctx).Client.DeleteOrgAccessToken(ctx, tokenId, orgName)
}

func (*OrgAccessToken) Create(
	ctx p.Context, name string, input OrgAccessTokenInput, preview bool,
) (string, OrgAccessTokenState, error) {
	if preview {
		return "", OrgAccessTokenState{OrgAccessTokenInput: input}, nil
	}

	accessToken, err := GetConfig(ctx).Client.CreateOrgAccessToken(ctx,
		input.Name, input.OrgName, input.Description, input.Admin)
	if err != nil {
		return "", OrgAccessTokenState{}, err
	}

	id := fmt.Sprintf(input.OrgName + "/" + input.Name + "/" + accessToken.ID)

	return id, OrgAccessTokenState{OrgAccessTokenInput: input, Value: accessToken.TokenValue}, nil

}

func (*OrgAccessToken) Read(
	ctx p.Context, id string, inputs OrgAccessTokenInput, state OrgAccessTokenState,
) (string, OrgAccessTokenInput, OrgAccessTokenState, error) {
	orgName, tokenName, tokenId, err := splitOrgAccessTokenId(id)

	// the org access token is immutable; if we get nil it got deleted, otherwise all data is the same
	accessToken, err := GetConfig(ctx).Client.GetOrgAccessToken(ctx, tokenId, orgName)
	if err != nil || accessToken == nil {
		return "", OrgAccessTokenInput{}, OrgAccessTokenState{}, err
	}

	inputs.OrgName = orgName
	inputs.Description = accessToken.Description
	inputs.Name = tokenName

	// TODO[BUG]: Admin and Value are not recoverable from a import.

	return id, inputs, OrgAccessTokenState{
		OrgAccessTokenInput: inputs,
		Value:               state.Value,
	}, nil
}

func splitOrgAccessTokenId(id string) (string, string, string, error) {
	// format: organization/name/tokenId
	s := strings.Split(id, "/")
	if len(s) != 3 {
		return "", "", "", fmt.Errorf("%q is invalid, must contain a single slash ('/')", id)
	}
	return s[0], s[1], s[2], nil
}
