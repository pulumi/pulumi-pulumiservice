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

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

const (
	gcInputs      = "__inputs"
	gcDescription = "description"
	gcValue       = "value"
)

type AccessToken struct{}

var (
	_ infer.CustomCheck[AccessTokenInput]                    = &AccessToken{}
	_ infer.CustomCreate[AccessTokenInput, AccessTokenState] = &AccessToken{}
	_ infer.CustomDelete[AccessTokenState]                   = &AccessToken{}
	_ infer.CustomRead[AccessTokenInput, AccessTokenState]   = &AccessToken{}
	_ infer.CustomStateMigrations[AccessTokenState]          = &AccessToken{}
)

func (*AccessToken) Annotate(a infer.Annotator) {
	a.Describe(
		&AccessToken{},
		"Creates a personal access token for the Pulumi Cloud user that the provider is authenticated as. "+
			"The token acts with that user's permissions.\n\n"+
			"To create a token that isn't tied to a user, use `OrgAccessToken` or `TeamAccessToken`."+
			util.ImportDocs("pulumiservice:index:AccessToken", "{tokenId}",
				"3f2b9c1e-7d4a-4e8b-9a61-0c5d2e8f7b13",
				util.APIIDHint("token ID", "/api/user/tokens", ".tokens[] | {name, id}"),
				"The token's `value` can't be read back from Pulumi Cloud, so it's empty after import."),
	)
}

type AccessTokenInput struct {
	Description string  `pulumi:"description"      provider:"replaceOnChanges"`
	Expires     *string `pulumi:"expires,optional" provider:"replaceOnChanges"`
}

func (i *AccessTokenInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Description, "Description of the access token.")
	a.Describe(
		&i.Expires,
		"When the token expires, as an RFC 3339 date-time such as `2027-06-30T17:00:00Z`. Omit it for a "+
			"token that never expires. Pulumi Cloud rejects an expiry that is in the past or more than two "+
			"years away.",
	)
}

type AccessTokenState struct {
	AccessTokenInput
	Value string `pulumi:"value" provider:"secret"`
}

func (s *AccessTokenState) Annotate(a infer.Annotator) {
	a.Describe(&s.Value, "The token's value.")
}

func (*AccessToken) Create(
	ctx context.Context,
	req infer.CreateRequest[AccessTokenInput],
) (infer.CreateResponse[AccessTokenState], error) {
	if req.DryRun {
		return infer.CreateResponse[AccessTokenState]{
			Output: AccessTokenState{AccessTokenInput: req.Inputs},
		}, nil
	}
	var expires int64
	if req.Inputs.Expires != nil {
		t, err := parseTokenExpiry(*req.Inputs.Expires)
		if err != nil {
			return infer.CreateResponse[AccessTokenState]{}, err
		}
		expires = t.Unix()
	}
	token, err := config.GetClient(ctx).CreateAccessToken(ctx, req.Inputs.Description, expires)
	if err != nil {
		return infer.CreateResponse[AccessTokenState]{}, fmt.Errorf(
			"error creating access token %q: %w", req.Inputs.Description, err,
		)
	}
	return infer.CreateResponse[AccessTokenState]{
		ID: token.ID,
		Output: AccessTokenState{
			AccessTokenInput: req.Inputs,
			Value:            token.TokenValue,
		},
	}, nil
}

func (*AccessToken) Delete(
	ctx context.Context,
	req infer.DeleteRequest[AccessTokenState],
) (infer.DeleteResponse, error) {
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteAccessToken(ctx, req.ID)
}

func (*AccessToken) Read(
	ctx context.Context,
	req infer.ReadRequest[AccessTokenInput, AccessTokenState],
) (infer.ReadResponse[AccessTokenInput, AccessTokenState], error) {
	token, err := config.GetClient(ctx).GetAccessToken(ctx, req.ID)
	if err != nil {
		return infer.ReadResponse[AccessTokenInput, AccessTokenState]{}, err
	}
	if token == nil {
		return infer.ReadResponse[AccessTokenInput, AccessTokenState]{}, nil
	}
	inputs := AccessTokenInput{
		Description: token.Description,
		Expires:     readTokenExpiry(token.Expires, req.State.Expires),
	}
	return infer.ReadResponse[AccessTokenInput, AccessTokenState]{
		ID:     req.ID,
		Inputs: inputs,
		State: AccessTokenState{
			AccessTokenInput: inputs,
			// The list-tokens API does not return token values; carry the existing
			// secret from state so refresh does not erase it.
			Value: req.State.Value,
		},
	}, nil
}

func (*AccessToken) Check(
	ctx context.Context,
	req infer.CheckRequest,
) (infer.CheckResponse[AccessTokenInput], error) {
	in, failures, err := infer.DefaultCheck[AccessTokenInput](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[AccessTokenInput]{}, err
	}
	if in.Expires != nil && !isUnknownInput(req.NewInputs, gcExpires) {
		if _, err := parseTokenExpiry(*in.Expires); err != nil {
			failures = append(failures, p.CheckFailure{Property: gcExpires, Reason: err.Error()})
		}
	}
	return infer.CheckResponse[AccessTokenInput]{Inputs: in, Failures: failures}, nil
}

// StateMigrations strips the legacy "__inputs" outputs property that the
// pre-infer AccessToken resource embedded in state. infer rejects unknown
// fields when decoding state, so without this migration a refresh against an
// existing stack errors with "Unrecognized field '__inputs'".
func (*AccessToken) StateMigrations(context.Context) []infer.StateMigrationFunc[AccessTokenState] {
	return []infer.StateMigrationFunc[AccessTokenState]{
		infer.StateMigration(migrateAccessTokenLegacyInputs),
	}
}

func migrateAccessTokenLegacyInputs(
	_ context.Context, old property.Map,
) (infer.MigrationResult[AccessTokenState], error) {
	if _, ok := old.GetOk(gcInputs); !ok {
		return infer.MigrationResult[AccessTokenState]{}, nil
	}
	state := AccessTokenState{}
	if v, ok := old.GetOk(gcDescription); ok && v.IsString() {
		state.Description = v.AsString()
	}
	if v, ok := old.GetOk(gcValue); ok && v.IsString() {
		state.Value = v.AsString()
	}
	return infer.MigrationResult[AccessTokenState]{Result: &state}, nil
}
