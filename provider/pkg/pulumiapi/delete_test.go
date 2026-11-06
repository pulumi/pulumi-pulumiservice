// Copyright 2026, Pulumi Corporation.

package pulumiapi

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// failingTransport fails every request before a response arrives, the way a
// network error or timeout does.
type failingTransport struct{}

func (failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("connection refused")
}

func newUnreachableClient(t *testing.T) *Client {
	c, err := NewClient(&http.Client{Transport: failingTransport{}}, "abc123", "http://unreachable")
	require.NoError(t, err)
	return c
}

func TestDelete(t *testing.T) {
	const (
		org     = "an-org"
		project = "a-project"
		team    = "a-team"
		token   = "a-token"
	)
	stack := StackIdentifier{OrgName: org, ProjectName: project, StackName: "the-stack"}
	env := EnvironmentIdentifier{OrgName: org, ProjectName: project, EnvName: "an-env"}

	deletes := map[string]func(context.Context, *Client) error{
		"DeleteAccessToken": func(ctx context.Context, c *Client) error {
			return c.DeleteAccessToken(ctx, token)
		},
		"DeleteAgentPool": func(ctx context.Context, c *Client) error {
			return c.DeleteAgentPool(ctx, "a-pool", org, false)
		},
		"DeleteEnvironmentApprovalRule": func(ctx context.Context, c *Client) error {
			return c.DeleteEnvironmentApprovalRule(ctx, org, "a-rule")
		},
		"DeleteDeploymentSettings": func(ctx context.Context, c *Client) error {
			return c.DeleteDeploymentSettings(ctx, stack)
		},
		"DeleteEnvironmentSchedule": func(ctx context.Context, c *Client) error {
			return c.DeleteEnvironmentSchedule(ctx, env, "a-schedule")
		},
		"DeleteInsightsAccount": func(ctx context.Context, c *Client) error {
			return c.DeleteInsightsAccount(ctx, org, "an-account")
		},
		"DeleteMemberFromOrg": func(ctx context.Context, c *Client) error {
			return c.DeleteMemberFromOrg(ctx, org, "a-user")
		},
		"DeleteOrgAccessToken": func(ctx context.Context, c *Client) error {
			return c.DeleteOrgAccessToken(ctx, token, org)
		},
		"DeleteOidcIssuer": func(ctx context.Context, c *Client) error {
			return c.DeleteOidcIssuer(ctx, org, "an-issuer")
		},
		"DeletePolicyGroup": func(ctx context.Context, c *Client) error {
			return c.DeletePolicyGroup(ctx, org, "a-group")
		},
		"DeletePolicyPack": func(ctx context.Context, c *Client) error {
			return c.DeletePolicyPack(ctx, org, "a-policy-pack")
		},
		"DeletePolicyPackVersion": func(ctx context.Context, c *Client) error {
			return c.DeletePolicyPackVersion(ctx, org, "a-policy-pack", "1.0.0")
		},
		"DeleteStack": func(ctx context.Context, c *Client) error {
			return c.DeleteStack(ctx, stack, false)
		},
		"DeleteStackSchedule": func(ctx context.Context, c *Client) error {
			return c.DeleteStackSchedule(ctx, stack, "a-schedule")
		},
		"DeleteStackTag": func(ctx context.Context, c *Client) error {
			return c.DeleteStackTag(ctx, stack, "a-tag")
		},
		"DeleteTeam": func(ctx context.Context, c *Client) error {
			return c.DeleteTeam(ctx, org, team)
		},
		"DeleteTeamAccessToken": func(ctx context.Context, c *Client) error {
			return c.DeleteTeamAccessToken(ctx, token, org, team)
		},
		"DeleteWebhook": func(ctx context.Context, c *Client) error {
			return c.DeleteWebhook(ctx, org, nil, nil, nil, "a-webhook")
		},
		"DeleteTemplateSource": func(ctx context.Context, c *Client) error {
			return c.DeleteTemplateSource(ctx, org, "a-template-source")
		},
		"RemoveEnvironmentSettings": func(ctx context.Context, c *Client) error {
			return c.RemoveEnvironmentSettings(ctx, TeamEnvironmentSettingsRequest{
				Organization: org, Team: team, Environment: "an-env", Project: project,
			})
		},
		"RemoveRoleFromTeam": func(ctx context.Context, c *Client) error {
			return c.RemoveRoleFromTeam(ctx, org, team, "a-role")
		},
		"RemoveStackPermission": func(ctx context.Context, c *Client) error {
			return c.RemoveStackPermission(ctx, stack, team)
		},
	}

	for name, del := range deletes {
		t.Run(name, func(t *testing.T) {
			for _, tc := range []struct {
				name    string
				status  int
				wantErr bool
			}{
				{"deleted", http.StatusNoContent, false},
				{"already gone", http.StatusNotFound, false},
				{"server error", http.StatusInternalServerError, true},
			} {
				t.Run(tc.name, func(t *testing.T) {
					c := startTestServerMulti(t, func(*http.Request) (int, any) {
						if tc.status == http.StatusNoContent {
							return tc.status, nil
						}
						return tc.status, ErrorResponse{StatusCode: tc.status, Message: http.StatusText(tc.status)}
					})
					err := del(t.Context(), c)
					if tc.wantErr {
						assert.Error(t, err, "a %d response should fail the delete", tc.status)
					} else {
						assert.NoError(t, err, "a %d response should count as deleted", tc.status)
					}
				})
			}

			t.Run("server unreachable", func(t *testing.T) {
				err := del(t.Context(), newUnreachableClient(t))
				assert.Error(t, err, "a request with no response should fail the delete")
			})
		})
	}
}
