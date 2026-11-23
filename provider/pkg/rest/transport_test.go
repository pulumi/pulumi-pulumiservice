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
	"testing"
)

func TestResolveTransport_ReturnsContextTransport(t *testing.T) {
	t.Parallel()
	want := &mockTransport{}
	got, err := resolveTransport(WithTransport(t.Context(), want))
	if err != nil {
		t.Fatalf("resolveTransport: %v", err)
	}
	if got != want {
		t.Errorf("resolveTransport returned %p; want %p", got, want)
	}
}

func TestResolveTransport_NoTransportReturnsError(t *testing.T) {
	t.Parallel()
	if _, err := resolveTransport(t.Context()); err == nil {
		t.Fatalf("expected error with no transport in context")
	}
}
