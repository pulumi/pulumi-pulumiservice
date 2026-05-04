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

// Command scaffold-metadata derives v2 resource operations from the
// embedded OpenAPI spec and merges them into metadata.json in place.
//
// metadata.json is the single source of truth at runtime. It is BOTH
// auto-generated (operations) AND hand-curated (idField, renames,
// fields, outputs, examples, etc.) — `go generate` re-runs this tool
// to refresh operations from the spec while preserving every other
// field on each entry.
//
// To exclude a derived token, add it to the top-level `_excluded`
// array in metadata.json. The scaffolder will skip those tokens on
// every regen so they never resurface.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"os"
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/rest"
)

// verb prefixes mapped to CRUD slots. Order matters: longest first so
// "BatchCreate" wins over "Create".
var verbPrefixes = []struct {
	prefix string
	slot   string
}{
	{"BatchCreate", "create"},
	{"BatchUpdate", "update"},
	{"BatchDelete", "delete"},
	{"Create", "create"},
	{"Register", "create"}, // RegisterOidcIssuer, RegisterX patterns
	{"Add", "create"},      // AddOrganizationMember, AddStackTag patterns
	{"New", "create"},      // NewPolicyGroup pattern
	{"Update", "update"},
	{"Delete", "delete"},
	{"Remove", "delete"}, // RemoveX as delete-shaped verb
	{"Patch", "update"},
	{"Replace", "update"},
	{"Put", "update"},
	{"Get", "read"},
	{"Read", "read"},
	{"Describe", "read"},
}

// nonCRUDVerbs are recognized so noun extraction works on them, but they
// don't map to a CRUD slot.
var nonCRUDVerbs = []string{
	"List", "Search", "Find", "Cancel", "Approve",
	"Reject", "Reset", "Refresh", "Restore", "Validate", "Trigger",
	"Poll", "Open", "Close", "Encrypt", "Decrypt", "Apply", "Complete",
	"Append", "Bulk", "Accept",
}

// scopePrefixes are noun prefixes that hint at the path context rather
// than the resource itself. When operations split into "Org{X}" + "{X}"
// nouns, we fold them together — the Pulumi resource is keyed off the
// bare noun. Order matters: longest first so "Organization" wins over
// "Org".
var scopePrefixes = []string{
	"Organization", "Org",
	"Pulumi",
	"Team",
	"Project",
	"Stack",
	"User",
}

// metadataDoc mirrors the on-disk shape of metadata.json. Operations
// are serialized via json.RawMessage so the scaffolder doesn't drop
// future fields it doesn't recognize.
type metadataDoc struct {
	Version   int                        `json:"version"`
	Package   string                     `json:"package,omitempty"`
	Note      string                     `json:"_note,omitempty"`
	Excluded  []string                   `json:"_excluded,omitempty"`
	Resources map[string]json.RawMessage `json:"resources"`
}

// derivedOps is the only field the scaffolder writes into a per-
// resource entry. Everything else on an entry is preserved as-is via
// json.RawMessage round-tripping.
type derivedOps struct {
	Create string `json:"create,omitempty"`
	Read   string `json:"read,omitempty"`
	Update string `json:"update,omitempty"`
	Delete string `json:"delete,omitempty"`
}

func main() {
	in := flag.String("in", "spec.json", "Input OpenAPI spec path")
	out := flag.String("out", "metadata.json", "Path to metadata.json (read + written in place)")
	flag.Parse()

	specBytes, err := os.ReadFile(*in)
	if err != nil {
		fail("read spec: %v", err)
	}
	var rawSpec struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(specBytes, &rawSpec); err != nil {
		fail("parse spec: %v", err)
	}
	parsedSpec, err := rest.ParseSpec(specBytes)
	if err != nil {
		fail("rest.ParseSpec: %v", err)
	}

	doc := loadMetadata(*out)

	candidates, totalOps, skipped, unmapped := derive(rawSpec.Paths)

	excluded := map[string]bool{}
	for _, tok := range doc.Excluded {
		excluded[tok] = true
	}

	added, updated := 0, 0
	for tok, derivedOpsForTok := range candidates {
		if excluded[tok] {
			continue
		}
		entry, exists := doc.Resources[tok]
		idField, renames := inferOverrides(parsedSpec, derivedOpsForTok)
		outputsExclude := inferOutputsExclude(parsedSpec, tok, derivedOpsForTok)
		merged, changed, err := mergeOperations(entry, derivedOpsForTok, idField, renames, outputsExclude)
		if err != nil {
			fail("merge %s: %v", tok, err)
		}
		if !exists {
			added++
		} else if changed {
			updated++
		}
		doc.Resources[tok] = merged
	}

	// Surface tokens that exist in metadata.json but the scaffolder no
	// longer derives — they may be stale (spec dropped the op) or hand-
	// added in error.
	var orphans []string
	for tok := range doc.Resources {
		if _, ok := candidates[tok]; !ok && !excluded[tok] {
			orphans = append(orphans, tok)
		}
	}
	sort.Strings(orphans)

	if err := writeMetadata(*out, doc); err != nil {
		fail("write %s: %v", *out, err)
	}

	fmt.Fprintf(os.Stderr, "scaffold-metadata: %s\n", *out)
	fmt.Fprintf(os.Stderr, "  operations in spec:        %d\n", totalOps)
	fmt.Fprintf(os.Stderr, "  candidates derived:        %d\n", len(candidates))
	fmt.Fprintf(os.Stderr, "  added new entries:         %d\n", added)
	fmt.Fprintf(os.Stderr, "  updated existing entries:  %d\n", updated)
	fmt.Fprintf(os.Stderr, "  excluded:                  %d\n", len(excluded))
	fmt.Fprintf(os.Stderr, "  skipped (no Create+Read|Delete): %d\n", len(skipped))
	if len(orphans) > 0 {
		fmt.Fprintf(os.Stderr, "  orphans (in metadata.json, not derived from spec): %d\n", len(orphans))
		for _, o := range orphans {
			fmt.Fprintf(os.Stderr, "    %s\n", o)
		}
	}
	if len(unmapped) > 0 {
		fmt.Fprintf(os.Stderr, "  noun groups with non-CRUD verbs only: %d\n", len(unmapped))
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "scaffold-metadata: "+format+"\n", args...)
	os.Exit(1)
}

func loadMetadata(path string) *metadataDoc {
	doc := &metadataDoc{
		Version:   1,
		Package:   "pulumiservice",
		Resources: map[string]json.RawMessage{},
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return doc
		}
		fail("read %s: %v", path, err)
	}
	if err := json.Unmarshal(data, doc); err != nil {
		fail("parse %s: %v", path, err)
	}
	if doc.Resources == nil {
		doc.Resources = map[string]json.RawMessage{}
	}
	return doc
}

// mergeOperations layers derived operations on top of an existing
// per-resource entry. The entry is decoded into a generic map, the
// `operations` key is replaced wholesale with the derived block, and
// the entry is re-encoded in deterministic key order.
//
// Inferred idField/renames fill in defaults the spec implies but a human
// can override: an entry that already declares `idField` keeps its
// value, and the existing `renames` map wins on per-key collision (so
// curators can add or correct the inference without it being clobbered
// on the next regen).
func mergeOperations(existing json.RawMessage, ops derivedOps, inferredIDField string, inferredRenames map[string]string, inferredOutputsExclude []string) (json.RawMessage, bool, error) {
	var entry map[string]any
	if len(existing) > 0 {
		if err := json.Unmarshal(existing, &entry); err != nil {
			return nil, false, err
		}
	}
	if entry == nil {
		entry = map[string]any{}
	}

	newOps := map[string]string{}
	if ops.Create != "" {
		newOps["create"] = ops.Create
	}
	if ops.Read != "" {
		newOps["read"] = ops.Read
	}
	if ops.Update != "" {
		newOps["update"] = ops.Update
	}
	if ops.Delete != "" {
		newOps["delete"] = ops.Delete
	}

	prev, _ := entry["operations"].(map[string]any)
	changed := !sameOps(prev, newOps)
	entry["operations"] = newOps

	// idField: keep human override, otherwise drop in the inference.
	if inferredIDField != "" {
		if _, has := entry["idField"]; !has {
			entry["idField"] = inferredIDField
		}
	}

	// renames: union of inferred + existing; existing keys win.
	if len(inferredRenames) > 0 {
		merged := map[string]any{}
		for k, v := range inferredRenames {
			merged[k] = v
		}
		if existingRenames, ok := entry["renames"].(map[string]any); ok {
			for k, v := range existingRenames {
				merged[k] = v
			}
		}
		entry["renames"] = merged
	}

	// outputsExclude: keep human override; drop in inference otherwise.
	if len(inferredOutputsExclude) > 0 {
		if _, has := entry["outputsExclude"]; !has {
			arr := make([]any, len(inferredOutputsExclude))
			for i, v := range inferredOutputsExclude {
				arr[i] = v
			}
			entry["outputsExclude"] = arr
		}
	}

	encoded, err := encodeStable(entry)
	if err != nil {
		return nil, false, err
	}
	return encoded, changed, nil
}

// inferOverrides walks the create + read operations and returns the
// idField and rename map the spec implies. Returns "" / nil when nothing
// can be inferred. The caller decides whether to apply them.
//
// Inference rules:
//
//  1. Composite idField when Create has no usable id in its response —
//     we synthesize "{p1}/{p2}/...{pN}" from the read (or update/delete)
//     path's parameters. Resources whose Create returns 204 No Content
//     (StackTag, OrganizationMember) need this, since extractID can't
//     pull "/id" from an empty body.
//
//  2. Server-id renames: a path parameter that appears in read/update/
//     delete but NOT in create is a server-generated identifier.
//     If Create's response body has a top-level `id` field, the
//     parameter on the wire is `id` while the path uses a more specific
//     name (issuerId, poolId, scheduleID, templateID, tokenId, etc.).
//     We emit `<wireParam> -> id` so response decoding renames the
//     server's `id` field to the parameter name in state.
//
//  3. Body-vs-path renames: same trigger as (2), but when the response
//     body has no `id` field, we look at Create's request body for a
//     plausible matching field. For "{noun}Name" style params (tagName,
//     hookName), if the request body has `name`, we emit `<wireParam>
//     -> name`.
func inferOverrides(spec *rest.Spec, ops derivedOps) (idField string, renames map[string]string) {
	createOp := opOrNil(spec, ops.Create)
	readOp := opOrNil(spec, ops.Read)
	updateOp := opOrNil(spec, ops.Update)
	deleteOp := opOrNil(spec, ops.Delete)

	idField = inferIDField(spec, createOp, readOp, updateOp, deleteOp)
	renames = inferRenames(spec, createOp, readOp, updateOp, deleteOp)
	return idField, renames
}

func opOrNil(spec *rest.Spec, id string) *rest.Operation {
	if id == "" {
		return nil
	}
	op, _ := spec.Op(id)
	return op
}

// pathParamsOf returns the path parameter names of an op in order. Empty
// when op is nil.
func pathParamsOf(op *rest.Operation) []string {
	if op == nil {
		return nil
	}
	out := make([]string, 0, len(op.Parameters))
	for _, p := range op.Parameters {
		if p.In == "path" {
			out = append(out, p.Name)
		}
	}
	return out
}

// flattenedProps walks a $ref into an object schema and returns its
// top-level properties, recursively flattening allOf composition.
// Mirrors rest/schema.go's flattenObjectSchema but lighter: we only
// need property names here, not the full schema entries.
func flattenedProps(spec *rest.Spec, ref string) map[string]any {
	if ref == "" {
		return nil
	}
	out := map[string]any{}
	visited := map[string]bool{}
	var walk func(node map[string]any)
	walk = func(node map[string]any) {
		if r, ok := node["$ref"].(string); ok {
			if visited[r] {
				return
			}
			visited[r] = true
			if resolved, ok := spec.ResolveSchema(r); ok {
				walk(resolved)
			}
			return
		}
		if all, ok := node["allOf"].([]any); ok {
			for _, m := range all {
				if mm, ok := m.(map[string]any); ok {
					walk(mm)
				}
			}
		}
		if props, ok := node["properties"].(map[string]any); ok {
			for k, v := range props {
				out[k] = v
			}
		}
	}
	if root, ok := spec.ResolveSchema(ref); ok {
		walk(root)
	}
	return out
}

// responseHasField reports whether the create response body schema
// (after flattening allOf) has a top-level field with the given name.
func responseHasField(spec *rest.Spec, op *rest.Operation, field string) bool {
	if op == nil {
		return false
	}
	_, has := flattenedProps(spec, op.ResponseRef)[field]
	return has
}

// requestHasField reports whether the create request body schema
// (after flattening allOf) has a top-level field with the given name.
func requestHasField(spec *rest.Spec, op *rest.Operation, field string) bool {
	if op == nil {
		return false
	}
	_, has := flattenedProps(spec, op.RequestRef)[field]
	return has
}

// inferOutputsExclude flags response-body fields that collide with the
// resource's own type name (a "service" field on the Service resource,
// "policyIssue" on PolicyIssue). Pulumi's per-language SDK codegen
// rejects these as member-vs-class collisions; the dispatcher doesn't
// need them either, since they're typically envelope wrappers around
// the same data already present at the top level.
func inferOutputsExclude(spec *rest.Spec, token string, ops derivedOps) []string {
	noun := token
	if i := strings.LastIndex(noun, ":"); i >= 0 {
		noun = noun[i+1:]
	}
	if noun == "" {
		return nil
	}
	lower := strings.ToLower(noun[:1]) + noun[1:]
	for _, opID := range []string{ops.Read, ops.Create} {
		if opID == "" {
			continue
		}
		op, ok := spec.Op(opID)
		if !ok || op.ResponseRef == "" {
			continue
		}
		schema, ok := spec.ResolveSchema(op.ResponseRef)
		if !ok {
			continue
		}
		props, _ := schema["properties"].(map[string]any)
		if _, has := props[lower]; has {
			return []string{lower}
		}
	}
	return nil
}

// splitTrailingSuffix breaks a CamelCased path param like "tagName"
// into ("tag", "name", true). Returns ok=false when there's no
// recognized trailing suffix.
func splitTrailingSuffix(name string) (prefix, bareSuffix string, ok bool) {
	for _, s := range []string{"Name", "ID", "Id"} {
		if strings.HasSuffix(name, s) && len(name) > len(s) {
			return name[:len(name)-len(s)], strings.ToLower(s), true
		}
	}
	return "", "", false
}

func inferIDField(spec *rest.Spec, createOp, readOp, updateOp, deleteOp *rest.Operation) string {
	// If create's response carries `id`, the default extractID lookup
	// at "/id" works — no override needed.
	if responseHasField(spec, createOp, "id") {
		return ""
	}
	// Otherwise fall back to a composite from the next-best path.
	for _, op := range []*rest.Operation{readOp, updateOp, deleteOp} {
		if op == nil {
			continue
		}
		params := pathParamsOf(op)
		if len(params) == 0 {
			continue
		}
		var b strings.Builder
		for i, p := range params {
			if i > 0 {
				b.WriteByte('/')
			}
			b.WriteString("{" + p + "}")
		}
		return b.String()
	}
	return ""
}

// inferRenames picks renames for path parameters that aren't already
// present in the create op's path. Two patterns:
//
//  1. Server-generated id: the create response carries `id` and the
//     path uses a noun-prefixed form (issuerId, poolId, scheduleID,
//     templateID, tokenId, gateID). The Pulumi-side name keeps the
//     descriptive path form; the rename teaches the dispatcher that
//     the wire `id` is that field. Output: {path-param → "id"}.
//
//  2. Body field mirrors the path: the create body has a field whose
//     name matches the path-param's stripped form (`projectName` →
//     body `project`, `tagName` → body `name`, `hookName` → body
//     `name`). The user types the body-side spelling; the rename
//     teaches the dispatcher that the wire path-param is that field.
//     Output: {body-field → path-param}.
func inferRenames(spec *rest.Spec, createOp, readOp, updateOp, deleteOp *rest.Operation) map[string]string {
	createParams := map[string]bool{}
	for _, p := range pathParamsOf(createOp) {
		createParams[p] = true
	}

	out := map[string]string{}
	seen := map[string]bool{}

	// Pass 1: precise matches (rules 1 + 2). These claim body keys
	// before the catch-all rule (3) gets a chance. Without two passes,
	// Service's `ownerType` (no recognized suffix) would steal `name`
	// before `serviceName` (suffix match) was visited.
	for _, op := range []*rest.Operation{readOp, updateOp, deleteOp} {
		for _, p := range pathParamsOf(op) {
			if createParams[p] || seen[p] {
				continue
			}
			// Path param that's also a body field is a duplicate
			// (Service's `ownerName` and `ownerType` appear in both).
			// They're not aliases — skip rename inference for them.
			if requestHasField(spec, createOp, p) {
				seen[p] = true
				continue
			}
			if responseHasField(spec, createOp, "id") && p != "id" {
				out[p] = "id"
				seen[p] = true
				continue
			}
			if prefix, bareSuffix, ok := splitTrailingSuffix(p); ok {
				if prefix != "" && requestHasField(spec, createOp, prefix) {
					if _, claimed := out[prefix]; !claimed {
						out[prefix] = p
						seen[p] = true
						continue
					}
				}
				if requestHasField(spec, createOp, bareSuffix) {
					if _, claimed := out[bareSuffix]; !claimed {
						out[bareSuffix] = p
						seen[p] = true
						continue
					}
				}
			}
		}
	}

	// Pass 2: rule (3) — bare path param matched by body `name`. Only
	// fires when the path param isn't itself a body field (in which
	// case it's a duplicate, not an alias) and `name` hasn't already
	// been claimed.
	for _, op := range []*rest.Operation{readOp, updateOp, deleteOp} {
		for _, p := range pathParamsOf(op) {
			if createParams[p] || seen[p] {
				continue
			}
			if requestHasField(spec, createOp, p) {
				seen[p] = true
				continue
			}
			if _, claimed := out["name"]; !claimed && requestHasField(spec, createOp, "name") {
				out["name"] = p
				seen[p] = true
				continue
			}
		}
	}

	// (4) Body fields that duplicate a path param under a verbose alias
	// (`organizationName` body ↔ `orgName` path). The user supplies one
	// value; the dispatcher fans it out to both wire fields. Detect by
	// matching trailing suffix and a stem prefix relationship.
	if createOp != nil {
		bodyProps := flattenedProps(spec, createOp.RequestRef)
		for _, p := range pathParamsOf(createOp) {
			pPrefix, pSuffix, pOK := splitTrailingSuffix(p)
			if !pOK || pPrefix == "" {
				continue
			}
			pPrefixLower := strings.ToLower(pPrefix)
			for body := range bodyProps {
				if body == p {
					continue
				}
				bPrefix, bSuffix, bOK := splitTrailingSuffix(body)
				if !bOK || bSuffix != pSuffix {
					continue
				}
				if !strings.HasPrefix(strings.ToLower(bPrefix), pPrefixLower) {
					continue
				}
				if _, exists := out[body]; !exists {
					out[body] = p
				}
			}
		}
	}
	return out
}

func sameOps(prev map[string]any, next map[string]string) bool {
	if len(prev) != len(next) {
		return false
	}
	for k, v := range next {
		pv, ok := prev[k].(string)
		if !ok || pv != v {
			return false
		}
	}
	return true
}

// encodeStable marshals a map with deterministic key ordering: a
// preferred order for known fields, then alphabetical for the rest.
// json.Marshal already sorts map[string]any keys alphabetically, but
// this helper makes diffs prefer "operations" near the top.
func encodeStable(entry map[string]any) (json.RawMessage, error) {
	preferred := []string{"operations", "idField", "aliases", "renames", "fields", "outputs", "outputsExclude", "description", "examples"}
	var keys []string
	seen := map[string]bool{}
	for _, k := range preferred {
		if _, ok := entry[k]; ok {
			keys = append(keys, k)
			seen[k] = true
		}
	}
	var rest []string
	for k := range entry {
		if !seen[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	keys = append(keys, rest...)

	var b strings.Builder
	b.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		kEnc, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		b.Write(kEnc)
		b.WriteByte(':')
		vEnc, err := json.Marshal(entry[k])
		if err != nil {
			return nil, err
		}
		b.Write(vEnc)
	}
	b.WriteByte('}')
	return json.RawMessage(b.String()), nil
}

func writeMetadata(path string, doc *metadataDoc) error {
	// Marshal the envelope, then re-encode the resources map in sorted
	// token order so diffs stay clean across regens.
	var b strings.Builder
	b.WriteString("{\n")
	fmt.Fprintf(&b, "  \"version\": %d,\n", doc.Version)
	if doc.Package != "" {
		fmt.Fprintf(&b, "  \"package\": %q,\n", doc.Package)
	}
	if doc.Note != "" {
		fmt.Fprintf(&b, "  \"_note\": %q,\n", doc.Note)
	}
	if len(doc.Excluded) > 0 {
		ex, err := json.MarshalIndent(doc.Excluded, "  ", "  ")
		if err != nil {
			return err
		}
		b.WriteString("  \"_excluded\": ")
		b.Write(ex)
		b.WriteString(",\n")
	}

	tokens := slices.Collect(maps.Keys(doc.Resources))
	sort.Strings(tokens)

	b.WriteString("  \"resources\": {")
	for i, tok := range tokens {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString("\n    ")
		kEnc, err := json.Marshal(tok)
		if err != nil {
			return err
		}
		b.Write(kEnc)
		b.WriteString(": ")
		entryEnc, err := indentJSON(doc.Resources[tok], "    ")
		if err != nil {
			return err
		}
		b.Write(entryEnc)
	}
	if len(tokens) > 0 {
		b.WriteString("\n  ")
	}
	b.WriteString("}\n}\n")

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// indentJSON re-renders a raw JSON value with 2-space indentation
// matching the surrounding metadata.json formatting.
func indentJSON(raw json.RawMessage, prefix string) ([]byte, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	formatted, err := json.MarshalIndent(v, prefix, "  ")
	if err != nil {
		return nil, err
	}
	return formatted, nil
}

// derive walks every operationId and returns the candidates we'd
// emit, plus diagnostics: total ops seen, skipped noun groups, and
// noun groups consisting solely of non-CRUD verbs.
func derive(paths map[string]map[string]any) (
	candidates map[string]derivedOps,
	totalOps int,
	skipped []string,
	unmapped map[string][]string,
) {
	type ops = map[string]string
	byNoun := map[string]ops{}
	otherOps := map[string][]string{}

	pathKeys := slices.Collect(maps.Keys(paths))
	sort.Strings(pathKeys)
	for _, pathKey := range pathKeys {
		item := paths[pathKey]
		methodKeys := slices.Collect(maps.Keys(item))
		sort.Strings(methodKeys)
		for _, method := range methodKeys {
			obj, ok := item[method].(map[string]any)
			if !ok {
				continue
			}
			id, _ := obj["operationId"].(string)
			if id == "" {
				continue
			}
			totalOps++

			verb, noun, slot := splitOperationID(id)
			if slot != "" {
				if _, exists := byNoun[noun]; !exists {
					byNoun[noun] = ops{}
				}
				if _, set := byNoun[noun][slot]; !set {
					byNoun[noun][slot] = id
				}
			} else if verb != "" {
				otherOps[noun] = append(otherOps[noun], id)
			}
		}
	}

	for _, scoped := range mapKeys(byNoun) {
		bare := stripScopePrefix(scoped)
		if bare == scoped {
			continue
		}
		if _, ok := byNoun[bare]; !ok {
			continue
		}
		for slot, opID := range byNoun[scoped] {
			if _, set := byNoun[bare][slot]; !set {
				byNoun[bare][slot] = opID
			}
		}
		otherOps[bare] = append(otherOps[bare], otherOps[scoped]...)
		delete(byNoun, scoped)
		delete(otherOps, scoped)
	}

	candidates = map[string]derivedOps{}
	for noun, o := range byNoun {
		// Drop update-kind suffixes (`_destroy`, `_preview`, `_update`,
		// `_refresh`). These are internal RPC variants of stack-update
		// operations, not resources. Real resources with disambiguating
		// path suffixes (`_esc_environments`, `_preview_environments`)
		// stay — they're hand-keyed by the example programs.
		if isUpdateKindNoun(noun) {
			continue
		}
		final := maps.Clone(o)
		// Upsert pattern: PUT /resource/{id} acts as create+update.
		if _, hasCreate := final["create"]; !hasCreate {
			if u, hasUpdate := final["update"]; hasUpdate {
				final["create"] = u
			}
		}
		_, hasCreate := final["create"]
		_, hasRead := final["read"]
		_, hasDelete := final["delete"]
		// Emit only if Create plus at least one of (Read, Delete).
		// Lone-Create endpoints are RPC-style actions, not resources.
		if !hasCreate || (!hasRead && !hasDelete) {
			slots := make([]string, 0, len(o))
			for k := range o {
				slots = append(slots, k)
			}
			sort.Strings(slots)
			skipped = append(skipped, fmt.Sprintf("%s [%s]", noun, strings.Join(slots, ",")))
			continue
		}
		tok := "pulumiservice:v2:" + noun
		candidates[tok] = derivedOps{
			Create: final["create"],
			Read:   final["read"],
			Update: final["update"],
			Delete: final["delete"],
		}
	}

	unmapped = map[string][]string{}
	for noun, ids := range otherOps {
		if _, derived := candidates["pulumiservice:v2:"+noun]; derived {
			continue
		}
		dedup := map[string]struct{}{}
		for _, x := range ids {
			dedup[x] = struct{}{}
		}
		uniq := make([]string, 0, len(dedup))
		for k := range dedup {
			uniq = append(uniq, k)
		}
		sort.Strings(uniq)
		unmapped[noun] = uniq
	}
	return candidates, totalOps, skipped, unmapped
}

// splitOperationID returns (verb, noun, slot). slot is "" when the verb
// isn't a CRUD verb (e.g., List, Cancel). verb is "" when no recognized
// prefix matches.
func splitOperationID(id string) (verb, noun, slot string) {
	for _, vp := range verbPrefixes {
		if matchPrefix(id, vp.prefix) {
			return vp.prefix, id[len(vp.prefix):], vp.slot
		}
	}
	for _, v := range nonCRUDVerbs {
		if matchPrefix(id, v) {
			return v, id[len(v):], ""
		}
	}
	return "", id, ""
}

// matchPrefix checks that prefix is at the start of id AND the next
// character is uppercase (so "Update" matches "UpdateStack" but not
// "Updater"). Returns false if prefix consumes the whole string.
func matchPrefix(id, prefix string) bool {
	if !strings.HasPrefix(id, prefix) {
		return false
	}
	if len(id) == len(prefix) {
		return false
	}
	return unicode.IsUpper(rune(id[len(prefix)]))
}

// isUpdateKindNoun reports whether a noun ends with an internal stack-
// update kind suffix. These are not user-facing resources.
func isUpdateKindNoun(noun string) bool {
	for _, suffix := range []string{"_destroy", "_preview", "_update", "_refresh"} {
		if strings.HasSuffix(noun, suffix) {
			return true
		}
	}
	return false
}

// stripScopePrefix removes a leading scope qualifier ("Org", "Team",
// "Project", "Stack", "User", "Organization") from a noun if doing so
// leaves a non-empty CamelCase remainder. Returns noun unchanged when
// no prefix matches.
func stripScopePrefix(noun string) string {
	for _, p := range scopePrefixes {
		if matchPrefix(noun, p) {
			return noun[len(p):]
		}
	}
	return noun
}

func mapKeys(m map[string]map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
