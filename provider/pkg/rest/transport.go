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
	"errors"
	"net/http"
	"sync"
)

// Transport executes HTTP requests. Implementations own scheme/host
// rewriting, authentication, timeouts, and retries.
type Transport interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

// TransportResolver returns the Transport to use for a given context.
type TransportResolver func(ctx context.Context) (Transport, error)

type transportCtxKey struct{}

// WithTransport returns a copy of ctx that carries t. resolveTransport reads
// this before falling back to the package-global registered via
// SetTransportResolver. Use this in tests (each goroutine has its own
// transport so t.Parallel is safe) and in any production code that needs
// multiple provider instances coexisting in one process.
func WithTransport(ctx context.Context, t Transport) context.Context {
	return context.WithValue(ctx, transportCtxKey{}, t)
}

var (
	resolverMu sync.RWMutex
	resolver   TransportResolver
)

// SetTransportResolver registers a process-global per-request resolver,
// replacing any previously registered one. Transitional — prefer
// [WithTransport] for new code; the global blocks multi-provider isolation
// and parallel testing.
func SetTransportResolver(r TransportResolver) {
	resolverMu.Lock()
	defer resolverMu.Unlock()
	resolver = r
}

// resolveTransport returns ctx's [WithTransport] value when set, otherwise
// falls back to the package-global from [SetTransportResolver].
func resolveTransport(ctx context.Context) (Transport, error) {
	if t, ok := ctx.Value(transportCtxKey{}).(Transport); ok {
		return t, nil
	}
	resolverMu.RLock()
	r := resolver
	resolverMu.RUnlock()
	if r == nil {
		return nil, errors.New("rest: no transport in context and no global resolver " +
			"(set via rest.WithTransport or rest.SetTransportResolver during Configure)")
	}
	return r(ctx)
}
