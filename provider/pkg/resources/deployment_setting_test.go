package resources

import (
	"context"
	"maps"
	"testing"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi-go-provider/integration"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/property"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

type getDeploymentSettingsFunc func() (*pulumiapi.DeploymentSettings, error)

type DeploymentSettingsClientMock struct {
	config.Client
	getDeploymentSettingsFunc getDeploymentSettingsFunc
	created, updated          []pulumiapi.DeploymentSettings
}

func (c *DeploymentSettingsClientMock) CreateDeploymentSettings(
	_ context.Context,
	_ pulumiapi.StackIdentifier,
	settings pulumiapi.DeploymentSettings,
) (*pulumiapi.DeploymentSettings, error) {
	c.created = append(c.created, settings)
	return &settings, nil
}

func (c *DeploymentSettingsClientMock) UpdateDeploymentSettings(
	_ context.Context,
	_ pulumiapi.StackIdentifier,
	settings pulumiapi.DeploymentSettings,
) (*pulumiapi.DeploymentSettings, error) {
	c.updated = append(c.updated, settings)
	return &settings, nil
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
		getDeploymentSettingsFunc: getDeploymentSettingsFunc,
	}
}

var deploymentSettingsURN = resource.NewURN("test", "proj", "", DeploymentSettingsToken, "settings")

func newDeploymentSettingsServer(t *testing.T, client *DeploymentSettingsClientMock) integration.Server {
	t.Helper()
	prov, err := infer.NewProviderBuilder().
		WithResources(infer.Resource(&DeploymentSettings{})).
		Build()
	require.NoError(t, err)
	server, err := integration.NewServer(config.WithMockClient(t.Context(), client),
		"pulumiservice", semver.MustParse("0.0.0"), integration.WithProvider(prov))
	require.NoError(t, err)
	return server
}

// checkDeploymentSettings runs Check over props and returns the checked inputs.
func checkDeploymentSettings(t *testing.T, props resource.PropertyMap) resource.PropertyMap {
	t.Helper()
	server := newDeploymentSettingsServer(t, &DeploymentSettingsClientMock{})
	resp, err := server.Check(p.CheckRequest{
		Urn:    deploymentSettingsURN,
		Inputs: resource.FromResourcePropertyValue(resource.NewObjectProperty(props)).AsMap(),
	})
	require.NoError(t, err)
	require.Empty(t, resp.Failures)
	return resourceMap(resp.Inputs)
}

func TestDeploymentSettings(t *testing.T) {
	t.Run("Read when the resource is not found", func(t *testing.T) {
		server := newDeploymentSettingsServer(t, buildDeploymentSettingsClientMock(
			func() (*pulumiapi.DeploymentSettings, error) { return nil, nil },
		))

		resp, err := server.Read(p.ReadRequest{ID: "abc/def/123", Urn: deploymentSettingsURN})

		require.NoError(t, err)
		assert.Empty(t, resp.ID)
	})

	t.Run("Read when the resource is found", func(t *testing.T) {
		server := newDeploymentSettingsServer(t, buildDeploymentSettingsClientMock(
			func() (*pulumiapi.DeploymentSettings, error) {
				return &pulumiapi.DeploymentSettings{
					Operation:     &pulumiapi.OperationContext{},
					GitHub:        &pulumiapi.DeploymentSettingsGitHub{},
					SourceContext: &pulumiapi.SourceContext{},
					Executor:      &pulumiapi.ExecutorContext{},
				}, nil
			},
		))

		resp, err := server.Read(p.ReadRequest{ID: "abc/def/123", Urn: deploymentSettingsURN})

		require.NoError(t, err)
		assert.Equal(t, "abc/def/123", resp.ID)
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

	checked := checkDeploymentSettings(t, propertyMap)

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
	require.NoError(t, err)
	return decoded
}

func secretOrString(prop resource.PropertyValue) string {
	if prop.IsSecret() {
		return prop.SecretValue().Element.StringValue()
	}
	if prop.IsNull() {
		return ""
	}
	return prop.StringValue()
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

		checked := checkDeploymentSettings(t, propertyMap)

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
	assert.Equal(t, testRedactedSecret, secretOrString(outputs[gcUsername]),
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

	checked := checkDeploymentSettings(t, propertyMap)
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
	assert.Equal(t, testRegistryPassword, secretOrString(inputs["password"]),
		"refresh must preserve the plaintext password from prior inputs")

	outputs := credentialsPropertyMap(t, input.ToPropertyMap(&plaintextInputs, &priorState, false))
	assert.Equal(t, testRedactedSecret, secretOrString(outputs["password"]),
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
	assert.Equal(t, "", secretOrString(inputs["password"]))
}

// Import has no inputs at all, so the password is replaced with the placeholder
// that prompts the user to fill it in.
func TestDeploymentSettingsExecutorImageCredentialsImport(t *testing.T) {
	settings := executorImageCredentialsSettings(redactedPassword())
	input := PulumiServiceDeploymentSettingsInput{DeploymentSettings: settings}

	inputs := credentialsPropertyMap(t, input.ToPropertyMap(nil, nil, true))
	assert.Equal(t, "<REPLACE WITH ACTUAL SECRET VALUE>", secretOrString(inputs["password"]))
	assert.Equal(t, "registry-user", secretOrString(inputs["username"]))
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
		return decodeSettings(t, resource.PropertyMap{gcOperationContext: oc}).Operation
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

const (
	gcPublicVar  = "PUBLIC_VAR"
	gcSecretVar  = "SECRET_VAR"
	gcSettingsID = "an-org/a-project/a-stack"
)

func deploymentSettingsInputs(extra map[string]property.Value) property.Map {
	m := map[string]property.Value{
		gcOrganization: property.New("an-org"),
		gcProject:      property.New("a-project"),
		gcStack:        property.New("a-stack"),
	}
	maps.Copy(m, extra)
	return property.NewMap(m)
}

func environmentVariables(vars map[string]property.Value) map[string]property.Value {
	return map[string]property.Value{
		gcOperationContext: property.New(map[string]property.Value{
			gcEnvironmentVariables: property.New(vars),
		}),
	}
}

func TestDeploymentSettingsCreateSendsEnvironmentVariableSecrecy(t *testing.T) {
	client := &DeploymentSettingsClientMock{}
	server := newDeploymentSettingsServer(t, client)

	resp, err := server.Create(p.CreateRequest{
		Urn: deploymentSettingsURN,
		Properties: deploymentSettingsInputs(environmentVariables(map[string]property.Value{
			gcPublicVar: property.New("visible"),
			gcSecretVar: property.New("hidden").WithSecret(true),
		})),
	})
	require.NoError(t, err)
	assert.Equal(t, gcSettingsID, resp.ID)

	require.Len(t, client.created, 1)
	assert.Equal(t, map[string]pulumiapi.SecretValue{
		gcPublicVar: {Value: "visible"},
		gcSecretVar: {Value: "hidden", Secret: true},
	}, client.created[0].Operation.EnvironmentVariables)

	vars := resp.Properties.Get(gcOperationContext).AsMap().Get(gcEnvironmentVariables).AsMap()
	assert.False(t, vars.Get(gcPublicVar).Secret(), "PLAIN should stay public in the state")
	assert.True(t, vars.Get(gcSecretVar).Secret(), "SECRET should stay secret in the state")
}

func TestDeploymentSettingsCreateSecretOperationContextMakesAllVariablesSecret(t *testing.T) {
	client := &DeploymentSettingsClientMock{}
	server := newDeploymentSettingsServer(t, client)

	_, err := server.Create(p.CreateRequest{
		Urn: deploymentSettingsURN,
		Properties: deploymentSettingsInputs(map[string]property.Value{
			gcOperationContext: property.New(map[string]property.Value{
				gcEnvironmentVariables: property.New(map[string]property.Value{
					"A": property.New("a"),
				}),
			}).WithSecret(true),
		}),
	})
	require.NoError(t, err)

	require.Len(t, client.created, 1)
	assert.True(t, client.created[0].Operation.EnvironmentVariables["A"].Secret)
}

func TestDeploymentSettingsPreviewDoesNotCallPulumiCloud(t *testing.T) {
	client := &DeploymentSettingsClientMock{}
	server := newDeploymentSettingsServer(t, client)
	inputs := deploymentSettingsInputs(nil)

	_, err := server.Create(p.CreateRequest{Urn: deploymentSettingsURN, Properties: inputs, DryRun: true})
	require.NoError(t, err)
	_, err = server.Update(p.UpdateRequest{
		ID: gcSettingsID, Urn: deploymentSettingsURN, State: inputs, Inputs: inputs, DryRun: true,
	})
	require.NoError(t, err)

	assert.Empty(t, client.created)
	assert.Empty(t, client.updated)
}

func TestDeploymentSettingsCheckDoesNotFillTrueDefaults(t *testing.T) {
	for _, block := range []string{"github", gcVcs} {
		t.Run(block+" flags left unset", func(t *testing.T) {
			settings := map[string]property.Value{"repository": property.New("org/repo")}
			if block == gcVcs {
				settings["provider"] = property.New("github")
			}
			checked := checkDeploymentSettings(t, resourceMap(deploymentSettingsInputs(map[string]property.Value{
				block: property.New(settings),
			})))

			got := checked[resource.PropertyKey(block)].ObjectValue()
			assert.NotContains(t, got, resource.PropertyKey("deployCommits"))
			assert.NotContains(t, got, resource.PropertyKey("previewPullRequests"))
		})

		t.Run(block+" flags set explicitly", func(t *testing.T) {
			settings := map[string]property.Value{
				"deployCommits":       property.New(true),
				"previewPullRequests": property.New(false),
			}
			if block == gcVcs {
				settings["provider"] = property.New("github")
			}
			checked := checkDeploymentSettings(t, resourceMap(deploymentSettingsInputs(map[string]property.Value{
				block: property.New(settings),
			})))

			got := checked[resource.PropertyKey(block)].ObjectValue()
			assert.Equal(t, resource.NewBoolProperty(true), got["deployCommits"])
			assert.Equal(t, resource.NewBoolProperty(false), got["previewPullRequests"])
		})
	}
}

func TestDeploymentSettingsCheckNormalizesAWSDuration(t *testing.T) {
	checked := checkDeploymentSettings(t, resourceMap(deploymentSettingsInputs(map[string]property.Value{
		gcOperationContext: property.New(map[string]property.Value{
			"oidc": property.New(map[string]property.Value{
				gcAWS: property.New(map[string]property.Value{
					"roleARN":     property.New("arn:aws:iam::123456789012:role/r"),
					"sessionName": property.New("s"),
					"duration":    property.New("90m"),
				}),
			}),
		}),
	})))

	aws := checked[gcOperationContext].ObjectValue()["oidc"].ObjectValue()[gcAWS].ObjectValue()
	assert.Equal(t, "1h30m0s", aws["duration"].StringValue())
}

func TestDeploymentSettingsDiff(t *testing.T) {
	server := newDeploymentSettingsServer(t, &DeploymentSettingsClientMock{})
	diff := func(t *testing.T, olds, news property.Map) p.DiffResponse {
		t.Helper()
		resp, err := server.Diff(p.DiffRequest{
			ID: gcSettingsID, Urn: deploymentSettingsURN, State: olds, OldInputs: olds, Inputs: news,
		})
		require.NoError(t, err)
		return resp
	}
	withVar := func(v property.Value) property.Map {
		return deploymentSettingsInputs(environmentVariables(map[string]property.Value{"A": v}))
	}

	t.Run("no changes", func(t *testing.T) {
		assert.False(t, diff(t, withVar(property.New("a")), withVar(property.New("a"))).HasChanges)
	})

	t.Run("making a value secret is not a change", func(t *testing.T) {
		assert.False(t, diff(t, withVar(property.New("a")), withVar(property.New("a").WithSecret(true))).HasChanges)
	})

	t.Run("changed variable updates in place", func(t *testing.T) {
		resp := diff(t, withVar(property.New("a")), withVar(property.New("b")))
		assert.True(t, resp.HasChanges)
		assert.Equal(t, p.Update, resp.DetailedDiff["operationContext.environmentVariables.A"].Kind)
	})

	t.Run("changed stack replaces, deleting first", func(t *testing.T) {
		news := deploymentSettingsInputs(map[string]property.Value{gcStack: property.New("other-stack")})
		resp := diff(t, deploymentSettingsInputs(nil), news)
		assert.True(t, resp.HasChanges)
		assert.True(t, resp.DeleteBeforeReplace)
		assert.Equal(t, p.UpdateReplace, resp.DetailedDiff[gcStack].Kind)
	})
}

func TestDeploymentSettingsRefreshKeepsSecretVariablesSecret(t *testing.T) {
	server := newDeploymentSettingsServer(t, buildDeploymentSettingsClientMock(
		func() (*pulumiapi.DeploymentSettings, error) {
			return &pulumiapi.DeploymentSettings{Operation: &pulumiapi.OperationContext{
				EnvironmentVariables: map[string]pulumiapi.SecretValue{
					gcPublicVar: {Value: "visible"},
					gcSecretVar: {Value: "ciphertext", Secret: true},
				},
			}}, nil
		},
	))
	inputs := deploymentSettingsInputs(environmentVariables(map[string]property.Value{
		gcPublicVar: property.New("visible"),
		gcSecretVar: property.New("hidden").WithSecret(true),
	}))
	state := deploymentSettingsInputs(environmentVariables(map[string]property.Value{
		gcPublicVar: property.New("visible"),
		gcSecretVar: property.New("ciphertext").WithSecret(true),
	}))

	resp, err := server.Read(p.ReadRequest{
		ID: gcSettingsID, Urn: deploymentSettingsURN, Properties: state, Inputs: inputs,
	})
	require.NoError(t, err)

	vars := resp.Inputs.Get(gcOperationContext).AsMap().Get(gcEnvironmentVariables).AsMap()
	assert.Equal(t, property.New("visible"), vars.Get(gcPublicVar))
	assert.Equal(t, property.New("hidden").WithSecret(true), vars.Get(gcSecretVar))
}
