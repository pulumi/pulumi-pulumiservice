// Copyright 2016-2025, Pulumi Corporation.
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
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/resources"
)

// GetInsightsAccountsFunction is an invoke function to list all Insights accounts
type GetInsightsAccountsFunction struct{}

type GetInsightsAccountsInput struct {
	OrganizationName string `pulumi:"organizationName"`
}

type GetInsightsAccountsOutput struct {
	Accounts []resources.InsightsAccountState `pulumi:"accounts"`
}

func (GetInsightsAccountsFunction) Annotate(a infer.Annotator) {
	a.Describe(&GetInsightsAccountsFunction{}, "Get a list of all Insights accounts for an organization.")
	a.SetToken("index", "getInsightsAccounts")
}

func (GetInsightsAccountsFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetInsightsAccountsInput],
) (infer.FunctionResponse[GetInsightsAccountsOutput], error) {
	client := config.GetClient(ctx)

	accounts, err := client.ListInsightsAccounts(ctx, req.Input.OrganizationName)
	if err != nil {
		return infer.FunctionResponse[GetInsightsAccountsOutput]{}, fmt.Errorf(
			"failed to list insights accounts: %w",
			err,
		)
	}

	outputAccounts := make([]resources.InsightsAccountState, len(accounts))
	for i, account := range accounts {
		outputAccounts[i] = resources.InsightsAccountStateFromAPI(req.Input.OrganizationName, account)
	}

	return infer.FunctionResponse[GetInsightsAccountsOutput]{
		Output: GetInsightsAccountsOutput{
			Accounts: outputAccounts,
		},
	}, nil
}

// GetInsightsAccountFunction is an invoke function to get a specific Insights account
type GetInsightsAccountFunction struct{}

type GetInsightsAccountInput struct {
	OrganizationName string `pulumi:"organizationName"`
	AccountName      string `pulumi:"accountName"`
}

func (GetInsightsAccountFunction) Annotate(a infer.Annotator) {
	a.Describe(&GetInsightsAccountFunction{}, "Get details about a specific Insights account.")
	a.SetToken("index", "getInsightsAccount")
}

func (GetInsightsAccountFunction) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetInsightsAccountInput],
) (infer.FunctionResponse[resources.InsightsAccountState], error) {
	client := config.GetClient(ctx)

	account, err := client.GetInsightsAccount(ctx, req.Input.OrganizationName, req.Input.AccountName)
	if err != nil {
		return infer.FunctionResponse[resources.InsightsAccountState]{}, fmt.Errorf(
			"failed to get insights account: %w",
			err,
		)
	}

	if account == nil {
		return infer.FunctionResponse[resources.InsightsAccountState]{}, fmt.Errorf(
			"insights account %q not found",
			req.Input.AccountName,
		)
	}

	return infer.FunctionResponse[resources.InsightsAccountState]{
		Output: resources.InsightsAccountStateFromAPI(req.Input.OrganizationName, *account),
	}, nil
}
