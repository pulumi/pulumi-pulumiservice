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

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

type ServicePrincipalAccessToken struct{}

type (
	spTokenInput = ServicePrincipalAccessTokenInput
	spTokenState = ServicePrincipalAccessTokenState
)

var (
	_ infer.CustomCheck[spTokenInput]                = &ServicePrincipalAccessToken{}
	_ infer.CustomCreate[spTokenInput, spTokenState] = &ServicePrincipalAccessToken{}
	_ infer.CustomDelete[spTokenState]               = &ServicePrincipalAccessToken{}
	_ infer.CustomRead[spTokenInput, spTokenState]   = &ServicePrincipalAccessToken{}
)

func (*ServicePrincipalAccessToken) Annotate(a infer.Annotator) {
	a.Describe(
		&ServicePrincipalAccessToken{},
		"Creates an access token that authenticates as a service principal. Requests made with the token "+
			"have the service principal's roles and entity access, and the token isn't tied to a user. It can "+
			"replace an `OrgAccessToken`: give a `ServicePrincipal` the roles the automation needs, then "+
			"create a token for it.\n\n"+
			"The provider's credentials need permission to create organization tokens. Pulumi Cloud refuses "+
			"delegated credentials, such as tokens minted by OIDC token exchange, here."+
			util.ImportDocs("pulumiservice:index:ServicePrincipalAccessToken", "{organization}/{login}/{tokenId}",
				"my-org/service-principal:d333a711-4aa0-402f-be6d-72af9665fc37/a81d4c3e-52f0-4b7a-b9e2-6d0f1c8e2a47",
				util.APIIDHint("token ID", "/api/orgs/my-org/service-principals/{login}/tokens",
					".tokens[] | {name, id}"),
				"The token's `value` can't be read back from Pulumi Cloud, so it's empty after import."),
	)
}

type ServicePrincipalAccessTokenInput struct {
	Name                  string  `pulumi:"name"                  provider:"replaceOnChanges"`
	OrganizationName      string  `pulumi:"organizationName"      provider:"replaceOnChanges"`
	ServicePrincipalLogin string  `pulumi:"servicePrincipalLogin" provider:"replaceOnChanges"`
	Description           *string `pulumi:"description,optional"  provider:"replaceOnChanges"`
	Expires               *string `pulumi:"expires,optional"      provider:"replaceOnChanges"`
}

func (i *ServicePrincipalAccessTokenInput) Annotate(a infer.Annotator) {
	a.Describe(
		&i.Name,
		"The name for the token, up to 40 characters. It must be unique among the service principal's "+
			"tokens, and Pulumi Cloud never frees a deleted token's name.\n\n"+
			"If omitted, Pulumi generates a name from the resource's logical name plus a random suffix, or "+
			"follows the stack's [autonaming configuration]"+
			"(https://www.pulumi.com/docs/iac/concepts/resources/names/#autonaming). A replacement gets a new "+
			"name.\n\n"+
			"Every input on this resource triggers a replacement, so when `name` is set explicitly, changing any "+
			"other input fails with `409 Conflict: token name already exists`. Omit `name`, or change it along "+
			"with the other input.",
	)
	a.Describe(&i.OrganizationName, "The organization's name.")
	a.Describe(
		&i.ServicePrincipalLogin,
		"The login of the service principal the token authenticates as: the `login` output of a "+
			"`ServicePrincipal` resource.",
	)
	a.Describe(&i.Description, "A description of the token.")
	a.Describe(
		&i.Expires,
		"When the token expires, as an RFC 3339 date-time such as `2027-06-30T17:00:00Z`. Omit it for a "+
			"token that never expires. Pulumi Cloud rejects an expiry that is in the past, more than two "+
			"years away, or beyond the organization's maximum token lifetime.",
	)
}

func (i *ServicePrincipalAccessTokenInput) AutoName(autoname infer.SetAutoName) {
	autoname(&i.Name, tokenAutoName)
}

type ServicePrincipalAccessTokenState struct {
	ServicePrincipalAccessTokenInput
	Value string `pulumi:"value" provider:"secret"`
}

func (s *ServicePrincipalAccessTokenState) Annotate(a infer.Annotator) {
	a.Describe(&s.Value, "The token's value.")
}

func (*ServicePrincipalAccessToken) Check(
	ctx context.Context,
	req infer.CheckRequest,
) (infer.CheckResponse[spTokenInput], error) {
	in, failures, err := infer.DefaultCheck[spTokenInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[spTokenInput]{}, err
	}
	if in.Expires != nil && !isUnknownInput(req.NewInputs, gcExpires) {
		if _, err := parseTokenExpiry(*in.Expires); err != nil {
			failures = append(failures, p.CheckFailure{Property: gcExpires, Reason: err.Error()})
		}
	}
	return infer.CheckResponse[spTokenInput]{Inputs: in, Failures: failures}, nil
}

func (*ServicePrincipalAccessToken) Create(
	ctx context.Context,
	req infer.CreateRequest[spTokenInput],
) (infer.CreateResponse[spTokenState], error) {
	if req.DryRun {
		return infer.CreateResponse[spTokenState]{
			Output: spTokenState{ServicePrincipalAccessTokenInput: req.Inputs},
		}, nil
	}
	createReq := pulumiapi.CreateServicePrincipalTokenRequest{
		Name:        req.Inputs.Name,
		Description: util.OrZero(req.Inputs.Description),
	}
	if req.Inputs.Expires != nil {
		t, err := parseTokenExpiry(*req.Inputs.Expires)
		if err != nil {
			return infer.CreateResponse[spTokenState]{}, err
		}
		utc := t.UTC()
		createReq.Expires = &utc
	}
	token, err := config.GetClient(ctx).CreateServicePrincipalToken(
		ctx, req.Inputs.OrganizationName, req.Inputs.ServicePrincipalLogin, createReq,
	)
	if err != nil {
		return infer.CreateResponse[spTokenState]{}, fmt.Errorf(
			"error creating service principal access token %q: %w", req.Inputs.Name, err,
		)
	}
	return infer.CreateResponse[spTokenState]{
		ID: spTokenID(req.Inputs.OrganizationName, req.Inputs.ServicePrincipalLogin, token.ID),
		Output: spTokenState{
			ServicePrincipalAccessTokenInput: req.Inputs,
			Value:                            token.TokenValue,
		},
	}, nil
}

func (*ServicePrincipalAccessToken) Delete(
	ctx context.Context,
	req infer.DeleteRequest[spTokenState],
) (infer.DeleteResponse, error) {
	org, login, tokenID, err := splitSPTokenID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteServicePrincipalToken(ctx, org, login, tokenID)
}

func (*ServicePrincipalAccessToken) Read(
	ctx context.Context,
	req infer.ReadRequest[spTokenInput, spTokenState],
) (infer.ReadResponse[spTokenInput, spTokenState], error) {
	type response = infer.ReadResponse[spTokenInput, spTokenState]
	org, login, tokenID, err := splitSPTokenID(req.ID)
	if err != nil {
		return response{}, err
	}
	token, err := config.GetClient(ctx).GetServicePrincipalToken(ctx, org, login, tokenID)
	if err != nil {
		return response{}, err
	}
	if token == nil {
		return response{}, nil
	}
	inputs := spTokenInput{
		Name:                  token.Name,
		OrganizationName:      org,
		ServicePrincipalLogin: login,
		Description:           stringPtrIfNonEmpty(token.Description),
		Expires:               readTokenExpiry(token.Expires, req.State.Expires),
	}
	return response{
		ID:     req.ID,
		Inputs: inputs,
		State: spTokenState{
			ServicePrincipalAccessTokenInput: inputs,
			// Token values aren't retrievable from the API after creation; carry
			// the existing secret from state so refresh does not erase it.
			Value: req.State.Value,
		},
	}, nil
}

func spTokenID(org, login, tokenID string) string {
	return strings.Join([]string{org, login, tokenID}, "/")
}

func splitSPTokenID(id string) (string, string, string, error) {
	s := strings.Split(id, "/")
	if len(s) != 3 || s[0] == "" || s[1] == "" || s[2] == "" {
		return "", "", "", fmt.Errorf("%q is invalid, must be of the form organization/login/tokenId", id)
	}
	return s[0], s[1], s[2], nil
}
