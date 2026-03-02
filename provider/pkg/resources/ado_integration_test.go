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

package resources

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

// --- mock ---

type AdoIntegrationClientMock struct {
	getFunc    func(ctx context.Context, orgName, integrationID string) (*pulumiapi.AzureDevOpsIntegration, error)
	updateFunc func(ctx context.Context, orgName, integrationID string, req pulumiapi.UpdateAzureDevOpsIntegrationRequest) error
	deleteFunc func(ctx context.Context, orgName, integrationID string) error
	listFunc   func(ctx context.Context, orgName string) ([]pulumiapi.AzureDevOpsIntegration, error)
}

func (m *AdoIntegrationClientMock) GetAzureDevOpsIntegration(ctx context.Context, orgName, integrationID string) (*pulumiapi.AzureDevOpsIntegration, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, orgName, integrationID)
	}
	return nil, nil
}

func (m *AdoIntegrationClientMock) UpdateAzureDevOpsIntegration(ctx context.Context, orgName, integrationID string, req pulumiapi.UpdateAzureDevOpsIntegrationRequest) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, orgName, integrationID, req)
	}
	return nil
}

func (m *AdoIntegrationClientMock) DeleteAzureDevOpsIntegration(ctx context.Context, orgName, integrationID string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, orgName, integrationID)
	}
	return nil
}

func (m *AdoIntegrationClientMock) ListAzureDevOpsIntegrations(ctx context.Context, orgName string) ([]pulumiapi.AzureDevOpsIntegration, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, orgName)
	}
	return nil, nil
}

// --- tests ---

func TestAdoIntegrationInputRoundtrip(t *testing.T) {
	input := PulumiServiceAdoIntegrationInput{
		Organization:        "my-org",
		AdoOrganizationName: "ado-org",
		ProjectID:           "proj-123",
		DisablePRComments:   true,
		DisableNeoSummaries: false,
		DisableDetailedDiff: true,
	}

	props := PulumiServiceAdoIntegrationProperties{
		PulumiServiceAdoIntegrationInput: input,
		IntegrationID:                    "int-456",
		Valid:                            true,
		AdoOrganizationID:                "ado-org-id",
		AdoOrganizationURL:               "https://dev.azure.com/ado-org",
		ProjectName:                      "My Project",
	}

	pm := props.ToPropertyMap()
	decoded := ToPulumiServiceAdoIntegrationInput(pm)
	assert.Equal(t, input, decoded)
}

func TestAdoIntegrationToPropertyMap(t *testing.T) {
	props := PulumiServiceAdoIntegrationProperties{
		PulumiServiceAdoIntegrationInput: PulumiServiceAdoIntegrationInput{
			Organization:        "my-org",
			AdoOrganizationName: "ado-org",
			ProjectID:           "proj-123",
			DisablePRComments:   true,
			DisableNeoSummaries: false,
			DisableDetailedDiff: true,
		},
		IntegrationID:      "int-456",
		Valid:              true,
		AdoOrganizationID:  "ado-org-id",
		AdoOrganizationURL: "https://dev.azure.com/ado-org",
		ProjectName:        "My Project",
	}

	pm := props.ToPropertyMap()

	expectedKeys := []resource.PropertyKey{
		"organization", "adoOrganizationName", "projectId",
		"disablePRComments", "disableNeoSummaries", "disableDetailedDiff",
		"integrationId", "valid", "adoOrganizationId", "adoOrganizationUrl", "projectName",
	}
	for _, key := range expectedKeys {
		assert.Truef(t, pm[key].HasValue(), "expected key %q in PropertyMap", key)
	}

	assert.Equal(t, "my-org", pm["organization"].StringValue())
	assert.Equal(t, "int-456", pm["integrationId"].StringValue())
	assert.True(t, pm["valid"].BoolValue())
}

func TestSplitAdoIntegrationID(t *testing.T) {
	t.Run("valid format", func(t *testing.T) {
		org, id, err := splitAdoIntegrationID("my-org/int-123")
		require.NoError(t, err)
		assert.Equal(t, "my-org", org)
		assert.Equal(t, "int-123", id)
	})

	t.Run("no slash", func(t *testing.T) {
		_, _, err := splitAdoIntegrationID("noslash")
		assert.Error(t, err)
	})

	t.Run("empty string", func(t *testing.T) {
		_, _, err := splitAdoIntegrationID("")
		assert.Error(t, err)
	})
}

func TestAdoIntegrationCheck(t *testing.T) {
	r := &PulumiServiceAdoIntegrationResource{
		Client: &AdoIntegrationClientMock{},
	}

	t.Run("missing required properties returns failures", func(t *testing.T) {
		// Empty property map — all three required fields missing
		props, err := plugin.MarshalProperties(
			resource.PropertyMap{},
			plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
		)
		require.NoError(t, err)

		resp, err := r.Check(&pulumirpc.CheckRequest{
			News: props,
		})
		require.NoError(t, err)
		assert.Len(t, resp.Failures, 3)

		reasons := map[string]bool{}
		for _, f := range resp.Failures {
			reasons[f.Property] = true
		}
		assert.True(t, reasons["organization"])
		assert.True(t, reasons["adoOrganizationName"])
		assert.True(t, reasons["projectId"])
	})

	t.Run("all required properties present returns no failures", func(t *testing.T) {
		pm := resource.PropertyMap{
			"organization":        resource.NewPropertyValue("my-org"),
			"adoOrganizationName": resource.NewPropertyValue("ado-org"),
			"projectId":           resource.NewPropertyValue("proj-123"),
		}
		props, err := plugin.MarshalProperties(
			pm,
			plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true},
		)
		require.NoError(t, err)

		resp, err := r.Check(&pulumirpc.CheckRequest{
			News: props,
		})
		require.NoError(t, err)
		assert.Empty(t, resp.Failures)
	})
}

func TestAdoIntegrationDiff(t *testing.T) {
	r := &PulumiServiceAdoIntegrationResource{
		Client: &AdoIntegrationClientMock{},
	}

	marshal := func(pm resource.PropertyMap) *pulumirpc.DiffRequest {
		old, err := plugin.MarshalProperties(pm, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
		require.NoError(t, err)
		return &pulumirpc.DiffRequest{
			OldInputs: old,
			News:      old,
		}
	}

	basePM := resource.PropertyMap{
		"organization":        resource.NewPropertyValue("my-org"),
		"adoOrganizationName": resource.NewPropertyValue("ado-org"),
		"projectId":           resource.NewPropertyValue("proj-123"),
		"disablePRComments":   resource.NewPropertyValue(false),
		"disableNeoSummaries": resource.NewPropertyValue(false),
		"disableDetailedDiff": resource.NewPropertyValue(false),
	}

	t.Run("no changes returns DIFF_NONE", func(t *testing.T) {
		req := marshal(basePM)
		resp, err := r.Diff(req)
		require.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, resp.Changes)
	})

	t.Run("change to replace property triggers replace", func(t *testing.T) {
		oldProps, err := plugin.MarshalProperties(basePM, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
		require.NoError(t, err)

		newPM := basePM.Copy()
		newPM["organization"] = resource.NewPropertyValue("other-org")
		newProps, err := plugin.MarshalProperties(newPM, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
		require.NoError(t, err)

		resp, err := r.Diff(&pulumirpc.DiffRequest{
			OldInputs: oldProps,
			News:      newProps,
		})
		require.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		assert.True(t, resp.HasDetailedDiff)

		orgDiff, ok := resp.DetailedDiff["organization"]
		require.True(t, ok, "expected organization in DetailedDiff")
		// Replace kinds are UPDATE_REPLACE (5) or ADD_REPLACE (1)
		assert.True(t, orgDiff.Kind == pulumirpc.PropertyDiff_UPDATE_REPLACE || orgDiff.Kind == pulumirpc.PropertyDiff_ADD_REPLACE,
			"expected replace kind, got %v", orgDiff.Kind)
	})

	t.Run("change to updatable property triggers update diff", func(t *testing.T) {
		oldProps, err := plugin.MarshalProperties(basePM, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
		require.NoError(t, err)

		newPM := basePM.Copy()
		newPM["disablePRComments"] = resource.NewPropertyValue(true)
		newProps, err := plugin.MarshalProperties(newPM, plugin.MarshalOptions{KeepUnknowns: true, SkipNulls: true})
		require.NoError(t, err)

		resp, err := r.Diff(&pulumirpc.DiffRequest{
			OldInputs: oldProps,
			News:      newProps,
		})
		require.NoError(t, err)
		assert.Equal(t, pulumirpc.DiffResponse_DIFF_SOME, resp.Changes)
		assert.True(t, resp.HasDetailedDiff)

		prDiff, ok := resp.DetailedDiff["disablePRComments"]
		require.True(t, ok, "expected disablePRComments in DetailedDiff")
		assert.Equal(t, pulumirpc.PropertyDiff_UPDATE, prDiff.Kind)
	})
}
