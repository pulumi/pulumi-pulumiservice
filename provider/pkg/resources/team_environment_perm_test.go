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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

// teamEnvPermClientMock satisfies pulumiapi.TeamClient for the handful of
// tests below. Only the methods that TeamEnvironmentPermission actually calls
// are wired up; the rest return zero values.
type teamEnvPermClientMock struct {
	getEnvSettingsFunc func() (*string, *pulumiapi.Duration, error)
}

func (*teamEnvPermClientMock) ListTeams(_ context.Context, _ string) ([]pulumiapi.Team, error) {
	return nil, nil
}

func (*teamEnvPermClientMock) GetTeam(_ context.Context, _, _ string) (*pulumiapi.Team, error) {
	return nil, nil
}

func (*teamEnvPermClientMock) CreateTeam(
	_ context.Context, _, _, _, _, _ string, _ int64,
) (*pulumiapi.Team, error) {
	return nil, nil
}

func (*teamEnvPermClientMock) UpdateTeam(_ context.Context, _, _, _, _ string) error {
	return nil
}

func (*teamEnvPermClientMock) DeleteTeam(_ context.Context, _, _ string) error { return nil }

func (*teamEnvPermClientMock) AddMemberToTeam(_ context.Context, _, _, _ string) error { return nil }

func (*teamEnvPermClientMock) DeleteMemberFromTeam(_ context.Context, _, _, _ string) error {
	return nil
}

func (*teamEnvPermClientMock) AddStackPermission(
	_ context.Context, _ pulumiapi.StackIdentifier, _ string, _ int,
) error {
	return nil
}

func (*teamEnvPermClientMock) RemoveStackPermission(
	_ context.Context, _ pulumiapi.StackIdentifier, _ string,
) error {
	return nil
}

func (*teamEnvPermClientMock) GetTeamStackPermission(
	_ context.Context, _ pulumiapi.StackIdentifier, _ string,
) (*int, error) {
	return nil, nil
}

func (*teamEnvPermClientMock) AddEnvironmentSettings(
	_ context.Context, _ pulumiapi.CreateTeamEnvironmentSettingsRequest,
) error {
	return nil
}

func (*teamEnvPermClientMock) RemoveEnvironmentSettings(
	_ context.Context, _ pulumiapi.TeamEnvironmentSettingsRequest,
) error {
	return nil
}

func (c *teamEnvPermClientMock) GetTeamEnvironmentSettings(
	_ context.Context, _ pulumiapi.TeamEnvironmentSettingsRequest,
) (*string, *pulumiapi.Duration, error) {
	return c.getEnvSettingsFunc()
}

func newTeamEnvPermResource(client pulumiapi.TeamClient) *PulumiServiceTeamEnvironmentPermissionResource {
	return &PulumiServiceTeamEnvironmentPermissionResource{Client: client}
}

func toStruct(t *testing.T, m resource.PropertyMap) *structpb.Struct {
	t.Helper()
	s, err := structpb.NewStruct(m.Mappable())
	require.NoError(t, err)
	return s
}

func TestTeamEnvironmentPermission_Check_OmitsUnsetMaxOpenDuration(t *testing.T) {
	// Regression: when upgrading from a provider version that did not have
	// maxOpenDuration, Check must not inject an empty-string maxOpenDuration
	// into the returned inputs. Otherwise Diff sees a "new" key against the
	// saved state and forces replacement.
	r := newTeamEnvPermResource(&teamEnvPermClientMock{})

	news := resource.PropertyMap{
		"organization": resource.NewStringProperty("org"),
		"team":         resource.NewStringProperty("team"),
		"project":      resource.NewStringProperty("proj"),
		"environment":  resource.NewStringProperty("env"),
		"permission":   resource.NewStringProperty("open"),
	}

	resp, err := r.Check(&pulumirpc.CheckRequest{News: toStruct(t, news)})
	require.NoError(t, err)
	assert.Empty(t, resp.Failures)

	out, err := plugin.UnmarshalProperties(resp.Inputs, plugin.MarshalOptions{})
	require.NoError(t, err)
	_, present := out["maxOpenDuration"]
	assert.False(t, present, "maxOpenDuration must not be emitted when the user did not set it")
}

func TestTeamEnvironmentPermission_Check_NormalizesMaxOpenDuration(t *testing.T) {
	r := newTeamEnvPermResource(&teamEnvPermClientMock{})

	news := resource.PropertyMap{
		"organization":    resource.NewStringProperty("org"),
		"team":            resource.NewStringProperty("team"),
		"project":         resource.NewStringProperty("proj"),
		"environment":     resource.NewStringProperty("env"),
		"permission":      resource.NewStringProperty("open"),
		"maxOpenDuration": resource.NewStringProperty("60m"),
	}

	resp, err := r.Check(&pulumirpc.CheckRequest{News: toStruct(t, news)})
	require.NoError(t, err)
	assert.Empty(t, resp.Failures)

	out, err := plugin.UnmarshalProperties(resp.Inputs, plugin.MarshalOptions{})
	require.NoError(t, err)
	require.True(t, out["maxOpenDuration"].IsString())
	assert.Equal(t, "1h0m0s", out["maxOpenDuration"].StringValue())
}

func TestTeamEnvironmentPermission_Check_EmptyStringMaxOpenDurationTreatedAsUnset(t *testing.T) {
	r := newTeamEnvPermResource(&teamEnvPermClientMock{})

	news := resource.PropertyMap{
		"organization":    resource.NewStringProperty("org"),
		"team":            resource.NewStringProperty("team"),
		"project":         resource.NewStringProperty("proj"),
		"environment":     resource.NewStringProperty("env"),
		"permission":      resource.NewStringProperty("open"),
		"maxOpenDuration": resource.NewStringProperty(""),
	}

	resp, err := r.Check(&pulumirpc.CheckRequest{News: toStruct(t, news)})
	require.NoError(t, err)
	assert.Empty(t, resp.Failures)

	out, err := plugin.UnmarshalProperties(resp.Inputs, plugin.MarshalOptions{})
	require.NoError(t, err)
	_, present := out["maxOpenDuration"]
	assert.False(t, present, "an explicit empty string must be stripped so it never triggers a spurious diff")
}

func TestTeamEnvironmentPermission_Check_InvalidMaxOpenDurationReportsFailure(t *testing.T) {
	r := newTeamEnvPermResource(&teamEnvPermClientMock{})

	news := resource.PropertyMap{
		"organization":    resource.NewStringProperty("org"),
		"team":            resource.NewStringProperty("team"),
		"project":         resource.NewStringProperty("proj"),
		"environment":     resource.NewStringProperty("env"),
		"permission":      resource.NewStringProperty("open"),
		"maxOpenDuration": resource.NewStringProperty("not-a-duration"),
	}

	resp, err := r.Check(&pulumirpc.CheckRequest{News: toStruct(t, news)})
	require.NoError(t, err)
	require.Len(t, resp.Failures, 1)
	assert.Equal(t, "maxOpenDuration", resp.Failures[0].Property)
}

func TestTeamEnvironmentPermission_Diff_UpgradeFromPreMaxOpenDuration(t *testing.T) {
	// End-to-end regression for the 0.29.2 -> 0.29.3 upgrade path:
	//   - OldInputs has no maxOpenDuration (was saved before the field existed)
	//   - The user's program still doesn't set maxOpenDuration
	//   - Check + Diff together must report no changes and no replaces
	r := newTeamEnvPermResource(&teamEnvPermClientMock{})

	inputs := resource.PropertyMap{
		"organization": resource.NewStringProperty("org"),
		"team":         resource.NewStringProperty("team"),
		"project":      resource.NewStringProperty("proj"),
		"environment":  resource.NewStringProperty("env"),
		"permission":   resource.NewStringProperty("open"),
	}

	checkResp, err := r.Check(&pulumirpc.CheckRequest{News: toStruct(t, inputs)})
	require.NoError(t, err)
	require.Empty(t, checkResp.Failures)

	diffResp, err := r.Diff(&pulumirpc.DiffRequest{
		OldInputs: toStruct(t, inputs),
		News:      checkResp.Inputs,
	})
	require.NoError(t, err)
	assert.Equal(t, pulumirpc.DiffResponse_DIFF_NONE, diffResp.Changes)
	assert.Empty(t, diffResp.Replaces, "maxOpenDuration should not drive a spurious replacement on upgrade")
}

func TestTeamEnvironmentPermission_Read_OmitsMaxOpenDurationWhenUnset(t *testing.T) {
	// When the Pulumi Cloud API does not return a maxOpenDuration for the
	// environment, Read must not bake an empty string into state. Otherwise
	// a follow-up up against a program that doesn't set the field would
	// observe a spurious diff (same root cause as the Check path).
	permission := "open"
	client := &teamEnvPermClientMock{
		getEnvSettingsFunc: func() (*string, *pulumiapi.Duration, error) {
			return &permission, nil, nil
		},
	}
	r := newTeamEnvPermResource(client)

	resp, err := r.Read(&pulumirpc.ReadRequest{Id: "org/team/proj+env"})
	require.NoError(t, err)

	out, err := plugin.UnmarshalProperties(resp.Inputs, plugin.MarshalOptions{})
	require.NoError(t, err)
	_, present := out["maxOpenDuration"]
	assert.False(t, present, "Read must omit maxOpenDuration when the API did not return one")
}

func TestTeamEnvironmentPermission_Read_IncludesMaxOpenDurationWhenSet(t *testing.T) {
	permission := "open"
	d := pulumiapi.Duration(30 * time.Minute)
	client := &teamEnvPermClientMock{
		getEnvSettingsFunc: func() (*string, *pulumiapi.Duration, error) {
			return &permission, &d, nil
		},
	}
	r := newTeamEnvPermResource(client)

	resp, err := r.Read(&pulumirpc.ReadRequest{Id: "org/team/proj+env"})
	require.NoError(t, err)

	out, err := plugin.UnmarshalProperties(resp.Inputs, plugin.MarshalOptions{})
	require.NoError(t, err)
	require.True(t, out["maxOpenDuration"].IsString())
	assert.Equal(t, "30m0s", out["maxOpenDuration"].StringValue())
}
