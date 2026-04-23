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
	case "query":
		s.Query[wireName] = fmt.Sprintf("%v", value)
	case "body", "":
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
		// the state persists the user's intent.
		if !present && (meta.WriteOnly || meta.Source == "path" || meta.Source == "query") {
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

// sortPropertyArray returns a new array property whose elements are sorted
// by their string representation — enough determinism for simple member/tag
// lists. Complex objects get sorted by their JSON-ish string form; this is
// fine for drift-prevention even if not strictly "semantic" sorting.
func sortPropertyArray(arr resource.PropertyValue) resource.PropertyValue {
	items := arr.ArrayValue()
	type indexed struct {
		key string
		val resource.PropertyValue
	}
	x := make([]indexed, 0, len(items))
	for _, it := range items {
		x = append(x, indexed{key: fmt.Sprintf("%v", it.Mappable()), val: it})
	}
	sort.Slice(x, func(i, j int) bool { return x[i].key < x[j].key })
	sorted := make([]resource.PropertyValue, 0, len(items))
	for _, it := range x {
		sorted = append(sorted, it.val)
	}
	return resource.NewArrayProperty(sorted)
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
