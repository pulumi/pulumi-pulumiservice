// Copyright 2016-2026, Pulumi Corporation.
//
// parse.go — load the pinned OpenAPI spec and the resource map, producing
// a single in-memory model both the coverage gate and (in subsequent PRs)
// the schema and metadata emitters consume.

package gen

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

// SpecOperation describes one operation (an HTTP method on a path) extracted
// from the OpenAPI spec. We only keep what the coverage gate and later
// emitters need; the raw spec is re-read by downstream steps that need type
// details.
type SpecOperation struct {
	OperationID string
	Method      string // uppercase: GET, POST, PUT, PATCH, DELETE
	Path        string
}

// Spec is the minimal projection of the OpenAPI spec the generator consumes.
type Spec struct {
	Operations []SpecOperation
	// operationID → operation, for fast lookup during mapping validation.
	ByID map[string]SpecOperation
}

// LoadSpec reads an OpenAPI 3.x JSON file and extracts operations. A minimal
// parser that doesn't pull in heavy dependencies; later emitters can switch
// to a typed OpenAPI library when they need component schemas.
func LoadSpec(path string) (*Spec, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading spec %s: %w", path, err)
	}
	return LoadSpecFromBytes(raw)
}

// LoadSpecFromBytes is the path-free form of LoadSpec — the runtime calls
// it with bytes from the embedded copy of the spec, while the disk-based
// LoadSpec is retained for tests that pass a path.
func LoadSpecFromBytes(raw []byte) (*Spec, error) {
	var doc struct {
		Paths map[string]map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parsing spec: %w", err)
	}
	httpVerbs := map[string]bool{"get": true, "post": true, "put": true, "patch": true, "delete": true}
	sp := &Spec{ByID: map[string]SpecOperation{}}
	for path, ops := range doc.Paths {
		for method, body := range ops {
			if !httpVerbs[strings.ToLower(method)] {
				continue
			}
			var op struct {
				OperationID string `json:"operationId"`
			}
			if err := json.Unmarshal(body, &op); err != nil {
				// Non-operation entry (parameters, summary, description). Skip.
				continue
			}
			if op.OperationID == "" {
				continue
			}
			entry := SpecOperation{
				OperationID: op.OperationID,
				Method:      strings.ToUpper(method),
				Path:        path,
			}
			sp.Operations = append(sp.Operations, entry)
			sp.ByID[op.OperationID] = entry
		}
	}
	return sp, nil
}

// ResourceMap is the in-memory form of resource-map.yaml. We deliberately
// only parse the fields the coverage gate cares about; the full schema is
// decoded into loosely-typed maps so additions don't require recompiling.
type ResourceMap struct {
	Modules    map[string]Module `yaml:"modules"`
	Methods    map[string]Method `yaml:"methods"`
	Exclusions []ExclusionEntry  `yaml:"exclusions"`
	// Types holds named complex types referenced by resource properties via
	// `ref:`. Key is the token fragment (e.g., "stacks/hooks:WebhookFilters");
	// the emitter prefixes with "pulumiservice:" when emitting.
	Types map[string]TypeDef `yaml:"types,omitempty"`
}

// TypeDef declares a named object/enum type usable as a `ref:` target from
// any property block. Kept narrow — only object types are emitted today;
// enums use the per-property `enum:` field.
type TypeDef struct {
	Type        string                      `yaml:"type,omitempty"`
	Description string                      `yaml:"description,omitempty"`
	Properties  map[string]ResourceProperty `yaml:"properties,omitempty"`
	Required    []string                    `yaml:"required,omitempty"`
}

// Module is the unit of organization in the map; corresponds to a Pulumi
// sub-module namespace.
type Module struct {
	Resources map[string]Resource `yaml:"resources"`
	Functions map[string]Function `yaml:"functions"`
	Methods   map[string]Method   `yaml:"methods"`
}

// Resource is one generated resource in the map. Operations stays as
// yaml.MapSlice because its shape is heterogeneous (flat CRUD / polymorphic
// case / readVia) — coverage.go walks it structurally. Other fields are
// typed because the schema emitter and runtime metadata emitter consume
// them directly.
type Resource struct {
	Operations     yaml.MapSlice               `yaml:"operations"`
	ID             *ResourceID                 `yaml:"id,omitempty"`
	ForceNew       []string                    `yaml:"forceNew,omitempty"`
	Properties     map[string]ResourceProperty `yaml:"properties,omitempty"`
	Discriminator  *ResourceDiscriminator      `yaml:"discriminator,omitempty"`
	Checks         []map[string]interface{}    `yaml:"checks,omitempty"`
	SortProperties []string                    `yaml:"sortProperties,omitempty"`
	AutoName       *ResourceAutoName           `yaml:"autoname,omitempty"`
	Doc            string                      `yaml:"doc,omitempty"`
}

// ResourceID holds the simple or polymorphic ID template configuration.
type ResourceID struct {
	Template  string            `yaml:"template,omitempty"`
	Params    []string          `yaml:"params,omitempty"`
	Case      string            `yaml:"case,omitempty"`
	Templates map[string]string `yaml:"templates,omitempty"`
}

// ResourceProperty mirrors the runtime.CloudAPIProperty shape at map-parse
// time. (We keep two separate types: this one is what appears in
// resource-map.yaml; runtime.CloudAPIProperty is what the generator emits
// into metadata.json.)
type ResourceProperty struct {
	From             string      `yaml:"from,omitempty"`
	Source           string      `yaml:"source,omitempty"`
	CreateSource     string      `yaml:"createSource,omitempty"`
	CreateFrom       string      `yaml:"createFrom,omitempty"`
	PathName         string      `yaml:"pathName,omitempty"`
	BodyFrom         string      `yaml:"bodyFrom,omitempty"`
	Type             string      `yaml:"type,omitempty"`
	Secret           bool        `yaml:"secret,omitempty"`
	Output           bool        `yaml:"output,omitempty"`
	WriteOnly        bool        `yaml:"writeOnly,omitempty"`
	DiffMode         string      `yaml:"diffMode,omitempty"`
	Default          interface{} `yaml:"default,omitempty"`
	DefaultFromField string      `yaml:"defaultFromField,omitempty"`
	SortOnRead       bool        `yaml:"sortOnRead,omitempty"`
	RequireIf        string      `yaml:"requireIf,omitempty"`
	Enum             []string    `yaml:"enum,omitempty"`
	Aliases          []string    `yaml:"aliases,omitempty"`
	Doc              string      `yaml:"doc,omitempty"`
	// Required explicitly marks an input as required (true) or optional
	// (false). If unset, the emitter falls back to "source: path" meaning
	// required. Use `required: false` on path-sourced properties that are
	// mutually-exclusive scope selectors (e.g., Webhook's
	// stackName/environmentName/organizationName trio).
	Required *bool `yaml:"required,omitempty"`
	// Items describes the element type for `type: array` properties.
	Items *ResourceProperty `yaml:"items,omitempty"`
	// AdditionalProperties describes the value type for `type: object`
	// properties that act as free-form maps (used sparingly).
	AdditionalProperties *ResourceProperty `yaml:"additionalProperties,omitempty"`
	// Ref names a type in the map's top-level `types:` section or a
	// primitive (e.g., "pulumi.json#/Any"). When set, overrides Type.
	Ref string `yaml:"ref,omitempty"`
}

// ResourceDiscriminator names the input property that selects a polymorphic
// variant (e.g. Team.type = pulumi|github).
type ResourceDiscriminator struct {
	Field  string   `yaml:"field"`
	Values []string `yaml:"values,omitempty"`
}

// ResourceAutoName configures automatic name generation.
type ResourceAutoName struct {
	Property string `yaml:"property"`
	Pattern  string `yaml:"pattern,omitempty"`
	Kind     string `yaml:"kind,omitempty"`
}

// Function / Method / ExclusionEntry are minimal views.
type Function struct {
	OperationID string                      `yaml:"operationId"`
	Doc         string                      `yaml:"doc,omitempty"`
	Inputs      map[string]ResourceProperty `yaml:"inputs,omitempty"`
	Outputs     map[string]ResourceProperty `yaml:"outputs,omitempty"`
}
type Method struct {
	Resource    string `yaml:"resource"`
	OperationID string `yaml:"operationId"`
}
type ExclusionEntry struct {
	OperationID string `yaml:"operationId"`
	Reason      string `yaml:"reason"`
}

// LoadResourceMap reads and parses the resource-map.yaml.
func LoadResourceMap(path string) (*ResourceMap, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading resource-map %s: %w", path, err)
	}
	return LoadResourceMapFromBytes(raw)
}

// LoadResourceMapFromBytes is the path-free form of LoadResourceMap —
// the runtime calls it with bytes from the embedded copy of the map.
func LoadResourceMapFromBytes(raw []byte) (*ResourceMap, error) {
	var rm ResourceMap
	if err := yaml.Unmarshal(raw, &rm); err != nil {
		return nil, fmt.Errorf("parsing resource-map: %w", err)
	}
	return &rm, nil
}

// ExtractOperationIDs walks the map and returns every operationId it claims,
// along with a human-readable citation of where the claim came from. Used by
// the coverage gate to attribute each operation to its owner.
func (rm *ResourceMap) ExtractOperationIDs() []OperationClaim {
	var claims []OperationClaim
	for modName, mod := range rm.Modules {
		for resName, res := range mod.Resources {
			for _, opKV := range res.Operations {
				collectOperationFromNode(opKV, fmt.Sprintf("modules.%s.resources.%s", modName, resName), &claims)
			}
		}
		for fnName, fn := range mod.Functions {
			if fn.OperationID != "" && !isTodoMarker(fn.OperationID) {
				claims = append(claims, OperationClaim{
					OperationID: fn.OperationID,
					ClaimedBy:   fmt.Sprintf("modules.%s.functions.%s", modName, fnName),
					Kind:        ClaimFunction,
				})
			}
		}
		for methodName, m := range mod.Methods {
			if m.OperationID != "" && !isTodoMarker(m.OperationID) {
				claims = append(claims, OperationClaim{
					OperationID: m.OperationID,
					ClaimedBy:   fmt.Sprintf("modules.%s.methods.%s", modName, methodName),
					Kind:        ClaimMethod,
				})
			}
		}
	}
	for methodName, m := range rm.Methods {
		if m.OperationID != "" && !isTodoMarker(m.OperationID) {
			claims = append(claims, OperationClaim{
				OperationID: m.OperationID,
				ClaimedBy:   fmt.Sprintf("methods.%s", methodName),
				Kind:        ClaimMethod,
			})
		}
	}
	for _, e := range rm.Exclusions {
		if e.OperationID != "" {
			claims = append(claims, OperationClaim{
				OperationID: e.OperationID,
				ClaimedBy:   "exclusions",
				Kind:        ClaimExclusion,
				Reason:      e.Reason,
			})
		}
	}
	return claims
}

// collectOperationFromNode walks a value that represents the `operations:`
// block for a resource and emits a claim for each operationId it finds.
//
// The block is heterogeneous: flat (`create: CreateFoo`), polymorphic
// (`case: type` with nested scopes), or using the `readVia:` fallback for
// resources whose read path isn't a simple GET. To avoid treating metadata
// values like `case: type` or `filterBy: id` as operationIds, we only
// claim values for keys in the allow-list below.
var operationIDKeys = map[string]bool{
	"create":      true,
	"postCreate":  true, // second-step op executed after create (e.g. ESC env YAML PATCH)
	"read":        true,
	"update":      true,
	"delete":      true,
	"operationId": true,
}

// metadataKeys hold configuration values inside an operations block that are
// never operationIds — we recurse into their sub-trees but never claim them
// as operations themselves.
var metadataKeys = map[string]bool{
	"bodyOverride":    true, // fixed request-body for tombstone-style ops
	"case":            true, // discriminator field name
	"contentType":     true, // op-level HTTP Content-Type override
	"extractField":    true, // readVia: parent response field to pluck
	"filterBy":        true, // key to filter by when using readVia
	"iterateKeyParam": true, // path placeholder for iterated delete key
	"iterateOver":     true, // delete iterates per-key over this property
	"keyBy":           true, // readVia: resource property used as the map key
	"rawBodyFrom":     true, // op-level: property feeding the raw HTTP body
	"rawBodyTo":       true, // op-level: property receiving the raw HTTP response
	"readVia":         true, // wrapper containing a nested operationId
	"scope":           true, // scope discriminator
	"scopes":          true, // wrapper containing scope → operations maps
	"valueProperty":   true, // readVia: SDK property to receive extracted value
	"values":          true, // enum value list
}

func collectOperationFromNode(kv yaml.MapItem, claimedBy string, claims *[]OperationClaim) {
	key, _ := kv.Key.(string)
	switch v := kv.Value.(type) {
	case string:
		// A string value is an operationId if its key is explicitly one of
		// the canonical CRUD verbs (create/read/update/delete/operationId)
		// OR it's any non-metadata key inside a nested block — the
		// discriminator-value case like `pulumi: CreatePulumiTeam` under a
		// polymorphic `create:` block.
		if isTodoMarker(v) {
			return
		}
		if metadataKeys[key] {
			return
		}
		if !operationIDKeys[key] && !looksLikeOperationID(v) {
			// Non-CRUD key with a value that doesn't look like an operationId;
			// probably pure metadata like `case: scope`. Ignore.
			return
		}
		*claims = append(*claims, OperationClaim{
			OperationID: v,
			ClaimedBy:   claimedBy + "." + key,
			Kind:        ClaimResource,
		})
	case yaml.MapSlice:
		// Recurse into nested structures (readVia, scopes, polymorphic create).
		for _, sub := range v {
			collectOperationFromNode(sub, claimedBy+"."+key, claims)
		}
	}
}

// looksLikeOperationID heuristically decides whether a string value is a
// plausible Pulumi Cloud operationId. Used only to disambiguate values of
// non-CRUD keys (e.g., discriminator values like `pulumi: CreatePulumiTeam`)
// from pure metadata values like `case: type`.
//
// The spec's operationIds are PascalCase or snake_case joined tokens, always
// starting with an uppercase letter and always at least 4 characters — whereas
// metadata values we've seen (`type`, `scope`, `id`) are lowercase single words.
func looksLikeOperationID(s string) bool {
	if len(s) < 4 {
		return false
	}
	c := s[0]
	return c >= 'A' && c <= 'Z'
}

// isTodoMarker returns true for placeholder values left in the map during v2
// bring-up (e.g. "TODO:CreateOidcIssuer"). These are not real claims and are
// reported as unmapped (but with a clearer explanation) by the coverage gate.
func isTodoMarker(s string) bool {
	return strings.HasPrefix(s, "TODO:") || s == "TODO"
}

// ClaimKind categorizes how an operation was claimed in the map.
type ClaimKind string

const (
	ClaimResource  ClaimKind = "resource"
	ClaimFunction  ClaimKind = "function"
	ClaimMethod    ClaimKind = "method"
	ClaimExclusion ClaimKind = "exclusion"
)

// OperationClaim links an operationId to the map entry that claims it.
type OperationClaim struct {
	OperationID string
	ClaimedBy   string
	Kind        ClaimKind
	Reason      string // only for exclusions
}
