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
	"encoding/json"
	"fmt"
)

// Metadata is the parsed contents of metadata.json: the Pulumi-only overrides
// (operationId → resource mapping, forceNew, secret, aliases, etc.) layered
// over the OpenAPI spec.
type Metadata struct {
	// Package is the default package name when a token lacks a prefix.
	Package string `json:"package,omitempty"`

	// Resources keyed by fully-qualified Pulumi token ("pkg:module:Name").
	Resources map[string]ResourceMeta `json:"resources"`

	// Functions keyed by fully-qualified Pulumi token. Reserved for future use.
	Functions map[string]ResourceMeta `json:"functions,omitempty"`
}

// ResourceMeta describes one Pulumi resource derived from OpenAPI operations.
type ResourceMeta struct {
	// Operations names the operationIds for each CRUD verb. Create is
	// required; Update and Delete may be empty.
	Operations Operations `json:"operations"`

	// Token overrides the user-facing Pulumi token. Keeps the map key as
	// the canonical OpenAPI-derived identifier for scaffolding while
	// exposing a clean module path (e.g. "pulumiservice:v1/esc:Environment").
	Token string `json:"token,omitempty"`

	// Aliases are tokens the engine treats as equivalent (for renames).
	Aliases []string `json:"aliases,omitempty"`

	// Fields holds Pulumi-only per-field overrides (Pulumi-side keys).
	Fields map[string]FieldMeta `json:"fields,omitempty"`

	// Renames maps Pulumi-side field names to OpenAPI-side names.
	Renames map[string]string `json:"renames,omitempty"`

	// Outputs is an allowlist of response fields exposed as State.
	// Empty means expose all.
	Outputs []string `json:"outputs,omitempty"`

	// OutputsExclude is a denylist. Outputs wins if both are set.
	OutputsExclude []string `json:"outputsExclude,omitempty"`

	// IDFormat is the resource-ID template ("{org}/{name}"). When unset,
	// path-parameter values from the most authoritative non-create op are
	// slash-joined.
	IDFormat string `json:"idFormat,omitempty"`

	// DeleteBeforeReplace destroys the old instance before creating the
	// new on a replacement. Use for resources whose names collide on
	// duplicate create and that aren't auto-named.
	DeleteBeforeReplace bool `json:"deleteBeforeReplace,omitempty"`

	// RequireImport gates Create on a pre-flight read; a 200 fails with
	// an "import this resource" error instead of silently upserting. Use
	// for PUT/PATCH-shaped singletons or when create and update share an
	// operationId.
	RequireImport bool `json:"requireImport,omitempty"`

	// Description overrides the generated resource description; empty
	// falls back to the create op's description.
	Description string `json:"description,omitempty"`

	// TODO
	// Examples are PCL snippets rendered as `## Example Usage` blocks.
	// SDK codegen runs `pulumi convert` per target language at gen time.
	Examples []string `json:"examples,omitempty"`
}

// Operations names the operationIds for each CRUD verb.
type Operations struct {
	Create string `json:"create,omitempty"`
	Read   string `json:"read,omitempty"`
	Update string `json:"update,omitempty"`
	Delete string `json:"delete,omitempty"`
}

// FieldMeta carries Pulumi-only overrides for a single field.
type FieldMeta struct {
	ForceNew    bool   `json:"forceNew,omitempty"`
	Secret      bool   `json:"secret,omitempty"`
	Description string `json:"description,omitempty"`

	// EmitOnCreate marks a field present only in the create response.
	// The runtime preserves it from prior state on refresh.
	EmitOnCreate bool `json:"emitOnCreate,omitempty"`

	// Unordered marks an array as set-like; Check sorts the values.
	Unordered bool `json:"unordered,omitempty"`

	// AutoName, when >0, makes the field optional on create and Check
	// generates one (from URN + random seed) capped at this max length.
	AutoName int `json:"autoName,omitempty"`
}

// ParseMetadata parses metadata.json.
func ParseMetadata(data []byte) (*Metadata, error) {
	var m Metadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("rest: parse metadata: %w", err)
	}
	return &m, nil
}
