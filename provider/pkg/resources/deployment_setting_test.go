package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/util"
)

type getDeploymentSettingsFunc func() (*pulumiapi.DeploymentSettings, error)

type DeploymentSettingsClientMock struct {
	getDeploymentSettingsFunc getDeploymentSettingsFunc
}

func (c *DeploymentSettingsClientMock) CreateDeploymentSettings(
	_ context.Context,
	_ pulumiapi.StackIdentifier,
	_ pulumiapi.DeploymentSettings,
) (*pulumiapi.DeploymentSettings, error) {
	return nil, nil
}

func (c *DeploymentSettingsClientMock) UpdateDeploymentSettings(
	_ context.Context,
	_ pulumiapi.StackIdentifier,
	_ pulumiapi.DeploymentSettings,
) (*pulumiapi.DeploymentSettings, error) {
	return nil, nil
}

func (c *DeploymentSettingsClientMock) GetDeploymentSettings(
	_ context.Context,
	_ pulumiapi.StackIdentifier,
) (*pulumiapi.DeploymentSettings, error) {
	return c.getDeploymentSettingsFunc()
}

func (c *DeploymentSettingsClientMock) DeleteDeploymentSettings(
	_ context.Context,
	_ pulumiapi.StackIdentifier,
) error {
	return nil
}

func buildDeploymentSettingsClientMock(
	getDeploymentSettingsFunc getDeploymentSettingsFunc,
) *DeploymentSettingsClientMock {
	return &DeploymentSettingsClientMock{
		getDeploymentSettingsFunc,
	}
}

func TestDeploymentSettings(t *testing.T) {
	t.Run("Read when the resource is not found", func(t *testing.T) {
		mockedClient := buildDeploymentSettingsClientMock(
			func() (*pulumiapi.DeploymentSettings, error) { return nil, nil },
		)

		provider := PulumiServiceDeploymentSettingsResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "abc/def/123",
			Urn: "urn:123",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "")
		assert.Nil(t, resp.Properties)
	})

	t.Run("Read when the resource is found", func(t *testing.T) {
		mockedClient := buildDeploymentSettingsClientMock(
			func() (*pulumiapi.DeploymentSettings, error) {
				return &pulumiapi.DeploymentSettings{
					Operation:     &pulumiapi.OperationContext{},
					GitHub:        &pulumiapi.DeploymentSettingsGitHub{},
					SourceContext: &pulumiapi.SourceContext{},
					Executor:      &pulumiapi.ExecutorContext{},
				}, nil
			},
		)

		provider := PulumiServiceDeploymentSettingsResource{
			Client: mockedClient,
		}

		req := pulumirpc.ReadRequest{
			Id:  "abc/def/123",
			Urn: "urn:123",
		}

		resp, err := provider.Read(&req)

		assert.NoError(t, err)
		assert.Equal(t, resp.Id, "abc/def/123")
	})
}

func TestDeploymentSettingsRoundtrip(t *testing.T) {
	initial := PulumiServiceDeploymentSettingsInput{
		DeploymentSettings: pulumiapi.DeploymentSettings{
			CacheOptions: &pulumiapi.CacheOptions{
				Enable: true,
			},
		}}

	encoded := initial.ToPropertyMap(nil, nil, true)
	decoded := (&PulumiServiceDeploymentSettingsResource{}).ToPulumiServiceDeploymentSettingsInput(encoded)

	assert.EqualValues(t, initial, decoded)
}

const (
	testRegistryUsername = "registry-user"
	testRegistryPassword = "registry-password"
	// What the settings API returns in place of a secret's plaintext.
	testRedactedSecret = "[secret]"
)

// executorImageCredentialsSettings builds deployment settings carrying a custom
// executor image with registry credentials whose password is `password`.
func executorImageCredentialsSettings(password pulumiapi.SecretValue) pulumiapi.DeploymentSettings {
	return pulumiapi.DeploymentSettings{
		Executor: &pulumiapi.ExecutorContext{
			ExecutorImage: &pulumiapi.DockerImage{
				Reference: "myreg.example.com/custom-pulumi:latest",
				Credentials: &pulumiapi.DockerImageCredentials{
					Username: testRegistryUsername,
					Password: password,
				},
			},
		},
	}
}

// plaintextPassword is what the user's program supplies.
func plaintextPassword() pulumiapi.SecretValue {
	return pulumiapi.SecretValue{Secret: true, Value: testRegistryPassword}
}

// redactedPassword is what the settings API hands back: the plaintext replaced
// by a marker, with the ciphertext alongside it.
func redactedPassword() pulumiapi.SecretValue {
	return pulumiapi.SecretValue{
		Secret:     true,
		Value:      testRedactedSecret,
		Ciphertext: []byte("ciphertext-from-the-service"),
	}
}

// credentialsPropertyMap digs the executor image credentials back out of an
// encoded property map.
func credentialsPropertyMap(t *testing.T, encoded resource.PropertyMap) resource.PropertyMap {
	t.Helper()
	executorContext := encoded["executorContext"]
	assert.True(t, executorContext.IsObject(), "executorContext should be present")
	credentials := executorContext.ObjectValue()["credentials"]
	assert.True(t, credentials.IsObject(), "credentials should be present")
	return credentials.ObjectValue()
}

func TestDeploymentSettingsExecutorImageCredentialsRoundtrip(t *testing.T) {
	settings := executorImageCredentialsSettings(plaintextPassword())
	initial := PulumiServiceDeploymentSettingsInput{DeploymentSettings: settings}

	// Create mode: plaintext inputs are available, no prior cipher state yet.
	encoded := initial.ToPropertyMap(&settings, nil, true)
	decoded := (&PulumiServiceDeploymentSettingsResource{}).ToPulumiServiceDeploymentSettingsInput(encoded)

	assert.EqualValues(t, initial, decoded)
	assert.Equal(t, testRegistryPassword, decoded.Executor.ExecutorImage.Credentials.Password.Value)
}

// An executor image without credentials must encode exactly as it did before
// credentials existed, so upgrading the provider produces no diff.
func TestDeploymentSettingsExecutorImageWithoutCredentialsRoundtrip(t *testing.T) {
	settings := pulumiapi.DeploymentSettings{
		Executor: &pulumiapi.ExecutorContext{
			ExecutorImage: &pulumiapi.DockerImage{Reference: "pulumi/pulumi-nodejs:latest"},
		},
	}
	initial := PulumiServiceDeploymentSettingsInput{DeploymentSettings: settings}

	encoded := initial.ToPropertyMap(&settings, nil, true)
	decoded := (&PulumiServiceDeploymentSettingsResource{}).ToPulumiServiceDeploymentSettingsInput(encoded)

	assert.EqualValues(t, initial, decoded)
	assert.NotContains(t, encoded["executorContext"].ObjectValue(), resource.PropertyKey("credentials"))
	assert.Nil(t, decoded.Executor.ExecutorImage.Credentials)
}

// On refresh the API hands back the password redacted to "[secret]" plus its
// ciphertext, never the plaintext. The user's plaintext input must be carried
// forward so refresh reports no change.
func TestDeploymentSettingsExecutorImageCredentialsRefresh(t *testing.T) {
	fromAPI := executorImageCredentialsSettings(redactedPassword())
	priorState := executorImageCredentialsSettings(redactedPassword())
	plaintextInputs := executorImageCredentialsSettings(plaintextPassword())

	input := PulumiServiceDeploymentSettingsInput{DeploymentSettings: fromAPI}

	inputs := credentialsPropertyMap(t, input.ToPropertyMap(&plaintextInputs, &priorState, true))
	assert.Equal(t, testRegistryPassword, util.GetSecretOrStringValue(inputs["password"]),
		"refresh must preserve the plaintext password from prior inputs")

	outputs := credentialsPropertyMap(t, input.ToPropertyMap(&plaintextInputs, &priorState, false))
	assert.Equal(t, testRedactedSecret, util.GetSecretOrStringValue(outputs["password"]),
		"outputs hold what the service returned, not the plaintext")
}

// With no prior cipher state to merge against, the plaintext cannot be trusted,
// so the input is blanked and the engine surfaces a diff.
func TestDeploymentSettingsExecutorImageCredentialsRefreshWithoutPriorState(t *testing.T) {
	fromAPI := executorImageCredentialsSettings(redactedPassword())
	priorState := pulumiapi.DeploymentSettings{Executor: &pulumiapi.ExecutorContext{}}
	plaintextInputs := executorImageCredentialsSettings(plaintextPassword())

	input := PulumiServiceDeploymentSettingsInput{DeploymentSettings: fromAPI}

	inputs := credentialsPropertyMap(t, input.ToPropertyMap(&plaintextInputs, &priorState, true))
	assert.Equal(t, "", util.GetSecretOrStringValue(inputs["password"]))
}

// Import has no inputs at all, so the password is replaced with the placeholder
// that prompts the user to fill it in.
func TestDeploymentSettingsExecutorImageCredentialsImport(t *testing.T) {
	settings := executorImageCredentialsSettings(redactedPassword())
	input := PulumiServiceDeploymentSettingsInput{DeploymentSettings: settings}

	inputs := credentialsPropertyMap(t, input.ToPropertyMap(nil, nil, true))
	assert.Equal(t, "<REPLACE WITH ACTUAL SECRET VALUE>", util.GetSecretOrStringValue(inputs["password"]))
	assert.Equal(t, "registry-user", util.GetSecretOrStringValue(inputs["username"]))
}

func TestDeploymentSettingsVcsRoundtrip(t *testing.T) {
	deployPR := int64(1)
	initial := PulumiServiceDeploymentSettingsInput{
		DeploymentSettings: pulumiapi.DeploymentSettings{
			Vcs: pulumiapi.DeploymentSettingsVCSAzureDevOpsBuilder{
				DeploymentSettingsVCSBuilder: pulumiapi.DeploymentSettingsVCSBuilder{
					Repository:          "my-org/my-repo",
					InstallationID:      "129444790",
					DeployCommits:       true,
					PreviewPullRequests: true,
					PullRequestTemplate: false,
					Paths:               []string{"infra/**"},
					DeployPullRequest:   &deployPR,
				},
			}.Build(),
		},
	}

	encoded := initial.ToPropertyMap(nil, nil, true)
	decoded := (&PulumiServiceDeploymentSettingsResource{}).ToPulumiServiceDeploymentSettingsInput(encoded)

	assert.EqualValues(t, initial, decoded)
}
