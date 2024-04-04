package provider

import (
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

var (
	// Life-cycle participation
	_ infer.CustomResource[AccessTokenInput, AccessTokenState] = (*AccessToken)(nil)
	_ infer.CustomDelete[AccessTokenState]                     = (*AccessToken)(nil)
	_ infer.CustomRead[AccessTokenInput, AccessTokenState]     = (*AccessToken)(nil)

	// Schema documentation
	_ infer.Annotated = (*AccessToken)(nil)
	_ infer.Annotated = (*AccessTokenInput)(nil)
	_ infer.Annotated = (*AccessTokenState)(nil)

	// Secret values
	_ infer.ExplicitDependencies[AccessTokenInput, AccessTokenState] = (*AccessToken)(nil)
)

type AccessToken struct{}

func (p *AccessToken) Annotate(a infer.Annotator) {
	a.Describe(p, "Access tokens allow a user to authenticate against the Pulumi Cloud")
}

type AccessTokenInput struct {
	Description string `pulumi:"description" provider:"replaceOnChanges"`
}

func (p *AccessTokenInput) Annotate(a infer.Annotator) {
	a.Describe(&p.Description, "Description of the access token.")
}

type AccessTokenState struct {
	AccessTokenInput
	TokenID string `pulumi:"tokenId"`
	Value   string `pulumi:"value"`
}

func (p *AccessTokenState) Annotate(a infer.Annotator) {
	a.Describe(&p.TokenID, "The token identifier.")
	a.Describe(&p.Value, "The token's value.")
}

func (*AccessToken) WireDependencies(f infer.FieldSelector, args *AccessTokenInput, state *AccessTokenState) {
	f.OutputField(&state.Value).AlwaysSecret()
}

func (*AccessToken) Delete(ctx p.Context, id string, props AccessTokenState) error {
	return GetConfig(ctx).Client.DeleteAccessToken(ctx, props.TokenID)
}

func (*AccessToken) Create(ctx p.Context, name string, input AccessTokenInput, preview bool) (id string, output AccessTokenState, err error) {
	tk, err := GetConfig(ctx).Client.CreateAccessToken(ctx, input.Description)
	if err != nil {
		return "", AccessTokenState{}, err
	}
	return tk.ID, AccessTokenState{
		AccessTokenInput: input,
		TokenID:          id,
		Value:            tk.TokenValue,
	}, nil
}

func (*AccessToken) Read(
	ctx p.Context, id string, _ AccessTokenInput, _ AccessTokenState,
) (string, AccessTokenInput, AccessTokenState, error) {
	// the access token is immutable; if we get nil it got deleted, otherwise all data is the same
	accessToken, err := GetConfig(ctx).Client.GetAccessToken(ctx, id)
	if err != nil {
		return "", AccessTokenInput{}, AccessTokenState{}, err
	}
	if accessToken == nil {
		return "", AccessTokenInput{}, AccessTokenState{}, nil
	}

	inputs := AccessTokenInput{
		Description: accessToken.Description,
	}

	return id, inputs, AccessTokenState{
		AccessTokenInput: inputs,
		TokenID:          accessToken.ID,
		Value:            accessToken.TokenValue,
	}, err
}
