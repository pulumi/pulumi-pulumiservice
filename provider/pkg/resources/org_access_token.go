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

type OrgAccessToken struct{}

var (
	_ infer.CustomCreate[OrgAccessTokenInput, OrgAccessTokenState] = &OrgAccessToken{}
	_ infer.CustomDelete[OrgAccessTokenState]                      = &OrgAccessToken{}
	_ infer.CustomRead[OrgAccessTokenInput, OrgAccessTokenState]   = &OrgAccessToken{}
	_ infer.CustomStateMigrations[OrgAccessTokenState]             = &OrgAccessToken{}
)

func (*OrgAccessToken) Annotate(a infer.Annotator) {
	a.Describe(
		&OrgAccessToken{},
		"The Pulumi Cloud allows users to create access tokens scoped to orgs. "+
			"Org access tokens is a resource to create them and assign them to an org",
	)
}

type OrgAccessTokenInput struct {
	Name             string  `pulumi:"name"                 provider:"replaceOnChanges"`
	OrganizationName string  `pulumi:"organizationName"     provider:"replaceOnChanges"`
	Description      *string `pulumi:"description,optional" provider:"replaceOnChanges"`
	Admin            *bool   `pulumi:"admin,optional"       provider:"replaceOnChanges"`
}

func (i *OrgAccessTokenInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Name, "The name for the token.")
	a.Describe(&i.OrganizationName, "The organization's name.")
	a.Describe(&i.Description, "Optional. Description for the token.")
	a.Describe(&i.Admin, "Optional. True if this is an admin token.")
	a.SetDefault(&i.Admin, false)
}

type OrgAccessTokenState struct {
	OrgAccessTokenInput
	Value string `pulumi:"value" provider:"secret"`
}

func (s *OrgAccessTokenState) Annotate(a infer.Annotator) {
	a.Describe(&s.Value, "The token's value.")
}

func (*OrgAccessToken) Create(
	ctx context.Context,
	req infer.CreateRequest[OrgAccessTokenInput],
) (infer.CreateResponse[OrgAccessTokenState], error) {
	if req.DryRun {
		return infer.CreateResponse[OrgAccessTokenState]{
			Output: OrgAccessTokenState{OrgAccessTokenInput: req.Inputs},
		}, nil
	}
	desc := ""
	if req.Inputs.Description != nil {
		desc = *req.Inputs.Description
	}
	admin := false
	if req.Inputs.Admin != nil {
		admin = *req.Inputs.Admin
	}
	token, err := config.GetClient(ctx).CreateOrgAccessToken(
		ctx, req.Inputs.Name, req.Inputs.OrganizationName, desc, admin,
	)
	if err != nil {
		return infer.CreateResponse[OrgAccessTokenState]{}, fmt.Errorf(
			"error creating org access token %q: %w", req.Inputs.Name, err,
		)
	}
	return infer.CreateResponse[OrgAccessTokenState]{
		ID: orgAccessTokenID(req.Inputs.OrganizationName, req.Inputs.Name, token.ID),
		Output: OrgAccessTokenState{
			OrgAccessTokenInput: req.Inputs,
			Value:               token.TokenValue,
		},
	}, nil
}

func (*OrgAccessToken) Delete(
	ctx context.Context,
	req infer.DeleteRequest[OrgAccessTokenState],
) (infer.DeleteResponse, error) {
	orgName, _, tokenID, err := splitOrgAccessTokenID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, err
	}
	return infer.DeleteResponse{}, config.GetClient(ctx).DeleteOrgAccessToken(ctx, tokenID, orgName)
}

func (*OrgAccessToken) Read(
	ctx context.Context,
	req infer.ReadRequest[OrgAccessTokenInput, OrgAccessTokenState],
) (infer.ReadResponse[OrgAccessTokenInput, OrgAccessTokenState], error) {
	orgName, _, tokenID, err := splitOrgAccessTokenID(req.ID)
	if err != nil {
		return infer.ReadResponse[OrgAccessTokenInput, OrgAccessTokenState]{}, err
	}

	token, err := config.GetClient(ctx).GetOrgAccessToken(ctx, tokenID, orgName)
	if err != nil {
		return infer.ReadResponse[OrgAccessTokenInput, OrgAccessTokenState]{}, err
	}
	if token == nil {
		return infer.ReadResponse[OrgAccessTokenInput, OrgAccessTokenState]{}, nil
	}

	admin := token.Admin
	inputs := OrgAccessTokenInput{
		Name:             token.Name,
		OrganizationName: orgName,
		Description:      stringPtrIfNonEmpty(token.Description),
		Admin:            &admin,
	}
	return infer.ReadResponse[OrgAccessTokenInput, OrgAccessTokenState]{
		ID:     req.ID,
		Inputs: inputs,
		State: OrgAccessTokenState{
			OrgAccessTokenInput: inputs,
			// Token values aren't retrievable from the API after creation; carry
			// the existing secret from state so refresh does not erase it.
			Value: req.State.Value,
		},
	}, nil
}

// StateMigrations strips the legacy "__inputs" outputs property that the
// pre-infer OrgAccessToken resource embedded in state. infer rejects unknown
// fields when decoding state, so without this migration a refresh against an
// existing stack errors with "Unrecognized field '__inputs'".
func (*OrgAccessToken) StateMigrations(context.Context) []infer.StateMigrationFunc[OrgAccessTokenState] {
	return []infer.StateMigrationFunc[OrgAccessTokenState]{
		infer.StateMigration(migrateOrgAccessTokenLegacyInputs),
	}
}

func migrateOrgAccessTokenLegacyInputs(
	_ context.Context, old property.Map,
) (infer.MigrationResult[OrgAccessTokenState], error) {
	if _, ok := old.GetOk("__inputs"); !ok {
		return infer.MigrationResult[OrgAccessTokenState]{}, nil
	}
	state := OrgAccessTokenState{}
	if v, ok := old.GetOk("name"); ok && v.IsString() {
		state.Name = v.AsString()
	}
	if v, ok := old.GetOk("organizationName"); ok && v.IsString() {
		state.OrganizationName = v.AsString()
	}
	if v, ok := old.GetOk("description"); ok && v.IsString() {
		s := v.AsString()
		state.Description = &s
	}
	if v, ok := old.GetOk("admin"); ok && v.IsBool() {
		b := v.AsBool()
		state.Admin = &b
	}
	if v, ok := old.GetOk("value"); ok && v.IsString() {
		state.Value = v.AsString()
	}
	return infer.MigrationResult[OrgAccessTokenState]{Result: &state}, nil
}

func orgAccessTokenID(org, name, tokenID string) string {
	return fmt.Sprintf("%s/%s/%s", org, name, tokenID)
}

func splitOrgAccessTokenID(id string) (string, string, string, error) {
	// format: organization/name/tokenID. Name may itself contain slashes,
	// so org is the first segment and tokenID is the last.
	s := strings.Split(id, "/")
	if len(s) < 3 {
		return "", "", "", fmt.Errorf(
			"%q is invalid, must be of the form organization/name/tokenId",
			id,
		)
	}
	org := s[0]
	tokenID := s[len(s)-1]
	name := strings.Join(s[1:len(s)-1], "/")
	return org, name, tokenID, nil
}
