// Copyright 2016-2026, Pulumi Corporation.
//
// metadata.go — emit the runtime metadata.json that drives CloudAPIMetadata
// lookups in provider/pkg/runtime at CRUD time. The same resource-map.yaml
// that produces schema.json produces this file; the two are always in lockstep.
//
// The emitter is intentionally dumb: translate each resource-map entry into
// its runtime.CloudAPIMetadata counterpart, filling in path templates from
// the OpenAPI spec. No type inference, no heuristics — the map is the truth.

package gen

import (
	"encoding/json"
	"fmt"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/runtime"
	"gopkg.in/yaml.v2"
)

// EmitMetadata produces the runtime metadata JSON.
func EmitMetadata(specPath, mapPath string) ([]byte, error) {
	spec, err := LoadSpec(specPath)
	if err != nil {
		return nil, err
	}
	rm, err := LoadResourceMap(mapPath)
	if err != nil {
		return nil, err
	}

	md := runtime.CloudAPIMetadata{
		Resources: map[string]runtime.CloudAPIResource{},
		Functions: map[string]runtime.CloudAPIFunction{},
		Methods:   map[string]runtime.CloudAPIMethod{},
	}

	for modName, mod := range rm.Modules {
		for resName, res := range mod.Resources {
			token := fmt.Sprintf("pulumiservice:%s:%s", modName, resName)
			r, ok := buildRuntimeResource(token, modName, res, spec)
			if !ok {
				continue // skipped: TODO or unresolvable operationIds
			}
			md.Resources[token] = r
		}
		for fnName, fn := range mod.Functions {
			if fn.OperationID == "" || isTodoMarker(fn.OperationID) {
				continue
			}
			op, ok := spec.ByID[fn.OperationID]
			if !ok {
				continue
			}
			token := fmt.Sprintf("pulumiservice:%s:%s", modName, fnName)
			md.Functions[token] = runtime.CloudAPIFunction{
				Token:  token,
				Module: modName,
				Operation: runtime.CloudAPIOperation{
					OperationID:  op.OperationID,
					Method:       op.Method,
					PathTemplate: op.Path,
				},
			}
		}
		for methodName, m := range mod.Methods {
			addMethod(md.Methods, methodName, m, spec)
		}
	}

	for methodToken, m := range rm.Methods {
		addMethod(md.Methods, methodToken, m, spec)
	}

	return json.MarshalIndent(md, "", "  ")
}

// addMethod inserts a method entry into the runtime metadata, skipping
// TODOs and operationIds the spec doesn't recognize.
func addMethod(out map[string]runtime.CloudAPIMethod, methodName string, m Method, spec *Spec) {
	if m.OperationID == "" || isTodoMarker(m.OperationID) {
		return
	}
	op, ok := spec.ByID[m.OperationID]
	if !ok {
		return
	}
	resToken := fmt.Sprintf("pulumiservice:%s", m.Resource)
	token := fmt.Sprintf("pulumiservice:%s", methodName)
	out[token] = runtime.CloudAPIMethod{
		Token:       methodName,
		ResourceRef: resToken,
		Operation: runtime.CloudAPIOperation{
			OperationID:  op.OperationID,
			Method:       op.Method,
			PathTemplate: op.Path,
		},
	}
}

// buildRuntimeResource translates one resource-map entry into runtime form.
// Returns (resource, ok); ok=false if essential operations are TODO/missing.
func buildRuntimeResource(token, modName string, res Resource, spec *Spec) (runtime.CloudAPIResource, bool) {
	out := runtime.CloudAPIResource{
		Token:          token,
		Module:         modName,
		ForceNew:       res.ForceNew,
		SortProperties: res.SortProperties,
		Properties:     map[string]runtime.CloudAPIProperty{},
	}
	if res.Discriminator != nil {
		out.Discriminator = res.Discriminator.Field
	}
	if res.ID != nil {
		out.ID = &runtime.CloudAPIID{
			Template:  res.ID.Template,
			Params:    res.ID.Params,
			Case:      res.ID.Case,
			Templates: res.ID.Templates,
		}
	}

	// Translate property metadata.
	for name, p := range res.Properties {
		out.Properties[name] = runtime.CloudAPIProperty{
			Type:             p.Type,
			From:             p.From,
			Source:           p.Source,
			CreateSource:     p.CreateSource,
			CreateFrom:       p.CreateFrom,
			Secret:           p.Secret,
			Output:           p.Output,
			WriteOnly:        p.WriteOnly,
			DiffMode:         p.DiffMode,
			Default:          p.Default,
			DefaultFromField: p.DefaultFromField,
			SortOnRead:       p.SortOnRead,
			Aliases:          p.Aliases,
			Doc:              p.Doc,
		}
	}

	if res.AutoName != nil {
		out.AutoName = &runtime.AutoNameConfig{
			Property: res.AutoName.Property,
			Pattern:  res.AutoName.Pattern,
			Kind:     res.AutoName.Kind,
		}
	}
	for _, c := range res.Checks {
		out.Checks = append(out.Checks, checkFromMap(c))
	}

	// Walk the operations block structurally. Two valid shapes:
	//   (a) flat: `create: CreateFoo` / `read: GetFoo` / ... / `readVia: {...}`
	//   (b) polymorphic scopes: `case: scope` + `scopes: {org: {...}, stack: {...}}`
	ok := buildOperations(&out, res.Operations, spec)
	if !ok {
		return runtime.CloudAPIResource{}, false
	}
	return out, true
}

// buildOperations decodes the heterogeneous `operations:` YAML into the
// runtime's CloudAPIOperation fields. Returns false if the entry's create
// operation is TODO / missing from the spec — the emitter skips those.
func buildOperations(res *runtime.CloudAPIResource, ops yaml.MapSlice, spec *Spec) bool {
	hasCanonicalCreate := false

	for _, kv := range ops {
		key, _ := kv.Key.(string)
		switch key {
		case "create":
			if op := buildOp(kv.Value, spec); op != nil {
				res.Create = op
				hasCanonicalCreate = true
			} else if hasPolymorphicCreate(kv.Value) {
				// Handled in the scopes branch; we still consider the
				// resource to have a create operation.
				hasCanonicalCreate = true
			}
		case "read":
			res.Read = buildOp(kv.Value, spec)
		case "update":
			res.Update = buildOp(kv.Value, spec)
		case "delete":
			res.Delete = buildOp(kv.Value, spec)
		case "postCreate":
			res.PostCreate = buildOp(kv.Value, spec)
		case "readVia":
			if rv := buildReadVia(kv.Value); rv != nil {
				res.ReadVia = rv
			}
		case "case":
			// Polymorphic discriminator; handled alongside `scopes`.
			if s, isStr := kv.Value.(string); isStr {
				if res.ID != nil && res.ID.Case == "" {
					res.ID.Case = s
				}
			}
		case "scopes":
			scopes, isSlice := kv.Value.(yaml.MapSlice)
			if !isSlice {
				continue
			}
			if res.PolymorphicScopes == nil {
				res.PolymorphicScopes = &runtime.PolymorphicScopes{
					Discriminator: "scope",
					Scopes:        map[string]runtime.CloudAPIResourceOps{},
				}
			}
			for _, scopeKV := range scopes {
				scopeName, _ := scopeKV.Key.(string)
				scopeOps, isSlice2 := scopeKV.Value.(yaml.MapSlice)
				if !isSlice2 {
					continue
				}
				opsBundle := runtime.CloudAPIResourceOps{}
				for _, sub := range scopeOps {
					sk, _ := sub.Key.(string)
					switch sk {
					case "create":
						opsBundle.Create = buildOp(sub.Value, spec)
						if opsBundle.Create != nil {
							hasCanonicalCreate = true
						}
					case "read":
						opsBundle.Read = buildOp(sub.Value, spec)
					case "update":
						opsBundle.Update = buildOp(sub.Value, spec)
					case "delete":
						opsBundle.Delete = buildOp(sub.Value, spec)
					}
				}
				res.PolymorphicScopes.Scopes[scopeName] = opsBundle
			}
		}
	}
	return hasCanonicalCreate
}

// buildOp turns an operation map entry into a runtime.CloudAPIOperation.
// Accepts two shapes:
//   - A bare string operationId (most common).
//   - A yaml.MapSlice with at least `operationId:` + optional modifiers
//     like `bodyOverride:` (for tombstone-style deletes via update ops).
// Returns nil for TODO markers or operationIds missing from the spec.
// Polymorphic `create:` blocks return nil here — the caller detects them
// separately via hasPolymorphicCreate and builds PolymorphicScopes instead.
func buildOp(v interface{}, spec *Spec) *runtime.CloudAPIOperation {
	switch x := v.(type) {
	case string:
		return buildOpFromID(x, spec)
	case yaml.MapSlice:
		opID := ""
		var bodyOverride map[string]interface{}
		iterateOver := ""
		iterateKeyParam := ""
		rawBodyFrom := ""
		rawBodyTo := ""
		contentType := ""
		for _, kv := range x {
			k, _ := kv.Key.(string)
			switch k {
			case "operationId":
				opID, _ = kv.Value.(string)
			case "bodyOverride":
				bodyOverride = yamlToJSONMap(kv.Value)
			case "iterateOver":
				iterateOver, _ = kv.Value.(string)
			case "iterateKeyParam":
				iterateKeyParam, _ = kv.Value.(string)
			case "rawBodyFrom":
				rawBodyFrom, _ = kv.Value.(string)
			case "rawBodyTo":
				rawBodyTo, _ = kv.Value.(string)
			case "contentType":
				contentType, _ = kv.Value.(string)
			}
		}
		if opID == "" {
			// Not an operation-with-modifiers shape — could be a polymorphic
			// create block. The caller handles that path.
			return nil
		}
		base := buildOpFromID(opID, spec)
		if base == nil {
			return nil
		}
		base.BodyOverride = bodyOverride
		base.IterateOver = iterateOver
		base.IterateKeyParam = iterateKeyParam
		base.RawBodyFrom = rawBodyFrom
		base.RawBodyTo = rawBodyTo
		base.ContentType = contentType
		return base
	}
	return nil
}

// buildOpFromID resolves an operationId string against the spec.
func buildOpFromID(s string, spec *Spec) *runtime.CloudAPIOperation {
	if isTodoMarker(s) {
		return nil
	}
	op, ok := spec.ByID[s]
	if !ok {
		return nil
	}
	return &runtime.CloudAPIOperation{
		OperationID:  op.OperationID,
		Method:       op.Method,
		PathTemplate: op.Path,
	}
}

// yamlToJSONMap converts a yaml.MapSlice (or similar) into a plain
// map[string]interface{} suitable for JSON encoding. Keeps the JSON
// encoder happy — yaml.MapSlice doesn't marshal to the right shape.
func yamlToJSONMap(v interface{}) map[string]interface{} {
	switch x := v.(type) {
	case yaml.MapSlice:
		out := map[string]interface{}{}
		for _, kv := range x {
			k, _ := kv.Key.(string)
			out[k] = yamlToJSONValue(kv.Value)
		}
		return out
	case map[interface{}]interface{}:
		out := map[string]interface{}{}
		for k, val := range x {
			ks, _ := k.(string)
			out[ks] = yamlToJSONValue(val)
		}
		return out
	}
	return nil
}

// yamlToJSONValue normalizes a decoded YAML value into JSON-marshalable form.
func yamlToJSONValue(v interface{}) interface{} {
	switch x := v.(type) {
	case yaml.MapSlice:
		return yamlToJSONMap(x)
	case map[interface{}]interface{}:
		return yamlToJSONMap(x)
	case []interface{}:
		out := make([]interface{}, len(x))
		for i, item := range x {
			out[i] = yamlToJSONValue(item)
		}
		return out
	}
	return v
}

// hasPolymorphicCreate returns true if a `create:` node is a discriminated
// MapSlice (as in Team: create has pulumi/github variants). The caller
// treats this as "create exists" without trying to resolve one operationId.
func hasPolymorphicCreate(v interface{}) bool {
	_, ok := v.(yaml.MapSlice)
	return ok
}

func buildReadVia(v interface{}) *runtime.CloudAPIReadVia {
	m, ok := v.(yaml.MapSlice)
	if !ok {
		return nil
	}
	var rv runtime.CloudAPIReadVia
	for _, kv := range m {
		k, _ := kv.Key.(string)
		s, _ := kv.Value.(string)
		switch k {
		case "operationId":
			rv.OperationID = s
		case "filterBy":
			rv.FilterBy = s
		case "extractField":
			rv.ExtractField = s
		case "keyBy":
			rv.KeyBy = s
		case "valueProperty":
			rv.ValueProperty = s
		}
	}
	if rv.OperationID == "" || isTodoMarker(rv.OperationID) {
		return nil
	}
	return &rv
}

// checkFromMap translates a declarative check entry into the runtime rule.
func checkFromMap(m map[string]interface{}) runtime.CheckRule {
	r := runtime.CheckRule{}
	if v, ok := m["requireOneOf"].([]interface{}); ok {
		for _, e := range v {
			if s, ok := e.(string); ok {
				r.RequireOneOf = append(r.RequireOneOf, s)
			}
		}
	}
	if v, ok := m["requireAtMostOne"].([]interface{}); ok {
		for _, e := range v {
			if s, ok := e.(string); ok {
				r.RequireAtMostOne = append(r.RequireAtMostOne, s)
			}
		}
	}
	if v, ok := m["requireTogether"].([]interface{}); ok {
		for _, e := range v {
			if s, ok := e.(string); ok {
				r.RequireTogether = append(r.RequireTogether, s)
			}
		}
	}
	if s, ok := m["requireIfSet"].(string); ok {
		r.RequireIfSet = s
	}
	if s, ok := m["requireIf"].(string); ok {
		r.RequireIf = s
	}
	if s, ok := m["field"].(string); ok {
		r.Field = s
	}
	if s, ok := m["message"].(string); ok {
		r.Message = s
	}
	return r
}
