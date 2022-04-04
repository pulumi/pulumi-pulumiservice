package provider

import (
	"fmt"
	"os"
)

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
	token := pc.getConfig("accessToken", "PULUMI_ACCESS_TOKEN")

	if len(token) == 0 {
		return nil, fmt.Errorf("no pulumi token found")
	}

	return &token, nil
}

func (pc *PulumiServiceConfig) getPulumiServiceUrl() (*string, error) {
	url := pc.getConfig("apiurl", "PULUMI_BACKEND_URL")

	baseurl := "api.pulumi.com"

	if len(url) == 0 {
		url = baseurl
	}

	return &url, nil
}
