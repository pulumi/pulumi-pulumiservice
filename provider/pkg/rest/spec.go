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
	"strings"
)

// Operation is the subset of an OpenAPI operation needed at runtime.
type Operation struct {
	ID                  string
	Path                string
	Method              string
	Parameters          []Parameter
	RequestRef          string // $ref into components.schemas, set only for application/json bodies with a $ref schema
	RequestContentType  string // wire content type of the request body, if any (e.g. "application/json", "application/x-yaml")
	ResponseRef         string // $ref of the 2xx response body, set only for application/json responses with a $ref schema
	ResponseContentType string // wire content type of the chosen 2xx response, if any
	Description         string
	Deprecated          bool           // true when the upstream spec marks this op deprecated
	Raw                 map[string]any // entire operation object, for callers that need fields we don't model
}

// Parameter is the subset of an OpenAPI parameter needed at runtime.
type Parameter struct {
	Name        string // wire name (matches the {placeholder} in Path for in=path)
	In          string // "path" | "query" | "header" | "cookie"
	Required    bool
	Description string
	SchemaType  string // "string" | "integer" | "number" | "boolean" | "" if absent
}

// Spec is a parsed OpenAPI 3 document indexed for fast lookup.
//
// We only model what BuildSchema and the dynamic resource handler need:
// the operation index by operationId, and components.schemas resolution
// (for following $ref chains in request/response bodies).
type Spec struct {
	Servers   []string
	ops       map[string]*Operation
	schemas   map[string]map[string]any // components/schemas by name
}

// ParseSchema parses an OpenAPI 3 JSON document.
func ParseSpec(data []byte) (*Spec, error) {
	var doc struct {
		Servers []struct {
			URL string `json:"url"`
		} `json:"servers"`
		Paths      map[string]map[string]any `json:"paths"`
		Components struct {
			Schemas map[string]map[string]any `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("rest: parse OpenAPI spec: %w", err)
	}

	s := &Spec{
		ops:     make(map[string]*Operation),
		schemas: doc.Components.Schemas,
	}
	for _, srv := range doc.Servers {
		s.Servers = append(s.Servers, srv.URL)
	}

	for path, pathItem := range doc.Paths {
		for method, raw := range pathItem {
			um := strings.ToUpper(method)
			if !isHTTPMethod(um) {
				continue
			}
			rawObj, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			id, _ := rawObj["operationId"].(string)
			if id == "" {
				continue
			}
			if existing, dup := s.ops[id]; dup {
				return nil, fmt.Errorf(
					"rest: duplicate operationId %q (paths %s and %s)",
					id, existing.Path, path)
			}
			s.ops[id] = parseOperation(id, path, um, rawObj)
		}
	}
	return s, nil
}

// Op returns the operation with the given operationId.
func (s *Spec) Op(id string) (*Operation, bool) {
	op, ok := s.ops[id]
	return op, ok
}

// AllOps returns a defensive copy of the operationId → Operation map,
// for callers that need to enumerate every op in the spec (e.g., a test
// that asserts every op is classified).
func (s *Spec) AllOps() map[string]*Operation {
	out := make(map[string]*Operation, len(s.ops))
	for id, op := range s.ops {
		out[id] = op
	}
	return out
}

// ResolveSchema looks up a $ref string of the form "#/components/schemas/Name"
// and returns the raw schema object. Returns false for $refs we can't resolve
// (external references, malformed strings).
func (s *Spec) ResolveSchema(ref string) (map[string]any, bool) {
	const prefix = "#/components/schemas/"
	if !strings.HasPrefix(ref, prefix) {
		return nil, false
	}
	got, ok := s.schemas[strings.TrimPrefix(ref, prefix)]
	return got, ok
}

// SchemaByName looks up a component schema directly.
func (s *Spec) SchemaByName(name string) (map[string]any, bool) {
	got, ok := s.schemas[name]
	return got, ok
}

func parseOperation(id, path, method string, raw map[string]any) *Operation {
	op := &Operation{
		ID:     id,
		Path:   path,
		Method: method,
		Raw:    raw,
	}
	if d, ok := raw["description"].(string); ok {
		op.Description = d
	}
	if dep, _ := raw["deprecated"].(bool); dep {
		op.Deprecated = true
	}
	if rp, ok := raw["x-pulumi-route-property"].(map[string]any); ok {
		if v, _ := rp["Visibility"].(string); v == "Deprecated" {
			op.Deprecated = true
		}
	}
	if params, ok := raw["parameters"].([]any); ok {
		for _, p := range params {
			pm, ok := p.(map[string]any)
			if !ok {
				continue
			}
			pp := Parameter{}
			pp.Name, _ = pm["name"].(string)
			pp.In, _ = pm["in"].(string)
			pp.Required, _ = pm["required"].(bool)
			pp.Description, _ = pm["description"].(string)
			if sch, ok := pm["schema"].(map[string]any); ok {
				pp.SchemaType, _ = sch["type"].(string)
			}
			op.Parameters = append(op.Parameters, pp)
		}
	}
	if rb, ok := raw["requestBody"].(map[string]any); ok {
		if ref := jsonContentRef(rb); ref != "" {
			op.RequestRef = ref
			op.RequestContentType = "application/json"
		} else if bodySchema(rb, "application/x-yaml") != nil {
			op.RequestContentType = "application/x-yaml"
		}
	}
	if resps, ok := raw["responses"].(map[string]any); ok {
		// Pick the first 2xx with a JSON $ref body, preserving the
		// existing JSON-first preference. Fall back to application/x-yaml
		// only when no JSON body is found at any 2xx code.
		for _, code := range []string{"200", "201", "202", "204"} {
			if r, ok := resps[code].(map[string]any); ok {
				if ref := jsonContentRef(r); ref != "" {
					op.ResponseRef = ref
					op.ResponseContentType = "application/json"
					break
				}
			}
		}
		if op.ResponseContentType == "" {
			for _, code := range []string{"200", "201", "202", "204"} {
				if r, ok := resps[code].(map[string]any); ok {
					if bodySchema(r, "application/x-yaml") != nil {
						op.ResponseContentType = "application/x-yaml"
						break
					}
				}
			}
		}
	}
	return op
}

// jsonContentRef extracts content["application/json"].schema.$ref from a
// requestBody or response object. Returns "" when the body is not JSON or
// has an inline (non-$ref) schema.
func jsonContentRef(o map[string]any) string {
	sch := bodySchema(o, "application/json")
	if sch == nil {
		return ""
	}
	ref, _ := sch["$ref"].(string)
	return ref
}

// bodySchema returns the schema object for the given content type from a
// requestBody or response object, or nil if absent.
func bodySchema(o map[string]any, contentType string) map[string]any {
	content, ok := o["content"].(map[string]any)
	if !ok {
		return nil
	}
	entry, ok := content[contentType].(map[string]any)
	if !ok {
		return nil
	}
	sch, _ := entry["schema"].(map[string]any)
	return sch
}

func isHTTPMethod(m string) bool {
	switch m {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return true
	}
	return false
}
