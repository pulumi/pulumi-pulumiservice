package pulumiapi

import (
	"net/http"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
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
			OperationContext: &OperationContext{},
			GitHub:           &GitHubConfiguration{},
			SourceContext:    &SourceContext{},
			ExecutorContext:  &apitype.ExecutorContext{},
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

		ds, err := c.GetDeploymentSettings(ctx, StackIdentifier{
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

		ds, err := c.GetDeploymentSettings(ctx, StackIdentifier{
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
			OperationContext: &OperationContext{},
			GitHub:           &GitHubConfiguration{},
			SourceContext:    &SourceContext{},
			ExecutorContext:  &apitype.ExecutorContext{},
			CacheOptions:     &CacheOptions{},
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

		response, err := c.CreateDeploymentSettings(ctx, StackIdentifier{
			OrgName:     orgName,
			ProjectName: projectName,
			StackName:   stackName,
		}, dsValue)

		assert.Nil(t, err)
		assert.Equal(t, dsValue, *response)
	})
}
