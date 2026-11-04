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
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const testGCPProjectNumber = "123456789012"

func gcpSettings(project resource.PropertyMap) resource.PropertyMap {
	gcp := resource.PropertyMap{
		"workloadPoolId": resource.NewStringProperty("pool"),
		"providerId":     resource.NewStringProperty("oidc-provider"),
		"serviceAccount": resource.NewStringProperty("deploy@example.iam.gserviceaccount.com"),
	}
	for k, v := range project {
		gcp[k] = v
	}
	return resource.PropertyMap{
		gcOrganization: resource.NewStringProperty("an-org"),
		gcProject:      resource.NewStringProperty("a-project"),
		gcStack:        resource.NewStringProperty("a-stack"),
		gcOperationContext: resource.NewObjectProperty(resource.PropertyMap{
			gcOIDC: resource.NewObjectProperty(resource.PropertyMap{
				gcGCP: resource.NewObjectProperty(gcp),
			}),
		}),
	}
}

func propertyMap(props resource.PropertyMap) property.Map {
	return resource.FromResourcePropertyValue(resource.NewObjectProperty(props)).AsMap()
}

func gcpProps(t *testing.T, props property.Map) resource.PropertyMap {
	t.Helper()
	v := resourceMap(props)[gcOperationContext]
	for _, key := range []resource.PropertyKey{gcOIDC, gcGCP} {
		require.Truef(t, v.IsObject(), "expected an object before %q in %v", key, props)
		v = v.ObjectValue()[key]
	}
	require.True(t, v.IsObject(), "gcp should be present")
	return v.ObjectValue()
}

func TestDeploymentSettingsCheckGCPProjectNumber(t *testing.T) {
	str := resource.NewStringProperty
	tests := []struct {
		name         string
		project      resource.PropertyMap
		failOn       string
		failContains string
	}{
		{name: "numeric projectNumber passes", project: resource.PropertyMap{gcProjectNumber: str(testGCPProjectNumber)}},
		{name: "numeric projectId passes", project: resource.PropertyMap{gcProjectID: str(testGCPProjectNumber)}},
		{
			name:    "unknown projectNumber passes",
			project: resource.PropertyMap{gcProjectNumber: resource.MakeComputed(str(""))},
		},
		{
			name:    "secret projectNumber passes",
			project: resource.PropertyMap{gcProjectNumber: resource.MakeSecret(str(testGCPProjectNumber))},
		},
		{
			name:         "project ID in projectId fails",
			project:      resource.PropertyMap{gcProjectID: str("my-project")},
			failOn:       "operationContext.oidc.gcp.projectId",
			failContains: "gcloud projects describe",
		},
		{
			name:         "project ID in projectNumber fails",
			project:      resource.PropertyMap{gcProjectNumber: str("my-project")},
			failOn:       "operationContext.oidc.gcp.projectNumber",
			failContains: `"my-project" is not a GCP project number`,
		},
		{
			name:         "neither key set fails",
			project:      resource.PropertyMap{},
			failOn:       "operationContext.oidc.gcp.projectNumber",
			failContains: "missing required property",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := newDeploymentSettingsServer(t, &DeploymentSettingsClientMock{})
			resp, err := server.Check(p.CheckRequest{
				Urn:    deploymentSettingsURN,
				Inputs: propertyMap(gcpSettings(tt.project)),
			})
			require.NoError(t, err)
			if tt.failOn == "" {
				assert.Empty(t, resp.Failures)
				return
			}
			require.Len(t, resp.Failures, 1, "failures: %v", resp.Failures)
			assert.Equal(t, tt.failOn, resp.Failures[0].Property)
			assert.Contains(t, resp.Failures[0].Reason, tt.failContains)
		})
	}
}

func TestDeploymentSettingsGCPProjectNumberSentToPulumiCloud(t *testing.T) {
	str := resource.NewStringProperty
	tests := []struct {
		name    string
		project resource.PropertyMap
	}{
		{name: gcProjectNumber, project: resource.PropertyMap{gcProjectNumber: str(testGCPProjectNumber)}},
		{name: gcProjectID, project: resource.PropertyMap{gcProjectID: str(testGCPProjectNumber)}},
		{name: "projectNumber wins over projectId", project: resource.PropertyMap{
			gcProjectNumber: str(testGCPProjectNumber),
			gcProjectID:     str("999"),
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded := decodeSettings(t, gcpSettings(tt.project))
			assert.Equal(t, testGCPProjectNumber, decoded.Operation.OIDC.GCP.ProjectID)
		})
	}
}

func TestDeploymentSettingsGCPProjectKeyKept(t *testing.T) {
	str := resource.NewStringProperty
	tests := []struct {
		name     string
		project  resource.PropertyMap
		wantKeys []resource.PropertyKey
	}{
		{
			name:     gcProjectNumber,
			project:  resource.PropertyMap{gcProjectNumber: str(testGCPProjectNumber)},
			wantKeys: []resource.PropertyKey{gcProjectNumber},
		},
		{
			name:     gcProjectID,
			project:  resource.PropertyMap{gcProjectID: str(testGCPProjectNumber)},
			wantKeys: []resource.PropertyKey{gcProjectID},
		},
		{
			name: "both",
			project: resource.PropertyMap{
				gcProjectNumber: str(testGCPProjectNumber),
				gcProjectID:     str(testGCPProjectNumber),
			},
			wantKeys: []resource.PropertyKey{gcProjectID, gcProjectNumber},
		},
	}
	assertKeys := func(t *testing.T, label string, gcp resource.PropertyMap, want []resource.PropertyKey) {
		t.Helper()
		for _, key := range []resource.PropertyKey{gcProjectID, gcProjectNumber} {
			if slices.Contains(want, key) {
				assert.Equalf(t, str(testGCPProjectNumber), gcp[key], "%s: %s", label, key)
			} else {
				assert.Falsef(t, gcp[key].HasValue(), "%s: %s should be absent, got %v", label, key, gcp[key])
			}
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputs := propertyMap(gcpSettings(tt.project))
			client := &DeploymentSettingsClientMock{}
			server := newDeploymentSettingsServer(t, client)

			created, err := server.Create(p.CreateRequest{Urn: deploymentSettingsURN, Properties: inputs})
			require.NoError(t, err)
			assertKeys(t, "create state", gcpProps(t, created.Properties), tt.wantKeys)

			client.getDeploymentSettingsFunc = func() (*pulumiapi.DeploymentSettings, error) {
				return &client.created[0], nil
			}
			read, err := server.Read(p.ReadRequest{
				ID: created.ID, Urn: deploymentSettingsURN, Properties: created.Properties, Inputs: inputs,
			})
			require.NoError(t, err)
			assertKeys(t, "refresh state", gcpProps(t, read.Properties), tt.wantKeys)
			assertKeys(t, "refresh inputs", gcpProps(t, read.Inputs), tt.wantKeys)
		})
	}

	t.Run("import uses projectNumber", func(t *testing.T) {
		server := newDeploymentSettingsServer(t, buildDeploymentSettingsClientMock(
			func() (*pulumiapi.DeploymentSettings, error) {
				return &pulumiapi.DeploymentSettings{Operation: &pulumiapi.OperationContext{
					OIDC: &pulumiapi.OperationContextOIDCConfiguration{
						GCP: &pulumiapi.OperationContextGCPOIDCConfiguration{ProjectID: testGCPProjectNumber},
					},
				}}, nil
			},
		))
		read, err := server.Read(p.ReadRequest{ID: "an-org/a-project/a-stack", Urn: deploymentSettingsURN})
		require.NoError(t, err)
		assertKeys(t, "import inputs", gcpProps(t, read.Inputs), []resource.PropertyKey{gcProjectNumber})
	})
}
