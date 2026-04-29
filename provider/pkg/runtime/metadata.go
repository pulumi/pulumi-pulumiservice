// Copyright 2016-2026, Pulumi Corporation.
//
// metadata.go — runtime metadata types for the v2 Pulumi Service Provider.
//
// These types mirror the shape produced by the generator (provider/pkg/gen)
// and consumed by the generic CRUD dispatcher (dispatch.go). One embedded
// metadata.json drives every resource — zero per-resource Go code.
// If a new Pulumi Cloud API pattern can't be expressed in this shape, the
// right move is to extend the metadata schema here, not to reintroduce a
// hand-coded escape hatch.
//
// This is intentionally simpler than pulumi-azure-native's AzureAPIResource:
// we have a single API version and no Azure-specific long-running-operation
// or x-ms-* machinery. When a pattern surfaces that can't be expressed here,
// we extend these types (not add an overlay file).

package runtime

// CloudAPIMetadata is the root of the runtime metadata. Emitted as
// metadata.json at generator time and embedded into the provider binary.
type CloudAPIMetadata struct {
	// Types is a registry of complex object types referenced by resources
	// and operations. Keyed by Pulumi type token (e.g. "pulumiservice:orgs/agents:AgentPoolConfig").
	Types map[string]CloudAPIType `json:"types,omitempty"`

	// Resources is the full set of generated resources, keyed by Pulumi resource
	// token (e.g. "pulumiservice:orgs/agents:AgentPool").
	Resources map[string]CloudAPIResource `json:"resources,omitempty"`

	// Functions (Pulumi "data sources") — read-only invokes, keyed by token
	// (e.g. "pulumiservice:orgs/agents:listAgentPools").
	Functions map[string]CloudAPIFunction `json:"functions,omitempty"`

	// Methods are non-CRUD actions surfaced on a resource (e.g. agentPool.rotateToken()).
	// Keyed by "<ResourceToken>.<methodName>".
	Methods map[string]CloudAPIMethod `json:"methods,omitempty"`
}

// CloudAPIResource describes one generated resource: its CRUD operations,
// how to compute its ID, which properties are force-new, auto-naming rules,
// and per-property metadata (secrets, defaults, etc.).
type CloudAPIResource struct {
	// Token is the Pulumi resource token.
	Token string `json:"token"`

	// Module is the Pulumi sub-module (e.g. "orgs/agents"). Carried here for
	// introspection; derivable from Token but saves string parsing at runtime.
	Module string `json:"module,omitempty"`

	// CRUD operations. Read may be nil if Read is served via ReadVia.
	Create *CloudAPIOperation `json:"create,omitempty"`
	Read   *CloudAPIOperation `json:"read,omitempty"`
	Update *CloudAPIOperation `json:"update,omitempty"`
	Delete *CloudAPIOperation `json:"delete,omitempty"`

	// PostCreate, if set, runs once after Create succeeds. Used for
	// resources whose bring-up requires two API calls (e.g., ESC
	// Environment: POST creates the empty env, then PATCH with raw YAML
	// sets the content). Dispatched with the update-style input handling.
	PostCreate *CloudAPIOperation `json:"postCreate,omitempty"`

	// ReadVia is an optional fallback for resources whose spec has no dedicated
	// single-resource GET (the dispatch lists and filters client-side).
	ReadVia *CloudAPIReadVia `json:"readVia,omitempty"`

	// PolymorphicScopes supports resources whose create/read/update/delete
	// endpoints vary by a discriminator (e.g. Webhook: org/stack/esc scopes).
	// Inputs are matched against the discriminator and the matching scope's
	// operations are used. Mutually exclusive with the top-level Create/Read/Update/Delete.
	PolymorphicScopes *PolymorphicScopes `json:"polymorphicScopes,omitempty"`

	// ID defines how the Pulumi resource ID is composed from inputs + response.
	// A simple RFC 6570 template for most resources; polymorphic for those
	// whose ID shape varies by scope.
	ID *CloudAPIID `json:"id,omitempty"`

	// ForceNew lists properties that require replacement when changed.
	ForceNew []string `json:"forceNew,omitempty"`

	// Properties is the per-property metadata table, keyed by Pulumi SDK
	// property name (e.g. "organizationName"). It carries the source/kind
	// annotations, secret/write-only flags, defaults, and rename bindings
	// from resource-map.yaml.
	Properties map[string]CloudAPIProperty `json:"properties,omitempty"`

	// RequiredContainers are nested empty object stubs the API demands even
	// when no properties are set within them (e.g. `properties: {}`). Each
	// entry is a path (top-down) from the request body root.
	RequiredContainers [][]string `json:"requiredContainers,omitempty"`

	// AutoName configures automatic name generation if the user doesn't
	// supply one. Nil means no autoname.
	AutoName *AutoNameConfig `json:"autoName,omitempty"`

	// Checks are declarative Check-phase validations — replaces the ad-hoc
	// Go code a hand-written provider would put in Check().
	Checks []CheckRule `json:"checks,omitempty"`

	// SortProperties lists properties whose list values should be sorted on
	// Read for determinism (ex-custom-Go sort logic).
	SortProperties []string `json:"sortProperties,omitempty"`

	// Discriminator names the input property whose value selects a polymorphic
	// operation set or body schema (e.g. Team.type: "pulumi" | "github").
	Discriminator string `json:"discriminator,omitempty"`

	// AsyncStyle hints at long-running operations. "none" or "" means synchronous.
	// Other values will be added as we encounter async endpoints in the spec.
	AsyncStyle *CloudAPIAsyncStyle `json:"asyncStyle,omitempty"`
}

// CloudAPIOperation describes a single HTTP call derived from an OpenAPI operation.
type CloudAPIOperation struct {
	// OperationID is the source operationId in the OpenAPI spec. Carried for
	// traceability (coverage reports, diagnostics) — NOT used to dispatch.
	OperationID string `json:"operationId"`

	// Method is the uppercase HTTP verb: GET, POST, PUT, PATCH, DELETE.
	Method string `json:"method"`

	// PathTemplate is the OpenAPI path with `{param}` placeholders.
	// Dispatch uses RFC 6570 expansion to substitute values at runtime.
	PathTemplate string `json:"path"`

	// Parameters is the ordered list of path/query parameters.
	// Request body properties are described by the Properties map on the
	// enclosing CloudAPIResource (not duplicated here).
	Parameters []CloudAPIParameter `json:"parameters,omitempty"`

	// BodyTypeRef is the Pulumi type reference to the request body schema
	// (e.g. "#/types/pulumiservice:orgs/agents:AgentPoolRequestBody").
	// Nil for operations without a body.
	BodyTypeRef string `json:"bodyTypeRef,omitempty"`

	// ResponseTypeRef is the Pulumi type reference for the response body.
	ResponseTypeRef string `json:"responseTypeRef,omitempty"`

	// BodyOverride, if non-nil, replaces the input-derived request body for
	// this operation with a fixed JSON object. Used for "delete via update
	// with tombstone value" patterns (e.g., TeamStackPermission's Delete is
	// PATCH permission:0 via the same endpoint as Update).
	BodyOverride map[string]interface{} `json:"bodyOverride,omitempty"`

	// IterateOver, when set on a delete op, names a map/object property in
	// the resource's current state whose KEYS are fired as individual
	// delete calls — one per key, substituted into PathTemplate via the
	// placeholder named by IterateKeyParam. Used for batch resources like
	// stack Tags where the single logical Pulumi resource owns N underlying
	// per-tag DELETEs.
	IterateOver string `json:"iterateOver,omitempty"`
	// IterateKeyParam is the path-template placeholder the iterated key
	// substitutes into (defaults to "tagName" for the stack-tag case; any
	// other iterate-based resource must set this explicitly).
	IterateKeyParam string `json:"iterateKeyParam,omitempty"`

	// RawBodyFrom names a property whose value is sent as the raw HTTP
	// request body (instead of the default JSON-encoded input split). Used
	// for ESC Environment's YAML PATCHes. When set, ContentType should be
	// set too.
	RawBodyFrom string `json:"rawBodyFrom,omitempty"`
	// RawBodyTo names a property that receives the raw HTTP response body
	// as a string. Used when the response is non-JSON (e.g., ESC
	// Environment GET returns application/x-yaml).
	RawBodyTo string `json:"rawBodyTo,omitempty"`
	// ContentType overrides the request's Content-Type header (default
	// application/json). Pairs with RawBodyFrom.
	ContentType string `json:"contentType,omitempty"`

	// BodyAs names a property whose JSON value becomes the entire
	// request body (instead of the default {prop1: val1, prop2: ...}
	// JSON object). Used for endpoints where the body is naturally
	// a top-level map (e.g. UpdateStackTags expects `{tagName: value}`
	// directly, not `{tags: {tagName: value}}`). Distinct from
	// RawBodyFrom — RawBodyFrom is for non-JSON bodies (YAML, etc.);
	// BodyAs still produces JSON.
	BodyAs string `json:"bodyAs,omitempty"`
}

// CloudAPIReadVia configures a Read that piggybacks on another operation
// instead of a dedicated single-resource GET. Two modes:
//
//   - **List + filter** (OperationID + FilterBy): call a list endpoint, pick
//     the one item whose FilterBy field matches the resource's ID. Used when
//     the spec has no per-item GET (e.g. OrgAccessToken).
//
//   - **Parent field extraction** (OperationID + ExtractField [+ KeyBy]): call
//     a parent-resource GET, pluck a nested field out of the response, and
//     (optionally) key into it by a property value. Used when a logical
//     resource is stored as a field on its parent (e.g. stack tags live on
//     the stack's GET — `response.tags[tagName]`). When KeyBy is set, the
//     matching entry's value is returned under the KeyBy-named property;
//     when absent, the whole extracted object is attached as the property
//     named by ExtractField.
type CloudAPIReadVia struct {
	OperationID string `json:"operationId"`
	// FilterBy is the field in each list item to match against the resource's ID.
	FilterBy string `json:"filterBy,omitempty"`
	// ExtractField is the top-level response field to pluck for
	// parent-field mode. Mutually exclusive with FilterBy in practice.
	ExtractField string `json:"extractField,omitempty"`
	// KeyBy, when set, names a resource property whose value is used as the
	// key into ExtractField (a map). Without KeyBy, the whole map is
	// returned under the ExtractField name.
	KeyBy string `json:"keyBy,omitempty"`
	// ValueProperty names the resource property to populate with the
	// extracted value when KeyBy is set (single-entry mode). Defaults to
	// "value".
	ValueProperty string `json:"valueProperty,omitempty"`
}

// PolymorphicScopes holds per-scope CRUD operation sets for resources whose
// underlying endpoints vary by discriminator (Webhook: org/stack/esc).
type PolymorphicScopes struct {
	// Discriminator is the input property that selects the scope.
	Discriminator string `json:"discriminator"`
	// Scopes is discriminator-value → CRUD operation set.
	Scopes map[string]CloudAPIResourceOps `json:"scopes"`
}

// CloudAPIResourceOps is the CRUD operation bundle for one scope of a
// polymorphic resource.
type CloudAPIResourceOps struct {
	Create *CloudAPIOperation `json:"create,omitempty"`
	Read   *CloudAPIOperation `json:"read,omitempty"`
	Update *CloudAPIOperation `json:"update,omitempty"`
	Delete *CloudAPIOperation `json:"delete,omitempty"`
}

// CloudAPIID describes how the Pulumi resource ID is composed.
type CloudAPIID struct {
	// Template is the RFC 6570 URI template for the single-ID case.
	// Nil for polymorphic IDs; use Templates instead.
	Template string `json:"template,omitempty"`

	// Params is the ordered list of input/response property names substituted
	// into Template (or Templates[scope]). Each maps to a path param or a
	// server-assigned field in the response.
	Params []string `json:"params,omitempty"`

	// Templates is the per-scope template table for polymorphic IDs
	// (e.g. Webhook has org/stack/esc shapes). Keyed by discriminator value.
	Templates map[string]string `json:"templates,omitempty"`

	// Case names the discriminator property when Templates is used.
	Case string `json:"case,omitempty"`
}

// AutoNameConfig specifies automatic name generation.
type AutoNameConfig struct {
	// Property is the property whose value is autogenerated if unset.
	Property string `json:"property"`
	// Pattern is an optional template (e.g. "{urn.name}-{random}").
	Pattern string `json:"pattern,omitempty"`
	// Kind is "random", "copy", "uuid", etc. — describes the generation strategy.
	Kind string `json:"kind,omitempty"`
}

// CheckRule is a declarative validation. Executed during the Check gRPC phase.
type CheckRule struct {
	// Exactly one of these fields is set per rule.
	RequireOneOf     []string `json:"requireOneOf,omitempty"`     // exactly one must be set
	RequireAtMostOne []string `json:"requireAtMostOne,omitempty"` // zero or one, never more
	RequireTogether  []string `json:"requireTogether,omitempty"`  // all-or-nothing
	RequireIfSet     string   `json:"requireIfSet,omitempty"`     // when this field is set, Field is required
	RequireIf        string   `json:"requireIf,omitempty"`        // e.g., "type == pulumi"
	Field            string   `json:"field,omitempty"`            // the field that RequireIf / RequireIfSet gates

	// Message is the user-facing error returned when the rule fails.
	Message string `json:"message,omitempty"`
}

// CloudAPIAsyncStyle captures long-running-operation polling behavior.
type CloudAPIAsyncStyle struct {
	Create string `json:"create,omitempty"` // e.g. "location", "operation-status"; "" or "none" = sync
	Update string `json:"update,omitempty"`
	Delete string `json:"delete,omitempty"`
}

// CloudAPIParameter describes a single path, query, or header parameter.
// (Body parameters live in the resource-level Properties map, with source="body".)
type CloudAPIParameter struct {
	Name     string `json:"name"`
	SDKName  string `json:"sdkName,omitempty"` // renamed public name if different
	Location string `json:"location"`          // "path", "query", "header"
	Required bool   `json:"required,omitempty"`
	// Value carries the type info for parameters (string, int, …).
	Value *CloudAPIProperty `json:"value,omitempty"`
}

// CloudAPIProperty is the per-property metadata table entry. Drives
// schema emission AND runtime conversion.
type CloudAPIProperty struct {
	// Type is the JSON schema type: string, integer, boolean, array, object.
	Type string `json:"type,omitempty"`

	// Ref is a reference to a complex type in Types (e.g. "#/types/pulumiservice:orgs/teams:TeamMember").
	Ref string `json:"$ref,omitempty"`

	// Items describes array element type (for Type=="array").
	Items *CloudAPIProperty `json:"items,omitempty"`

	// AdditionalProperties describes map value type (for Type=="object" with
	// open-ended keys, i.e. OpenAPI additionalProperties).
	AdditionalProperties *CloudAPIProperty `json:"additionalProperties,omitempty"`

	// Enum is the allowed value set for constrained string/int properties.
	Enum []interface{} `json:"enum,omitempty"`

	// From is the wire-serialized name in the underlying HTTP API, if it
	// differs from the property's SDK-facing name.
	From string `json:"from,omitempty"`

	// Source is "path" | "query" | "body" | "response". Drives how the value
	// is plumbed into the request and extracted from the response.
	Source string `json:"source,omitempty"`

	// CreateSource, if set, overrides Source specifically for the create verb.
	// Used when a property appears in the body of POST but in the path of
	// subsequent GET/PATCH/DELETE (e.g., Stack's `stackName` is body for
	// CreateStack — POST /api/stacks/{org}/{project} — but path for the
	// resource-scoped verbs).
	CreateSource string `json:"createSource,omitempty"`

	// CreateFrom, if set, overrides From during the create verb. Needed when
	// the wire-level rename differs between Source locations — e.g., Tag's
	// `name` is body field `name` at POST /.../tags but path param
	// `tagName` at PATCH /.../tags/{tagName}.
	CreateFrom string `json:"createFrom,omitempty"`

	// PathName, if set, is the URL path-placeholder name for this
	// property when expanding path templates. Needed when the spec
	// uses one name in the request/response body and a different
	// name in the URL path — e.g., AgentPool's `id` field vs the
	// `{poolId}` path placeholder, or any other case where the
	// generator can't infer the path parameter name from From alone.
	PathName string `json:"pathName,omitempty"`

	// BodyFrom, if set, overrides From specifically for the request
	// body. Needed when the API uses one name in writes (e.g.
	// `newEnabled` for set-this-to semantics) and a different name
	// in reads (`enabled` as the canonical state). Default From
	// continues to drive response decoding.
	BodyFrom string `json:"bodyFrom,omitempty"`

	// Secret wraps the output value in resource.MakeSecret.
	Secret bool `json:"secret,omitempty"`

	// Output-only (never an input).
	Output bool `json:"output,omitempty"`

	// WriteOnly — user provides on create/update; API never echoes it back.
	// The diff engine doesn't drift-alert when WriteOnly properties differ
	// from their input on Read (since there's nothing to compare against).
	WriteOnly bool `json:"writeOnly,omitempty"`

	// DiffMode controls how the diff engine compares input vs observed state.
	// Values: "" (default), "ciphertext" (three-way plaintext/ciphertext/import).
	DiffMode string `json:"diffMode,omitempty"`

	// Default is the default value injected on create/update if the user
	// doesn't supply one (and on Read if the API returns the property missing).
	Default interface{} `json:"default,omitempty"`

	// DefaultFromField injects the value of another property if unset.
	// e.g. Team.displayName defaults to Team.name.
	DefaultFromField string `json:"defaultFromField,omitempty"`

	// SortOnRead — if true, this list's elements are sorted on Read for
	// deterministic state. Prevents API-ordering-churn from surfacing as diffs.
	SortOnRead bool `json:"sortOnRead,omitempty"`

	// Aliases are prior SDK names for this property. Drift-free migration
	// when a property is renamed between v1 and v2.
	Aliases []string `json:"aliases,omitempty"`

	// Doc overrides the spec description in the generated SDK. Omit to
	// inherit the spec description.
	Doc string `json:"doc,omitempty"`
}

// CloudAPIType is a complex object type — typically one of the OpenAPI
// components.schemas entries, or a synthesized type for a request/response body.
type CloudAPIType struct {
	Properties map[string]CloudAPIProperty `json:"properties,omitempty"`
	Required   []string                    `json:"required,omitempty"`
	Doc        string                      `json:"doc,omitempty"`
}

// CloudAPIFunction is a read-only invoke (Pulumi "data source").
type CloudAPIFunction struct {
	Token     string            `json:"token"`
	Module    string            `json:"module,omitempty"`
	Operation CloudAPIOperation `json:"operation"`
}

// CloudAPIMethod is a non-CRUD action surfaced on a resource.
type CloudAPIMethod struct {
	Token        string            `json:"token"` // "<ResourceToken>.<methodName>"
	ResourceRef  string            `json:"resourceRef"`
	Operation    CloudAPIOperation `json:"operation"`
}
