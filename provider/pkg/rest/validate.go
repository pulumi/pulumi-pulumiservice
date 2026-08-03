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
	"fmt"
	"slices"
	"strings"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi/sdk/v3/go/property"
)

// ValidateInputs walks the create-op request body schema over the checked
// inputs and reports mistakes the schema can prove: unknown discriminator
// tags, out-of-subset variants, unknown object fields, and object/scalar
// shape mismatches. It stays lenient everywhere the service is lenient:
// required fields and enum membership are not enforced, and unknown
// (computed) values skip their subtree.
func ValidateInputs(
	spec *Spec, typeMeta map[string]TypeMeta, op *Operation, meta ResourceMeta, inputs property.Map,
) []p.CheckFailure {
	if op == nil || op.RequestRef == "" {
		return nil
	}
	bodyProps, _, err := flattenObjectSchema(spec, op.RequestRef)
	if err != nil {
		return nil
	}
	v := &inputValidator{spec: spec, typeMeta: typeMeta}
	inputs.AllStable(func(name string, val property.Value) bool {
		if node, ok := bodyProps[wireSideName(name, meta.Renames)].(map[string]any); ok {
			v.walk(name, val, node)
		}
		return true
	})
	return v.failures
}

type inputValidator struct {
	spec     *Spec
	typeMeta map[string]TypeMeta
	failures []p.CheckFailure
}

func (v *inputValidator) failf(path, format string, args ...any) {
	v.failures = append(v.failures, p.CheckFailure{Property: path, Reason: fmt.Sprintf(format, args...)})
}

// walk validates one value against one OpenAPI schema node.
func (v *inputValidator) walk(path string, val property.Value, node map[string]any) {
	if val.IsNull() || val.IsComputed() {
		return
	}
	if ref, ok := node[refKey].(string); ok {
		v.walkRef(path, val, refSchemaName(ref))
		return
	}
	if _, ok := node[allOfKey]; ok {
		// Inline compositions degrade to Any in the schema; nothing provable.
		return
	}
	switch t, _ := node["type"].(string); t {
	case typeString, typeInteger, typeNumber, typeBoolean:
		if val.IsMap() || val.IsArray() {
			v.failf(path, "expected a %s, got %s", t, valueKind(val))
		}
	case typeArray:
		if !val.IsArray() {
			if val.IsMap() {
				v.failf(path, "expected an array, got an object")
			}
			return
		}
		items, ok := node["items"].(map[string]any)
		if !ok {
			return
		}
		val.AsArray().All(func(i int, elem property.Value) bool {
			v.walk(fmt.Sprintf("%s[%d]", path, i), elem, items)
			return true
		})
	case typeObject, "":
		ap, ok := node["additionalProperties"].(map[string]any)
		if !ok || !val.IsMap() {
			// Anonymous free-form object.
			return
		}
		val.AsMap().AllStable(func(k string, elem property.Value) bool {
			v.walk(path+"."+k, elem, ap)
			return true
		})
	}
}

// walkRef validates a value against a named component schema, dispatching
// unions through the same resolution rules the schema generator uses.
func (v *inputValidator) walkRef(path string, val property.Value, name string) {
	tm := v.typeMeta[name]
	if tm.Any {
		return
	}
	target, ok := v.spec.ResolveSchema(schemaRefPrefix + name)
	if !ok {
		return
	}
	if tm.ScalarShorthand != "" && !val.IsMap() {
		// The shorthand scalar form.
		return
	}
	if tagProp, mapping := discriminatorOf(target); tagProp != "" && len(mapping) > 0 {
		v.walkUnion(path, val, tagProp, mapping)
		return
	}
	if tagProp, tag, ok := variantTagFor(v.spec, name); ok {
		v.walkVariant(path, val, tagProp, tag, name)
		return
	}
	if base, tagProp, mapping := markerUnionBase(v.spec, name, target); base != "" && len(mapping) > 0 {
		if subset := markerSubset(v.spec, name, mapping); len(subset) > 0 {
			v.walkUnion(path, val, tagProp, subset)
			return
		}
	}
	if isObjectSchema(target) {
		v.walkObject(path, val, name, "")
		return
	}
	if val.IsMap() {
		v.failf(path, "expected a scalar %s value, got an object", name)
	}
}

// walkUnion validates a discriminated position: the tag must be present and
// name one of the allowed variants, then the value is checked as that
// variant.
func (v *inputValidator) walkUnion(path string, val property.Value, tagProp string, mapping map[string]string) {
	if !val.IsMap() {
		v.failf(path, "expected an object with a %q discriminator, got %s", tagProp, valueKind(val))
		return
	}
	allowed := sortedStringKeys(mapping)
	tv, ok := val.AsMap().GetOk(tagProp)
	if !ok {
		v.failf(path+"."+tagProp, "missing %q; expected one of: %s", tagProp, strings.Join(allowed, ", "))
		return
	}
	if tv.IsComputed() {
		// Tag unknown until apply; nothing more to prove at preview.
		return
	}
	if !tv.IsString() {
		v.failf(path+"."+tagProp, "%q must be a string; expected one of: %s", tagProp, strings.Join(allowed, ", "))
		return
	}
	tag := tv.AsString()
	ref, ok := mapping[tag]
	if !ok {
		msg := fmt.Sprintf("unknown %s %q; expected one of: %s", tagProp, tag, strings.Join(allowed, ", "))
		if s := nearestName(tag, allowed); s != "" {
			msg += fmt.Sprintf(" (did you mean %q?)", s)
		}
		v.failf(path+"."+tagProp, "%s", msg)
		return
	}
	v.walkObject(path, val, refSchemaName(ref), tagProp)
}

// walkVariant validates a definite-variant position: the tag, when present,
// must be the variant's own.
func (v *inputValidator) walkVariant(path string, val property.Value, tagProp, tag, name string) {
	if !val.IsMap() {
		v.failf(path, "expected an object (%s), got %s", name, valueKind(val))
		return
	}
	tv, ok := val.AsMap().GetOk(tagProp)
	switch {
	case !ok:
		v.failf(path+"."+tagProp, "missing %q; expected %q", tagProp, tag)
	case tv.IsComputed():
	case !tv.IsString() || tv.AsString() != tag:
		v.failf(path+"."+tagProp, "%q must be %q here", tagProp, tag)
	}
	v.walkObject(path, val, name, tagProp)
}

// walkObject checks a value's fields against a named object schema: unknown
// fields fail (the service drops them silently), known fields recurse.
func (v *inputValidator) walkObject(path string, val property.Value, name, tagProp string) {
	if v.typeMeta[name].Any {
		return
	}
	if !val.IsMap() {
		v.failf(path, "expected an object (%s), got %s", name, valueKind(val))
		return
	}
	props, _, err := flattenObjectSchema(v.spec, schemaRefPrefix+name)
	if err != nil {
		return
	}
	known := make([]string, 0, len(props))
	for k := range props {
		known = append(known, k)
	}
	slices.Sort(known)
	val.AsMap().AllStable(func(k string, elem property.Value) bool {
		node, ok := props[k].(map[string]any)
		if !ok {
			msg := fmt.Sprintf("unknown field %q on %s (the service drops unknown fields silently)", k, name)
			if s := nearestName(k, known); s != "" {
				msg += fmt.Sprintf("; did you mean %q?", s)
			}
			v.failf(path+"."+k, "%s", msg)
			return true
		}
		if k != tagProp {
			v.walk(path+"."+k, elem, node)
		}
		return true
	})
}

func valueKind(v property.Value) string {
	switch {
	case v.IsArray():
		return "an array"
	case v.IsMap():
		return "an object"
	case v.IsString():
		return "a string"
	case v.IsBool():
		return "a boolean"
	case v.IsNumber():
		return "a number"
	default:
		return "an unsupported value"
	}
}

func sortedStringKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

// nearestName returns the candidate within edit distance 3 of s, preferring
// the closest; empty when none is close enough.
func nearestName(s string, candidates []string) string {
	best, bestDist := "", 4
	for _, c := range candidates {
		if d := editDistance(strings.ToLower(s), strings.ToLower(c), bestDist); d < bestDist {
			best, bestDist = c, d
		}
	}
	return best
}

// editDistance is the Levenshtein distance between a and b, short-circuiting
// to limit when the distance is provably >= limit.
func editDistance(a, b string, limit int) int {
	if d := len(a) - len(b); d >= limit || -d >= limit {
		return limit
	}
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		rowMin := i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, min(cur[j-1]+1, prev[j-1]+cost))
			rowMin = min(rowMin, cur[j])
		}
		if rowMin >= limit {
			return limit
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}
