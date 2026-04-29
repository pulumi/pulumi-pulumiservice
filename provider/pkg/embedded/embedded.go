// Copyright 2016-2026, Pulumi Corporation.
//
// Package embedded owns the canonical inputs to the v2 provider:
// the pinned Pulumi Cloud OpenAPI spec and the resource map. These
// are //go:embed-ed directly into the runtime binary so that
// `go build` produces a fully working provider without a separate
// generator step. `update_spec` (in .mk/v2.mk) writes new bytes
// into openapi_public.json here; the resource-map.yaml is the
// repo's editable source of truth and is hand-edited in place.

package embedded

import _ "embed"

//go:embed openapi_public.json
var spec []byte

//go:embed resource-map.yaml
var resourceMap []byte

// Spec returns the embedded Pulumi Cloud OpenAPI 3.0.3 spec bytes.
// The slice is shared across calls; callers must not mutate it.
func Spec() []byte { return spec }

// ResourceMap returns the embedded resource-map.yaml bytes that
// classify each Pulumi Cloud operation as a Pulumi resource,
// function, method, or deliberate exclusion.
// The slice is shared across calls; callers must not mutate it.
func ResourceMap() []byte { return resourceMap }
