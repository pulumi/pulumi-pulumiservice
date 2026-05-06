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

// Package rest builds Pulumi resources from a pair of inputs:
//
//   - an OpenAPI 3 spec describing the wire shapes
//   - a metadata document declaring which operationIds form which Pulumi
//     resources, plus per-field overrides (forceNew, secret, renames,
//     output allow/deny lists) and per-resource aliases
//
// Both are loaded at runtime; nothing is generated at build time. The
// schema returned by the provider is computed from (spec, metadata) on
// demand. Mapping errors surface at GetSchema time, not at startup —
// this keeps `go build` simple and makes the provider amenable to
// runtime parameterization (a future feature: replacing the embedded
// spec/metadata via the Pulumi engine's Parameterize RPC).
//
// The two main entry points:
//
//   - BuildSchema(spec, meta, pkg) emits a complete schema.PackageSpec.
//   - Resources(spec, meta, pkg) emits a slice of mw.CustomResource
//     handlers ready to register with dispatch.Wrap.
//
// Resource decoded I/O uses resource.PropertyMap throughout — there are
// no per-resource Go types to maintain.
package rest
