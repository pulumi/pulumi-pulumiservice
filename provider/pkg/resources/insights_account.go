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

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type InsightsAccount struct{}

var (
	_ infer.CustomCreate[InsightsAccountInput, InsightsAccountState] = &InsightsAccount{}
	_ infer.CustomDelete[InsightsAccountState]                       = &InsightsAccount{}
	_ infer.CustomRead[InsightsAccountInput, InsightsAccountState]   = &InsightsAccount{}
	_ infer.CustomUpdate[InsightsAccountInput, InsightsAccountState] = &InsightsAccount{}
)

func (ia *InsightsAccount) Annotate(a infer.Annotator) {
	a.Describe(ia, "Insights Account for cloud resource scanning and analysis across AWS, Azure, and GCP.")
}

// CloudProvider enum for supported cloud providers
type CloudProvider string

const (
	CloudProviderAWS        CloudProvider = "aws"
	CloudProviderAzure      CloudProvider = "azure-native"
	CloudProviderGCP        CloudProvider = "gcp"
	CloudProviderKubernetes CloudProvider = "kubernetes"
	CloudProviderOCI        CloudProvider = "oci"
)

func (CloudProvider) Values() []infer.EnumValue[CloudProvider] {
	return []infer.EnumValue[CloudProvider]{
		{Name: "aws", Value: CloudProviderAWS, Description: "Amazon Web Services"},
		{Name: "azure-native", Value: CloudProviderAzure, Description: "Microsoft Azure"},
		{Name: "gcp", Value: CloudProviderGCP, Description: "Google Cloud Platform"},
		{Name: "kubernetes", Value: CloudProviderKubernetes, Description: "Kubernetes"},
		{Name: "oci", Value: CloudProviderOCI, Description: "Oracle Cloud Infrastructure"},
	}
}

// ScanSchedule enum for automated scanning frequency
type ScanSchedule string

const (
	ScanScheduleNone  ScanSchedule = "none"
	ScanScheduleDaily ScanSchedule = "daily"
)

func (ScanSchedule) Values() []infer.EnumValue[ScanSchedule] {
	return []infer.EnumValue[ScanSchedule]{
		{Name: "none", Value: ScanScheduleNone, Description: "Disable automated scanning."},
		{Name: "daily", Value: ScanScheduleDaily, Description: "Run automated scans once per day."},
	}
}

// InsightsAccountCore contains the core fields for an insights account
type InsightsAccountCore struct {
	OrganizationName string                 `pulumi:"organizationName"        provider:"replaceOnChanges"`
	AccountName      string                 `pulumi:"accountName"             provider:"replaceOnChanges"`
	Provider         CloudProvider          `pulumi:"provider"                provider:"replaceOnChanges"`
	Environment      string                 `pulumi:"environment"`
	ScanSchedule     ScanSchedule           `pulumi:"scanSchedule"`
	ProviderConfig   map[string]interface{} `pulumi:"providerConfig,optional"`
	Tags             map[string]string      `pulumi:"tags,optional"`
}

func (c *InsightsAccountCore) Annotate(a infer.Annotator) {
	a.Describe(&c.OrganizationName, "The organization's name.")
	a.Describe(&c.AccountName, "Name of the insights account.")
	a.Describe(&c.Provider, "The cloud provider for scanning.")
	a.Describe(
		&c.Environment,
		"The ESC environment used for provider credentials. Format: 'project/environment' with optional "+
			"'@version' suffix (e.g., 'my-project/prod-env' or 'my-project/prod-env@v1.0').",
	)
	a.Describe(
		&c.ScanSchedule,
		"Schedule for automated scanning. Use 'daily' to enable daily scans, or 'none' to disable scheduled "+
			"scanning. Defaults to 'none'.",
	)
	a.SetDefault(&c.ScanSchedule, ScanScheduleNone)
	a.Describe(
		&c.ProviderConfig,
		"Provider-specific configuration as a JSON object. For AWS, specify regions to scan: "+
			"{\"regions\": [\"us-west-1\", \"us-west-2\"]}.",
	)
	a.Describe(&c.Tags, "Key-value tags to associate with the insights account.")
}

// InsightsAccountInput represents the input properties for creating an insights account
type InsightsAccountInput struct {
	InsightsAccountCore
}

// InsightsAccountState represents the output properties of an insights account
type InsightsAccountState struct {
	InsightsAccountCore
	InsightsAccountID    string `pulumi:"insightsAccountId"`
	ScheduledScanEnabled bool   `pulumi:"scheduledScanEnabled"`
}

func (s *InsightsAccountState) Annotate(a infer.Annotator) {
	a.Describe(&s.InsightsAccountID, "The insights account identifier.")
	a.Describe(&s.ScheduledScanEnabled, "Whether scheduled scanning is enabled.")
}

// InsightsAccountStateFromAPI converts a pulumiapi.InsightsAccount to an InsightsAccountState.
func InsightsAccountStateFromAPI(orgName string, account pulumiapi.InsightsAccount) InsightsAccountState {
	scanSchedule := ScanScheduleNone
	if account.ScheduledScanEnabled {
		scanSchedule = ScanScheduleDaily
	}
	return InsightsAccountState{
		InsightsAccountCore: InsightsAccountCore{
			OrganizationName: orgName,
			AccountName:      account.Name,
			Provider:         CloudProvider(account.Provider),
			Environment:      account.ProviderEnvRef,
			ProviderConfig:   account.ProviderConfig,
			ScanSchedule:     scanSchedule,
		},
		InsightsAccountID:    account.ID,
		ScheduledScanEnabled: account.ScheduledScanEnabled,
	}
}

func (*InsightsAccount) Create(
	ctx context.Context,
	req infer.CreateRequest[InsightsAccountInput],
) (infer.CreateResponse[InsightsAccountState], error) {
	accountID := fmt.Sprintf("%s/%s", req.Inputs.OrganizationName, req.Inputs.AccountName)
	if req.DryRun {
		return infer.CreateResponse[InsightsAccountState]{
			ID: accountID,
			Output: InsightsAccountState{
				InsightsAccountCore:  req.Inputs.InsightsAccountCore,
				InsightsAccountID:    "",
				ScheduledScanEnabled: req.Inputs.ScanSchedule != ScanScheduleNone,
			},
		}, nil
	}

	client := config.GetClient(ctx)

	createReq := pulumiapi.CreateInsightsAccountRequest{
		Provider:       string(req.Inputs.Provider),
		Environment:    req.Inputs.Environment,
		ProviderConfig: req.Inputs.ProviderConfig,
		ScanSchedule:   string(req.Inputs.ScanSchedule),
	}

	err := client.CreateInsightsAccount(ctx, req.Inputs.OrganizationName, req.Inputs.AccountName, createReq)
	if err != nil {
		return infer.CreateResponse[InsightsAccountState]{}, fmt.Errorf(
			"error creating insights account '%s': %w",
			req.Inputs.AccountName,
			err,
		)
	}

	// Set tags if provided
	if len(req.Inputs.Tags) > 0 {
		err = client.SetInsightsAccountTags(ctx, req.Inputs.OrganizationName, req.Inputs.AccountName, req.Inputs.Tags)
		if err != nil {
			return infer.CreateResponse[InsightsAccountState]{
				ID: accountID,
				Output: InsightsAccountState{
					InsightsAccountCore: req.Inputs.InsightsAccountCore,
				},
			}, infer.ResourceInitFailedError{Reasons: []string{fmt.Sprintf("failed to set tags: %s", err.Error())}}
		}
	}

	account, err := client.GetInsightsAccount(ctx, req.Inputs.OrganizationName, req.Inputs.AccountName)
	if err != nil {
		return infer.CreateResponse[InsightsAccountState]{
			ID: accountID,
			Output: InsightsAccountState{
				InsightsAccountCore: req.Inputs.InsightsAccountCore,
			},
		}, infer.ResourceInitFailedError{Reasons: []string{err.Error()}}
	}
	if account == nil {
		return infer.CreateResponse[InsightsAccountState]{
				ID: accountID,
				Output: InsightsAccountState{
					InsightsAccountCore: req.Inputs.InsightsAccountCore,
				},
			}, infer.ResourceInitFailedError{
				Reasons: []string{
					fmt.Sprintf("insights account '%s' not found after creation", req.Inputs.AccountName),
				},
			}
	}

	return infer.CreateResponse[InsightsAccountState]{
		ID: accountID,
		Output: InsightsAccountState{
			InsightsAccountCore:  req.Inputs.InsightsAccountCore,
			InsightsAccountID:    account.ID,
			ScheduledScanEnabled: account.ScheduledScanEnabled,
		},
	}, nil
}

func (*InsightsAccount) Delete(
	ctx context.Context,
	req infer.DeleteRequest[InsightsAccountState],
) (infer.DeleteResponse, error) {
	client := config.GetClient(ctx)
	return infer.DeleteResponse{}, client.DeleteInsightsAccount(ctx, req.State.OrganizationName, req.State.AccountName)
}

func (*InsightsAccount) Read(
	ctx context.Context,
	req infer.ReadRequest[InsightsAccountInput, InsightsAccountState],
) (infer.ReadResponse[InsightsAccountInput, InsightsAccountState], error) {
	client := config.GetClient(ctx)
	orgName, accountName, err := splitInsightsAccountID(req.ID)
	if err != nil {
		return infer.ReadResponse[InsightsAccountInput, InsightsAccountState]{}, err
	}

	account, err := client.GetInsightsAccount(ctx, orgName, accountName)
	if err != nil {
		return infer.ReadResponse[InsightsAccountInput, InsightsAccountState]{}, fmt.Errorf(
			"failed to read InsightsAccount (%q): %w",
			req.ID,
			err,
		)
	}
	if account == nil {
		return infer.ReadResponse[InsightsAccountInput, InsightsAccountState]{}, nil
	}

	// Preserve input ProviderConfig to avoid spurious diffs
	// when API returns empty map or single region (default configuration)
	providerConfig := req.Inputs.ProviderConfig
	isInputDefault := isDefaultProviderConfig(req.Inputs.Provider, req.Inputs.ProviderConfig)
	isAPIDefault := isDefaultProviderConfig(req.Inputs.Provider, account.ProviderConfig)
	if isInputDefault != isAPIDefault {
		providerConfig = account.ProviderConfig
	}

	// Fetch tags from API
	tags, err := client.GetInsightsAccountTags(ctx, orgName, accountName)
	if err != nil {
		return infer.ReadResponse[InsightsAccountInput, InsightsAccountState]{}, fmt.Errorf(
			"failed to get tags for InsightsAccount (%q): %w",
			req.ID,
			err,
		)
	}

	core := InsightsAccountCore{
		OrganizationName: orgName,
		AccountName:      accountName,
		Provider:         CloudProvider(account.Provider),
		Environment:      account.ProviderEnvRef,
		ProviderConfig:   providerConfig,
		ScanSchedule:     req.Inputs.ScanSchedule, // Preserve input since API doesn't return this
		Tags:             tags,
	}

	return infer.ReadResponse[InsightsAccountInput, InsightsAccountState]{
		ID: req.ID,
		Inputs: InsightsAccountInput{
			InsightsAccountCore: core,
		},
		State: InsightsAccountState{
			InsightsAccountCore:  core,
			InsightsAccountID:    account.ID,
			ScheduledScanEnabled: account.ScheduledScanEnabled,
		},
	}, nil
}

func (*InsightsAccount) Update(
	ctx context.Context,
	req infer.UpdateRequest[InsightsAccountInput, InsightsAccountState],
) (infer.UpdateResponse[InsightsAccountState], error) {
	if req.DryRun {
		return infer.UpdateResponse[InsightsAccountState]{
			Output: InsightsAccountState{
				InsightsAccountCore:  req.Inputs.InsightsAccountCore,
				InsightsAccountID:    req.State.InsightsAccountID,
				ScheduledScanEnabled: req.State.ScheduledScanEnabled,
			},
		}, nil
	}

	client := config.GetClient(ctx)

	providerConfig := req.Inputs.ProviderConfig
	// If provider config is default (empty or single region for AWS), pass
	// config with empty regions to ensure API updates correctly
	if isDefaultProviderConfig(req.Inputs.Provider, providerConfig) {
		providerConfig = map[string]interface{}{"regions": []string{}}
	}

	updateReq := pulumiapi.UpdateInsightsAccountRequest{
		Environment:    req.Inputs.Environment,
		ProviderConfig: providerConfig,
		ScanSchedule:   string(req.Inputs.ScanSchedule),
	}

	err := client.UpdateInsightsAccount(ctx, req.State.OrganizationName, req.State.AccountName, updateReq)
	if err != nil {
		return infer.UpdateResponse[InsightsAccountState]{}, fmt.Errorf(
			"error updating insights account '%s': %w",
			req.State.AccountName,
			err,
		)
	}

	// Update tags - SetInsightsAccountTags replaces all tags, so we always call it
	// to ensure the desired state matches (including removing tags if empty)
	err = client.SetInsightsAccountTags(ctx, req.State.OrganizationName, req.State.AccountName, req.Inputs.Tags)
	if err != nil {
		return infer.UpdateResponse[InsightsAccountState]{
			Output: InsightsAccountState{
				InsightsAccountCore: req.Inputs.InsightsAccountCore,
				InsightsAccountID:   req.State.InsightsAccountID,
			},
		}, infer.ResourceInitFailedError{Reasons: []string{fmt.Sprintf("failed to set tags: %s", err.Error())}}
	}

	account, err := client.GetInsightsAccount(ctx, req.State.OrganizationName, req.State.AccountName)
	if err != nil {
		return infer.UpdateResponse[InsightsAccountState]{
			Output: InsightsAccountState{
				InsightsAccountCore: req.Inputs.InsightsAccountCore,
				InsightsAccountID:   req.State.InsightsAccountID,
			},
		}, infer.ResourceInitFailedError{Reasons: []string{err.Error()}}
	}
	if account == nil {
		return infer.UpdateResponse[InsightsAccountState]{
				Output: InsightsAccountState{
					InsightsAccountCore: req.Inputs.InsightsAccountCore,
					InsightsAccountID:   req.State.InsightsAccountID,
				},
			}, infer.ResourceInitFailedError{
				Reasons: []string{fmt.Sprintf("insights account '%s' not found after update", req.State.AccountName)},
			}
	}

	return infer.UpdateResponse[InsightsAccountState]{
		Output: InsightsAccountState{
			InsightsAccountCore:  req.Inputs.InsightsAccountCore,
			InsightsAccountID:    account.ID,
			ScheduledScanEnabled: account.ScheduledScanEnabled,
		},
	}, nil
}

func isDefaultProviderConfig(provider CloudProvider, config map[string]interface{}) bool {
	if len(config) == 0 {
		return true
	}
	if provider == CloudProviderAWS {
		if regions, ok := config["regions"]; ok {
			if regionsSlice, ok := regions.([]interface{}); ok {
				return len(regionsSlice) == 0
			}
		}
	}
	return false
}

func splitInsightsAccountID(id string) (string, string, error) {
	// format: organization/accountName
	s := strings.Split(id, "/")
	if len(s) != 2 {
		return "", "", fmt.Errorf("%q is invalid, must be in the format: organization/accountName", id)
	}
	return s[0], s[1], nil
}
