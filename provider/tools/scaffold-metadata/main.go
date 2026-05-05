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
// auto-generated (operations, idField, renames, outputsExclude, token)
// AND hand-curated (examples, descriptions, aliases, _excluded). The
// scaffolder rewrites the auto-generated fields on every regen and
// preserves the hand-curated ones via json.RawMessage round-tripping.
//
// To exclude a derived token, add it to the top-level `_excluded`
// array in metadata.json. The scaffolder also drops anything tagged
// with `x-pulumi-route-property.Visibility = "Deprecated"`.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"os"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/rest"
)

// pathParamPattern matches "{name}" placeholders in OpenAPI path strings.
var pathParamPattern = regexp.MustCompile(`\{([^/{}]+)\}`)

// unorderedFieldNames is the heuristic set of Pulumi-side field names that the
// scaffolder treats as set-like by default. Conservative — fields here have
// names that strongly imply order-insensitivity (tag lists, scope lists,
// thumbprints). Humans can flip individual fields off (or other fields on)
// via metadata.json's fields[name].unordered override.
var unorderedFieldNames = map[string]bool{
	"tags":          true,
	"scopes":        true,
	"groups":        true,
	"thumbprints":   true,
	"permissions":   true,
	"allowedRoles":  true,
	"allowedScopes": true,
}

// nameLikeBodyFields is the set of Pulumi-side body-field names the scaffolder
// recognizes as user-supplied unique identifiers — the signal for inferring
// deleteBeforeReplace. Path-param renames have already been applied by the
// time we check, so these names match what the user writes in code.
var nameLikeBodyFields = map[string]bool{
	"name":        true,
	"displayName": true,
	"hookName":    true,
	"tagName":     true,
	"teamName":    true,
}

// secretLikeNameSubstrings matches the runtime's looksSecret heuristic in
// rest/schema.go. Duplicated here to avoid widening the rest package's
// public API for a build-time tool.
var secretLikeNameSubstrings = []string{"secret", "tokenvalue", "password", "apikey", "accesstoken", "ciphertext"}

func looksSecretLike(name string) bool {
	lower := strings.ToLower(name)
	for _, sub := range secretLikeNameSubstrings {
		if strings.Contains(lower, sub) {
			return true
		}
	}
	return false
}

// verbPrefixes are operationId prefixes the scaffolder recognizes when
// extracting nouns. The slot column is a fallback used only when the
// HTTP method doesn't disambiguate (e.g., POST on an instance path that
// is logically an update). Order matters: longest first so "BatchCreate"
// wins over "Create".
var verbPrefixes = []struct {
	prefix string
	slot   string
}{
	{"BatchCreate", "create"}, {"BatchUpdate", "update"}, {"BatchDelete", "delete"},
	{"Create", "create"}, {"Register", "create"}, {"Add", "create"}, {"New", "create"},
	{"Update", "update"}, {"Patch", "update"}, {"Replace", "update"}, {"Put", "update"},
	{"Delete", "delete"}, {"Remove", "delete"},
	{"Get", "read"}, {"Read", "read"}, {"Describe", "read"},
	// Non-CRUD verbs: recognized so noun extraction works on action ops.
	{"List", ""}, {"Search", ""}, {"Find", ""}, {"Cancel", ""}, {"Approve", ""},
	{"Reject", ""}, {"Reset", ""}, {"Refresh", ""}, {"Restore", ""}, {"Validate", ""},
	{"Trigger", ""}, {"Poll", ""}, {"Open", ""}, {"Close", ""}, {"Encrypt", ""},
	{"Decrypt", ""}, {"Apply", ""}, {"Complete", ""}, {"Append", ""}, {"Bulk", ""},
	{"Accept", ""},
}

// scopePrefixes fold qualified nouns onto their bare form (OrgAgentPool →
// AgentPool, PulumiTeam → Team) when the unqualified resource also exists.
// Order matters: longest first so "Organization" wins over "Org".
var scopePrefixes = []string{
	"Organization", "Org", "Pulumi", "Team", "Project", "Stack", "User",
}

// moduleAliases maps URL-derived module paths to the user-facing Pulumi
// module name. Lives here rather than in metadata.json since it's a
// scaffolder-internal mapping driven by service URL conventions.
var moduleAliases = map[string]string{
	"agent-pools":          "agents",
	"auth/policies":        "auth",
	"esc/environments":     "esc",
	"oidc/issuers":         "auth",
	"preview/agents":       "agents",
	"preview/environments": "preview",
	"preview/insights":     "insights",
	"saml":                 "auth",
	"stacks/deployments":   "deployments",
	"teams/tokens":         "tokens",
}

// metadataDoc mirrors metadata.json. Resources serialize via RawMessage so
// the scaffolder doesn't drop fields it doesn't recognize.
type metadataDoc struct {
	Version   int                        `json:"version"`
	Package   string                     `json:"package,omitempty"`
	Note      string                     `json:"_note,omitempty"`
	Excluded  []string                   `json:"_excluded,omitempty"`
	Resources map[string]json.RawMessage `json:"resources"`
}

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
	candidates, stats := derive(rawSpec.Paths, true)
	// Re-derive without the deprecation filter so we can distinguish two
	// orphan classes: dropped-because-deprecated (auto-prunable) vs.
	// dropped-because-spec-changed-or-heuristic-miss (warn only).
	candidatesIfKept, _ := derive(rawSpec.Paths, false)

	excluded := map[string]bool{}
	for _, tok := range doc.Excluded {
		excluded[tok] = true
	}

	modules := deriveModules(candidates, parsedSpec)

	added, updated := 0, 0
	autoNameRecommendations := map[string]map[string]int{}
	for tok, ops := range candidates {
		if excluded[tok] {
			continue
		}
		entry, exists := doc.Resources[tok]
		renames := inferRenames(parsedSpec, opOrNil(parsedSpec, ops.Create), opOrNil(parsedSpec, ops.Read), opOrNil(parsedSpec, ops.Update), opOrNil(parsedSpec, ops.Delete))
		d := derivations{
			Renames:             renames,
			OutputsExclude:      inferOutputsExclude(parsedSpec, tok, ops),
			Token:               deriveToken(doc.Package, tok, modules[tok]),
			IDFormat:            inferIDFormat(parsedSpec, ops, renames),
			DeleteBeforeReplace: inferDeleteBeforeReplace(parsedSpec, ops, renames),
			EmitOnCreateFields:  inferEmitOnCreate(parsedSpec, ops, renames),
			UnorderedFields:     inferUnordered(parsedSpec, ops, renames),
		}
		if rec := inferAutoNameRecommendations(parsedSpec, ops, renames); len(rec) > 0 {
			autoNameRecommendations[tok] = rec
		}
		merged, changed, err := mergeOperations(entry, ops, d)
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

	// Split tokens that survived in metadata.json but didn't make it into
	// candidates: deprecation-induced ones are pruned; the rest just warn.
	var orphans, prunedDeprecated []string
	for tok := range doc.Resources {
		if _, ok := candidates[tok]; ok {
			continue
		}
		if excluded[tok] {
			continue
		}
		if _, wouldBeIfKept := candidatesIfKept[tok]; wouldBeIfKept {
			delete(doc.Resources, tok)
			prunedDeprecated = append(prunedDeprecated, tok)
		} else {
			orphans = append(orphans, tok)
		}
	}
	sort.Strings(orphans)
	sort.Strings(prunedDeprecated)

	if err := writeMetadata(*out, doc); err != nil {
		fail("write %s: %v", *out, err)
	}

	fmt.Fprintf(os.Stderr, "scaffold-metadata: %s\n", *out)
	fmt.Fprintf(os.Stderr, "  operations in spec:        %d\n", stats.totalOps)
	fmt.Fprintf(os.Stderr, "  candidates derived:        %d\n", len(candidates))
	fmt.Fprintf(os.Stderr, "  added new entries:         %d\n", added)
	fmt.Fprintf(os.Stderr, "  updated existing entries:  %d\n", updated)
	fmt.Fprintf(os.Stderr, "  excluded (_excluded):      %d\n", len(excluded))
	fmt.Fprintf(os.Stderr, "  excluded (Deprecated):     %d\n", len(stats.deprecated))
	fmt.Fprintf(os.Stderr, "  skipped (no Create+Read|Delete): %d\n", len(stats.skipped))
	if len(prunedDeprecated) > 0 {
		fmt.Fprintf(os.Stderr, "  auto-pruned (deprecated upstream): %d\n", len(prunedDeprecated))
		for _, o := range prunedDeprecated {
			fmt.Fprintf(os.Stderr, "    %s\n", o)
		}
	}
	if len(orphans) > 0 {
		fmt.Fprintf(os.Stderr, "  orphans (in metadata.json, not derived from spec): %d\n", len(orphans))
		for _, o := range orphans {
			fmt.Fprintf(os.Stderr, "    %s\n", o)
		}
	}
	if len(autoNameRecommendations) > 0 {
		fmt.Fprintf(os.Stderr, "  autoName candidates (maxLength from spec; opt in by hand-setting fields[<name>].autoName): %d\n", len(autoNameRecommendations))
		toks := make([]string, 0, len(autoNameRecommendations))
		for k := range autoNameRecommendations {
			toks = append(toks, k)
		}
		sort.Strings(toks)
		for _, tok := range toks {
			rec := autoNameRecommendations[tok]
			fields := make([]string, 0, len(rec))
			for f := range rec {
				fields = append(fields, f)
			}
			sort.Strings(fields)
			for _, f := range fields {
				fmt.Fprintf(os.Stderr, "    %s.%s -> %d\n", tok, f, rec[f])
			}
		}
	}
}

// inferIDFormat builds a "{paramA}/{paramB}" template from the most
// authoritative non-create op's path (read → update → delete → create).
// Path-param wire names are translated through Renames so the template
// matches what the user writes in metadata. Returns "" when no op carries
// path params (true singletons; rare).
func inferIDFormat(spec *rest.Spec, ops derivedOps, renames map[string]string) string {
	for _, opID := range []string{ops.Read, ops.Update, ops.Delete, ops.Create} {
		op := opOrNil(spec, opID)
		if op == nil {
			continue
		}
		matches := pathParamPattern.FindAllStringSubmatch(op.Path, -1)
		if len(matches) == 0 {
			continue
		}
		var parts []string
		for _, m := range matches {
			parts = append(parts, "{"+wireToPulumi(m[1], renames)+"}")
		}
		return strings.Join(parts, "/")
	}
	return ""
}

// inferEmitOnCreate finds fields in the create response that don't appear in
// the read response and that match the secret-naming heuristic. Returns
// Pulumi-side names so the metadata's fields[name] key matches what users
// write.
func inferEmitOnCreate(spec *rest.Spec, ops derivedOps, renames map[string]string) []string {
	createOp := opOrNil(spec, ops.Create)
	if createOp == nil || createOp.ResponseRef == "" {
		return nil
	}
	createFields := flattenedProps(spec, createOp.ResponseRef)
	if len(createFields) == 0 {
		return nil
	}
	var readFields map[string]any
	if readOp := opOrNil(spec, ops.Read); readOp != nil {
		readFields = flattenedProps(spec, readOp.ResponseRef)
	}
	var out []string
	for wireName := range createFields {
		if _, inRead := readFields[wireName]; inRead {
			continue
		}
		pulName := wireToPulumi(wireName, renames)
		if looksSecretLike(pulName) {
			out = append(out, pulName)
		}
	}
	sort.Strings(out)
	return out
}

// inferDeleteBeforeReplace returns true when a duplicate create would collide
// upstream — i.e., POST method with a free-form name in the body or with a
// resource-name path param. PUT/PATCH-as-create resources (configuration
// singletons) skip this because they're idempotent by design.
func inferDeleteBeforeReplace(spec *rest.Spec, ops derivedOps, renames map[string]string) bool {
	createOp := opOrNil(spec, ops.Create)
	if createOp == nil || createOp.Method != "POST" {
		return false
	}
	if createOp.RequestRef != "" {
		bodyProps := flattenedProps(spec, createOp.RequestRef)
		for wireName := range bodyProps {
			if nameLikeBodyFields[wireToPulumi(wireName, renames)] {
				return true
			}
		}
	}
	for _, p := range pathParamsOf(createOp) {
		pul := wireToPulumi(p, renames)
		if pul == "orgName" || pul == "organizationName" {
			continue
		}
		if nameLikeBodyFields[pul] || strings.HasSuffix(pul, "Name") || strings.HasSuffix(pul, "name") {
			return true
		}
	}
	return false
}

// inferUnordered scans the create-op body for array fields whose Pulumi-side
// name matches the unorderedFieldNames heuristic. Conservative — known
// set-like names only; humans add others via the metadata flag.
func inferUnordered(spec *rest.Spec, ops derivedOps, renames map[string]string) []string {
	createOp := opOrNil(spec, ops.Create)
	if createOp == nil || createOp.RequestRef == "" {
		return nil
	}
	bodyProps := flattenedProps(spec, createOp.RequestRef)
	var out []string
	for wireName, raw := range bodyProps {
		prop, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if t, _ := prop["type"].(string); t != "array" {
			continue
		}
		pulName := wireToPulumi(wireName, renames)
		if unorderedFieldNames[pulName] {
			out = append(out, pulName)
		}
	}
	sort.Strings(out)
	return out
}

// inferAutoNameRecommendations returns a map of name-like body field → its
// declared maxLength. The scaffolder doesn't auto-enable autoName (UX choice
// per resource), but reporting the cap lets a human opt in without grepping
// the spec.
func inferAutoNameRecommendations(spec *rest.Spec, ops derivedOps, renames map[string]string) map[string]int {
	createOp := opOrNil(spec, ops.Create)
	if createOp == nil || createOp.RequestRef == "" {
		return nil
	}
	bodyProps := flattenedProps(spec, createOp.RequestRef)
	out := map[string]int{}
	for wireName, raw := range bodyProps {
		prop, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if t, _ := prop["type"].(string); t != "string" {
			continue
		}
		pulName := wireToPulumi(wireName, renames)
		if !nameLikeBodyFields[pulName] {
			continue
		}
		switch v := prop["maxLength"].(type) {
		case float64:
			out[pulName] = int(v)
		case json.Number:
			if n, err := v.Int64(); err == nil {
				out[pulName] = int(n)
			}
		}
	}
	return out
}

// wireToPulumi inverts a Pulumi→wire renames map: given a wire-side OpenAPI
// name, return the matching Pulumi-side name when a rename targets it,
// otherwise the input unchanged.
func wireToPulumi(wireName string, renames map[string]string) string {
	for pul, wire := range renames {
		if wire == wireName {
			return pul
		}
	}
	return wireName
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

// derivations bundles the auto-derived per-resource metadata the scaffolder
// fills in. Every field follows write-if-absent semantics in mergeOperations —
// humans can override any of them in metadata.json and the override survives
// regen.
type derivations struct {
	Renames             map[string]string
	OutputsExclude      []string
	Token               string
	IDFormat            string
	DeleteBeforeReplace bool
	EmitOnCreateFields  []string // Pulumi-side field names
	UnorderedFields     []string // Pulumi-side field names
}

// mergeOperations layers derived fields onto an existing per-resource entry.
// The `operations` block is replaced wholesale; everything else is written
// only when absent so hand-edits survive regen.
func mergeOperations(existing json.RawMessage, ops derivedOps, d derivations) (json.RawMessage, bool, error) {
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

	if len(d.Renames) > 0 {
		merged := map[string]any{}
		for k, v := range d.Renames {
			merged[k] = v
		}
		if existing, ok := entry["renames"].(map[string]any); ok {
			maps.Copy(merged, existing)
		}
		entry["renames"] = merged
	}
	if len(d.OutputsExclude) > 0 {
		if _, has := entry["outputsExclude"]; !has {
			arr := make([]any, len(d.OutputsExclude))
			for i, v := range d.OutputsExclude {
				arr[i] = v
			}
			entry["outputsExclude"] = arr
		}
	}
	if d.Token != "" {
		if _, has := entry["token"]; !has {
			entry["token"] = d.Token
		}
	}
	if d.IDFormat != "" {
		if _, has := entry["idFormat"]; !has {
			entry["idFormat"] = d.IDFormat
		}
	}
	if d.DeleteBeforeReplace {
		if _, has := entry["deleteBeforeReplace"]; !has {
			entry["deleteBeforeReplace"] = true
		}
	}
	for _, name := range d.EmitOnCreateFields {
		setFieldFlag(entry, name, "emitOnCreate", true)
	}
	for _, name := range d.UnorderedFields {
		setFieldFlag(entry, name, "unordered", true)
	}

	encoded, err := encodeStable(entry)
	if err != nil {
		return nil, false, err
	}
	return encoded, changed, nil
}

// setFieldFlag sets entry.fields[fieldName].flag = value, but only when the
// flag isn't already present on that field. Creates the nested maps lazily.
func setFieldFlag(entry map[string]any, fieldName, flag string, value any) {
	fields, _ := entry["fields"].(map[string]any)
	if fields == nil {
		fields = map[string]any{}
	}
	fm, _ := fields[fieldName].(map[string]any)
	if fm == nil {
		fm = map[string]any{}
	}
	if _, has := fm[flag]; has {
		return
	}
	fm[flag] = value
	fields[fieldName] = fm
	entry["fields"] = fields
}

func opOrNil(spec *rest.Spec, id string) *rest.Operation {
	if id == "" {
		return nil
	}
	op, _ := spec.Op(id)
	return op
}

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
			maps.Copy(out, props)
		}
	}
	if root, ok := spec.ResolveSchema(ref); ok {
		walk(root)
	}
	return out
}

func responseHasField(spec *rest.Spec, op *rest.Operation, field string) bool {
	if op == nil {
		return false
	}
	_, has := flattenedProps(spec, op.ResponseRef)[field]
	return has
}

func requestHasField(spec *rest.Spec, op *rest.Operation, field string) bool {
	if op == nil {
		return false
	}
	_, has := flattenedProps(spec, op.RequestRef)[field]
	return has
}

// inferOutputsExclude flags response-body envelope fields that collide with
// the resource's own type name (a "service" field on the Service resource).
// Pulumi's per-language SDK codegen rejects these as member-vs-class
// collisions.
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

// splitTrailingSuffix breaks a CamelCased path param like "tagName" into
// ("tag", "name", true). Returns ok=false when no recognized suffix.
func splitTrailingSuffix(name string) (prefix, bareSuffix string, ok bool) {
	for _, s := range []string{"Name", "ID", "Id"} {
		if strings.HasSuffix(name, s) && len(name) > len(s) {
			return name[:len(name)-len(s)], strings.ToLower(s), true
		}
	}
	return "", "", false
}

// inferRenames picks renames for path parameters that aren't already on the
// create op's path. Three patterns:
//
//  1. Server-id (path uses noun-prefixed form, body returns "id"): emit
//     {path-param → "id"}.
//  2. Body field mirrors path under suffix-strip ({tagName} ↔ body `name`):
//     emit {body-field → path-param}.
//  3. Bare path param matched by body `name` (catch-all when (1)/(2) miss).
//
// Plus: rule (4) for verbose body aliases (`organizationName` body ↔
// `orgName` path), detected by suffix match + stem prefix relationship.
func inferRenames(spec *rest.Spec, createOp, readOp, updateOp, deleteOp *rest.Operation) map[string]string {
	createParams := map[string]bool{}
	for _, p := range pathParamsOf(createOp) {
		createParams[p] = true
	}

	out := map[string]string{}
	seen := map[string]bool{}

	// Pass 1: precise matches (rules 1, 2). These claim body keys before the
	// catch-all rule (3) gets a chance.
	for _, op := range []*rest.Operation{readOp, updateOp, deleteOp} {
		for _, p := range pathParamsOf(op) {
			if createParams[p] || seen[p] {
				continue
			}
			// Path param that's also a body field is a duplicate, not an alias.
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

	// Pass 2: rule (3) — bare path param matched by body `name`.
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

	// Rule (4): body field that duplicates a path param under a verbose alias.
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

// encodeStable marshals a map with deterministic key ordering: a preferred
// order for known fields, then alphabetical for the rest.
func encodeStable(entry map[string]any) (json.RawMessage, error) {
	preferred := []string{"operations", "idField", "idFormat", "deleteBeforeReplace", "token", "aliases", "renames", "fields", "outputs", "outputsExclude", "description", "examples"}
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

func indentJSON(raw json.RawMessage, prefix string) ([]byte, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	return json.MarshalIndent(v, prefix, "  ")
}

type deriveStats struct {
	totalOps   int
	skipped    []string
	deprecated []string
}

// derive walks every operation and returns the candidate resources we'd
// emit. Resources are grouped by operationId noun; each candidate's slot
// (create/read/update/delete) comes from the HTTP method, with the verb
// prefix as a tiebreaker for ambiguous POSTs. When filterDeprecated is
// false, ops marked deprecated upstream still contribute to candidates —
// used by the orphan-classification pass to detect which entries dropped
// out specifically because of the deprecation filter.
func derive(paths map[string]map[string]any, filterDeprecated bool) (map[string]derivedOps, deriveStats) {
	type ops = map[string]string
	byNoun := map[string]ops{}
	otherOps := map[string][]string{}
	stats := deriveStats{}

	pathKeys := slices.Sorted(maps.Keys(paths))
	for _, pathKey := range pathKeys {
		item := paths[pathKey]
		methodKeys := slices.Sorted(maps.Keys(item))
		for _, method := range methodKeys {
			obj, ok := item[method].(map[string]any)
			if !ok {
				continue
			}
			id, _ := obj["operationId"].(string)
			if id == "" {
				continue
			}
			stats.totalOps++

			if isDeprecated(obj) {
				stats.deprecated = append(stats.deprecated, id)
				if filterDeprecated {
					continue
				}
			}

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

	// Scope-prefix folding: OrgAgentPool → AgentPool when both nouns exist
	// but the unqualified one lacks a Create op.
	for _, longer := range mapKeys(byNoun) {
		for shorter := range byNoun {
			if shorter == longer || !strings.HasPrefix(longer, shorter+"_") {
				continue
			}
			if _, hasCreate := byNoun[shorter]["create"]; hasCreate {
				continue
			}
			c, ok := byNoun[longer]["create"]
			if !ok {
				continue
			}
			byNoun[shorter]["create"] = c
			otherOps[shorter] = append(otherOps[shorter], otherOps[longer]...)
			delete(byNoun, longer)
			delete(otherOps, longer)
			break
		}
	}

	// Plural folding: "Tasks" plural feeds non-read slots into "Task" singular.
	for _, plural := range mapKeys(byNoun) {
		if !strings.HasSuffix(plural, "s") || len(plural) <= 1 {
			continue
		}
		singular := strings.TrimSuffix(plural, "s")
		if _, ok := byNoun[singular]; !ok {
			continue
		}
		for slot, opID := range byNoun[plural] {
			if slot == "read" {
				continue
			}
			if _, set := byNoun[singular][slot]; !set {
				byNoun[singular][slot] = opID
			}
		}
		hasOnlyReadLeft := true
		for slot := range byNoun[plural] {
			if slot != "read" {
				hasOnlyReadLeft = false
				break
			}
		}
		otherOps[singular] = append(otherOps[singular], otherOps[plural]...)
		if hasOnlyReadLeft {
			otherOps[singular] = append(otherOps[singular], byNoun[plural]["read"])
		}
		delete(byNoun, plural)
		delete(otherOps, plural)
	}

	// Scope-prefix-strip: OrgX → X when both exist and X has no own Create.
	for _, scoped := range mapKeys(byNoun) {
		bare := stripScopePrefix(scoped)
		if bare == scoped {
			continue
		}
		if _, ok := byNoun[bare]; !ok {
			continue
		}
		if _, hasOwnCreate := byNoun[bare]["create"]; hasOwnCreate {
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

	candidates := map[string]derivedOps{}
	for noun, o := range byNoun {
		if isUpdateKindNoun(noun) {
			continue
		}
		final := maps.Clone(o)
		// Upsert: PUT /resource/{id} acts as create+update.
		if _, hasCreate := final["create"]; !hasCreate {
			if u, hasUpdate := final["update"]; hasUpdate {
				final["create"] = u
			}
		}
		_, hasCreate := final["create"]
		_, hasRead := final["read"]
		_, hasDelete := final["delete"]
		if !hasCreate || (!hasRead && !hasDelete) {
			slots := make([]string, 0, len(o))
			for k := range o {
				slots = append(slots, k)
			}
			sort.Strings(slots)
			stats.skipped = append(stats.skipped, fmt.Sprintf("%s [%s]", noun, strings.Join(slots, ",")))
			continue
		}
		candidates["pulumiservice:v2:"+noun] = derivedOps{
			Create: final["create"],
			Read:   final["read"],
			Update: final["update"],
			Delete: final["delete"],
		}
	}

	return candidates, stats
}

// isDeprecated returns true when an operation is marked deprecated, either via
// the OpenAPI standard `deprecated: true` boolean or the legacy Pulumi-custom
// `x-pulumi-route-property.Visibility = "Deprecated"` extension. Preview
// endpoints still scaffold so the provider can expose them behind their own
// resources; only deprecated routes are filtered.
func isDeprecated(op map[string]any) bool {
	if dep, _ := op["deprecated"].(bool); dep {
		return true
	}
	rp, _ := op["x-pulumi-route-property"].(map[string]any)
	if rp == nil {
		return false
	}
	v, _ := rp["Visibility"].(string)
	return v == "Deprecated"
}

// deriveModules assigns each candidate a (module, originalPrefix) pair when
// at least two candidates share the same first-or-aliased URL segment. The
// module ends up in the user-facing token (pulumiservice:v2/<module>:Type).
func deriveModules(candidates map[string]derivedOps, spec *rest.Spec) map[string]moduleAssignment {
	type pair struct{ canonical, original string }
	prefixes := map[string]pair{}
	for tok, ops := range candidates {
		op, ok := spec.Op(ops.Create)
		if !ok {
			continue
		}
		original := routePrefix(op.Path)
		canonical := original
		if alias, ok := moduleAliases[canonical]; ok {
			canonical = alias
		} else if i := strings.Index(canonical, "/"); i > 0 {
			canonical = canonical[:i]
		}
		prefixes[tok] = pair{canonical, original}
	}
	counts := map[string]int{}
	for _, p := range prefixes {
		counts[p.canonical]++
	}
	out := map[string]moduleAssignment{}
	for tok, p := range prefixes {
		if counts[p.canonical] >= 2 {
			out[tok] = moduleAssignment{Module: p.canonical, OriginalPrefix: p.original}
		}
	}
	return out
}

type moduleAssignment struct {
	Module         string
	OriginalPrefix string
}

func routePrefix(path string) string {
	var parts []string
	for _, seg := range strings.Split(strings.Trim(path, "/"), "/") {
		if seg == "" || strings.HasPrefix(seg, "{") {
			continue
		}
		switch seg {
		case "api", "orgs", "user", "console":
			continue
		}
		parts = append(parts, seg)
		if len(parts) == 2 {
			break
		}
	}
	return strings.Join(parts, "/")
}

// deriveToken constructs the user-facing Pulumi token. Empty when the
// candidate is a singleton (no shared module); resource then stays at the
// metadata-key form. Type name is shortened by dropping a redundant module
// prefix (AgentPool in module agents → Pool) and a trailing route-mangled
// suffix (Environment_esc_environments in module esc → Environment).
func deriveToken(pkg, key string, ma moduleAssignment) string {
	if ma.Module == "" {
		return ""
	}
	typ := key
	if i := strings.LastIndex(key, ":"); i >= 0 {
		typ = key[i+1:]
	}
	typ = stripRouteSuffix(typ, ma.OriginalPrefix)
	typ = stripModulePrefix(typ, ma.Module)
	return fmt.Sprintf("%s:v2/%s:%s", pkg, ma.Module, typ)
}

func stripRouteSuffix(typ, originalPrefix string) string {
	if originalPrefix == "" {
		return typ
	}
	suffix := "_" + strings.ReplaceAll(originalPrefix, "/", "_")
	if strings.HasSuffix(typ, suffix) && len(typ) > len(suffix) {
		return typ[:len(typ)-len(suffix)]
	}
	return typ
}

func stripModulePrefix(typ, module string) string {
	base := module
	if i := strings.LastIndex(module, "/"); i >= 0 {
		base = module[i+1:]
	}
	for _, c := range []string{base, strings.TrimSuffix(base, "s")} {
		if c == "" {
			continue
		}
		cap := strings.ToUpper(c[:1]) + c[1:]
		if !strings.HasPrefix(typ, cap) || len(typ) <= len(cap) {
			continue
		}
		if r := rune(typ[len(cap)]); r >= 'A' && r <= 'Z' {
			return typ[len(cap):]
		}
	}
	return typ
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
	return "", id, ""
}

// matchPrefix checks that prefix is at the start of id AND the next char
// is uppercase (so "Update" matches "UpdateStack" not "Updater"). Returns
// false if prefix consumes the whole string.
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
// update kind suffix. These are RPC variants, not user-facing resources.
func isUpdateKindNoun(noun string) bool {
	for _, suffix := range []string{"_destroy", "_preview", "_update", "_refresh"} {
		if strings.HasSuffix(noun, suffix) {
			return true
		}
	}
	return false
}

// stripScopePrefix removes a leading scope qualifier ("Org", "Team", etc.)
// from a noun if doing so leaves a non-empty CamelCase remainder.
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
