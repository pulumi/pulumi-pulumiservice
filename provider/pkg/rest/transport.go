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

var (
	resolverMu sync.RWMutex
	resolver   TransportResolver
)

// SetTransportResolver registers the per-request Transport resolver,
// replacing any previously registered one.
func SetTransportResolver(r TransportResolver) {
	resolverMu.Lock()
	defer resolverMu.Unlock()
	resolver = r
}

func resolveTransport(ctx context.Context) (Transport, error) {
	resolverMu.RLock()
	r := resolver
	resolverMu.RUnlock()
	if r == nil {
		return nil, errors.New("rest: no transport resolver registered (call rest.SetTransportResolver during Configure)")
	}
	return r(ctx)
}
