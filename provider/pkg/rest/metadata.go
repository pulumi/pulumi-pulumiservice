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

// Metadata is the parsed contents of metadata.json.
//
// This file is the human-curated bridge between OpenAPI operations and Pulumi
// resources. It declares which operationIds form which resources, plus
// Pulumi-only metadata that the OpenAPI spec doesn't carry (forceNew, secret,
// aliases, etc.). Field shapes are derived from the OpenAPI spec; this
// document only carries Pulumi-only overrides.
type Metadata struct {
	// Version pins the schema version of this file. Currently 1.
	Version int `json:"version"`

	// Package is the default package name for any token that doesn't already
	// include a package prefix (most metadata uses fully-qualified tokens
	// instead).
	Package string `json:"package,omitempty"`

	// Resources keyed by fully-qualified Pulumi token ("pkg:module:Name").
	Resources map[string]ResourceMeta `json:"resources"`

	// Functions keyed by fully-qualified Pulumi token. Reserved for future use.
	Functions map[string]ResourceMeta `json:"functions,omitempty"`
}

// ResourceMeta describes one Pulumi resource derived from OpenAPI operations.
type ResourceMeta struct {
	// Operations names the operationIds that implement each CRUD verb. Create
	// and Read are required; Update may be empty (replace-on-change); Delete
	// may be empty (no deletion semantics on the API side).
	Operations Operations `json:"operations"`

	// Token, when non-empty, overrides the metadata-key as the user-facing
	// Pulumi token for this resource. Lets us expose a clean module path
	// (e.g. "pulumiservice:v2/esc:Environment") while keeping the map key
	// as the canonical OpenAPI-derived identifier the scaffolder rebuilds
	// from spec.json (so go generate stays a no-op on these entries).
	Token string `json:"token,omitempty"`

	// Aliases are fully-qualified Pulumi tokens that the engine should treat as
	// equivalent to this resource (used for in-place migration after renames).
	Aliases []string `json:"aliases,omitempty"`

	// Fields applies Pulumi-only overrides per field. Keys are Pulumi-side
	// field names.
	Fields map[string]FieldMeta `json:"fields,omitempty"`

	// Renames maps Pulumi-side field names to OpenAPI-side names. Used when
	// the SDK-facing name should differ from the wire name.
	Renames map[string]string `json:"renames,omitempty"`

	// Outputs is an optional allowlist of response field names exposed as
	// State outputs. If empty, all response fields are exposed. Corresponds to
	// design option B from the brainstorm.
	Outputs []string `json:"outputs,omitempty"`

	// OutputsExclude is an optional denylist subtracted from the response
	// schema. Mutually exclusive with Outputs (Outputs wins if both set).
	OutputsExclude []string `json:"outputsExclude,omitempty"`

	// IDFormat is a template controlling resource-ID synthesis and import-time
	// parsing. Use "{paramName}" placeholders for path-parameter values
	// (Pulumi-side names after Renames). Example: "{org}/{name}". When unset,
	// the synthesizer slash-joins path-parameter values from the most
	// authoritative non-create op.
	IDFormat string `json:"idFormat,omitempty"`

	// DeleteBeforeReplace makes Pulumi delete the old instance before creating
	// the new one on a replacement (instead of the default create-new-then-
	// delete-old). Use for resources whose names collide on duplicate create
	// and that aren't auto-named.
	DeleteBeforeReplace bool `json:"deleteBeforeReplace,omitempty"`

	// Description is an optional override of the resource description in the
	// generated schema. If empty, the schema builder falls back to the create
	// operation's description.
	Description string `json:"description,omitempty"`

	// Examples are canonical PCL (Pulumi Configuration Language) snippets
	// rendered into the resource's description as `## Example Usage` blocks.
	// Pulumi's SDK codegen runs `pulumi convert` from PCL to each target
	// language at gen time, so a single PCL example becomes per-language
	// snippets in TypeScript/Python/Go/.NET/Java docstrings and on the
	// Registry doc page.
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

	// EmitOnCreate marks a field that's only present in the create response,
	// never readable thereafter. The runtime preserves it from prior state
	// during refresh; the schema builder includes it in outputs even if the
	// read-response schema doesn't carry it.
	EmitOnCreate bool `json:"emitOnCreate,omitempty"`

	// Unordered marks an array field as set-like; Check sorts the values
	// before returning so user-side reordering doesn't trigger spurious diffs.
	Unordered bool `json:"unordered,omitempty"`

	// AutoName, when >0, makes the field optional on create. If the user
	// doesn't provide a value, Check generates one from the resource URN and
	// the engine-supplied random seed, capped at this max length.
	AutoName int `json:"autoName,omitempty"`
}

// ParseMetadata parses metadata.json.
func ParseMetadata(data []byte) (*Metadata, error) {
	var m Metadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("rest: parse metadata: %w", err)
	}
	if m.Version != 1 {
		return nil, fmt.Errorf("rest: unsupported metadata version %d (expected 1)", m.Version)
	}
	return &m, nil
}

