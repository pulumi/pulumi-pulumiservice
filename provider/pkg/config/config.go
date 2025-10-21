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

package config

import (
	"context"
	"fmt"
	"net/http"
	"time"

	esc_client "github.com/pulumi/esc/cmd/esc/cli/client"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

const (
	EnvVarPulumiAccessToken = "PULUMI_ACCESS_TOKEN"
	EnvVarPulumiBackendUrl  = "PULUMI_BACKEND_URL"
)

var ErrAccessTokenNotFound = fmt.Errorf("pulumi access token not found")

type testClientKey struct{}

// TestClientKey is the key used during client lookups to override the normal key.
var TestClientKey = testClientKey{}

// GetClient gets the client for a particular service, or a test client if it is
// available.
func GetClient[T any](ctx context.Context) T {
	// Check test client first
	if v := ctx.Value(TestClientKey); v != nil {
		return v.(T)
	}

	cfg := GetConfig(ctx)

	// Try to return the ESC client first if it matches the requested type
	if escClient, ok := any(cfg.escClient).(T); ok {
		return escClient
	}

	// Fall back to the Pulumi API client
	return any(cfg.client).(T)
}

// GetConfig accesses the config associated with the current request.
func GetConfig(ctx context.Context) Config { return *infer.GetConfig[*Config](ctx) }

// Config defines the provider configuration for pulumi-pulumiservice.
// It implements infer.Annotated and infer.CustomConfigure to integrate with
// the pulumi-go-provider infer framework.
type Config struct {
	AccessToken string `pulumi:"accessToken,optional" provider:"secret"`
	ServiceURL  string `pulumi:"serviceURL,optional"`
	client      *pulumiapi.Client
	escClient   esc_client.Client
}

var (
	_ infer.Annotated       = (*Config)(nil)
	_ infer.CustomConfigure = (*Config)(nil)
)

func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c.AccessToken, "Access Token to authenticate with Pulumi Cloud.")
	a.SetDefault(&c.AccessToken, nil, EnvVarPulumiAccessToken)

	a.Describe(&c.ServiceURL, "The service URL used to reach Pulumi Cloud.")
	a.SetDefault(&c.ServiceURL, "https://api.pulumi.com", EnvVarPulumiBackendUrl)
}

func (c *Config) Configure(context.Context) error {
	// Ensure that we have an access token
	if len(c.AccessToken) == 0 {
		creds, err := workspace.GetStoredCredentials()
		if err != nil {
			return ErrAccessTokenNotFound
		}
		if token, ok := creds.AccessTokens[creds.Current]; ok {
			c.AccessToken = token
		} else {
			return ErrAccessTokenNotFound
		}
	}

	// Construct the PulumiService client
	client, err := pulumiapi.NewClient(&http.Client{
		Timeout: 60 * time.Second,
	}, c.AccessToken, c.ServiceURL)

	c.client = client

	// Initialize ESC client for Environment resources
	escAPIUrl := c.ServiceURL
	if escAPIUrl == "" {
		escAPIUrl = "https://api.pulumi.com"
	}
	c.escClient = esc_client.New("pulumi-pulumiservice/infer", escAPIUrl, c.AccessToken, false /* insecure */)

	return err
}
