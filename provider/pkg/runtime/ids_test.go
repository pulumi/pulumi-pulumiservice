// Copyright 2016-2026, Pulumi Corporation.

package runtime

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandPath(t *testing.T) {
	got, err := ExpandPath("/api/orgs/{orgName}/agent-pools/{poolId}", map[string]string{
		"orgName": "acme-corp",
		"poolId":  "pool-abc",
	})
	require.NoError(t, err)
	assert.Equal(t, "/api/orgs/acme-corp/agent-pools/pool-abc", got)
}

func TestExpandPath_MissingValueFails(t *testing.T) {
	_, err := ExpandPath("/api/orgs/{orgName}", map[string]string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "orgName")
}

func TestExtractPathParams(t *testing.T) {
	got := ExtractPathParams("/api/orgs/{orgName}/stacks/{projectName}/{stackName}")
	assert.Equal(t, []string{"orgName", "projectName", "stackName"}, got)
}

func TestBuildID_Simple(t *testing.T) {
	spec := &CloudAPIID{
		Template: "{orgName}/{name}/{agentPoolId}",
	}
	id, err := BuildID(spec, "", map[string]string{
		"orgName":     "acme-corp",
		"name":        "vpc-isolated",
		"agentPoolId": "pool-abc",
	})
	require.NoError(t, err)
	assert.Equal(t, "acme-corp/vpc-isolated/pool-abc", id)
}

func TestBuildID_Polymorphic(t *testing.T) {
	spec := &CloudAPIID{
		Case: "scope",
		Templates: map[string]string{
			"org":   "{orgName}/{name}",
			"stack": "{orgName}/{projectName}/{stackName}/{name}",
		},
	}
	id, err := BuildID(spec, "stack", map[string]string{
		"orgName":     "acme-corp",
		"projectName": "infra",
		"stackName":   "prod",
		"name":        "pagerduty-alerts",
	})
	require.NoError(t, err)
	assert.Equal(t, "acme-corp/infra/prod/pagerduty-alerts", id)
}

func TestDecomposeID_Simple(t *testing.T) {
	spec := &CloudAPIID{
		Template: "{orgName}/{name}/{agentPoolId}",
	}
	got, err := DecomposeID(spec, "", "acme-corp/vpc-isolated/pool-abc")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"orgName":     "acme-corp",
		"name":        "vpc-isolated",
		"agentPoolId": "pool-abc",
	}, got)
}

func TestDecomposeID_WrongSegmentCountFails(t *testing.T) {
	spec := &CloudAPIID{
		Template: "{orgName}/{name}",
	}
	_, err := DecomposeID(spec, "", "acme-corp")
	assert.Error(t, err)
}

func TestDecomposeID_RoundTripPolymorphic(t *testing.T) {
	spec := &CloudAPIID{
		Case: "scope",
		Templates: map[string]string{
			"esc": "{orgName}/environment/{projectName}/{envName}/{name}",
		},
	}
	original := map[string]string{
		"orgName":     "acme-corp",
		"projectName": "platform",
		"envName":     "prod-secrets",
		"name":        "slack-rotation",
	}
	id, err := BuildID(spec, "esc", original)
	require.NoError(t, err)
	parsed, err := DecomposeID(spec, "esc", id)
	require.NoError(t, err)
	assert.Equal(t, original, parsed)
}
