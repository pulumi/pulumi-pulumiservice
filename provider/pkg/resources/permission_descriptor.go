// Copyright 2016-2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import "fmt"

// kindToWireType maps the user-facing `kind` discriminator to the Pulumi Cloud
// API's `__type` discriminator. Pulumi Cloud uses `__type` on the wire; the
// SDK exposes `kind` because Python's runtime strips `__`-prefixed keys from
// invoke responses (a long-standing convention; see `pulumi/sdk` Python
// `runtime/rpc.py:deserialize_property`). Translating at the provider boundary
// keeps users out of that footgun.
var kindToWireType = map[string]string{
	"allow":                     "PermissionDescriptorAllow",
	"group":                     "PermissionDescriptorGroup",
	"condition":                 "PermissionDescriptorCondition",
	"equal":                     "PermissionExpressionEqual",
	"expressionEnvironment":     "PermissionExpressionEnvironment",
	"expressionStack":           "PermissionExpressionStack",
	"expressionInsightsAccount": "PermissionExpressionInsightsAccount",
	"literalEnvironment":        "PermissionLiteralExpressionEnvironment",
	"literalStack":              "PermissionLiteralExpressionStack",
	"literalInsightsAccount":    "PermissionLiteralExpressionInsightsAccount",
}

// wireTypeToKind is the reverse direction, built from kindToWireType so the
// two stay in lockstep.
var wireTypeToKind = func() map[string]string {
	m := make(map[string]string, len(kindToWireType))
	for k, v := range kindToWireType {
		m[v] = k
	}
	return m
}()

// permissionsKindToWire walks a user-supplied PermissionDescriptor tree and
// returns a wire-shaped copy with `kind` rewritten to `__type` everywhere it
// appears. The descriptor tree is recursive — `group.entries[*]`,
// `condition.{condition,subNode}`, and `equal.{left,right}`
// each carry another descriptor or expression — so the walk recurses through
// every map-typed value it encounters.
//
// Returns an error if any node has a missing or unrecognised `kind`. Leaves
// non-discriminator fields (e.g. `permissions: ["stack:read"]`,
// `identity: "<uuid>"`) untouched.
func permissionsKindToWire(node map[string]interface{}) (map[string]interface{}, error) {
	rawKind, ok := node["kind"]
	if !ok {
		return nil, fmt.Errorf("permissions descriptor node missing required `kind` field: %v", node)
	}
	kind, ok := rawKind.(string)
	if !ok {
		return nil, fmt.Errorf("permissions descriptor `kind` must be a string, got %T", rawKind)
	}
	wireType, ok := kindToWireType[kind]
	if !ok {
		return nil, fmt.Errorf("unknown permissions descriptor kind %q", kind)
	}
	out := make(map[string]interface{}, len(node))
	for k, v := range node {
		if k == "kind" {
			continue
		}
		translated, err := translateValueKindToWire(v)
		if err != nil {
			return nil, err
		}
		out[k] = translated
	}
	out["__type"] = wireType
	return out, nil
}

// translateValueKindToWire recurses into nested maps and slices, applying
// permissionsKindToWire to anything that has a `kind` field.
func translateValueKindToWire(v interface{}) (interface{}, error) {
	switch t := v.(type) {
	case map[string]interface{}:
		if _, hasKind := t["kind"]; hasKind {
			return permissionsKindToWire(t)
		}
		// A non-discriminated nested object — copy through as-is.
		out := make(map[string]interface{}, len(t))
		for k, val := range t {
			tv, err := translateValueKindToWire(val)
			if err != nil {
				return nil, err
			}
			out[k] = tv
		}
		return out, nil
	case []interface{}:
		out := make([]interface{}, len(t))
		for i, item := range t {
			ti, err := translateValueKindToWire(item)
			if err != nil {
				return nil, err
			}
			out[i] = ti
		}
		return out, nil
	default:
		return v, nil
	}
}

// permissionsWireToKind is the inverse of permissionsKindToWire: it walks a
// wire-shaped descriptor returned by the Pulumi Cloud API and rewrites every
// `__type` back to `kind`. Used in Read so the SDK-facing state matches what
// the user originally supplied (no refresh drift).
func permissionsWireToKind(node map[string]interface{}) (map[string]interface{}, error) {
	rawType, ok := node["__type"]
	if !ok {
		return nil, fmt.Errorf("permissions descriptor node missing required `__type` field: %v", node)
	}
	wireType, ok := rawType.(string)
	if !ok {
		return nil, fmt.Errorf("permissions descriptor `__type` must be a string, got %T", rawType)
	}
	kind, ok := wireTypeToKind[wireType]
	if !ok {
		return nil, fmt.Errorf("unknown permissions descriptor __type %q", wireType)
	}
	out := make(map[string]interface{}, len(node))
	for k, v := range node {
		if k == "__type" {
			continue
		}
		translated, err := translateValueWireToKind(v)
		if err != nil {
			return nil, err
		}
		out[k] = translated
	}
	out["kind"] = kind
	return out, nil
}

func translateValueWireToKind(v interface{}) (interface{}, error) {
	switch t := v.(type) {
	case map[string]interface{}:
		if _, hasType := t["__type"]; hasType {
			return permissionsWireToKind(t)
		}
		out := make(map[string]interface{}, len(t))
		for k, val := range t {
			tv, err := translateValueWireToKind(val)
			if err != nil {
				return nil, err
			}
			out[k] = tv
		}
		return out, nil
	case []interface{}:
		out := make([]interface{}, len(t))
		for i, item := range t {
			ti, err := translateValueWireToKind(item)
			if err != nil {
				return nil, err
			}
			out[i] = ti
		}
		return out, nil
	default:
		return v, nil
	}
}
