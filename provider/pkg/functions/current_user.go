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

	"github.com/pulumi/pulumi-go-provider/infer"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

// GetCurrentUserFunction returns the Pulumi Cloud user that the provider's
// configured access token belongs to. Intended for cases where a Pulumi
// program needs to reference the caller — e.g. seeding a new `Team` with
// the creator as a member so refresh doesn't drift against the service's
// automatic team-creator membership.
type GetCurrentUserFunction struct{}

type GetCurrentUserInput struct{}

type GetCurrentUserOutput struct {
	Username  string `pulumi:"username"`
	Name      string `pulumi:"name"`
	Email     string `pulumi:"email"`
	AvatarUrl string `pulumi:"avatarUrl"`
}

func (GetCurrentUserFunction) Annotate(a infer.Annotator) {
	a.Describe(
		&GetCurrentUserFunction{},
		"Returns the Pulumi Cloud user that the provider's access token belongs to. "+
			"Useful for seeding a newly-created `Team` with the creator as a member, since "+
			"Pulumi Cloud auto-adds the creator and omitting them causes a refresh drift.",
	)
	a.SetToken("index", "getCurrentUser")
}

func (o *GetCurrentUserOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.Username, "The user's Pulumi Cloud username.")
	a.Describe(&o.Name, "The user's display name.")
	a.Describe(&o.Email, "The user's email address.")
	a.Describe(&o.AvatarUrl, "URL of the user's avatar image.")
}

func (GetCurrentUserFunction) Invoke(
	ctx context.Context,
	_ infer.FunctionRequest[GetCurrentUserInput],
) (infer.FunctionResponse[GetCurrentUserOutput], error) {
	user, err := config.GetClient(ctx).GetCurrentUser(ctx)
	if err != nil {
		return infer.FunctionResponse[GetCurrentUserOutput]{}, fmt.Errorf(
			"failed to resolve the current Pulumi Cloud user: %w",
			err,
		)
	}
	return infer.FunctionResponse[GetCurrentUserOutput]{
		Output: GetCurrentUserOutput{
			Username:  user.GithubLogin,
			Name:      user.Name,
			Email:     user.Email,
			AvatarUrl: user.AvatarURL,
		},
	}, nil
}
