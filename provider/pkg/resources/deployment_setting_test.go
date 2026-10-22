package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
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

// checkGitAuth runs Check over deployment settings carrying gitAuth and hands
// back the checked gitAuth property map.
func checkGitAuth(t *testing.T, gitAuth resource.PropertyMap) resource.PropertyMap {
	t.Helper()

	propertyMap := resource.PropertyMap{
		gcOrganization: resource.NewStringProperty("an-org"),
		gcProject:      resource.NewStringProperty("a-project"),
		gcStack:        resource.NewStringProperty("a-stack"),
		gcSourceContext: resource.NewObjectProperty(resource.PropertyMap{
			gcGit: resource.NewObjectProperty(resource.PropertyMap{
				"repoUrl": resource.NewStringProperty("https://github.com/pulumi/deploy-demos.git"),
				gcGitAuth: resource.NewObjectProperty(gitAuth),
			}),
		}),
	}

	news, err := plugin.MarshalProperties(propertyMap, util.StandardMarshal)
	assert.NoError(t, err)

	resp, err := (&PulumiServiceDeploymentSettingsResource{}).Check(&pulumirpc.CheckRequest{News: news})
	assert.NoError(t, err)
	assert.Empty(t, resp.Failures)

	checked, err := plugin.UnmarshalProperties(resp.GetInputs(), util.KeepSecretsUnmarshal)
	assert.NoError(t, err)

	sourceContext := checked[gcSourceContext]
	assert.True(t, sourceContext.IsObject(), "sourceContext should be present")
	git := sourceContext.ObjectValue()[gcGit]
	assert.True(t, git.IsObject(), "git should be present")
	checkedGitAuth := git.ObjectValue()[gcGitAuth]
	assert.True(t, checkedGitAuth.IsObject(), "gitAuth should be present")

	return checkedGitAuth.ObjectValue()
}

// decodeSettings decodes a property map the way Create, Update and Read do.
func decodeSettings(t *testing.T, props resource.PropertyMap) PulumiServiceDeploymentSettingsInput {
	t.Helper()
	decoded, err := decodeDeploymentSettingsInput(props)
	assert.NoError(t, err)
	return decoded
}

// assertSecret asserts that prop is a secret wrapping the given plaintext.
func assertSecret(t *testing.T, prop resource.PropertyValue, want string) {
	t.Helper()
	assert.True(t, prop.IsSecret(), "expected a secret, got %v", prop)
	assert.Equal(t, want, prop.SecretValue().Element.StringValue())
}

// The schema's `"secret": true` on these fields is codegen-only, and for a
// property nested inside a type only the .NET SDK acts on it, so Check has to
// promote them. Without this, a plaintext literal lands in the state file
// verbatim for every other language.
func TestDeploymentSettingsCheckGitAuthSecrets(t *testing.T) {
	t.Run("promotes plaintext sshAuth credentials", func(t *testing.T) {
		gitAuth := checkGitAuth(t, resource.PropertyMap{
			gcSSHAuth: resource.NewObjectProperty(resource.PropertyMap{
				gcSSHPrivateKey: resource.NewStringProperty("a-private-key"),
				gcPassword:      resource.NewStringProperty("a-key-password"),
			}),
		})

		sshAuth := gitAuth[gcSSHAuth].ObjectValue()
		assertSecret(t, sshAuth[gcSSHPrivateKey], "a-private-key")
		assertSecret(t, sshAuth[gcPassword], "a-key-password")
	})

	t.Run("promotes plaintext basicAuth credentials", func(t *testing.T) {
		gitAuth := checkGitAuth(t, resource.PropertyMap{
			gcBasicAuth: resource.NewObjectProperty(resource.PropertyMap{
				gcUsername: resource.NewStringProperty("alice"),
				gcPassword: resource.NewStringProperty("hunter2"),
			}),
		})

		basicAuth := gitAuth[gcBasicAuth].ObjectValue()
		assertSecret(t, basicAuth[gcUsername], "alice")
		assertSecret(t, basicAuth[gcPassword], "hunter2")
	})

	t.Run("leaves already secret credentials untouched", func(t *testing.T) {
		gitAuth := checkGitAuth(t, resource.PropertyMap{
			gcSSHAuth: resource.NewObjectProperty(resource.PropertyMap{
				gcSSHPrivateKey: resource.MakeSecret(resource.NewStringProperty("a-private-key")),
			}),
			gcBasicAuth: resource.NewObjectProperty(resource.PropertyMap{
				gcUsername: resource.MakeSecret(resource.NewStringProperty("alice")),
				gcPassword: resource.MakeSecret(resource.NewStringProperty("hunter2")),
			}),
		})

		// Not double-wrapped: the element is still the string, not another secret.
		assertSecret(t, gitAuth[gcSSHAuth].ObjectValue()[gcSSHPrivateKey], "a-private-key")
		assertSecret(t, gitAuth[gcBasicAuth].ObjectValue()[gcUsername], "alice")
		assertSecret(t, gitAuth[gcBasicAuth].ObjectValue()[gcPassword], "hunter2")
	})

	t.Run("optional sshAuth password stays absent", func(t *testing.T) {
		gitAuth := checkGitAuth(t, resource.PropertyMap{
			gcSSHAuth: resource.NewObjectProperty(resource.PropertyMap{
				gcSSHPrivateKey: resource.NewStringProperty("a-private-key"),
			}),
		})

		sshAuth := gitAuth[gcSSHAuth].ObjectValue()
		assertSecret(t, sshAuth[gcSSHPrivateKey], "a-private-key")
		assert.NotContains(t, sshAuth, resource.PropertyKey(gcPassword))
	})
}

// The executor image registry password is nested inside a type too, so it needs
// the same promotion in Check as the git auth credentials.
func TestDeploymentSettingsCheckExecutorImageCredentialsSecret(t *testing.T) {
	credentials := func(password resource.PropertyValue) resource.PropertyMap {
		t.Helper()

		propertyMap := resource.PropertyMap{
			gcOrganization: resource.NewStringProperty("an-org"),
			gcProject:      resource.NewStringProperty("a-project"),
			gcStack:        resource.NewStringProperty("a-stack"),
			gcExecutorContext: resource.NewObjectProperty(resource.PropertyMap{
				"executorImage": resource.NewStringProperty("myreg.example.com/custom-pulumi:latest"),
				gcCredentials: resource.NewObjectProperty(resource.PropertyMap{
					gcUsername: resource.NewStringProperty(testRegistryUsername),
					gcPassword: password,
				}),
			}),
		}

		news, err := plugin.MarshalProperties(propertyMap, util.StandardMarshal)
		assert.NoError(t, err)

		resp, err := (&PulumiServiceDeploymentSettingsResource{}).Check(&pulumirpc.CheckRequest{News: news})
		assert.NoError(t, err)
		assert.Empty(t, resp.Failures)

		checked, err := plugin.UnmarshalProperties(resp.GetInputs(), util.KeepSecretsUnmarshal)
		assert.NoError(t, err)

		executorContext := checked[gcExecutorContext]
		assert.True(t, executorContext.IsObject(), "executorContext should be present")
		checkedCredentials := executorContext.ObjectValue()[gcCredentials]
		assert.True(t, checkedCredentials.IsObject(), "credentials should be present")

		return checkedCredentials.ObjectValue()
	}

	t.Run("promotes a plaintext password", func(t *testing.T) {
		checked := credentials(resource.NewStringProperty(testRegistryPassword))

		assertSecret(t, checked[gcPassword], testRegistryPassword)
		// The username is not a credential the schema marks secret.
		assert.Equal(t, resource.NewStringProperty(testRegistryUsername), checked[gcUsername])
	})

	t.Run("leaves an already secret password untouched", func(t *testing.T) {
		checked := credentials(resource.MakeSecret(resource.NewStringProperty(testRegistryPassword)))

		// Not double-wrapped: the element is still the string, not another secret.
		assertSecret(t, checked[gcPassword], testRegistryPassword)
	})
}

// basicAuthSettings builds deployment settings carrying git basic auth.
func basicAuthSettings(username, password pulumiapi.SecretValue) pulumiapi.DeploymentSettings {
	return pulumiapi.DeploymentSettings{
		SourceContext: &pulumiapi.SourceContext{
			Git: &pulumiapi.SourceContextGit{
				RepoURL: "https://github.com/pulumi/deploy-demos.git",
				GitAuth: &pulumiapi.GitAuthConfig{
					BasicAuth: &pulumiapi.BasicAuth{UserName: username, Password: password},
				},
			},
		},
	}
}

// plaintextBasicAuth is what the user's program supplies.
func plaintextBasicAuth() pulumiapi.DeploymentSettings {
	return basicAuthSettings(
		pulumiapi.SecretValue{Secret: true, Value: "alice"},
		pulumiapi.SecretValue{Secret: true, Value: "hunter2"},
	)
}

// redactedBasicAuth is what the settings API hands back: the plaintext replaced
// by a marker, with the ciphertext alongside it.
func redactedBasicAuth() pulumiapi.DeploymentSettings {
	redacted := pulumiapi.SecretValue{
		Secret:     true,
		Value:      testRedactedSecret,
		Ciphertext: []byte("ciphertext-from-the-service"),
	}
	return basicAuthSettings(redacted, redacted)
}

// basicAuthPropertyMap digs the git basic auth back out of an encoded property map.
func basicAuthPropertyMap(t *testing.T, encoded resource.PropertyMap) resource.PropertyMap {
	t.Helper()

	value := encoded[gcSourceContext]
	for _, key := range []resource.PropertyKey{gcGit, gcGitAuth, gcBasicAuth} {
		assert.Truef(t, value.IsObject(), "expected an object before %q", key)
		value = value.ObjectValue()[key]
	}
	assert.True(t, value.IsObject(), "basicAuth should be present")
	return value.ObjectValue()
}

// The username is a twin-value secret like the password: Pulumi Cloud encrypts
// it and never hands the plaintext back, so it has to survive the same
// create/refresh/import paths.
func TestDeploymentSettingsBasicAuthRoundtrip(t *testing.T) {
	settings := plaintextBasicAuth()
	initial := PulumiServiceDeploymentSettingsInput{DeploymentSettings: settings}

	// Create mode: plaintext inputs are available, no prior cipher state yet.
	encoded := initial.ToPropertyMap(&settings, nil, true)
	decoded := decodeSettings(t, encoded)

	assert.EqualValues(t, initial, decoded)
	assertSecret(t, basicAuthPropertyMap(t, encoded)[gcUsername], "alice")
}

// On refresh the API hands back the credentials redacted plus their ciphertext,
// never the plaintext. The user's plaintext input must be carried forward so
// refresh reports no change.
func TestDeploymentSettingsBasicAuthRefresh(t *testing.T) {
	fromAPI := redactedBasicAuth()
	priorState := redactedBasicAuth()
	plaintextInputs := plaintextBasicAuth()

	input := PulumiServiceDeploymentSettingsInput{DeploymentSettings: fromAPI}

	inputs := basicAuthPropertyMap(t, input.ToPropertyMap(&plaintextInputs, &priorState, true))
	assertSecret(t, inputs[gcUsername], "alice")
	assertSecret(t, inputs[gcPassword], "hunter2")

	outputs := basicAuthPropertyMap(t, input.ToPropertyMap(&plaintextInputs, &priorState, false))
	assert.Equal(t, testRedactedSecret, util.GetSecretOrStringValue(outputs[gcUsername]),
		"outputs hold what the service returned, not the plaintext")
}

// Import has no inputs at all, so both credentials are replaced with the
// placeholder that prompts the user to fill them in.
func TestDeploymentSettingsBasicAuthImport(t *testing.T) {
	settings := redactedBasicAuth()
	input := PulumiServiceDeploymentSettingsInput{DeploymentSettings: settings}

	inputs := basicAuthPropertyMap(t, input.ToPropertyMap(nil, nil, true))
	assertSecret(t, inputs[gcUsername], "<REPLACE WITH ACTUAL SECRET VALUE>")
	assertSecret(t, inputs[gcPassword], "<REPLACE WITH ACTUAL SECRET VALUE>")
}

// Settings with no sourceContext at all must pass through Check unchanged.
func TestDeploymentSettingsCheckWithoutSourceContext(t *testing.T) {
	propertyMap := resource.PropertyMap{
		gcOrganization: resource.NewStringProperty("an-org"),
		gcProject:      resource.NewStringProperty("a-project"),
		gcStack:        resource.NewStringProperty("a-stack"),
	}

	news, err := plugin.MarshalProperties(propertyMap, util.StandardMarshal)
	assert.NoError(t, err)

	resp, err := (&PulumiServiceDeploymentSettingsResource{}).Check(&pulumirpc.CheckRequest{News: news})
	assert.NoError(t, err)
	assert.Empty(t, resp.Failures)

	checked, err := plugin.UnmarshalProperties(resp.GetInputs(), util.KeepSecretsUnmarshal)
	assert.NoError(t, err)
	assert.Equal(t, propertyMap, checked)
}

func TestDeploymentSettingsRoundtrip(t *testing.T) {
	initial := PulumiServiceDeploymentSettingsInput{
		DeploymentSettings: pulumiapi.DeploymentSettings{
			CacheOptions: &pulumiapi.CacheOptions{
				Enable: true,
			},
		}}

	encoded := initial.ToPropertyMap(nil, nil, true)
	decoded := decodeSettings(t, encoded)

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
	decoded := decodeSettings(t, encoded)

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
	decoded := decodeSettings(t, encoded)

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
					DeployTags:          true,
					TagFilters:          []string{"v*", "!*-rc*"},
				},
			}.Build(),
		},
	}

	encoded := initial.ToPropertyMap(nil, nil, true)
	decoded := decodeSettings(t, encoded)

	assert.EqualValues(t, initial, decoded)
}

func TestDeploymentSettingsGitHubRoundtrip(t *testing.T) {
	initial := PulumiServiceDeploymentSettingsInput{
		DeploymentSettings: pulumiapi.DeploymentSettings{
			GitHub: &pulumiapi.DeploymentSettingsGitHub{
				Repository:          "my-org/my-repo",
				DeployCommits:       true,
				PreviewPullRequests: true,
				PullRequestTemplate: false,
				Paths:               []string{"infra/**"},
				DeployTags:          true,
				TagFilters:          []string{"v*", "!*-rc*"},
			},
		},
	}

	encoded := initial.ToPropertyMap(nil, nil, true)
	decoded := decodeSettings(t, encoded)

	assert.EqualValues(t, initial, decoded)
}

func TestDecodeDeploymentSettingsVcsProviders(t *testing.T) {
	vcsProps := func(provider string) resource.PropertyMap {
		return resource.PropertyMap{
			"vcs": resource.NewObjectProperty(resource.PropertyMap{
				"provider":   resource.NewStringProperty(provider),
				"repository": resource.NewStringProperty("my-org/my-repo"),
			}),
		}
	}

	for _, want := range []pulumiapi.DeploymentSettingsVCS{
		pulumiapi.DeploymentSettingsVCSAzureDevOpsBuilder{}.Build(),
		pulumiapi.DeploymentSettingsVCSBitbucketBuilder{}.Build(),
		pulumiapi.DeploymentSettingsVCSCustomBuilder{}.Build(),
		pulumiapi.DeploymentSettingsVCSGitHubBuilder{}.Build(),
		pulumiapi.DeploymentSettingsVCSGitLabBuilder{}.Build(),
	} {
		provider, err := want.GetDiscriminatorValue()
		assert.NoError(t, err)
		t.Run(provider, func(t *testing.T) {
			decoded := decodeSettings(t, vcsProps(provider))
			got, err := decoded.Vcs.GetDiscriminatorValue()
			assert.NoError(t, err)
			assert.Equal(t, provider, got)
			assert.Equal(t, "my-org/my-repo", decoded.Vcs.Repository())
		})
	}

	t.Run("unknown provider", func(t *testing.T) {
		assert.Nil(t, decodeSettings(t, vcsProps("svn")).Vcs)
	})
}

func TestDecodeDeploymentSettingsOperationContext(t *testing.T) {
	const (
		plainKey    = "PLAIN"
		plainValue  = "plain-value"
		secretKey   = "SECRET"
		secretValue = "secret-value"
	)
	operationContext := func(envVars resource.PropertyValue) resource.PropertyMap {
		return resource.PropertyMap{
			"preRunCommands":       resource.NewArrayProperty([]resource.PropertyValue{resource.NewStringProperty("make")}),
			"environmentVariables": envVars,
			"options": resource.NewObjectProperty(resource.PropertyMap{
				"skipInstallDependencies":     resource.NewBoolProperty(true),
				"skipIntermediateDeployments": resource.NewBoolProperty(true),
				"deleteAfterDestroy":          resource.NewBoolProperty(true),
				"shell":                       resource.NewStringProperty("/bin/bash"),
			}),
			"oidc": resource.NewObjectProperty(resource.PropertyMap{
				"aws": resource.NewObjectProperty(resource.PropertyMap{
					"roleARN":     resource.NewStringProperty("arn:aws:iam::123456789012:role/deploy"),
					"sessionName": resource.NewStringProperty("deploy"),
					"duration":    resource.NewStringProperty("1h"),
					"policyARNs": resource.NewArrayProperty([]resource.PropertyValue{
						resource.NewStringProperty("arn:aws:iam::aws:policy/ReadOnlyAccess"),
					}),
				}),
				"gcp": resource.NewObjectProperty(resource.PropertyMap{
					"projectId":      resource.NewStringProperty("123456789"),
					"workloadPoolId": resource.NewStringProperty("pool"),
					"providerId":     resource.NewStringProperty("oidc-provider"),
					"serviceAccount": resource.NewStringProperty("deploy@example.iam.gserviceaccount.com"),
					"tokenLifetime":  resource.NewStringProperty("3600s"),
				}),
				"azure": resource.NewObjectProperty(resource.PropertyMap{
					"clientId":       resource.NewStringProperty("client"),
					"tenantId":       resource.NewStringProperty("tenant"),
					"subscriptionId": resource.NewStringProperty("subscription"),
				}),
			}),
		}
	}
	envVars := resource.NewObjectProperty(resource.PropertyMap{
		plainKey:  resource.NewStringProperty(plainValue),
		secretKey: resource.MakeSecret(resource.NewStringProperty(secretValue)),
	})
	allSecret := map[string]pulumiapi.SecretValue{
		plainKey:  {Secret: true, Value: plainValue},
		secretKey: {Secret: true, Value: secretValue},
	}
	decode := func(t *testing.T, oc resource.PropertyValue) *pulumiapi.OperationContext {
		return decodeSettings(t, resource.PropertyMap{"operationContext": oc}).Operation
	}

	t.Run("per-variable secrets", func(t *testing.T) {
		oc := decode(t, resource.NewObjectProperty(operationContext(envVars)))
		assert.Equal(t, []string{"make"}, oc.PreRunCommands)
		assert.Equal(t, map[string]pulumiapi.SecretValue{
			plainKey:  {Value: plainValue},
			secretKey: {Secret: true, Value: secretValue},
		}, oc.EnvironmentVariables)
		// Shell is deliberately not sent; see DeploymentSettingsOperationContext.toAPI.
		assert.Equal(t, &pulumiapi.OperationContextOptions{
			SkipInstallDependencies:     true,
			SkipIntermediateDeployments: true,
			DeleteAfterDestroy:          true,
		}, oc.Options)
		assert.Equal(t, &pulumiapi.OperationContextAWSOIDCConfiguration{
			RoleARN:     "arn:aws:iam::123456789012:role/deploy",
			SessionName: "deploy",
			Duration:    "1h",
			PolicyARNs:  []string{"arn:aws:iam::aws:policy/ReadOnlyAccess"},
		}, oc.OIDC.AWS)
		assert.Equal(t, &pulumiapi.OperationContextGCPOIDCConfiguration{
			ProjectID:      "123456789",
			WorkloadPoolID: "pool",
			ProviderID:     "oidc-provider",
			ServiceAccount: "deploy@example.iam.gserviceaccount.com",
			TokenLifetime:  "3600s",
		}, oc.OIDC.GCP)
		assert.Equal(t, &pulumiapi.OperationContextAzureOIDCConfiguration{
			ClientID:       "client",
			TenantID:       "tenant",
			SubscriptionID: "subscription",
		}, oc.OIDC.Azure)
	})

	t.Run("secret operationContext makes every variable secret", func(t *testing.T) {
		oc := decode(t, resource.MakeSecret(resource.NewObjectProperty(operationContext(envVars))))
		assert.Equal(t, allSecret, oc.EnvironmentVariables)
	})

	t.Run("secret environmentVariables makes every variable secret", func(t *testing.T) {
		oc := decode(t, resource.NewObjectProperty(operationContext(resource.MakeSecret(envVars))))
		assert.Equal(t, allSecret, oc.EnvironmentVariables)
	})
}

func TestDecodeDeploymentSettingsInputWrongType(t *testing.T) {
	_, err := decodeDeploymentSettingsInput(resource.PropertyMap{
		"github": resource.NewObjectProperty(resource.PropertyMap{
			"deployCommits": resource.NewStringProperty("yes"),
		}),
	})
	assert.Error(t, err)
}
