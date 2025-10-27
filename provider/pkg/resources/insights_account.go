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

// Call implements the Call RPC for resource methods
func (ia *PulumiServiceInsightsAccountResource) Call(req *pulumirpc.CallRequest) (*pulumirpc.CallResponse, error) {
	ctx := context.Background()

	// Get the resource URN from __self__
	selfArg, ok := req.GetArgs().GetFields()["__self__"]
	if !ok {
		return nil, fmt.Errorf("missing required __self__ argument")
	}

	resourceRef := selfArg.GetStructValue()
	if resourceRef == nil {
		return nil, fmt.Errorf("__self__ must be a resource reference")
	}

	urn := resourceRef.GetFields()["urn"].GetStringValue()
	if urn == "" {
		return nil, fmt.Errorf("__self__ resource reference missing URN")
	}

	// During preview (dry run), return computed/unknown values for all outputs
	// Check req.DryRun first (proper way to detect preview mode)
	if req.DryRun {
		return ia.returnUnknownOutputs(req.GetTok())
	}

	// Fallback: check for empty ID (for older SDKs/tests that don't set DryRun)
	id := resourceRef.GetFields()["id"].GetStringValue()
	if id == "" {
		return ia.returnUnknownOutputs(req.GetTok())
	}

	// Split the ID to get org and account name
	orgName, accountName, err := splitInsightsAccountId(id)
	if err != nil {
		return nil, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Route to the appropriate method
	switch req.GetTok() {
	case "pulumiservice:index:InsightsAccount/triggerScan":
		return ia.triggerScan(ctx, orgName, accountName)
	case "pulumiservice:index:InsightsAccount/getStatus":
		return ia.getStatus(ctx, orgName, accountName)
	default:
		return nil, fmt.Errorf("unknown method: %s", req.GetTok())
	}
}

// returnUnknownOutputs returns computed/unknown values for all outputs during preview
func (ia *PulumiServiceInsightsAccountResource) returnUnknownOutputs(tok string) (*pulumirpc.CallResponse, error) {
	var outputMap resource.PropertyMap

	switch tok {
	case "pulumiservice:index:InsightsAccount/triggerScan":
		outputMap = resource.PropertyMap{
			"scanId":    resource.MakeComputed(resource.NewStringProperty("")),
			"status":    resource.MakeComputed(resource.NewStringProperty("")),
			"timestamp": resource.MakeComputed(resource.NewStringProperty("")),
		}
	case "pulumiservice:index:InsightsAccount/getStatus":
		outputMap = resource.PropertyMap{
			"accountId":     resource.MakeComputed(resource.NewStringProperty("")),
			"accountName":   resource.MakeComputed(resource.NewStringProperty("")),
			"status":        resource.MakeComputed(resource.NewStringProperty("")),
			"lastScanId":    resource.MakeComputed(resource.NewStringProperty("")),
			"lastScanTime":  resource.MakeComputed(resource.NewStringProperty("")),
			"nextScanTime":  resource.MakeComputed(resource.NewStringProperty("")),
			"resourceCount": resource.MakeComputed(resource.NewNumberProperty(0)),
		}
	default:
		return nil, fmt.Errorf("unknown method: %s", tok)
	}

	returnValue, err := plugin.MarshalProperties(outputMap, plugin.MarshalOptions{
		KeepUnknowns: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal unknown outputs: %w", err)
	}

	return &pulumirpc.CallResponse{
		Return: returnValue,
	}, nil
}

// triggerScan implements the triggerScan method
func (ia *PulumiServiceInsightsAccountResource) triggerScan(ctx context.Context, orgName, accountName string) (*pulumirpc.CallResponse, error) {
	response, err := ia.Client.TriggerScan(ctx, orgName, accountName)
	if err != nil {
		return nil, fmt.Errorf("failed to trigger scan: %w", err)
	}

	// Convert WorkflowRun response to property map
	outputMap := resource.PropertyMap{
		"status": resource.NewPropertyValue(response.Status),
	}

	// Only include scanId if it's available (may be empty for HTTP 204 responses)
	if response.ID != "" {
		outputMap["scanId"] = resource.NewPropertyValue(response.ID)
	}

	if response.StartedAt != "" {
		outputMap["timestamp"] = resource.NewPropertyValue(response.StartedAt)
	}

	// Marshal to protobuf struct
	returnValue, err := plugin.MarshalProperties(outputMap, plugin.MarshalOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trigger scan response: %w", err)
	}

	return &pulumirpc.CallResponse{
		Return: returnValue,
	}, nil
}

// getStatus implements the getStatus method
func (ia *PulumiServiceInsightsAccountResource) getStatus(ctx context.Context, orgName, accountName string) (*pulumirpc.CallResponse, error) {
	status, err := ia.Client.GetScanStatus(ctx, orgName, accountName)
	if err != nil {
		return nil, fmt.Errorf("failed to get scan status: %w", err)
	}

	// Handle case where no scan exists yet (404 response)
	if status == nil {
		return nil, fmt.Errorf("no scan has been initiated for insights account %q yet", accountName)
	}

	// Convert ScanStatusResponse to property map
	outputMap := resource.PropertyMap{
		"accountId":   resource.NewPropertyValue(orgName + "/" + accountName),
		"accountName": resource.NewPropertyValue(accountName),
		"status":      resource.NewPropertyValue(status.Status),
	}

	if status.ID != "" {
		outputMap["lastScanId"] = resource.NewPropertyValue(status.ID)
	}

	if status.FinishedAt != "" {
		outputMap["lastScanTime"] = resource.NewPropertyValue(status.FinishedAt)
	}

	if status.NextScan != "" {
		outputMap["nextScanTime"] = resource.NewPropertyValue(status.NextScan)
	}

	if status.ResourceCount > 0 {
		outputMap["resourceCount"] = resource.NewPropertyValue(status.ResourceCount)
	}

	// Marshal to protobuf struct
	returnValue, err := plugin.MarshalProperties(outputMap, plugin.MarshalOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal scan status response: %w", err)
	}

	return &pulumirpc.CallResponse{
		Return: returnValue,
	}, nil
}

func splitInsightsAccountId(id string) (string, string, error) {
	// format: organization/accountName
	s := strings.Split(id, "/")
	if len(s) != 2 {
		return "", "", fmt.Errorf("%q is invalid, must be in the format: organization/accountName", id)
	}
	return s[0], s[1], nil
}
