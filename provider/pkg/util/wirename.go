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

package util

import "strings"

// Pulumi Cloud names its discriminators with a `__` prefix (`__type`), which a
// Pulumi schema cannot carry: Go codegen emits an unexported struct field,
// Python mangles the constructor parameter, and the engine treats
// `__`-prefixed keys as internal — hiding them from diffs and stripping them
// in the .NET deserializer. The schema therefore exposes the suffix form
// (`type__`) and the provider rewrites names in both directions at every wire
// boundary. Python's mangling rule needs two or more *leading* underscores, so
// the suffix form is immune by language definition.
//
// The mapping is injective because no Pulumi Cloud property ends in `__`. It
// is not surjective: a *schema-side* name that already starts with `__` has no
// wire preimage, so ToWireName passes it through rather than inventing one.
// Such a name cannot reach the provider anyway — the engine strips
// `__`-prefixed keys before Check sees them.
const (
	wireNamePrefix   = "__"
	schemaNameSuffix = "__"
)

// ToSchemaName maps a wire property name to the name the Pulumi schema exposes.
func ToSchemaName(wire string) string {
	if base, ok := strings.CutPrefix(wire, wireNamePrefix); ok && base != "" {
		return base + schemaNameSuffix
	}
	return wire
}

// ToWireName inverts ToSchemaName. Names that already carry the wire prefix
// pass through, so the function is idempotent — the request path applies it
// once when resolving declared renames and again when walking the body tree.
func ToWireName(schema string) string {
	if strings.HasPrefix(schema, wireNamePrefix) {
		return schema
	}
	if base, ok := strings.CutSuffix(schema, schemaNameSuffix); ok && base != "" {
		return wireNamePrefix + base
	}
	return schema
}

// ToWireTree rewrites schema-side property names to wire names throughout a
// JSON-ready request body, and ToSchemaTree does the reverse for a decoded
// response. Both are purely lexical and recurse through nested objects and
// arrays: the runtime only resolves the top-level body schema, and the
// discriminated trees that carry `__type` sit arbitrarily deep inside slots
// the schema may have degraded to Any. A free-form map key that happens to
// match the pattern would be rewritten too, but no Pulumi Cloud property or
// key ends in `__`.
func ToWireTree(m map[string]any) map[string]any {
	return renameTree(m, ToWireName)
}

func ToSchemaTree(m map[string]any) map[string]any {
	return renameTree(m, ToSchemaName)
}

func renameTree(m map[string]any, rename func(string) string) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[rename(k)] = renameTreeValue(v, rename)
	}
	return out
}

func renameTreeValue(v any, rename func(string) string) any {
	switch x := v.(type) {
	case map[string]any:
		return renameTree(x, rename)
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = renameTreeValue(e, rename)
		}
		return out
	default:
		return v
	}
}
