// Copyright 2016-2026, Pulumi Corporation.
//
// schema.go — emit Pulumi schema.json from the resource-map + OpenAPI spec.
//
// v2.0 MVP: sufficient to round-trip a resource like AgentPool from the map
// to a valid schema that `pulumi package gen-sdk` can consume. Nested
// types, enums, and complex body schemas follow as we expand coverage.

package gen

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v2"
)

// parseResourceModule splits "modName:ResName" → modName. Falls back to
// the caller's modName if the input has no colon.
func parseResourceModule(s, fallback string) string {
	if i := strings.LastIndex(s, ":"); i >= 0 {
		return s[:i]
	}
	return fallback
}

// parseResourceName splits "modName:ResName" → ResName. Falls back to the
// caller's resName if the input has no colon.
func parseResourceName(s, fallback string) string {
	if i := strings.LastIndex(s, ":"); i >= 0 {
		return s[i+1:]
	}
	return fallback
}

// pulumiSchema is a minimal view of the Pulumi package schema format we
// need to emit. Matches the JSON shape `pulumi package gen-sdk` consumes.
// Using our own struct (rather than pulumi/pkg/codegen/schema) keeps this
// package self-contained and easy to test.
type pulumiSchema struct {
	Name              string                      `json:"name"`
	Description       string                      `json:"description,omitempty"`
	DisplayName       string                      `json:"displayName,omitempty"`
	Publisher         string                      `json:"publisher,omitempty"`
	Repository        string                      `json:"repository,omitempty"`
	Homepage          string                      `json:"homepage,omitempty"`
	License           string                      `json:"license,omitempty"`
	Keywords          []string                    `json:"keywords,omitempty"`
	PluginDownloadURL string                      `json:"pluginDownloadURL,omitempty"`
	Config            *pulumiConfig               `json:"config,omitempty"`
	Provider          *pulumiResource             `json:"provider,omitempty"`
	Resources         map[string]pulumiResource   `json:"resources,omitempty"`
	Functions         map[string]pulumiFunction   `json:"functions,omitempty"`
	Types             map[string]pulumiObjectType `json:"types,omitempty"`
	Language          map[string]json.RawMessage  `json:"language,omitempty"`
}

type pulumiConfig struct {
	Variables map[string]pulumiProperty `json:"variables,omitempty"`
	Required  []string                  `json:"required,omitempty"`
}

// pulumiDefaultInfo carries the `defaultInfo` block on a config/property —
// currently used for environment-variable defaults (PULUMI_BACKEND_URL).
type pulumiDefaultInfo struct {
	Environment []string `json:"environment,omitempty"`
}

type pulumiResource struct {
	Description       string                    `json:"description,omitempty"`
	Properties        map[string]pulumiProperty `json:"properties,omitempty"`
	Required          []string                  `json:"required,omitempty"`
	InputProperties   map[string]pulumiProperty `json:"inputProperties,omitempty"`
	RequiredInputs    []string                  `json:"requiredInputs,omitempty"`
	Methods           map[string]string         `json:"methods,omitempty"`
}

type pulumiFunction struct {
	Description string                    `json:"description,omitempty"`
	Inputs      *pulumiObjectType         `json:"inputs,omitempty"`
	Outputs     *pulumiObjectType         `json:"outputs,omitempty"`
}

type pulumiObjectType struct {
	Description string                    `json:"description,omitempty"`
	Type        string                    `json:"type,omitempty"` // "object"
	Properties  map[string]pulumiProperty `json:"properties,omitempty"`
	Required    []string                  `json:"required,omitempty"`
}

type pulumiProperty struct {
	Description          string             `json:"description,omitempty"`
	Type                 string             `json:"type,omitempty"`
	Items                *pulumiProperty    `json:"items,omitempty"`
	AdditionalProperties *pulumiProperty    `json:"additionalProperties,omitempty"`
	Ref                  string             `json:"$ref,omitempty"`
	Secret               bool               `json:"secret,omitempty"`
	WillReplaceOnChanges bool               `json:"willReplaceOnChanges,omitempty"`
	Default              interface{}        `json:"default,omitempty"`
	DefaultInfo          *pulumiDefaultInfo `json:"defaultInfo,omitempty"`
	Enum                 []pulumiEnumValue  `json:"enum,omitempty"`
}

type pulumiEnumValue struct {
	Name        string      `json:"name,omitempty"`
	Value       interface{} `json:"value"`
	Description string      `json:"description,omitempty"`
}

// EmitSchema reads the spec + resource-map from disk and returns the
// Pulumi schema as a formatted JSON document. Errors short-circuit — a
// partial emit isn't useful and masks the underlying problem.
func EmitSchema(specPath, mapPath string) ([]byte, error) {
	specBytes, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("reading spec %s: %w", specPath, err)
	}
	mapBytes, err := os.ReadFile(mapPath)
	if err != nil {
		return nil, fmt.Errorf("reading resource-map %s: %w", mapPath, err)
	}
	return EmitSchemaFromBytes(specBytes, mapBytes)
}

// EmitSchemaFromBytes is the path-free form of EmitSchema — the runtime
// calls it with bytes from the embedded copies of the spec and map.
func EmitSchemaFromBytes(specBytes, mapBytes []byte) ([]byte, error) {
	spec, err := LoadSpecFromBytes(specBytes)
	if err != nil {
		return nil, err
	}
	rm, err := LoadResourceMapFromBytes(mapBytes)
	if err != nil {
		return nil, err
	}

	sch := pulumiSchema{
		Name:        "pulumiservice",
		DisplayName: "Pulumi Cloud",
		Description: "A Pulumi package for creating and managing Pulumi Cloud resources.",
		Publisher:   "Pulumi",
		Homepage:    "https://pulumi.com",
		Repository:  "https://github.com/pulumi/pulumi-pulumiservice",
		License:     "Apache-2.0",
		Keywords:    []string{"pulumi", "kind/native", "category/infrastructure"},
		Config:      providerConfig(),
		Provider:    providerResource(),
		Resources:   map[string]pulumiResource{},
		Functions:   map[string]pulumiFunction{},
		Types:       map[string]pulumiObjectType{},
		Language:    languageBlocks(),
	}

	// Build an index of component schemas for description lookup.
	compSchemas, err := loadComponentSchemasFromBytes(specBytes)
	if err != nil {
		return nil, fmt.Errorf("loading component schemas: %w", err)
	}

	for modName, mod := range rm.Modules {
		for resName, res := range mod.Resources {
			token, r, skipped := buildResourceSpec(modName, resName, res, spec, compSchemas)
			if skipped {
				// Missing operations (e.g., TODO-placeholder create op). Skip
				// emission; the coverage gate separately calls this out.
				continue
			}
			sch.Resources[token] = r
		}
		for fnName, fn := range mod.Functions {
			token := fmt.Sprintf("pulumiservice:%s:%s", modName, fnName)
			sch.Functions[token] = buildFunctionSpec(fn, spec)
		}
		// Methods are emitted as a function entry (the call surface) plus a
		// pointer from the owning resource's `methods` map. Method tokens
		// use the form pulumiservice:<modName>:<resource>/<methodName>.
		// Skip methods whose owning resource was itself skipped (e.g.,
		// read-only resources without a create op are not emitted, and
		// dangling method references would fail schema binding).
		for fullName, m := range mod.Methods {
			parts := strings.SplitN(fullName, ".", 2)
			if len(parts) != 2 {
				continue
			}
			resName, methodName := parts[0], parts[1]
			resourceToken := fmt.Sprintf("pulumiservice:%s:%s", parseResourceModule(m.Resource, modName), parseResourceName(m.Resource, resName))
			r, ok := sch.Resources[resourceToken]
			if !ok {
				continue // owning resource not emitted; skip the method too
			}
			fnToken := fmt.Sprintf("pulumiservice:%s:%s/%s", modName, resName, methodName)
			fn := buildFunctionSpec(Function{OperationID: m.OperationID}, spec)
			// Inject __self__ as a required input pointing at the owning
			// resource — Pulumi method calling convention.
			if fn.Inputs == nil {
				fn.Inputs = &pulumiObjectType{Type: "object", Properties: map[string]pulumiProperty{}}
			}
			fn.Inputs.Properties["__self__"] = pulumiProperty{Ref: "#/resources/" + resourceToken}
			fn.Inputs.Required = append([]string{"__self__"}, fn.Inputs.Required...)
			sch.Functions[fnToken] = fn
			if r.Methods == nil {
				r.Methods = map[string]string{}
			}
			r.Methods[methodName] = fnToken
			sch.Resources[resourceToken] = r
		}
	}

	// Emit author-declared types from the map's `types:` section. Keys in
	// the map are the token fragment ("stacks/hooks:WebhookFilters"); the
	// emitted token is "pulumiservice:<fragment>".
	for frag, td := range rm.Types {
		token := fmt.Sprintf("pulumiservice:%s", frag)
		sch.Types[token] = buildTypeSpec(td)
	}

	raw, err := json.MarshalIndent(sch, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling schema: %w", err)
	}
	return raw, nil
}

// buildResourceSpec produces the Pulumi resource entry for one map entry.
// Returns (token, spec, skip). skip=true if the entry has incomplete
// operations (e.g., TODO markers), in which case the caller omits it.
func buildResourceSpec(modName, resName string, res Resource, spec *Spec, compSchemas map[string]componentSchema) (string, pulumiResource, bool) {
	token := fmt.Sprintf("pulumiservice:%s:%s", modName, resName)

	// Skip resources whose canonical create operation is still a TODO
	// placeholder or missing from the spec — the coverage gate surfaces
	// them separately.
	createOpID := extractCanonicalOpID(res.Operations, "create")
	if createOpID == "" || isTodoMarker(createOpID) {
		return token, pulumiResource{}, true
	}
	if _, ok := spec.ByID[createOpID]; !ok {
		return token, pulumiResource{}, true
	}

	description := ""
	if cs, ok := compSchemas[resName]; ok {
		description = cs.Description
	}

	forceNew := map[string]bool{}
	for _, f := range res.ForceNew {
		forceNew[f] = true
	}

	inputs := map[string]pulumiProperty{}
	outputs := map[string]pulumiProperty{}
	var requiredInputs, requiredOutputs []string
	for name, p := range res.Properties {
		pp := propertyFromMap(p, compSchemas)
		if forceNew[name] {
			pp.WillReplaceOnChanges = true
		}
		if !p.Output {
			inputs[name] = pp
			// Required-input logic:
			//   - Explicit `required: true/false` wins.
			//   - Else default: `source: path` (or `pathAndBody`) is required.
			//   - requireIf/requireOneOf membership overrides to optional
			//     (those are runtime-enforced by Check).
			required := p.Source == "path" || p.Source == "pathAndBody"
			if p.Required != nil {
				required = *p.Required
			} else if p.RequireIf != "" || isInCheckSet(res.Checks, name) {
				required = false
			}
			if required {
				requiredInputs = append(requiredInputs, name)
			}
		}
		outputs[name] = pp
		if p.Source == "path" || p.Source == "response" {
			requiredOutputs = append(requiredOutputs, name)
		}
	}
	sort.Strings(requiredInputs)
	sort.Strings(requiredOutputs)

	return token, pulumiResource{
		Description:     description,
		Properties:      outputs,
		Required:        requiredOutputs,
		InputProperties: inputs,
		RequiredInputs:  requiredInputs,
	}, false
}

// extractCanonicalOpID pulls *an* operationId for a given CRUD verb from a
// resource's operations block. Handles three shapes:
//   - Flat: `create: SomeOp`
//   - Per-verb polymorphic: `create: { case: …, scopes: { a: op1, b: op2 } }`
//   - Top-level polymorphic: `operations: { case: scope, scopes: { a: {create: op1, …}, b: {create: op2, …} } }`
// The schema emitter only uses this to confirm the resource has a valid
// create path and to fetch a description — the resource's properties come
// from the map, not from the picked op.
func extractCanonicalOpID(ops yaml.MapSlice, verb string) string {
	// First try direct: "create: …" at the top level.
	for _, kv := range ops {
		if key, _ := kv.Key.(string); key == verb {
			return firstOpIDFromValue(kv.Value)
		}
	}
	// Fall back: top-level polymorphic, dive into scopes and look for
	// `verb:` under each scope, returning the first hit.
	for _, kv := range ops {
		if key, _ := kv.Key.(string); key == "scopes" || key == "values" {
			if s := firstVerbInPolymorphic(kv.Value, verb); s != "" {
				return s
			}
		}
	}
	return ""
}

// firstVerbInPolymorphic walks a scopes/values block (polymorphic operation
// variants keyed by discriminator value) and returns the first operationId
// found under `verb:` in any variant.
func firstVerbInPolymorphic(v interface{}, verb string) string {
	switch x := v.(type) {
	case yaml.MapSlice:
		for _, variant := range x {
			if s := firstVerbInNode(variant.Value, verb); s != "" {
				return s
			}
		}
	case map[interface{}]interface{}:
		for _, vv := range x {
			if s := firstVerbInNode(vv, verb); s != "" {
				return s
			}
		}
	}
	return ""
}

// firstVerbInNode inspects one scope variant (a map of verb→opID or nested
// blocks) and returns the opID under `verb:`, recursing if needed.
func firstVerbInNode(v interface{}, verb string) string {
	switch x := v.(type) {
	case yaml.MapSlice:
		for _, kv := range x {
			if key, _ := kv.Key.(string); key == verb {
				return firstOpIDFromValue(kv.Value)
			}
		}
	case map[interface{}]interface{}:
		for k, vv := range x {
			if ks, _ := k.(string); ks == verb {
				return firstOpIDFromValue(vv)
			}
		}
	}
	return ""
}

// firstOpIDFromValue finds the first string operationId reachable from a
// polymorphic/readVia operation value. Uses the same allow-list logic as
// coverage's collectOperationFromNode to avoid picking up metadata keys
// like "filterBy" or "case" as operationIds.
func firstOpIDFromValue(v interface{}) string {
	switch x := v.(type) {
	case string:
		if looksLikeOperationID(x) {
			return x
		}
	case yaml.MapSlice:
		for _, inner := range x {
			if s := firstOpIDFromValue(inner.Value); s != "" {
				return s
			}
		}
	case map[interface{}]interface{}:
		for _, vv := range x {
			if s := firstOpIDFromValue(vv); s != "" {
				return s
			}
		}
	case []interface{}:
		for _, vv := range x {
			if s := firstOpIDFromValue(vv); s != "" {
				return s
			}
		}
	}
	return ""
}

// propertyFromMap builds a pulumiProperty from a resource-map property
// entry. Resolves:
//   - `ref:` → $ref to a named type in the map's `types:` section, or a
//     primitive like "pulumi.json#/Any" (verbatim).
//   - `type: array` with `items:` → recursively emits the element shape.
//   - `type: object` with `additionalProperties:` → free-form map type.
//   - Scalar types otherwise (string default).
func propertyFromMap(p ResourceProperty, compSchemas map[string]componentSchema) pulumiProperty {
	if p.Ref != "" {
		return pulumiProperty{
			Description: p.Doc,
			Ref:         resolveRef(p.Ref),
			Secret:      p.Secret,
			Default:     p.Default,
		}
	}
	typ := p.Type
	if typ == "" {
		typ = "string"
	}
	out := pulumiProperty{
		Description: p.Doc,
		Type:        typ,
		Secret:      p.Secret,
		Default:     p.Default,
	}
	if p.Items != nil {
		inner := propertyFromMap(*p.Items, compSchemas)
		out.Items = &inner
	}
	if p.AdditionalProperties != nil {
		inner := propertyFromMap(*p.AdditionalProperties, compSchemas)
		out.AdditionalProperties = &inner
	}
	if len(p.Enum) > 0 {
		out.Enum = make([]pulumiEnumValue, 0, len(p.Enum))
		for _, v := range p.Enum {
			out.Enum = append(out.Enum, pulumiEnumValue{Value: v})
		}
	}
	return out
}

// isInCheckSet reports whether a property name participates in any of the
// resource's declarative checks (requireOneOf / requireTogether / requireIf).
// Participating properties are *not* schema-required — their conditional
// requirement is enforced at runtime in Check().
func isInCheckSet(checks []map[string]interface{}, name string) bool {
	for _, c := range checks {
		for _, key := range []string{"requireOneOf", "requireAtMostOne", "requireTogether"} {
			if v, ok := c[key]; ok {
				if slice, ok := v.([]interface{}); ok {
					for _, item := range slice {
						if s, _ := item.(string); s == name {
							return true
						}
					}
				}
			}
		}
		// requireIfSet / requireIf name scalar fields; both the trigger
		// field and the `field:` target are conditionally required.
		for _, key := range []string{"requireIfSet", "field"} {
			if s, ok := c[key].(string); ok && s == name {
				return true
			}
		}
	}
	return false
}

// resolveRef normalizes a `ref:` value. Raw refs starting with "pulumi.json"
// or "#/" pass through; bare tokens like "stacks/hooks:WebhookFilters" are
// prefixed with the package-local type namespace.
func resolveRef(ref string) string {
	if len(ref) > 0 && (ref[0] == '#' || (len(ref) > 11 && ref[:11] == "pulumi.json")) {
		return ref
	}
	return "#/types/pulumiservice:" + ref
}

// buildTypeSpec renders a TypeDef into the schema's `types:` map entry.
func buildTypeSpec(td TypeDef) pulumiObjectType {
	props := map[string]pulumiProperty{}
	for name, p := range td.Properties {
		props[name] = propertyFromMap(p, nil)
	}
	kind := td.Type
	if kind == "" {
		kind = "object"
	}
	required := append([]string(nil), td.Required...)
	sort.Strings(required)
	return pulumiObjectType{
		Description: td.Description,
		Type:        kind,
		Properties:  props,
		Required:    required,
	}
}

// buildFunctionSpec emits a Pulumi function spec from a map entry.
// Inputs/outputs are declared inline on the Function (the map is the
// authoritative source). The spec argument is reserved for future fallback
// behavior (e.g., pulling descriptions from OpenAPI when a map doc is
// omitted); today it is unused.
func buildFunctionSpec(fn Function, spec *Spec) pulumiFunction {
	_ = spec
	out := pulumiFunction{Description: fn.Doc}
	if len(fn.Inputs) > 0 {
		out.Inputs = buildInlineObject(fn.Inputs)
	}
	if len(fn.Outputs) > 0 {
		out.Outputs = buildInlineObject(fn.Outputs)
	}
	return out
}

// buildInlineObject wraps a property block as an inline object type (used
// for function inputs/outputs).
func buildInlineObject(props map[string]ResourceProperty) *pulumiObjectType {
	out := map[string]pulumiProperty{}
	var required []string
	for name, p := range props {
		out[name] = propertyFromMap(p, nil)
		if !p.Output {
			if p.Source == "path" || p.Source == "query" {
				required = append(required, name)
			}
		}
	}
	sort.Strings(required)
	return &pulumiObjectType{
		Type:       "object",
		Properties: out,
		Required:   required,
	}
}

// componentSchema is a narrow projection of an OpenAPI components.schemas entry.
type componentSchema struct {
	Description string
	Properties  map[string]componentProperty
	Required    []string
}

type componentProperty struct {
	Description string
	Type        string
	Ref         string
}

// loadComponentSchemasFromBytes parses the spec a second time to extract
// just the component schema descriptions we use for doc comments. Kept
// as a separate narrow parse (rather than expanding LoadSpec) so the
// coverage gate stays cheap.
func loadComponentSchemasFromBytes(raw []byte) (map[string]componentSchema, error) {
	var doc struct {
		Components struct {
			Schemas map[string]struct {
				Description string                    `json:"description"`
				Type        string                    `json:"type"`
				Required    []string                  `json:"required"`
				Properties  map[string]componentProperty `json:"properties"`
			} `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	out := make(map[string]componentSchema, len(doc.Components.Schemas))
	for name, s := range doc.Components.Schemas {
		out[name] = componentSchema{
			Description: s.Description,
			Properties:  s.Properties,
			Required:    s.Required,
		}
	}
	return out, nil
}

