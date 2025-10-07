// Package provider implements the Pulumi Service resource provider.
package provider

import (
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

const (
	// EnvVarPulumiAccessToken is the environment variable name for the Pulumi access token.
	EnvVarPulumiAccessToken = "PULUMI_ACCESS_TOKEN"
	// EnvVarPulumiBackendURL is the environment variable name for the Pulumi backend URL.
	EnvVarPulumiBackendURL = "PULUMI_BACKEND_URL"
)

// ErrAccessTokenNotFound is returned when a Pulumi access token cannot be found.
var ErrAccessTokenNotFound = fmt.Errorf("pulumi access token not found")

// PulumiServiceConfig holds configuration for the Pulumi Service provider.
type PulumiServiceConfig struct {
	Config map[string]string
}

func (pc *PulumiServiceConfig) getConfig(configName, envName string) string {
	if val, ok := pc.Config[configName]; ok {
		return val
	}

	return os.Getenv(envName)
}

func (pc *PulumiServiceConfig) getPulumiAccessToken() (*string, error) {
	token := pc.getConfig("accessToken", EnvVarPulumiAccessToken)

	if len(token) > 0 {
		// found the token
		return &token, nil
	}

	// attempt to grab credentials directly from the pulumi configuration on the machine
	creds, err := workspace.GetStoredCredentials()
	if err != nil {
		return nil, ErrAccessTokenNotFound
	}
	if token, ok := creds.AccessTokens[creds.Current]; ok {
		return &token, nil
	}
	return nil, ErrAccessTokenNotFound
}

func (pc *PulumiServiceConfig) getPulumiServiceURL() (*string, error) {
	url := pc.getConfig("apiUrl", EnvVarPulumiBackendURL)
	baseurl := "https://api.pulumi.com"

	if len(url) == 0 {
		url = baseurl
	}

	return &url, nil
}
