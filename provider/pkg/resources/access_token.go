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

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

type AccessToken struct{}

var (
	_ infer.CustomCreate[AccessTokenInput, AccessTokenState] = &AccessToken{}
	_ infer.CustomDelete[AccessTokenState]                   = &AccessToken{}
	_ infer.CustomRead[AccessTokenInput, AccessTokenState]   = &AccessToken{}
	_ infer.CustomStateMigrations[AccessTokenState]          = &AccessToken{}
)

func (*AccessToken) Annotate(a infer.Annotator) {
	a.Describe(&AccessToken{}, "Access tokens allow a user to authenticate against the Pulumi Cloud.")
}

type AccessTokenInput struct {
	Description string `pulumi:"description" provider:"replaceOnChanges"`
}

func (i *AccessTokenInput) Annotate(a infer.Annotator) {
	a.Describe(&i.Description, "Description of the access token.")
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
	token, err := config.GetClient(ctx).CreateAccessToken(ctx, req.Inputs.Description)
	if err != nil {
		return infer.CreateResponse[AccessTokenState]{}, fmt.Errorf(
			"error creating access token %q: %w", req.Inputs.Description, err,
		)
	}
	return infer.CreateResponse[AccessTokenState]{
		ID: token.ID,
		Output: AccessTokenState{
			AccessTokenInput: AccessTokenInput{Description: req.Inputs.Description},
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
	inputs := AccessTokenInput{Description: token.Description}
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
	if _, ok := old.GetOk("__inputs"); !ok {
		return infer.MigrationResult[AccessTokenState]{}, nil
	}
	state := AccessTokenState{}
	if v, ok := old.GetOk("description"); ok && v.IsString() {
		state.Description = v.AsString()
	}
	if v, ok := old.GetOk("value"); ok && v.IsString() {
		state.Value = v.AsString()
	}
	return infer.MigrationResult[AccessTokenState]{Result: &state}, nil
}

// diffAccessTokenProperties is shared with the legacy gRPC-style access-token
// resources (org_access_token.go, team_access_token.go) which still rely on the
// "__inputs" property layout. Remove this when those resources are migrated.
func diffAccessTokenProperties(req *pulumirpc.DiffRequest, replaceProps []string) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOlds(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: false})
	if err != nil {
		return nil, err
	}

	inputs, ok := olds["__inputs"]
	if !ok {
		return nil, fmt.Errorf("missing __inputs property")
	}
	diffs := inputs.ObjectValue().Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	changes, replaces := pulumirpc.DiffResponse_DIFF_NONE, []string(nil)
	for _, k := range replaceProps {
		if diffs.Changed(resource.PropertyKey(k)) {
			changes = pulumirpc.DiffResponse_DIFF_SOME
			replaces = append(replaces, k)
		}
	}

	return &pulumirpc.DiffResponse{
		Changes:  changes,
		Replaces: replaces,
	}, nil
}
