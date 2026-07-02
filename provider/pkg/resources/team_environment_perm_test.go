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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type teamEnvPermClientMock struct {
	config.Client
	getFunc func() (*string, *pulumiapi.Duration, error)
}

func (c *teamEnvPermClientMock) GetTeamEnvironmentSettings(
	_ context.Context, _ pulumiapi.TeamEnvironmentSettingsRequest,
) (*string, *pulumiapi.Duration, error) {
	return c.getFunc()
}

func newTeamEnvPermCheckRequest(t *testing.T, props map[string]property.Value) infer.CheckRequest {
	t.Helper()
	return infer.CheckRequest{NewInputs: property.NewMap(props)}
}

func TestTeamEnvironmentPermissionCheck(t *testing.T) {
	r := &TeamEnvironmentPermission{}
	base := func() map[string]property.Value {
		return map[string]property.Value{
			gcOrganization: property.New(gcOrg),
			gcTeam:         property.New(gcTeam),
			gcProject:      property.New("proj"),
			gcEnvironment:  property.New(gcEnv),
			"permission":   property.New("open"),
		}
	}

	t.Run("omits unset maxOpenDuration", func(t *testing.T) {
		// Regression: when upgrading from a provider version that did not have
		// maxOpenDuration, Check must not inject a maxOpenDuration into the
		// returned inputs.
		resp, err := r.Check(t.Context(), newTeamEnvPermCheckRequest(t, base()))
		require.NoError(t, err)
		assert.Empty(t, resp.Failures)
		assert.Nil(t, resp.Inputs.MaxOpenDuration)
	})

	t.Run("normalizes maxOpenDuration", func(t *testing.T) {
		props := base()
		props["maxOpenDuration"] = property.New("60m")

		resp, err := r.Check(t.Context(), newTeamEnvPermCheckRequest(t, props))
		require.NoError(t, err)
		assert.Empty(t, resp.Failures)
		require.NotNil(t, resp.Inputs.MaxOpenDuration)
		assert.Equal(t, "1h0m0s", *resp.Inputs.MaxOpenDuration)
	})

	t.Run("treats empty-string maxOpenDuration as unset", func(t *testing.T) {
		props := base()
		props["maxOpenDuration"] = property.New("")

		resp, err := r.Check(t.Context(), newTeamEnvPermCheckRequest(t, props))
		require.NoError(t, err)
		assert.Empty(t, resp.Failures)
		assert.Nil(t, resp.Inputs.MaxOpenDuration)
	})

	t.Run("rejects invalid maxOpenDuration", func(t *testing.T) {
		props := base()
		props["maxOpenDuration"] = property.New("not-a-duration")

		resp, err := r.Check(t.Context(), newTeamEnvPermCheckRequest(t, props))
		require.NoError(t, err)
		require.Len(t, resp.Failures, 1)
		assert.Equal(t, "maxOpenDuration", resp.Failures[0].Property)
	})
}

func TestTeamEnvironmentPermissionDiff(t *testing.T) {
	r := &TeamEnvironmentPermission{}
	state := func() TeamEnvironmentPermissionState {
		return TeamEnvironmentPermissionState{
			TeamEnvironmentPermissionInput: TeamEnvironmentPermissionInput{
				Organization: gcOrg,
				Team:         gcTeam,
				Project:      "proj",
				Environment:  gcEnv,
				Permission:   EnvironmentPermissionOpen,
			},
		}
	}

	t.Run("no changes when nothing changed", func(t *testing.T) {
		// Regression: 0.29.2 -> 0.29.3 upgrade path. State has no
		// maxOpenDuration; inputs don't set it. Diff must report no changes.
		resp, err := r.Diff(t.Context(), infer.DiffRequest[
			TeamEnvironmentPermissionInput, TeamEnvironmentPermissionState,
		]{
			State:  state(),
			Inputs: state().TeamEnvironmentPermissionInput,
		})
		require.NoError(t, err)
		assert.False(t, resp.HasChanges)
	})

	t.Run("treats empty-string maxOpenDuration in state as unset", func(t *testing.T) {
		// 0.29.3-0.36.0 wrote `""` into state when the user did not set the
		// field. After the migration, refreshes against that state must not
		// force a spurious replacement.
		empty := ""
		oldState := state()
		oldState.MaxOpenDuration = &empty

		resp, err := r.Diff(t.Context(), infer.DiffRequest[
			TeamEnvironmentPermissionInput, TeamEnvironmentPermissionState,
		]{
			State:  oldState,
			Inputs: state().TeamEnvironmentPermissionInput,
		})
		require.NoError(t, err)
		assert.False(t, resp.HasChanges)
	})

	t.Run("detects real maxOpenDuration change", func(t *testing.T) {
		// Guard against over-aggressive normalization.
		oldDur := "30m0s"
		newDur := "1h0m0s"
		oldState := state()
		oldState.MaxOpenDuration = &oldDur
		newInputs := state().TeamEnvironmentPermissionInput
		newInputs.MaxOpenDuration = &newDur

		resp, err := r.Diff(t.Context(), infer.DiffRequest[
			TeamEnvironmentPermissionInput, TeamEnvironmentPermissionState,
		]{State: oldState, Inputs: newInputs})
		require.NoError(t, err)
		assert.True(t, resp.HasChanges)
		assert.Contains(t, resp.DetailedDiff, "maxOpenDuration")
		assert.True(t, resp.DeleteBeforeReplace)
	})
}

func TestTeamEnvironmentPermissionRead(t *testing.T) {
	r := &TeamEnvironmentPermission{}

	t.Run("omits maxOpenDuration when API returns none", func(t *testing.T) {
		permission := "open"
		mock := &teamEnvPermClientMock{
			getFunc: func() (*string, *pulumiapi.Duration, error) {
				return &permission, nil, nil
			},
		}
		ctx := config.WithMockClient(t.Context(), mock)

		resp, err := r.Read(ctx, infer.ReadRequest[
			TeamEnvironmentPermissionInput, TeamEnvironmentPermissionState,
		]{ID: "org/team/proj+env"})
		require.NoError(t, err)
		assert.Nil(t, resp.Inputs.MaxOpenDuration)
	})

	t.Run("includes maxOpenDuration when API returns one", func(t *testing.T) {
		permission := "open"
		d := pulumiapi.Duration(30 * time.Minute)
		mock := &teamEnvPermClientMock{
			getFunc: func() (*string, *pulumiapi.Duration, error) {
				return &permission, &d, nil
			},
		}
		ctx := config.WithMockClient(t.Context(), mock)

		resp, err := r.Read(ctx, infer.ReadRequest[
			TeamEnvironmentPermissionInput, TeamEnvironmentPermissionState,
		]{ID: "org/team/proj+env"})
		require.NoError(t, err)
		require.NotNil(t, resp.Inputs.MaxOpenDuration)
		assert.Equal(t, "30m0s", *resp.Inputs.MaxOpenDuration)
	})
}

func TestSplitTeamEnvironmentPermissionID(t *testing.T) {
	t.Run("project+environment form", func(t *testing.T) {
		got, err := splitTeamEnvironmentPermissionID("org/team/proj+env")
		require.NoError(t, err)
		assert.Equal(t, teamEnvironmentPermissionID{
			Organization: gcOrg,
			Team:         gcTeam,
			Project:      "proj",
			Environment:  gcEnv,
		}, got)
	})

	t.Run("legacy environment-only form defaults project", func(t *testing.T) {
		got, err := splitTeamEnvironmentPermissionID("org/team/env")
		require.NoError(t, err)
		assert.Equal(t, teamEnvironmentPermissionID{
			Organization: gcOrg,
			Team:         gcTeam,
			Project:      "default",
			Environment:  gcEnv,
		}, got)
	})

	t.Run("rejects malformed id", func(t *testing.T) {
		_, err := splitTeamEnvironmentPermissionID("org/team")
		assert.Error(t, err)
	})
}
