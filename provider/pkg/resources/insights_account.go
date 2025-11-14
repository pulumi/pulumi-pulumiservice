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

package resources

import (
	"context"
	"fmt"
	"strings"

	pbempty "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

type PulumiServiceInsightsAccountResource struct {
	Client pulumiapi.InsightsAccountClient
}

type PulumiServiceInsightsAccountInput struct {
	OrgName        string
	AccountName    string
	Provider       string
	Environment    string
	Cron           string
	ProviderConfig map[string]interface{}
}

func (ia *PulumiServiceInsightsAccountResource) Name() string {
	return "pulumiservice:index:InsightsAccount"
}

func (ia *PulumiServiceInsightsAccountResource) ToPulumiServiceInsightsAccountInput(inputMap resource.PropertyMap) PulumiServiceInsightsAccountInput {
	input := PulumiServiceInsightsAccountInput{}

	if inputMap["organizationName"].HasValue() && inputMap["organizationName"].IsString() {
		input.OrgName = inputMap["organizationName"].StringValue()
	}

	if inputMap["accountName"].HasValue() && inputMap["accountName"].IsString() {
		input.AccountName = inputMap["accountName"].StringValue()
	}

	if inputMap["provider"].HasValue() && inputMap["provider"].IsString() {
		input.Provider = inputMap["provider"].StringValue()
	}

	if inputMap["environment"].HasValue() && inputMap["environment"].IsString() {
		input.Environment = inputMap["environment"].StringValue()
	}

	if inputMap["cron"].HasValue() && inputMap["cron"].IsString() {
		input.Cron = inputMap["cron"].StringValue()
	}

	if inputMap["providerConfig"].HasValue() && inputMap["providerConfig"].IsObject() {
		input.ProviderConfig = inputMap["providerConfig"].ObjectValue().Mappable()
	}

	return input
}

func GenerateInsightsAccountProperties(input PulumiServiceInsightsAccountInput, account pulumiapi.InsightsAccount) (*structpb.Struct, *structpb.Struct, error) {
	inputMap := resource.PropertyMap{}
	inputMap["organizationName"] = resource.NewPropertyValue(input.OrgName)
	inputMap["accountName"] = resource.NewPropertyValue(input.AccountName)
	inputMap["provider"] = resource.NewPropertyValue(input.Provider)
	inputMap["environment"] = resource.NewPropertyValue(input.Environment)

	if input.Cron != "" {
		inputMap["cron"] = resource.NewPropertyValue(input.Cron)
	}

	if input.ProviderConfig != nil {
		inputMap["providerConfig"] = resource.NewPropertyValue(input.ProviderConfig)
	}

	outputMap := resource.PropertyMap{}
	outputMap["insightsAccountId"] = resource.NewPropertyValue(account.ID)
	outputMap["organizationName"] = inputMap["organizationName"]
	outputMap["accountName"] = inputMap["accountName"]
	outputMap["provider"] = inputMap["provider"]
	outputMap["environment"] = inputMap["environment"]
	outputMap["scheduledScanEnabled"] = resource.NewPropertyValue(account.ScheduledScanEnabled)

	if input.Cron != "" {
		outputMap["cron"] = inputMap["cron"]
	}

	if account.ProviderVersion != "" {
		outputMap["providerVersion"] = resource.NewPropertyValue(account.ProviderVersion)
	}

	if input.ProviderConfig != nil {
		outputMap["providerConfig"] = inputMap["providerConfig"]
	}

	inputs, err := plugin.MarshalProperties(inputMap, plugin.MarshalOptions{})
	if err != nil {
		return nil, nil, err
	}

	outputs, err := plugin.MarshalProperties(outputMap, plugin.MarshalOptions{})
	if err != nil {
		return nil, nil, err
	}

	return outputs, inputs, nil
}

func (ia *PulumiServiceInsightsAccountResource) Check(req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return &pulumirpc.CheckResponse{Inputs: req.News, Failures: nil}, nil
}

func (ia *PulumiServiceInsightsAccountResource) Diff(req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	olds, err := plugin.UnmarshalProperties(req.GetOldInputs(), plugin.MarshalOptions{KeepUnknowns: false, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	diffs := olds.Diff(news)
	if diffs == nil {
		return &pulumirpc.DiffResponse{
			Changes: pulumirpc.DiffResponse_DIFF_NONE,
		}, nil
	}

	dd := plugin.NewDetailedDiffFromObjectDiff(diffs, false)

	detailedDiffs := map[string]*pulumirpc.PropertyDiff{}
	replaceProperties := map[string]bool{
		"organizationName": true,
		"accountName":      true,
		"provider":         true,
	}
	for k, v := range dd {
		if _, ok := replaceProperties[k]; ok {
			v.Kind = v.Kind.AsReplace()
		}
		detailedDiffs[k] = &pulumirpc.PropertyDiff{
			Kind:      pulumirpc.PropertyDiff_Kind(v.Kind),
			InputDiff: v.InputDiff,
		}
	}

	changes := pulumirpc.DiffResponse_DIFF_NONE
	if len(detailedDiffs) > 0 {
		changes = pulumirpc.DiffResponse_DIFF_SOME
	}

	return &pulumirpc.DiffResponse{
		Changes:         changes,
		DetailedDiff:    detailedDiffs,
		HasDetailedDiff: true,
	}, nil
}

func (ia *PulumiServiceInsightsAccountResource) Create(req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	ctx := context.Background()
	inputMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	input := ia.ToPulumiServiceInsightsAccountInput(inputMap)

	createReq := pulumiapi.CreateInsightsAccountRequest{
		Provider:       input.Provider,
		Environment:    input.Environment,
		Cron:           input.Cron,
		ProviderConfig: input.ProviderConfig,
	}

	err = ia.Client.CreateInsightsAccount(ctx, input.OrgName, input.AccountName, createReq)
	if err != nil {
		return nil, fmt.Errorf("error creating insights account '%s': %w", input.AccountName, err)
	}

	account, err := ia.Client.GetInsightsAccount(ctx, input.OrgName, input.AccountName)
	if err != nil {
		return nil, fmt.Errorf("error reading insights account after creation: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("insights account '%s' not found after creation", input.AccountName)
	}

	outputProperties, _, err := GenerateInsightsAccountProperties(input, *account)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.CreateResponse{
		Id:         input.OrgName + "/" + input.AccountName,
		Properties: outputProperties,
	}, nil
}

func (ia *PulumiServiceInsightsAccountResource) Read(req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	ctx := context.Background()
	urn := req.GetId()

	orgName, accountName, err := splitInsightsAccountId(urn)
	if err != nil {
		return nil, err
	}

	account, err := ia.Client.GetInsightsAccount(ctx, orgName, accountName)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return &pulumirpc.ReadResponse{}, nil
	}

	propertyMap, err := plugin.UnmarshalProperties(req.GetProperties(), plugin.MarshalOptions{KeepSecrets: true})
	if err != nil {
		return nil, err
	}

	input := PulumiServiceInsightsAccountInput{
		OrgName:     orgName,
		AccountName: accountName,
		Provider:    account.Provider,
		Environment: account.ProviderEnvRef,
	}

	// Only include providerConfig if it was in the original inputs
	if propertyMap["providerConfig"].HasValue() {
		input.ProviderConfig = account.ProviderConfig
	}

	if propertyMap["cron"].HasValue() && propertyMap["cron"].IsString() {
		input.Cron = propertyMap["cron"].StringValue()
	}

	outputProperties, inputs, err := GenerateInsightsAccountProperties(input, *account)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.ReadResponse{
		Id:         req.GetId(),
		Properties: outputProperties,
		Inputs:     inputs,
	}, nil
}

func (ia *PulumiServiceInsightsAccountResource) Update(req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	ctx := context.Background()

	orgName, accountName, err := splitInsightsAccountId(req.GetId())
	if err != nil {
		return nil, fmt.Errorf("invalid resource id: %w", err)
	}

	news, err := plugin.UnmarshalProperties(req.GetNews(), plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
	if err != nil {
		return nil, err
	}

	input := ia.ToPulumiServiceInsightsAccountInput(news)

	updateReq := pulumiapi.UpdateInsightsAccountRequest{
		Environment:    input.Environment,
		Cron:           input.Cron,
		ProviderConfig: input.ProviderConfig,
	}

	err = ia.Client.UpdateInsightsAccount(ctx, orgName, accountName, updateReq)
	if err != nil {
		return nil, fmt.Errorf("error updating insights account '%s': %w", accountName, err)
	}

	account, err := ia.Client.GetInsightsAccount(ctx, orgName, accountName)
	if err != nil {
		return nil, fmt.Errorf("error reading insights account after update: %w", err)
	}
	if account == nil {
		return nil, fmt.Errorf("insights account '%s' not found after update", accountName)
	}

	outputProperties, _, err := GenerateInsightsAccountProperties(input, *account)
	if err != nil {
		return nil, err
	}

	return &pulumirpc.UpdateResponse{
		Properties: outputProperties,
	}, nil
}

func (ia *PulumiServiceInsightsAccountResource) Delete(req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	ctx := context.Background()

	orgName, accountName, err := splitInsightsAccountId(req.GetId())
	if err != nil {
		return nil, err
	}

	err = ia.Client.DeleteInsightsAccount(ctx, orgName, accountName)
	if err != nil {
		return &pbempty.Empty{}, err
	}

	return &pbempty.Empty{}, nil
}

func splitInsightsAccountId(id string) (string, string, error) {
	// format: organization/accountName
	s := strings.Split(id, "/")
	if len(s) != 2 {
		return "", "", fmt.Errorf("%q is invalid, must be in the format: organization/accountName", id)
	}
	return s[0], s[1], nil
}
