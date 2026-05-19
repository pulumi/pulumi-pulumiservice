// Copyright 2016-2026, Pulumi Corporation.
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

package rest

import (
	"context"
	"testing"
)

// TestWithTransport_BeatsGlobalResolver: a ctx-scoped transport wins over the
// package-global resolver, so future test parallelism + multi-provider don't
// race on the global.
func TestWithTransport_BeatsGlobalResolver(t *testing.T) {
	globalCalled := false
	ctxOnly := &mockTransport{}
	SetTransportResolver(func(_ context.Context) (Transport, error) {
		globalCalled = true
		return &mockTransport{}, nil
	})
	defer SetTransportResolver(nil)

	ctx := WithTransport(context.Background(), ctxOnly)
	got, err := resolveTransport(ctx)
	if err != nil {
		t.Fatalf("resolveTransport: %v", err)
	}
	if got != ctxOnly {
		t.Errorf("resolveTransport returned %p; want ctx-scoped transport %p", got, ctxOnly)
	}
	if globalCalled {
		t.Errorf("global resolver must not run when ctx has its own transport")
	}
}

// TestResolveTransport_FallsBackToGlobal: with no ctx-scoped transport, the
// package-global resolver still runs (transitional compatibility).
func TestResolveTransport_FallsBackToGlobal(t *testing.T) {
	wanted := &mockTransport{}
	SetTransportResolver(func(_ context.Context) (Transport, error) {
		return wanted, nil
	})
	defer SetTransportResolver(nil)

	got, err := resolveTransport(context.Background())
	if err != nil {
		t.Fatalf("resolveTransport: %v", err)
	}
	if got != wanted {
		t.Errorf("global fallback returned %p; want %p", got, wanted)
	}
}

// TestResolveTransport_NoTransportReturnsError: neither path set → clear error.
func TestResolveTransport_NoTransportReturnsError(t *testing.T) {
	SetTransportResolver(nil)
	_, err := resolveTransport(context.Background())
	if err == nil {
		t.Fatalf("expected error with no transport configured")
	}
}
