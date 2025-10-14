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

	pbempty "google.golang.org/protobuf/types/known/emptypb"

	inferResources "github.com/pulumi/pulumi-pulumiservice/provider/pkg/infer"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
)

// hybridManualProvider wraps pulumiserviceProvider and injects API clients into context
// for infer resources to access.
type hybridManualProvider struct {
	*pulumiserviceProvider
}

// injectClientContext adds the API clients to the context so infer resources can access them.
func (h *hybridManualProvider) injectClientContext(ctx context.Context) context.Context {
	if h.client != nil {
		ctx = inferResources.WithClient(ctx, h.client)
	}
	// TODO: Add ESC client injection when migrating environment resources
	return ctx
}

// Override methods to inject client context

func (h *hybridManualProvider) Check(ctx context.Context, req *pulumirpc.CheckRequest) (*pulumirpc.CheckResponse, error) {
	return h.pulumiserviceProvider.Check(h.injectClientContext(ctx), req)
}

func (h *hybridManualProvider) Diff(ctx context.Context, req *pulumirpc.DiffRequest) (*pulumirpc.DiffResponse, error) {
	return h.pulumiserviceProvider.Diff(h.injectClientContext(ctx), req)
}

func (h *hybridManualProvider) Create(ctx context.Context, req *pulumirpc.CreateRequest) (*pulumirpc.CreateResponse, error) {
	return h.pulumiserviceProvider.Create(h.injectClientContext(ctx), req)
}

func (h *hybridManualProvider) Read(ctx context.Context, req *pulumirpc.ReadRequest) (*pulumirpc.ReadResponse, error) {
	return h.pulumiserviceProvider.Read(h.injectClientContext(ctx), req)
}

func (h *hybridManualProvider) Update(ctx context.Context, req *pulumirpc.UpdateRequest) (*pulumirpc.UpdateResponse, error) {
	return h.pulumiserviceProvider.Update(h.injectClientContext(ctx), req)
}

func (h *hybridManualProvider) Delete(ctx context.Context, req *pulumirpc.DeleteRequest) (*pbempty.Empty, error) {
	return h.pulumiserviceProvider.Delete(h.injectClientContext(ctx), req)
}

func (h *hybridManualProvider) Invoke(ctx context.Context, req *pulumirpc.InvokeRequest) (*pulumirpc.InvokeResponse, error) {
	return h.pulumiserviceProvider.Invoke(h.injectClientContext(ctx), req)
}
