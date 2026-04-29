// Copyright 2016-2026, Pulumi Corporation.
//
// convert.go — translate between Pulumi's resource.PropertyMap and the
// request/response JSON shapes of the Pulumi Cloud API.
//
// Two directions:
//   - InputsToRequest: split a PropertyMap into (pathValues, queryValues,
//     bodyObject) per the resource's property metadata.
//   - ResponseToOutputs: wrap an API response (already unmarshaled as a
//     map[string]interface{}) back into a PropertyMap, applying renames,
//     secrets, sort-on-read, and default backfills.
//
// Kept intentionally simple in the v2.0 bring-up; scalar types and shallow
// objects are handled inline. Deeper nested schema handling follows as we
// expand coverage.

package runtime

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

// Split decomposes resource inputs into the three request destinations
// (path, query, body) using the per-property `Source` annotation.
type Split struct {
	Path  map[string]string
	Query map[string]string
	Body  map[string]interface{}
	// Scope is the value of the discriminator field, if any. Used by
	// polymorphic resources (e.g. Webhook.scope) to pick which operation set
	// to dispatch to.
	Scope string
}

// InputsToRequest partitions a PropertyMap according to the resource's
// property metadata. Unknown / computed values are left out — Pulumi is
// expected to have already resolved them before CRUD dispatch.
//
// `forCreate` tells the splitter to honor each property's CreateSource
// override (typically used for identity fields that sit in the body during
// POST but in the path afterwards — e.g., Stack's stackName).
func InputsToRequest(props map[string]CloudAPIProperty, inputs resource.PropertyMap, discriminator string, forCreate bool) (*Split, error) {
	out := &Split{
		Path:  map[string]string{},
		Query: map[string]string{},
		Body:  map[string]interface{}{},
	}
	for name, meta := range props {
		if meta.Output {
			continue // output-only properties are never inputs
		}
		if meta.Source == "rawBody" {
			// rawBody properties are never plumbed into the JSON split; the
			// dispatcher pulls them out separately via the operation's
			// RawBodyFrom hook.
			continue
		}
		effective := meta
		if forCreate {
			if meta.CreateSource != "" {
				effective.Source = meta.CreateSource
			}
			if meta.CreateFrom != "" {
				effective.From = meta.CreateFrom
			}
		}
		key := resource.PropertyKey(name)
		v, ok := inputs[key]
		if !ok || v.IsNull() {
			// Backfill defaults where configured.
			if meta.Default != nil {
				out.writeTo(effective, name, meta.Default)
			} else if meta.DefaultFromField != "" {
				if src, ok := inputs[resource.PropertyKey(meta.DefaultFromField)]; ok && !src.IsNull() {
					out.writeTo(effective, name, src.Mappable())
				}
			}
			continue
		}
		// Unwrap secrets — the wire body carries the plaintext; the secret
		// marker is a Pulumi-side concept. We rewrap on the response side.
		mv := v
		if mv.IsSecret() {
			mv = mv.SecretValue().Element
		}
		if mv.IsComputed() {
			return nil, fmt.Errorf("property %q is still computed at dispatch time; Pulumi should have resolved it", name)
		}
		out.writeTo(effective, name, mv.Mappable())
	}
	// Discriminator is itself a body/path value — read it back so the
	// caller knows which scope to select.
	if discriminator != "" {
		if v, ok := inputs[resource.PropertyKey(discriminator)]; ok && v.IsString() {
			out.Scope = v.StringValue()
		}
	}
	return out, nil
}


// writeTo places a value into the appropriate destination based on the
// property metadata's Source field.
func (s *Split) writeTo(meta CloudAPIProperty, name string, value interface{}) {
	wireName := meta.From
	if wireName == "" {
		wireName = name
	}
	switch meta.Source {
	case "path":
		s.Path[wireName] = fmt.Sprintf("%v", value)
	case "pathAndBody":
		// Some endpoints — Webhook in particular — require the
		// identity fields in BOTH the URL path and the request body
		// (the API validates they agree). The path-side wire name
		// comes from From; the body-side wire name comes from
		// BodyFrom (or the SDK name when unset).
		s.Path[wireName] = fmt.Sprintf("%v", value)
		bodyName := meta.BodyFrom
		if bodyName == "" {
			bodyName = name
		}
		s.Body[bodyName] = value
	case "query":
		s.Query[wireName] = fmt.Sprintf("%v", value)
	case "body", "":
		// BodyFrom overrides the body-side wire name when the API
		// uses different names for writes vs reads (LogExport:
		// `newEnabled` body / `enabled` response).
		if meta.BodyFrom != "" {
			s.Body[meta.BodyFrom] = value
			return
		}
		// Default to body when unspecified. Body is a nested object; we
		// populate top-level keys only for now — nested flattening is
		// handled once the schema emitter lands richer type info.
		s.Body[wireName] = value
	case "response":
		// Response-only properties are outputs; should have been filtered
		// by the meta.Output check above. Defensive no-op.
	}
}

// ResponseToOutputs wraps an API response body (decoded from JSON into a
// map[string]interface{}) back into a Pulumi PropertyMap, applying the
// per-property output metadata: renames, secret wrapping, sort-on-read,
// and backfill of write-only inputs the API doesn't echo back.
func ResponseToOutputs(props map[string]CloudAPIProperty, response map[string]interface{}, originalInputs resource.PropertyMap) resource.PropertyMap {
	out := resource.PropertyMap{}
	for name, meta := range props {
		wireName := meta.From
		if wireName == "" {
			wireName = name
		}

		var raw interface{}
		var present bool

		// Response-sourced properties and default body/path/query properties
		// come from the API response.
		if v, ok := response[wireName]; ok {
			raw = v
			present = true
		}

		// Secret fields with `diffMode: ciphertext`: the API redacts them
		// on read (returns stars/empty/cipher). Prefer the user's original
		// plaintext input so state stays stable across refresh and the diff
		// engine doesn't see spurious churn. This is the MVP three-way-diff:
		// trust inputs, not response.
		if meta.Secret && meta.DiffMode == "ciphertext" {
			if v, ok := originalInputs[resource.PropertyKey(name)]; ok && !v.IsNull() {
				out[resource.PropertyKey(name)] = v
				continue
			}
		}

		// WriteOnly inputs aren't echoed; carry the input value forward so
		// the state persists the user's intent. Also covers identity
		// properties that aren't always echoed by the response (path,
		// query, pathAndBody).
		if !present && (meta.WriteOnly || meta.Source == "path" || meta.Source == "query" || meta.Source == "pathAndBody") {
			if v, ok := originalInputs[resource.PropertyKey(name)]; ok && !v.IsNull() {
				out[resource.PropertyKey(name)] = v
				continue
			}
		}

		// Write-once secret outputs: the API returns the value on Create
		// but not on subsequent reads (e.g. {Org,Team,AccessToken}.value
		// issued by create/list endpoints that omit the token secret from
		// list payloads). Without this fallback the secret silently
		// disappears from state on the first refresh. `originalInputs` here
		// is the prior full state passed to Read, so the value is present
		// whenever it was ever written.
		if !present && meta.Secret && meta.Output {
			if v, ok := originalInputs[resource.PropertyKey(name)]; ok && !v.IsNull() {
				out[resource.PropertyKey(name)] = v
				continue
			}
		}
		if !present {
			continue
		}

		propVal := resource.NewPropertyValue(raw)
		if meta.SortOnRead && propVal.IsArray() {
			propVal = sortPropertyArray(propVal)
		}
		if meta.Secret {
			propVal = resource.MakeSecret(propVal)
		}
		out[resource.PropertyKey(name)] = propVal
	}
	return out
}

// HasSortOnRead reports whether any property in the map is marked
// sortOnRead. Cheap probe for the provider's Check RPC to decide whether
// it needs to unmarshal inputs at all.
func HasSortOnRead(props map[string]CloudAPIProperty) bool {
	for _, meta := range props {
		if meta.SortOnRead {
			return true
		}
	}
	return false
}

// CanonicalizeSortedInputs sorts any input property marked sortOnRead,
// using the same comparator ResponseToOutputs uses for the Read side.
// Called from the provider's Check RPC so that user input (however they
// wrote the array) and server response (however it returns the array)
// end up in the same canonical order in state. Without this, a property
// like OidcIssuer.policies that the server reorders produces a perpetual
// diff — the bug originally fixed in pulumi-pulumiservice#542.
func CanonicalizeSortedInputs(props map[string]CloudAPIProperty, inputs resource.PropertyMap) resource.PropertyMap {
	out := make(resource.PropertyMap, len(inputs))
	for k, v := range inputs {
		meta, known := props[string(k)]
		if !known || !meta.SortOnRead {
			out[k] = v
			continue
		}
		// Unwrap secret if needed, sort, re-wrap.
		inner := v
		wasSecret := false
		if inner.IsSecret() {
			wasSecret = true
			inner = inner.SecretValue().Element
		}
		if inner.IsArray() {
			inner = sortPropertyArray(inner)
		}
		if wasSecret {
			inner = resource.MakeSecret(inner)
		}
		out[k] = inner
	}
	return out
}

// sortPropertyArray returns a new array property whose elements are sorted
// by their canonical JSON representation. Using json.Marshal (which
// sorts map keys alphabetically) — rather than fmt.Sprintf("%v", …) on
// the Mappable() form — keeps sort keys stable across runs for complex
// object elements; map iteration order would otherwise produce different
// sort keys for the same logical object and re-introduce the drift the
// sort is meant to prevent.
func sortPropertyArray(arr resource.PropertyValue) resource.PropertyValue {
	items := arr.ArrayValue()
	type indexed struct {
		key string
		val resource.PropertyValue
	}
	x := make([]indexed, 0, len(items))
	for _, it := range items {
		keyBytes, err := json.Marshal(it.Mappable())
		var key string
		if err != nil {
			// Marshal should never fail on a Mappable()'d value, but if it
			// ever does, fall back to %v; the sort is approximate rather
			// than unstable.
			key = fmt.Sprintf("%v", it.Mappable())
		} else {
			key = string(keyBytes)
		}
		x = append(x, indexed{key: key, val: it})
	}
	sort.Slice(x, func(i, j int) bool { return x[i].key < x[j].key })
	sorted := make([]resource.PropertyValue, 0, len(items))
	for _, it := range x {
		sorted = append(sorted, it.val)
	}
	return resource.NewArrayProperty(sorted)
}

// InferScopeFromInputs picks the polymorphic scope whose
// user-settable identity properties are all satisfied by the given
// inputs. Used for resources like Webhook where scope isn't a single
// discriminator field but is implied by which identity properties
// the user sets: `organizationName` alone → org scope, plus
// `stackName` → stack scope, plus `environmentName` + `projectName`
// → esc scope.
//
// Server-assigned (output-only) placeholders in the ID template are
// ignored — they're not present at dispatch time. Picks the most
// specific scope (most user-input placeholders) when multiple match.
// Returns "" when no scope's user-input placeholders are all
// present.
func InferScopeFromInputs(res *CloudAPIResource, inputs resource.PropertyMap) string {
	if res.PolymorphicScopes == nil || res.ID == nil || len(res.ID.Templates) == 0 {
		return ""
	}
	bestScope := ""
	bestSpecificity := -1
	for scope, tmpl := range res.ID.Templates {
		placeholders := extractPathPlaceholders(tmpl)
		matched := true
		userPresent := 0
		for _, ph := range placeholders {
			sdkName := sdkNameForPlaceholder(res.Properties, ph)
			meta := res.Properties[sdkName]
			// Output-only properties (server-assigned IDs/names) are
			// not in inputs at dispatch time. They don't contribute
			// to scope selection.
			if meta.Output {
				continue
			}
			v, ok := inputs[resource.PropertyKey(sdkName)]
			if !ok || v.IsNull() {
				matched = false
				break
			}
			userPresent++
		}
		if matched && userPresent > bestSpecificity {
			bestScope = scope
			bestSpecificity = userPresent
		}
	}
	return bestScope
}

// extractPathPlaceholders pulls the {name} segments from a path or
// ID template. Returns the inner names without braces.
func extractPathPlaceholders(template string) []string {
	var out []string
	for i := 0; i < len(template); i++ {
		if template[i] != '{' {
			continue
		}
		j := i + 1
		for j < len(template) && template[j] != '}' {
			j++
		}
		if j >= len(template) {
			break
		}
		out = append(out, template[i+1:j])
		i = j
	}
	return out
}

// sdkNameForPlaceholder maps a path/ID-template placeholder back to
// the resource's SDK property name. Tries direct match first (the
// common case where the placeholder is the SDK name), then walks the
// property map looking for a `From` or `PathName` match.
func sdkNameForPlaceholder(props map[string]CloudAPIProperty, placeholder string) string {
	if _, ok := props[placeholder]; ok {
		return placeholder
	}
	for name, p := range props {
		if p.From == placeholder || p.PathName == placeholder {
			return name
		}
	}
	return placeholder
}

// SelectOperations resolves which CRUD operation set applies based on the
// discriminator (for polymorphic resources) or the top-level operations
// (for simple ones).
func SelectOperations(res *CloudAPIResource, scope string) (*CloudAPIResourceOps, error) {
	if res.PolymorphicScopes == nil {
		return &CloudAPIResourceOps{
			Create: res.Create,
			Read:   res.Read,
			Update: res.Update,
			Delete: res.Delete,
		}, nil
	}
	ops, ok := res.PolymorphicScopes.Scopes[scope]
	if !ok {
		return nil, fmt.Errorf("no operation set for scope %q; valid scopes: %v",
			scope, polymorphicKeys(res.PolymorphicScopes.Scopes))
	}
	return &ops, nil
}

func polymorphicKeys(m map[string]CloudAPIResourceOps) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
