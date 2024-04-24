package provider

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	"github.com/stretchr/testify/assert"
)

const (
	EnvVarPulumiHome = "PULUMI_HOME"
)

func WithClient[T any](client T) context.Context {
	return context.WithValue(context.Background(), TestClientKey, client)
}

func TestGetPulumiAccessToken(t *testing.T) {
	wantToken := "pul-1234abcd"

	t.Run("Uses Saved Credential", func(t *testing.T) {
		c := Config{}
		// ensure env var isn't set
		t.Setenv(EnvVarPulumiAccessToken, "")

		dir := t.TempDir()
		// set home directory so that workspace writes to temp dir
		t.Setenv(EnvVarPulumiHome, dir)

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
		err = c.Configure(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, wantToken, c.AccessToken)
	})

	t.Run("Returns Error", func(t *testing.T) {
		c := Config{}

		// point PULUMI_HOME to empty dir so there's no credentials available
		dir := t.TempDir()
		t.Setenv(EnvVarPulumiHome, dir)
		// explicitly unset access token variable
		t.Setenv(EnvVarPulumiAccessToken, "")
		err := c.Configure(context.Background())

		assert.Empty(t, c.AccessToken)
		assert.Equal(t, ErrAccessTokenNotFound, err)
	})
}
