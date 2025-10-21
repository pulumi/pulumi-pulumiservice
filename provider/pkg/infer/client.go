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

// Package infer provides utilities for resources using the pulumi-go-provider infer framework.
//
// This package manages client injection via context, allowing infer-based resources to access
// the Pulumi Service API client and ESC client that are configured during provider initialization.
package infer

import (
	"context"

	esc_client "github.com/pulumi/esc/cmd/esc/cli/client"
	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/pulumiapi"
)

// clientContextKey is a context key for storing the Pulumi Service API client.
//
// We use an unexported empty struct type instead of a string to prevent key collisions
// with other packages that might use the same string key name. This is the recommended
// pattern for context keys in Go. See https://pkg.go.dev/context#WithValue for details.
type clientContextKey struct{}

// escClientContextKey is a context key for storing the ESC (Environments, Secrets, Config) client.
//
// We use an unexported empty struct type instead of a string to prevent key collisions.
// This ensures that only this package can access values stored with this key.
type escClientContextKey struct{}

// WithClient stores the Pulumi Service API client in the context.
func WithClient(ctx context.Context, client *pulumiapi.Client) context.Context {
	return context.WithValue(ctx, clientContextKey{}, client)
}

// GetClient retrieves the Pulumi Service API client from the context.
// Returns nil if the client is not present in the context.
func GetClient(ctx context.Context) *pulumiapi.Client {
	if client, ok := ctx.Value(clientContextKey{}).(*pulumiapi.Client); ok {
		return client
	}
	return nil
}

// WithESCClient stores the ESC client in the context.
func WithESCClient(ctx context.Context, escClient esc_client.Client) context.Context {
	return context.WithValue(ctx, escClientContextKey{}, escClient)
}

// GetESCClient retrieves the ESC client from the context.
// Returns nil if the client is not present in the context.
func GetESCClient(ctx context.Context) esc_client.Client {
	if client, ok := ctx.Value(escClientContextKey{}).(esc_client.Client); ok {
		return client
	}
	return nil
}
