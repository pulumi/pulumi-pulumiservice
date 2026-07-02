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
	"net/http"
	"sort"
	"strings"
)

// Centralized wire literals so goconst doesn't keep flagging the same
// strings as new call sites appear. HTTP methods come from net/http.
// Use these in place of the bare strings everywhere in this package.
const (
	contentJSON = "application/json"
	contentYAML = "application/x-yaml"
	inPath      = "path"  // Parameter.In value for path parameters
	inQuery     = "query" // Parameter.In value for query parameters

	// descriptionKey is the operation/property field name carrying a description.
	descriptionKey = "description"
)

// Operation is the subset of an OpenAPI operation needed at runtime.
type Operation struct {
	ID                  string
	Path                string
	Method              string
	Parameters          []Parameter
	RequestRef          string // $ref of the JSON body schema, if any
	RequestContentType  string // e.g. "application/json", "application/x-yaml"
	ResponseRef         string // $ref of the 2xx JSON body schema, if any
	ResponseContentType string
	Description         string
	Deprecated          bool
	Raw                 map[string]any // raw operation object
}

// Parameter is the subset of an OpenAPI parameter needed at runtime.
type Parameter struct {
	Name        string // wire name (matches {placeholder} for in=path)
	In          string // "path" | "query" | "header" | "cookie"
	Required    bool
	Description string
	SchemaType  string // "string" | "integer" | "number" | "boolean" | ""
}

// Spec is a parsed OpenAPI 3 document indexed by operationId.
type Spec struct {
	Servers []string
	ops     map[string]*Operation
	schemas map[string]map[string]any // components/schemas by name
}

// ParseSpec parses an OpenAPI 3 JSON document.
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

	// Iterate paths in sorted order so duplicate-operationId errors name
	// the same (existing, conflicting) pair every run.
	paths := make([]string, 0, len(doc.Paths))
	for path := range doc.Paths {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		pathItem := doc.Paths[path]
		methods := make([]string, 0, len(pathItem))
		for method := range pathItem {
			methods = append(methods, method)
		}
		sort.Strings(methods)
		for _, method := range methods {
			raw := pathItem[method]
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

// AllOps returns a defensive copy of the operationId → Operation map.
func (s *Spec) AllOps() map[string]*Operation {
	out := make(map[string]*Operation, len(s.ops))
	for id, op := range s.ops {
		out[id] = op
	}
	return out
}

// ResolveSchema resolves a "#/components/schemas/Name" $ref. Returns false
// for external or malformed refs.
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
	if d, ok := raw[descriptionKey].(string); ok {
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
			pp.Description, _ = pm[descriptionKey].(string)
			if sch, ok := pm["schema"].(map[string]any); ok {
				pp.SchemaType, _ = sch["type"].(string)
			}
			op.Parameters = append(op.Parameters, pp)
		}
	}
	if rb, ok := raw["requestBody"].(map[string]any); ok {
		if ref := jsonContentRef(rb); ref != "" {
			op.RequestRef = ref
			op.RequestContentType = contentJSON
		} else if bodySchema(rb, contentYAML) != nil {
			op.RequestContentType = contentYAML
		}
	}
	if resps, ok := raw["responses"].(map[string]any); ok {
		// JSON $ref body wins; fall back to application/x-yaml only if no JSON
		// body is found at any 2xx code.
		for _, code := range []string{"200", "201", "202", "204"} {
			if r, ok := resps[code].(map[string]any); ok {
				if ref := jsonContentRef(r); ref != "" {
					op.ResponseRef = ref
					op.ResponseContentType = contentJSON
					break
				}
			}
		}
		if op.ResponseContentType == "" {
			for _, code := range []string{"200", "201", "202", "204"} {
				if r, ok := resps[code].(map[string]any); ok {
					if bodySchema(r, contentYAML) != nil {
						op.ResponseContentType = contentYAML
						break
					}
				}
			}
		}
	}
	return op
}

// jsonContentRef returns content[contentJSON].schema.$ref, or "".
func jsonContentRef(o map[string]any) string {
	sch := bodySchema(o, contentJSON)
	if sch == nil {
		return ""
	}
	ref, _ := sch["$ref"].(string)
	return ref
}

// bodySchema returns the schema for the given content type, or nil.
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
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodDelete, http.MethodHead, http.MethodOptions:
		return true
	}
	return false
}
