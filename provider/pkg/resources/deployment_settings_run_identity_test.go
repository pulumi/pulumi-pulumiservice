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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const (
	testRunAsLogin = "service-principal:d333a711"
	testRunAsTeam  = "platform"
	testRunAsRole  = "role-id"
)

func runAsUser() *pulumiapi.AssignmentPrincipal {
	return &pulumiapi.AssignmentPrincipal{User: &pulumiapi.ServicePrincipalUser{GitHubLogin: testRunAsLogin}}
}

func runAsTeam() *pulumiapi.AssignmentPrincipal {
	return &pulumiapi.AssignmentPrincipal{Team: &pulumiapi.AssignmentTeam{Name: testRunAsTeam}}
}

func operationContextInputs(oc map[string]property.Value) property.Map {
	return deploymentSettingsInputs(map[string]property.Value{gcOperationContext: property.New(oc)})
}

func servicePrincipalInput(key, value string) map[string]property.Value {
	return map[string]property.Value{
		gcServicePrincipal: property.New(map[string]property.Value{key: property.New(value)}),
	}
}

func TestDeploymentSettingsRunIdentityRoundtrip(t *testing.T) {
	for name, operation := range map[string]pulumiapi.OperationContext{
		"service principal": {ServicePrincipal: runAsUser()},
		"team":              {ServicePrincipal: runAsTeam()},
		"role":              {Role: &pulumiapi.DeploymentRole{ID: testRunAsRole}},
	} {
		t.Run(name, func(t *testing.T) {
			initial := PulumiServiceDeploymentSettingsInput{
				DeploymentSettings: pulumiapi.DeploymentSettings{Operation: &operation},
			}
			decoded := decodeSettings(t, initial.ToPropertyMap(nil, nil, true))
			assert.EqualValues(t, initial, decoded)
		})
	}
}

func TestDeploymentSettingsCreateRunsAsServicePrincipal(t *testing.T) {
	// Pulumi Cloud answers with the service principal resolved.
	resolved := &pulumiapi.AssignmentPrincipal{User: &pulumiapi.ServicePrincipalUser{
		Name:        "ci-deployer",
		GitHubLogin: testRunAsLogin,
	}}
	client := &DeploymentSettingsClientMock{
		respond: func(s pulumiapi.DeploymentSettings) pulumiapi.DeploymentSettings {
			op := *s.Operation
			op.ServicePrincipal = resolved
			s.Operation = &op
			return s
		},
	}
	server := newDeploymentSettingsServer(t, client)

	resp, err := server.Create(p.CreateRequest{
		Urn:        deploymentSettingsURN,
		Properties: operationContextInputs(servicePrincipalInput(gcLogin, testRunAsLogin)),
	})
	require.NoError(t, err)

	require.Len(t, client.created, 1)
	assert.Equal(t, runAsUser(), client.created[0].Operation.ServicePrincipal)
	assert.Nil(t, client.created[0].Operation.Role)

	sp := resp.Properties.Get(gcOperationContext).AsMap().Get(gcServicePrincipal).AsMap()
	assert.Equal(t, property.NewMap(map[string]property.Value{gcLogin: property.New(testRunAsLogin)}), sp)
}

func TestDeploymentSettingsCreateRunsWithRole(t *testing.T) {
	client := &DeploymentSettingsClientMock{}
	server := newDeploymentSettingsServer(t, client)

	resp, err := server.Create(p.CreateRequest{
		Urn: deploymentSettingsURN,
		Properties: operationContextInputs(map[string]property.Value{
			gcRoleID: property.New(testRunAsRole),
		}),
	})
	require.NoError(t, err)

	require.Len(t, client.created, 1)
	assert.Equal(t, &pulumiapi.DeploymentRole{ID: testRunAsRole}, client.created[0].Operation.Role)
	assert.Nil(t, client.created[0].Operation.ServicePrincipal)
	assert.Equal(t, property.New(testRunAsRole), resp.Properties.Get(gcOperationContext).AsMap().Get(gcRoleID))
}

// A service principal selected outside the program shows up on refresh, so the
// next update doesn't drop it unnoticed.
func TestDeploymentSettingsRefreshReportsServicePrincipal(t *testing.T) {
	for name, tc := range map[string]struct {
		selected *pulumiapi.AssignmentPrincipal
		want     map[string]property.Value
	}{
		"service principal": {runAsUser(), servicePrincipalInput(gcLogin, testRunAsLogin)},
		"team":              {runAsTeam(), servicePrincipalInput(gcTeamName, testRunAsTeam)},
	} {
		t.Run(name, func(t *testing.T) {
			server := newDeploymentSettingsServer(t, buildDeploymentSettingsClientMock(
				func() (*pulumiapi.DeploymentSettings, error) {
					return &pulumiapi.DeploymentSettings{Operation: &pulumiapi.OperationContext{
						PreRunCommands:   []string{"make"},
						ServicePrincipal: tc.selected,
					}}, nil
				},
			))
			old := operationContextInputs(map[string]property.Value{
				"preRunCommands": property.New([]property.Value{property.New("make")}),
			})

			resp, err := server.Read(p.ReadRequest{
				ID: gcSettingsID, Urn: deploymentSettingsURN, Properties: old, Inputs: old,
			})
			require.NoError(t, err)

			oc := resp.Inputs.Get(gcOperationContext).AsMap()
			assert.Equal(t, property.New(tc.want[gcServicePrincipal].AsMap()), oc.Get(gcServicePrincipal))
		})
	}
}

func TestDeploymentSettingsCheckRunIdentity(t *testing.T) {
	unknown := property.New(property.Computed)
	for _, tt := range []struct {
		name   string
		oc     map[string]property.Value
		failOn string
		reason string
	}{
		{name: "service principal login passes", oc: servicePrincipalInput(gcLogin, testRunAsLogin)},
		{name: "team passes", oc: servicePrincipalInput(gcTeamName, testRunAsTeam)},
		{name: "role passes", oc: map[string]property.Value{gcRoleID: property.New(testRunAsRole)}},
		{
			name: "unknown login passes",
			oc: map[string]property.Value{
				gcServicePrincipal: property.New(map[string]property.Value{gcLogin: unknown}),
			},
		},
		{
			name: "login and team fail",
			oc: map[string]property.Value{
				gcServicePrincipal: property.New(map[string]property.Value{
					gcLogin:    property.New(testRunAsLogin),
					gcTeamName: property.New(testRunAsTeam),
				}),
			},
			failOn: "operationContext.servicePrincipal.teamName",
			reason: "set only one of login and teamName",
		},
		{
			name: "neither login nor team fails",
			oc: map[string]property.Value{
				gcServicePrincipal: property.New(map[string]property.Value{}),
			},
			failOn: "operationContext.servicePrincipal",
			reason: "set login or teamName",
		},
		{
			name: "role and service principal fail",
			oc: map[string]property.Value{
				gcRoleID: property.New(testRunAsRole),
				gcServicePrincipal: property.New(map[string]property.Value{
					gcLogin: unknown,
				}),
			},
			failOn: "operationContext.roleId",
			reason: "set only one of roleId and servicePrincipal",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := newDeploymentSettingsServer(t, &DeploymentSettingsClientMock{})
			resp, err := server.Check(p.CheckRequest{Urn: deploymentSettingsURN, Inputs: operationContextInputs(tt.oc)})
			require.NoError(t, err)
			if tt.failOn == "" {
				assert.Empty(t, resp.Failures)
				return
			}
			require.Len(t, resp.Failures, 1, "failures: %v", resp.Failures)
			assert.Equal(t, tt.failOn, resp.Failures[0].Property)
			assert.Equal(t, tt.reason, resp.Failures[0].Reason)
		})
	}
}
