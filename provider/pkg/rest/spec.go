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
	ID         string
	Path       string
	Method     string
	Parameters []Parameter
	RequestRef string         // $ref into components.schemas, if the body is a $ref
	ResponseRef string        // $ref of the 2xx response body, if any
	Description string
	Raw        map[string]any // entire operation object, for callers that need fields we don't model
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
		op.RequestRef = jsonContentRef(rb)
	}
	if resps, ok := raw["responses"].(map[string]any); ok {
		// Pick the first 2xx response with a JSON content body.
		for _, code := range []string{"200", "201", "202", "204"} {
			if r, ok := resps[code].(map[string]any); ok {
				if ref := jsonContentRef(r); ref != "" {
					op.ResponseRef = ref
					break
				}
			}
		}
	}
	return op
}

// jsonContentRef extracts content["application/json"].schema.$ref from a
// requestBody or response object.
func jsonContentRef(o map[string]any) string {
	content, ok := o["content"].(map[string]any)
	if !ok {
		return ""
	}
	js, ok := content["application/json"].(map[string]any)
	if !ok {
		return ""
	}
	sch, ok := js["schema"].(map[string]any)
	if !ok {
		return ""
	}
	ref, _ := sch["$ref"].(string)
	return ref
}

func isHTTPMethod(m string) bool {
	switch m {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return true
	}
	return false
}
