// Copyright 2016-2025, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"context"
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/config"
)

// Re-export config types and functions for backward compatibility
// and to maintain clean imports within the provider package.

type Config = config.Config

// GetClient gets the client for a particular service, or a test client if it is available.
func GetClient[T any](ctx context.Context) T {
	return config.GetClient[T](ctx)
}

// GetConfig accesses the config associated with the current request.
func GetConfig(ctx context.Context) Config {
	return config.GetConfig(ctx)
}

var (
	TestClientKey          = config.TestClientKey
	ErrAccessTokenNotFound = config.ErrAccessTokenNotFound
)

const (
	EnvVarPulumiAccessToken = config.EnvVarPulumiAccessToken
	EnvVarPulumiBackendUrl  = config.EnvVarPulumiBackendUrl
)

// PulumiServiceConfig is a legacy configuration type used by the manual provider.
// This will be removed once all resources are migrated to the infer framework.
//
// Deprecated: Use config.Config instead.
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

func (pc *PulumiServiceConfig) getPulumiServiceUrl() (*string, error) {
	url := pc.getConfig("apiUrl", EnvVarPulumiBackendUrl)
	baseurl := "https://api.pulumi.com"

	if len(url) == 0 {
		url = baseurl
	}

	return &url, nil
}
