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

	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
)

const schemaRefPrefix = "#/components/schemas/"

const refKey = "$ref"

const (
	anyRef      = "pulumi.json#/Any"
	typeObject  = "object"
	typeArray   = "array"
	typeString  = "string"
	typeInteger = "integer"
	typeNumber  = "number"
	typeBoolean = "boolean"
)

// typeRegistry emits named Pulumi types for component schemas on demand and
// renders discriminated bases as inline oneOf unions over const-tagged
// variants (the azure-native shape). Emission is memoized and cycle-safe:
// a schema referenced while it is being emitted just yields its token.
type typeRegistry struct {
	spec           *Spec
	pkg            string
	typeMeta       map[string]TypeMeta
	modules        map[string]map[string]bool // schema name -> referencing modules
	tokens         map[string]string          // schema name -> resolved token
	types          map[string]schema.ComplexTypeSpec
	state          map[string]int // 0 unseen, 1 emitting, 2 done
	resourceTokens map[string]bool
	errs           []string
}

const (
	stateEmitting = 1
	stateDone     = 2
)

func newTypeRegistry(spec *Spec, metadata *Metadata, pkg string) *typeRegistry {
	r := &typeRegistry{
		spec:           spec,
		pkg:            pkg,
		typeMeta:       metadata.Types,
		modules:        map[string]map[string]bool{},
		tokens:         map[string]string{},
		types:          map[string]schema.ComplexTypeSpec{},
		state:          map[string]int{},
		resourceTokens: map[string]bool{},
	}
	for key, rm := range metadata.Resources {
		token := key
		if rm.Token != "" {
			token = rm.Token
		}
		r.resourceTokens[token] = true
		module := moduleOfToken(token)
		for _, opID := range resourceOpIDs(rm) {
			op, ok := spec.Op(opID)
			if !ok {
				continue
			}
			for _, ref := range []string{op.RequestRef, op.ResponseRef} {
				if ref != "" {
					r.markReachable(refSchemaName(ref), module, map[string]bool{})
				}
			}
		}
	}
	return r
}

func resourceOpIDs(rm ResourceMeta) []string {
	ids := []string{rm.Operations.Create, rm.Operations.Read, rm.Operations.Update, rm.Operations.Delete}
	if rm.Attachment != nil {
		ids = append(ids, rm.Attachment.MutationOp, rm.Attachment.ReadOp)
	}
	out := ids[:0]
	for _, id := range ids {
		if id != "" {
			out = append(out, id)
		}
	}
	return out
}

func moduleOfToken(token string) string {
	parts := strings.Split(token, ":")
	if len(parts) == 3 {
		return parts[1]
	}
	return "api"
}

func refSchemaName(ref string) string {
	return strings.TrimPrefix(ref, schemaRefPrefix)
}

// markReachable records module ownership for name and every schema reachable
// from it, including discriminator variants.
func (r *typeRegistry) markReachable(name, module string, seen map[string]bool) {
	if name == "" || seen[name] {
		return
	}
	seen[name] = true
	if r.modules[name] == nil {
		r.modules[name] = map[string]bool{}
	}
	r.modules[name][module] = true
	node, ok := r.spec.ResolveSchema(schemaRefPrefix + name)
	if !ok {
		return
	}
	var visit func(n map[string]any)
	visit = func(n map[string]any) {
		if ref, ok := n[refKey].(string); ok {
			r.markReachable(refSchemaName(ref), module, seen)
			return
		}
		for _, key := range []string{"allOf", "oneOf", "anyOf"} {
			if list, ok := n[key].([]any); ok {
				for _, m := range list {
					if mm, ok := m.(map[string]any); ok {
						visit(mm)
					}
				}
			}
		}
		if items, ok := n["items"].(map[string]any); ok {
			visit(items)
		}
		if ap, ok := n["additionalProperties"].(map[string]any); ok {
			visit(ap)
		}
		if props, ok := n["properties"].(map[string]any); ok {
			for _, v := range props {
				if vm, ok := v.(map[string]any); ok {
					visit(vm)
				}
			}
		}
		if disc, ok := n["discriminator"].(map[string]any); ok {
			if mapping, ok := disc["mapping"].(map[string]any); ok {
				for _, v := range mapping {
					if s, ok := v.(string); ok {
						r.markReachable(refSchemaName(s), module, seen)
					}
				}
			}
		}
	}
	visit(node)
}

func (r *typeRegistry) errorf(format string, args ...any) {
	r.errs = append(r.errs, fmt.Sprintf(format, args...))
}

// moduleFor places a type in its sole referencing module, hoisting types
// shared across modules to the api root so tokens stay stable as resources
// are added.
func (r *typeRegistry) moduleFor(name string) string {
	mods := r.modules[name]
	if len(mods) == 1 {
		for m := range mods {
			return m
		}
	}
	return "api"
}

// tokenFor resolves the Pulumi token for a component schema, suffixing
// "Properties" when the natural token collides with a resource token.
func (r *typeRegistry) tokenFor(name string) string {
	if tok, ok := r.tokens[name]; ok {
		return tok
	}
	tok := fmt.Sprintf("%s:%s:%s", r.pkg, r.moduleFor(name), name)
	if r.resourceTokens[tok] {
		tok += "Properties"
	}
	r.tokens[name] = tok
	return tok
}

func discriminatorOf(node map[string]any) (string, map[string]string) {
	disc, ok := node["discriminator"].(map[string]any)
	if !ok {
		return "", nil
	}
	prop, _ := disc["propertyName"].(string)
	mapping := map[string]string{}
	if mm, ok := disc["mapping"].(map[string]any); ok {
		for tag, ref := range mm {
			if s, ok := ref.(string); ok {
				mapping[tag] = s
			}
		}
	}
	if prop == "" {
		return "", nil
	}
	return prop, mapping
}

// property converts an OpenAPI property node into a Pulumi PropertySpec,
// emitting named types for any $ref it encounters.
func (r *typeRegistry) property(node any) schema.PropertySpec {
	nm, ok := node.(map[string]any)
	if !ok {
		return schema.PropertySpec{TypeSpec: schema.TypeSpec{Ref: anyRef}}
	}
	ts := r.typeSpec(nm)
	desc, _ := nm["description"].(string)
	return schema.PropertySpec{TypeSpec: ts, Description: desc}
}

func anyTypeSpec() schema.TypeSpec {
	return schema.TypeSpec{Ref: anyRef}
}

func (r *typeRegistry) refTo(token string) schema.TypeSpec {
	return schema.TypeSpec{Ref: "#/types/" + token}
}

func (r *typeRegistry) typeSpec(nm map[string]any) schema.TypeSpec {
	if ref, ok := nm[refKey].(string); ok {
		name := refSchemaName(ref)
		target, found := r.spec.ResolveSchema(ref)
		if !found {
			r.errorf("unresolvable $ref %q", ref)
			return anyTypeSpec()
		}
		if tm, ok := r.typeMeta[name]; ok {
			if tm.Any {
				return anyTypeSpec()
			}
			if tm.ScalarShorthand != "" {
				return schema.TypeSpec{OneOf: []schema.TypeSpec{
					{Type: tm.ScalarShorthand},
					r.refTo(r.emitNamed(name)),
				}}
			}
		}
		if tagProp, mapping := discriminatorOf(target); tagProp != "" {
			switch len(mapping) {
			case 0:
				// No variants declared: treat the base as a plain object.
				return r.refTo(r.emitNamed(name))
			case 1:
				// Pulumi oneOf requires >= 2 members: emit the sole variant
				// as the definite type.
				tag := soleKey(mapping)
				return r.refTo(r.emitVariant(refSchemaName(mapping[tag]), tagProp, tag))
			default:
				return r.unionTypeSpec(name, tagProp, mapping)
			}
		}
		if isObjectSchema(target) {
			return r.refTo(r.emitNamed(name))
		}
		return scalarTypeSpec(target)
	}
	if _, ok := nm["oneOf"]; ok {
		return anyTypeSpec()
	}
	if _, ok := nm["anyOf"]; ok {
		return anyTypeSpec()
	}
	if _, ok := nm["allOf"]; ok {
		return anyTypeSpec()
	}
	t, _ := nm["type"].(string)
	switch t {
	case typeString, typeInteger, typeNumber, typeBoolean:
		return schema.TypeSpec{Type: t}
	case typeArray:
		items, _ := nm["items"].(map[string]any)
		var itemTS schema.TypeSpec
		if items != nil {
			itemTS = r.typeSpec(items)
		} else {
			itemTS = anyTypeSpec()
		}
		return schema.TypeSpec{Type: typeArray, Items: &itemTS}
	case typeObject, "":
		if ap, ok := nm["additionalProperties"].(map[string]any); ok {
			elem := r.typeSpec(ap)
			return schema.TypeSpec{Type: typeObject, AdditionalProperties: &elem}
		}
		// Anonymous inline object: stays free-form (phase 1 rule).
		return schema.TypeSpec{Type: typeObject, AdditionalProperties: &schema.TypeSpec{Ref: anyRef}}
	default:
		return anyTypeSpec()
	}
}

func isObjectSchema(node map[string]any) bool {
	if t, ok := node["type"].(string); ok && t != typeObject && t != "" {
		return false
	}
	if _, ok := node["properties"]; ok {
		return true
	}
	if _, ok := node["allOf"]; ok {
		return true
	}
	t, _ := node["type"].(string)
	return t == typeObject
}

// scalarTypeSpec renders a named non-object schema (a string/array alias)
// structurally. Ref-typed array items degrade to Any (rare).
func scalarTypeSpec(node map[string]any) schema.TypeSpec {
	t, _ := node["type"].(string)
	switch t {
	case typeString, typeInteger, typeNumber, typeBoolean:
		return schema.TypeSpec{Type: t}
	case typeArray:
		if items, ok := node["items"].(map[string]any); ok {
			if _, isRef := items[refKey]; !isRef {
				it := scalarTypeSpec(items)
				return schema.TypeSpec{Type: typeArray, Items: &it}
			}
		}
	}
	return anyTypeSpec()
}

func soleKey(m map[string]string) string {
	for k := range m {
		return k
	}
	return ""
}

// unionTypeSpec renders a discriminated base as an inline oneOf +
// discriminator over its const-tagged variants. The base type itself is
// never registered. Two-member object unions are rejected: the .NET and
// Java SDK deserializers pick the wrong member / null at that arity.
func (r *typeRegistry) unionTypeSpec(baseName, tagProp string, mapping map[string]string) schema.TypeSpec {
	if len(mapping) == 2 {
		r.errorf("%s: 2-member object unions are not allowed "+
			"(broken .NET/Java output deserialization); flatten or wait for upstream fixes", baseName)
		return anyTypeSpec()
	}
	tags := make([]string, 0, len(mapping))
	for tag := range mapping {
		tags = append(tags, tag)
	}
	slices.Sort(tags)

	oneOf := make([]schema.TypeSpec, 0, len(mapping))
	discMapping := make(map[string]string, len(mapping))
	for _, tag := range tags {
		variantName := refSchemaName(mapping[tag])
		token := r.emitVariant(variantName, tagProp, tag)
		oneOf = append(oneOf, r.refTo(token))
		discMapping[tag] = "#/types/" + token
	}
	return schema.TypeSpec{
		OneOf: oneOf,
		Discriminator: &schema.DiscriminatorSpec{
			PropertyName: tagProp,
			Mapping:      discMapping,
		},
	}
}

// emitNamed registers a named object type for a component schema and returns
// its token. allOf chains are flattened; recursion terminates via the
// emitting state.
func (r *typeRegistry) emitNamed(name string) string {
	return r.emit(name, "", "")
}

// emitVariant registers a union variant: the flattened schema (allOf pulls in
// the base) with the tag property const-ed and required.
func (r *typeRegistry) emitVariant(name, tagProp, tag string) string {
	return r.emit(name, tagProp, tag)
}

func (r *typeRegistry) emit(name, tagProp, tag string) string {
	token := r.tokenFor(name)
	if r.state[name] != 0 {
		return token
	}
	r.state[name] = stateEmitting

	ref := schemaRefPrefix + name
	props, required, err := flattenObjectSchema(r.spec, ref)
	if err != nil {
		r.errorf("type %s: %v", name, err)
		r.state[name] = stateDone
		return token
	}

	pprops := map[string]schema.PropertySpec{}
	for k, v := range props {
		ps := r.property(v)
		if looksSecret(k) {
			ps.Secret = true
		}
		pprops[k] = ps
	}
	reqSet := map[string]bool{}
	for _, rr := range required {
		if _, ok := pprops[rr]; ok {
			reqSet[rr] = true
		}
	}
	for _, opt := range r.typeMeta[name].Optional {
		delete(reqSet, opt)
	}

	if tagProp != "" {
		desc := fmt.Sprintf("Expected value is '%s'.", tag)
		if prev, ok := pprops[tagProp]; ok && prev.Description != "" {
			desc = prev.Description + " " + desc
		}
		pprops[tagProp] = schema.PropertySpec{
			TypeSpec:    schema.TypeSpec{Type: typeString},
			Const:       tag,
			Description: desc,
		}
		reqSet[tagProp] = true
	}

	desc := ""
	if node, ok := r.spec.ResolveSchema(ref); ok {
		desc, _ = node["description"].(string)
	}

	r.types[token] = schema.ComplexTypeSpec{
		ObjectTypeSpec: schema.ObjectTypeSpec{
			Type:        typeObject,
			Description: desc,
			Properties:  pprops,
			Required:    sortedKeys(reqSet),
		},
	}
	r.state[name] = stateDone
	return token
}
