package pulumiapi

import (
	"encoding/json"
	"net/http"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testDeploymentSettingsOrgName     = "an-organization"
	testDeploymentSettingsProjectName = "a-project"
	testDeploymentSettingsStackName   = "a-stack"
)

func TestGetDeploymentSettings(t *testing.T) {

	orgName := testDeploymentSettingsOrgName
	projectName := testDeploymentSettingsProjectName
	stackName := testDeploymentSettingsStackName

	t.Run("Happy Path", func(t *testing.T) {
		dsValue := DeploymentSettings{
			Operation:     &OperationContext{},
			GitHub:        &DeploymentSettingsGitHub{},
			SourceContext: &SourceContext{},
			Executor:      &ExecutorContext{},
		}

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath: "/" + path.Join(
				"api",
				"stacks",
				orgName,
				projectName,
				stackName,
				"deployments",
				"settings",
			),
			ResponseCode: 200,
			ResponseBody: dsValue,
		})

		ds, err := c.GetDeploymentSettings(t.Context(), StackIdentifier{
			OrgName:     orgName,
			ProjectName: projectName,
			StackName:   stackName,
		})

		assert.Nil(t, err)
		assert.Equal(t, *ds, dsValue)
	})

	t.Run("404", func(t *testing.T) {
		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodGet,
			ExpectedReqPath: "/" + path.Join(
				"api",
				"stacks",
				orgName,
				projectName,
				stackName,
				"deployments",
				"settings",
			),
			ResponseCode: 404,
			ResponseBody: ErrorResponse{
				StatusCode: 404,
				Message:    "not found",
			},
		})

		ds, err := c.GetDeploymentSettings(t.Context(), StackIdentifier{
			OrgName:     orgName,
			ProjectName: projectName,
			StackName:   stackName,
		})

		assert.Nil(t, ds, "deployment settings should be nil since error was returned")
		assert.Nil(t, err, "err should be nil since error was returned")
	})
}

func TestCreateDeploymentSettings(t *testing.T) {

	orgName := testDeploymentSettingsOrgName
	projectName := testDeploymentSettingsProjectName
	stackName := testDeploymentSettingsStackName

	t.Run("Happy Path", func(t *testing.T) {
		dsValue := DeploymentSettings{
			Operation:     &OperationContext{},
			GitHub:        &DeploymentSettingsGitHub{},
			SourceContext: &SourceContext{},
			Executor:      &ExecutorContext{},
			CacheOptions:  &CacheOptions{},
		}

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPut,
			ExpectedReqPath: "/" + path.Join(
				"api",
				"stacks",
				orgName,
				projectName,
				stackName,
				"deployments",
				"settings",
			),
			ResponseCode:    201,
			ExpectedReqBody: dsValue,
			ResponseBody:    dsValue,
		})

		response, err := c.CreateDeploymentSettings(t.Context(), StackIdentifier{
			OrgName:     orgName,
			ProjectName: projectName,
			StackName:   stackName,
		}, dsValue)

		assert.Nil(t, err)
		assert.Equal(t, dsValue, *response)
	})

	t.Run("Executor image credentials", func(t *testing.T) {
		dsValue := DeploymentSettings{
			Executor: &ExecutorContext{
				ExecutorImage: &DockerImage{
					Reference: "myreg.example.com/custom-pulumi:latest",
					Credentials: &DockerImageCredentials{
						Username: "registry-user",
						Password: SecretValue{Secret: true, Value: "registry-password"},
					},
				},
			},
		}

		// Pin the wire format. Two things matter here: executorImage switches
		// from a bare string to an object as soon as credentials are set, and
		// the password must go out in the {"secret": ...} form. Pulumi Cloud
		// only encrypts secret values that are flagged as such, so a password
		// sent as a bare string would be persisted in plaintext.
		expectedBody := json.RawMessage(`{
			"executorContext": {
				"executorImage": {
					"reference": "myreg.example.com/custom-pulumi:latest",
					"credentials": {
						"username": "registry-user",
						"password": { "secret": "registry-password" }
					}
				}
			}
		}`)

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPut,
			ExpectedReqPath: "/" + path.Join(
				"api",
				"stacks",
				orgName,
				projectName,
				stackName,
				"deployments",
				"settings",
			),
			ResponseCode:    201,
			ExpectedReqBody: expectedBody,
			ResponseBody:    dsValue,
		})

		_, err := c.CreateDeploymentSettings(t.Context(), StackIdentifier{
			OrgName:     orgName,
			ProjectName: projectName,
			StackName:   stackName,
		}, dsValue)

		assert.Nil(t, err)
	})

	t.Run("Executor image without credentials keeps the bare string form", func(t *testing.T) {
		dsValue := DeploymentSettings{
			Executor: &ExecutorContext{
				ExecutorImage: &DockerImage{Reference: "pulumi/pulumi-nodejs:latest"},
			},
		}

		// Existing programs that only set an image must keep serializing the
		// image as a plain string.
		expectedBody := json.RawMessage(
			`{"executorContext": {"executorImage": "pulumi/pulumi-nodejs:latest"}}`,
		)

		c := startTestServer(t, testServerConfig{
			ExpectedReqMethod: http.MethodPut,
			ExpectedReqPath: "/" + path.Join(
				"api", "stacks", orgName, projectName, stackName, "deployments", "settings",
			),
			ResponseCode:    201,
			ExpectedReqBody: expectedBody,
			ResponseBody:    dsValue,
		})

		_, err := c.CreateDeploymentSettings(t.Context(), StackIdentifier{
			OrgName:     orgName,
			ProjectName: projectName,
			StackName:   stackName,
		}, dsValue)

		assert.Nil(t, err)
	})
}

// Pulumi Cloud returns the identity deployments run as resolved, but takes a
// service principal by reference. The writes must send the reference form.
func TestDeploymentSettingsRunIdentityWireFormat(t *testing.T) {
	stack := StackIdentifier{
		OrgName:     testDeploymentSettingsOrgName,
		ProjectName: testDeploymentSettingsProjectName,
		StackName:   testDeploymentSettingsStackName,
	}
	settingsPath := "/" + path.Join("api", "stacks", stack.OrgName, stack.ProjectName, stack.StackName,
		"deployments", "settings")

	for _, tc := range []struct {
		name      string
		operation OperationContext
		body      string
	}{
		{
			name: "user-defined service principal",
			operation: OperationContext{ServicePrincipal: &AssignmentPrincipal{
				User: &ServicePrincipalUser{GitHubLogin: "service-principal:d333a711"},
			}},
			body: `{"operationContext": {"servicePrincipal": {"userId": "service-principal:d333a711"}}}`,
		},
		{
			name: "team",
			operation: OperationContext{ServicePrincipal: &AssignmentPrincipal{
				Team: &AssignmentTeam{Name: "platform"},
			}},
			body: `{"operationContext": {"servicePrincipal": {"teamName": "platform"}}}`,
		},
		{
			name:      "role",
			operation: OperationContext{Role: &DeploymentRole{ID: "role-id"}},
			body:      `{"operationContext": {"role": {"id": "role-id"}}}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			settings := DeploymentSettings{Operation: &tc.operation}
			for _, write := range []struct {
				name string
				call func(*Client) (*DeploymentSettings, error)
			}{
				{"create", func(c *Client) (*DeploymentSettings, error) {
					return c.CreateDeploymentSettings(t.Context(), stack, settings)
				}},
				{"update", func(c *Client) (*DeploymentSettings, error) {
					return c.UpdateDeploymentSettings(t.Context(), stack, settings)
				}},
			} {
				t.Run(write.name, func(t *testing.T) {
					c := startTestServer(t, testServerConfig{
						ExpectedReqMethod: http.MethodPut,
						ExpectedReqPath:   settingsPath,
						ExpectedReqBody:   json.RawMessage(tc.body),
						ResponseCode:      http.StatusOK,
						ResponseBody:      settings,
					})
					got, err := write.call(c)
					assert.NoError(t, err)
					assert.Equal(t, &settings, got)
				})
			}
		})
	}
}
