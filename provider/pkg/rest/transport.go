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
)

// Transport executes HTTP requests. Implementations own scheme/host
// rewriting, authentication, timeouts, and retries.
type Transport interface {
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

type transportCtxKey struct{}

// WithTransport returns a copy of ctx that carries t. Requests made with the
// returned context use t, so each provider instance (or test) can have its own
// transport without process-global state.
func WithTransport(ctx context.Context, t Transport) context.Context {
	return context.WithValue(ctx, transportCtxKey{}, t)
}

// resolveTransport returns the Transport that [WithTransport] put in ctx.
func resolveTransport(ctx context.Context) (Transport, error) {
	if t, ok := ctx.Value(transportCtxKey{}).(Transport); ok {
		return t, nil
	}
	return nil, errors.New("rest: no transport in context (set via rest.WithTransport during Configure)")
}
