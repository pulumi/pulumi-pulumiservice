package provider

import (
	"fmt"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

const (
	EnvVarPulumiAccessToken = "PULUMI_ACCESS_TOKEN"
	EnvVarPulumiBackendUrl  = "PULUMI_BACKEND_URL"
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

func (pc *PulumiServiceConfig) getPulumiServiceUrl() (*string, error) {
	url := pc.getConfig("apiUrl", EnvVarPulumiBackendUrl)
	baseurl := "https://api.pulumi.com"

	if len(url) == 0 {
		url = baseurl
	}

	return &url, nil
}
