package provider

import (
	"os"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/stretchr/testify/assert"
)

const (
	EnvVarPulumiHome = "PULUMI_HOME"
)

func TestGetPulumiAccessToken(t *testing.T) {
	wantToken := "pul-1234abcd"

	// setEnv sets environment variable and returns a function to reset
	// environment variable back to previous value
	setEnv := func(envVar, val string) func() {
		oldVal := os.Getenv(envVar)
		os.Setenv(envVar, val)
		return func() {
			os.Setenv(envVar, oldVal)
		}
	}

	t.Run("Uses Config Variable", func(t *testing.T) {
		defer setEnv(EnvVarPulumiAccessToken, "")()
		c := PulumiServiceConfig{
			Config: map[string]string{
				"accessToken": wantToken,
			},
		}
		gotToken, err := c.getPulumiAccessToken()
		assert.NoError(t, err)
		assert.Equal(t, wantToken, *gotToken)

	})

	t.Run("Uses Environment Variable", func(t *testing.T) {

		c := PulumiServiceConfig{}
		defer setEnv(EnvVarPulumiAccessToken, wantToken)()

		gotToken, err := c.getPulumiAccessToken()
		assert.NoError(t, err)
		assert.Equal(t, wantToken, *gotToken)
	})

	t.Run("Uses Saved Credential", func(t *testing.T) {
		c := PulumiServiceConfig{}
		// ensure env var isn't set
		defer setEnv(EnvVarPulumiAccessToken, "")()

		dir := t.TempDir()
		// set home directory so that workspace writes to temp dir
		defer setEnv(EnvVarPulumiHome, dir)()

		account := "https://api.pulumi.com"

		err := workspace.StoreCredentials(workspace.Credentials{
			Current: account,
			AccessTokens: map[string]string{
				account: wantToken,
			},
		})
		if err != nil {
			t.Fatalf("failed to store test credentials: %v", err)
		}
		gotToken, err := c.getPulumiAccessToken()
		assert.NoError(t, err)
		assert.Equal(t, wantToken, *gotToken)
	})

	t.Run("Returns Error", func(t *testing.T) {
		c := PulumiServiceConfig{}

		// point PULUMI_HOME to empty dir so there's no credentials available
		dir := t.TempDir()
		defer setEnv(EnvVarPulumiHome, dir)()
		// explicitly unset access token variable
		defer setEnv(EnvVarPulumiAccessToken, "")()

		gotToken, err := c.getPulumiAccessToken()

		assert.Nil(t, gotToken)
		assert.Equal(t, ErrAccessTokenNotFound, err)
	})
}
