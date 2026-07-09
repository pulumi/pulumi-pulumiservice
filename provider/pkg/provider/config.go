package provider

import (
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

const (
	EnvVarPulumiAccessToken = "PULUMI_ACCESS_TOKEN"
	EnvVarPulumiBackendURL  = "PULUMI_BACKEND_URL"
	// EnvVarPulumiAPI is an internal alias the Pulumi CLI sets on certain login
	// flows (e.g. `pl login devstack`). We honor it as a fallback so a token
	// scoped to a non-prod backend doesn't silently route to api.pulumi.com.
	EnvVarPulumiAPI = "PULUMI_API"
	accessTokenKey  = "accessToken"
	apiURLKey       = "apiUrl"
)

var ErrAccessTokenNotFound = fmt.Errorf("pulumi access token not found")

type PulumiServiceConfig struct {
	Config map[string]string
}

func (pc *PulumiServiceConfig) getConfig(configName, envName string) string {
	if val := pc.Config[configName]; val != "" {
		return val
	}

	return os.Getenv(envName)
}

func (pc *PulumiServiceConfig) getPulumiAccessToken() (*string, error) {
	token := pc.getConfig(accessTokenKey, EnvVarPulumiAccessToken)

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
	url := pc.getConfig(apiURLKey, EnvVarPulumiBackendURL)
	if url == "" {
		url = os.Getenv(EnvVarPulumiAPI)
	}
	if url == "" {
		url = "https://api.pulumi.com"
	}
	return &url, nil
}
