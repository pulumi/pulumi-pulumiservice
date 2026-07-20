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

// Command scaffold-metadata derives api resource operations from the
// embedded OpenAPI spec and merges them into metadata.json in place.
//
// metadata.json is the single source of truth at runtime. It is BOTH
// auto-generated (operations, idFormat, renames, outputsExclude, token)
// AND hand-curated (examples, descriptions, aliases, _excluded). The
// scaffolder rewrites the auto-generated fields on every regen and
// preserves the hand-curated ones via json.RawMessage round-tripping.
//
// To exclude a derived token, add it to the top-level `_excluded`
// array in metadata.json.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"net/http"
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

// OpenAPI schema "type" discriminator values.
const (
	typeArray    = "array"
	typeString   = "string"
	opCreate     = "create"
	opRead       = "read"
	opUpdate     = "update"
	opDelete     = "delete"
	authModule   = "auth"
	nameFieldKey = "name"
)

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
	nameFieldKey:  true,
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
	{"BatchCreate", opCreate}, {"BatchUpdate", opUpdate}, {"BatchDelete", opDelete},
	{"Create", opCreate}, {"Register", opCreate}, {"Add", opCreate}, {"New", opCreate},
	{"Update", opUpdate}, {"Patch", opUpdate}, {"Replace", opUpdate}, {"Put", opUpdate},
	{"Delete", opDelete}, {"Remove", opDelete},
	{"Get", opRead}, {"Read", opRead}, {"Describe", opRead},
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

// verboseBodyAliases pairs the lower-cased stem of a path-param prefix with
// every body-field prefix that's recognized as a verbose alias of it (rule 4
// in inferRenames). Strict allowlist — a HasPrefix match would also pair
// "organizationalUnitName" against "{orgName}", inventing a rename out of
// thin air.
var verboseBodyAliases = map[string][]string{
	"org": {"organization"},
}

// moduleAliases maps URL-derived module paths to the user-facing Pulumi
// module name. Lives here rather than in metadata.json since it's a
// scaffolder-internal mapping driven by service URL conventions.
var moduleAliases = map[string]string{
	"agent-pools":        "agents",
	"auth/policies":      authModule,
	"esc/environments":   "esc",
	"oidc/issuers":       authModule,
	"preview/agents":     "agents",
	"preview/insights":   "insights",
	"saml":               authModule,
	"stacks/deployments": "deployments",
	"teams/tokens":       "tokens",
}

// metadataDoc mirrors metadata.json. Resources serialize via RawMessage so
// the scaffolder doesn't drop fields it doesn't recognize.
type metadataDoc struct {
	Package   string                     `json:"package,omitempty"`
	Note      string                     `json:"_note,omitempty"`
	Excluded  []string                   `json:"_excluded,omitempty"`
	Resources map[string]json.RawMessage `json:"resources"`
}

// unmappedFieldSet holds update-body wire fields no input can populate,
// split by whether the schema marks them required.
type unmappedFieldSet struct {
	required, optional []string
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
	candidates, stats := derive(rawSpec.Paths)

	excluded := map[string]bool{}
	for _, tok := range doc.Excluded {
		excluded[tok] = true
	}

	modules := deriveModules(candidates, parsedSpec)

	added, updated := 0, 0
	autoNameRecommendations := map[string]map[string]int{}
	unmappedUpdates := map[string]unmappedFieldSet{}
	for tok, ops := range candidates {
		if excluded[tok] {
			continue
		}
		entry, exists := doc.Resources[tok]
		renames := inferRenames(parsedSpec,
			opOrNil(parsedSpec, ops.Create),
			opOrNil(parsedSpec, ops.Read),
			opOrNil(parsedSpec, ops.Update),
			opOrNil(parsedSpec, ops.Delete))
		d := derivations{
			Renames:             renames,
			OutputsExclude:      inferOutputsExclude(parsedSpec, tok, ops),
			Token:               deriveToken(doc.Package, tok, modules[tok]),
			IDFormat:            inferIDFormat(parsedSpec, ops, renames),
			DeleteBeforeReplace: inferDeleteBeforeReplace(parsedSpec, ops, renames),
			RequireImport:       inferRequireImport(parsedSpec, ops),
			UpdateEnvelope:      inferUpdateEnvelope(parsedSpec, ops),
			EmitOnCreateFields:  inferEmitOnCreate(parsedSpec, ops, renames),
			UnorderedFields:     inferUnordered(parsedSpec, ops, renames),
		}
		if rec := inferAutoNameRecommendations(parsedSpec, ops, renames); len(rec) > 0 {
			autoNameRecommendations[tok] = rec
		}
		if d.UpdateEnvelope == nil && !rawHasKey(entry, "updateEnvelope") {
			if reqU, optU := unmappedUpdateFields(parsedSpec, ops, renames); len(reqU)+len(optU) > 0 {
				unmappedUpdates[tok] = unmappedFieldSet{required: reqU, optional: optU}
			}
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

	// Attachment pass: any resource whose update op carries symmetric
	// add<X>/remove<X> body fields (an imperative membership edge) gets a
	// synthesized attachment resource per pair. Spec-wide and parent-agnostic
	attachmentTokens := map[string]bool{}
	attachAdded, attachChanged, attachSkipped := 0, 0, 0
	for _, tok := range slices.Sorted(maps.Keys(candidates)) {
		if excluded[tok] {
			continue
		}
		ops := candidates[tok]
		updOp, readOp := opOrNil(parsedSpec, ops.Update), opOrNil(parsedSpec, ops.Read)
		if updOp == nil || readOp == nil {
			continue
		}
		parentToken := deriveToken(doc.Package, tok, modules[tok])
		pairs, skips := detectAttachmentPairs(parsedSpec, updOp, readOp)
		attachSkipped += skips
		for _, pair := range pairs {
			atTok := tok + pair.EdgeStem + "Attachment"
			if excluded[atTok] {
				continue
			}
			_, existed := doc.Resources[atTok]
			merged, changed, err := mergeAttachment(doc.Resources[atTok], attachmentDerivation{
				Token:           parentToken + pair.EdgeStem + "Attachment",
				MutationOp:      ops.Update,
				ReadOp:          ops.Read,
				AddField:        pair.AddField,
				RemoveField:     pair.RemoveField,
				MembershipField: pair.MembershipField,
				MatchKey:        pair.MatchKey,
				IDFormat:        attachmentIDFormat(updOp, pair.MatchKey),
			})
			if err != nil {
				fail("merge attachment %s: %v", atTok, err)
			}
			doc.Resources[atTok] = merged
			attachmentTokens[atTok] = true
			if !existed {
				added++
				attachAdded++
			} else if changed {
				updated++
				attachChanged++
			}
		}
	}

	// Tokens that survived in metadata.json but didn't make it into candidates
	// are reported as orphans — typically a spec change or heuristic miss.
	var orphans []string
	for tok := range doc.Resources {
		if _, ok := candidates[tok]; ok {
			continue
		}
		if excluded[tok] || attachmentTokens[tok] {
			continue
		}
		orphans = append(orphans, tok)
	}
	sort.Strings(orphans)

	if err := writeMetadata(*out, doc); err != nil {
		fail("write %s: %v", *out, err)
	}

	fmt.Fprintf(os.Stderr, "scaffold-metadata: %s\n", *out)
	fmt.Fprintf(os.Stderr, "  operations in spec:        %d\n", stats.totalOps)
	fmt.Fprintf(os.Stderr, "  candidates derived:        %d\n", len(candidates))
	fmt.Fprintf(os.Stderr, "  added new entries:         %d\n", added)
	fmt.Fprintf(os.Stderr, "  updated existing entries:  %d\n", updated)
	fmt.Fprintf(os.Stderr, "  excluded (_excluded):      %d\n", len(excluded))
	fmt.Fprintf(os.Stderr, "  deprecated (kept):         %d\n", len(stats.deprecated))
	fmt.Fprintf(os.Stderr, "  skipped (no Create+Read|Delete): %d\n", len(stats.skipped))
	fmt.Fprintf(os.Stderr, "  attachment resources emitted: %d new, %d updated (add/remove pairs skipped: %d)\n",
		attachAdded, attachChanged, attachSkipped)
	if len(orphans) > 0 {
		fmt.Fprintf(os.Stderr, "  orphans (in metadata.json, not derived from spec): %d\n", len(orphans))
		for _, o := range orphans {
			fmt.Fprintf(os.Stderr, "    %s\n", o)
		}
	}
	if len(unmappedUpdates) > 0 {
		fmt.Fprintf(os.Stderr,
			"  update-body fields no input can populate (omitted from PATCHes; REQUIRED ones break every update): %d resources\n",
			len(unmappedUpdates))
		for _, tok := range slices.Sorted(maps.Keys(unmappedUpdates)) {
			u := unmappedUpdates[tok]
			for _, f := range u.required {
				fmt.Fprintf(os.Stderr, "    %s.%s (REQUIRED)\n", tok, f)
			}
			for _, f := range u.optional {
				fmt.Fprintf(os.Stderr, "    %s.%s\n", tok, f)
			}
		}
	}
	if len(autoNameRecommendations) > 0 {
		fmt.Fprintf(os.Stderr,
			"  autoName candidates (maxLength from spec; opt in by hand-setting fields[<name>].autoName): %d\n",
			len(autoNameRecommendations))
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

// inferRequireImport returns true when the create op is upsert-shaped:
//   - HTTP method PUT or PATCH (configuration singleton), OR
//   - create operationId equals update operationId (e.g., the upsert
//     promotion in derive() below, or a hand-curated mapping).
//
// Both signals indicate the upstream API has no separate "create" — calling
// it twice silently overwrites instead of erroring on conflict (Pulumi#9925).
// The runtime uses this flag to gate Create on a pre-flight read probe.
func inferRequireImport(spec *rest.Spec, ops derivedOps) bool {
	if ops.Create != "" && ops.Create == ops.Update {
		return true
	}
	createOp := opOrNil(spec, ops.Create)
	if createOp == nil {
		return false
	}
	switch createOp.Method {
	case http.MethodPut, http.MethodPatch:
		return true
	}
	return false
}

// inferUpdateEnvelope detects update request bodies that pair the prior
// values with the desired ones for optimistic concurrency — e.g.
// UpdateEnvironmentTag's {currentTag, newTag}: exactly two object-typed
// properties named current<Stem>/new<Stem> with a shared stem. The generic
// flat input→body mapping matches neither wrapper (it would send an empty
// body), so the runtime hand-shapes the body from prior state and new
// inputs; see rest.UpdateEnvelopeMeta.
func inferUpdateEnvelope(spec *rest.Spec, ops derivedOps) *rest.UpdateEnvelopeMeta {
	upd := opOrNil(spec, ops.Update)
	if upd == nil || ops.Update == ops.Create || upd.RequestRef == "" {
		return nil
	}
	props := flattenedProps(spec, upd.RequestRef)
	if len(props) != 2 {
		return nil
	}
	var currentField, currentStem, newField, newStem string
	for name := range props {
		if stem, ok := attachmentStem(name, "current"); ok {
			currentField, currentStem = name, stem
		} else if stem, ok := attachmentStem(name, "new"); ok {
			newField, newStem = name, stem
		}
	}
	if currentField == "" || newField == "" || currentStem != newStem {
		return nil
	}
	if !isObjectProp(spec, props[currentField]) || !isObjectProp(spec, props[newField]) {
		return nil
	}
	return &rest.UpdateEnvelopeMeta{CurrentField: currentField, NewField: newField}
}

// isObjectProp reports whether prop resolves to an object schema with at
// least one property, via $ref or inline.
func isObjectProp(spec *rest.Spec, prop any) bool {
	if ref := refOf(prop); ref != "" {
		return len(flattenedProps(spec, ref)) > 0
	}
	m, ok := prop.(map[string]any)
	if !ok {
		return false
	}
	props, ok := m["properties"].(map[string]any)
	return ok && len(props) > 0
}

// inferDeleteBeforeReplace returns true when a duplicate create would collide
// upstream — i.e., POST method with a free-form name in the body or with a
// resource-name path param. PUT/PATCH-as-create resources (configuration
// singletons) skip this because they're idempotent by design.
func inferDeleteBeforeReplace(spec *rest.Spec, ops derivedOps, renames map[string]string) bool {
	createOp := opOrNil(spec, ops.Create)
	if createOp == nil || createOp.Method != http.MethodPost {
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
		if nameLikeBodyFields[pul] || strings.HasSuffix(pul, "Name") || strings.HasSuffix(pul, nameFieldKey) {
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
	RequireImport       bool
	UpdateEnvelope      *rest.UpdateEnvelopeMeta
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
		newOps[opCreate] = ops.Create
	}
	if ops.Read != "" {
		newOps[opRead] = ops.Read
	}
	if ops.Update != "" {
		newOps[opUpdate] = ops.Update
	}
	if ops.Delete != "" {
		newOps[opDelete] = ops.Delete
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
		if existing, has := entry["token"].(string); has {
			if existing != d.Token {
				fmt.Fprintf(os.Stderr,
					"  INFO: token pinned at %q; heuristic now suggests %q (preserving pin)\n",
					existing, d.Token)
			}
		} else {
			entry["token"] = d.Token
		}
	}

	if d.IDFormat != "" {
		if existing, has := entry["idFormat"].(string); has {
			if existing != d.IDFormat {
				fmt.Fprintf(os.Stderr,
					"  INFO: idFormat pinned at %q; heuristic now suggests %q (preserving pin)\n",
					existing, d.IDFormat)
			}
		} else {
			entry["idFormat"] = d.IDFormat
		}
	}
	if d.DeleteBeforeReplace {
		if _, has := entry["deleteBeforeReplace"]; !has {
			entry["deleteBeforeReplace"] = true
		}
	}
	if d.RequireImport {
		if _, has := entry["requireImport"]; !has {
			entry["requireImport"] = true
		}
	}
	if d.UpdateEnvelope != nil {
		if _, has := entry["updateEnvelope"]; !has {
			entry["updateEnvelope"] = map[string]any{
				"currentField": d.UpdateEnvelope.CurrentField,
				"newField":     d.UpdateEnvelope.NewField,
			}
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
	props, _ := flattenedPropsRequired(spec, ref)
	return props
}

// flattenedPropsRequired is flattenedProps plus the union of required-field
// names encountered during the walk.
func flattenedPropsRequired(spec *rest.Spec, ref string) (map[string]any, map[string]bool) {
	if ref == "" {
		return nil, nil
	}
	out := map[string]any{}
	required := map[string]bool{}
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
		if rr, ok := node["required"].([]any); ok {
			for _, r := range rr {
				if rs, ok := r.(string); ok {
					required[rs] = true
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
	return out, required
}

// unmappedUpdateFields returns the update-op request-body fields (wire-side,
// sorted) that no resource input can populate. The runtime's flat body
// mapping silently omits them from PATCHes: a REQUIRED one means every
// update fails upstream, an optional one is a server capability updates can
// never exercise. Callers exempt updateEnvelope resources — their body is
// hand-shaped from prior state and new inputs, not the flat mapping.
func unmappedUpdateFields(spec *rest.Spec, ops derivedOps, renames map[string]string) (required, optional []string) {
	upd := opOrNil(spec, ops.Update)
	if upd == nil || ops.Update == ops.Create || upd.RequestRef == "" {
		return nil, nil
	}
	inputs := map[string]bool{}
	if cr := opOrNil(spec, ops.Create); cr != nil {
		for wire := range flattenedProps(spec, cr.RequestRef) {
			inputs[wireToPulumi(wire, renames)] = true
		}
	}
	for _, opID := range []string{ops.Create, ops.Read, ops.Update, ops.Delete} {
		for _, name := range pathParamsOf(opOrNil(spec, opID)) {
			inputs[wireToPulumi(name, renames)] = true
		}
	}
	props, requiredSet := flattenedPropsRequired(spec, upd.RequestRef)
	for wire := range props {
		if inputs[wireToPulumi(wire, renames)] {
			continue
		}
		if requiredSet[wire] {
			required = append(required, wire)
		} else {
			optional = append(optional, wire)
		}
	}
	sort.Strings(required)
	sort.Strings(optional)
	return required, optional
}

// rawHasKey reports whether a raw JSON object contains key at the top level.
func rawHasKey(raw json.RawMessage, key string) bool {
	if len(raw) == 0 {
		return false
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return false
	}
	_, ok := m[key]
	return ok
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

// --- Attachment detection ------------------------------------------------
//
// An attachment edge is an imperative membership pattern: an update op whose
// request body carries paired add<X>/remove<X> fields of the same type, where
// that type also appears as a list in the read op's response. The pair becomes
// a synthesized attachment resource (Create=add<X>, Delete=remove<X>,
// Read=membership test). This is parent-agnostic — it fires on any resource
// in the spec that matches the shape, not just PolicyGroup.

type attachmentPair struct {
	EdgeStem        string   // CamelCase stem, e.g. "Stack" from "addStack"
	AddField        string   // "addStack"
	RemoveField     string   // "removeStack"
	MembershipField string   // read-response list field, matched by element type
	MatchKey        []string // edge fields identifying one element (sorted)
}

type attachmentDerivation struct {
	Token           string
	MutationOp      string
	ReadOp          string
	AddField        string
	RemoveField     string
	MembershipField string
	MatchKey        []string
	IDFormat        string
}

// detectAttachmentPairs scans an update op's request body for symmetric
// add/remove field pairs and returns the ones that resolve to a clean edge.
// The second return is the count of near-miss pairs deliberately skipped
// (logged to stderr) so partial coverage is never silent.
func detectAttachmentPairs(spec *rest.Spec, updOp, readOp *rest.Operation) ([]attachmentPair, int) {
	body := flattenedProps(spec, updOp.RequestRef)
	resp := flattenedProps(spec, readOp.ResponseRef)

	addFields := make([]string, 0, len(body))
	for k := range body {
		if _, ok := attachmentStem(k, "add"); ok {
			addFields = append(addFields, k)
		}
	}
	sort.Strings(addFields)

	var pairs []attachmentPair
	skipped := 0
	for _, addField := range addFields {
		stem, _ := attachmentStem(addField, "add")
		removeField := "remove" + stem
		if _, ok := body[removeField]; !ok {
			continue // not a paired add/remove; not an edge at all
		}
		addRef, removeRef := refOf(body[addField]), refOf(body[removeField])
		if addRef == "" || addRef != removeRef {
			fmt.Fprintf(os.Stderr, "  attachment skip %s on %s: add/remove types differ\n", stem, updOp.ID)
			skipped++
			continue
		}
		// Resolve the membership list + the fields that identify one element.
		// Two shapes, both derived from the spec — no hand-tuning per resource:
		//   object: the read response has a list whose element type IS the add
		//     type, and that type is a flat all-required-string identity (e.g.
		//     stacks -> AppPulumiStackReference{name,routingProject}).
		//   scalar: no such typed list, but the add type carries a `name` and
		//     the response has a lone []string list holding those names (e.g.
		//     accounts -> []string). The spec types these membership lists as
		//     bare strings, so the link is by-name, not by-$ref.
		var membership string
		var matchKey []string
		if m := findListFieldByItemRef(resp, addRef); m != "" {
			mk, ok := pureStringIdentity(spec, addRef)
			if !ok {
				fmt.Fprintf(os.Stderr, "  attachment skip %s on %s: edge type isn't a flat all-required-string identity\n",
					stem, updOp.ID)
				skipped++
				continue
			}
			membership, matchKey = m, mk
		} else if nameField := stringField(spec, addRef, nameFieldKey); nameField != "" {
			if scalarList := findLoneScalarList(resp); scalarList != "" {
				membership, matchKey = scalarList, []string{nameField}
			}
		}
		if membership == "" {
			fmt.Fprintf(os.Stderr, "  attachment skip %s on %s: no membership list in %s matches the edge\n",
				stem, updOp.ID, readOp.ID)
			skipped++
			continue
		}
		// An edge field sharing a parent path-param name can't round-trip: the
		// path value would shadow the edge value in both body and match.
		if clash := pathParamClash(updOp, matchKey); clash != "" {
			fmt.Fprintf(os.Stderr, "  attachment skip %s on %s: edge field %q collides with a path parameter\n",
				stem, updOp.ID, clash)
			skipped++
			continue
		}
		pairs = append(pairs, attachmentPair{
			EdgeStem:        stem,
			AddField:        addField,
			RemoveField:     removeField,
			MembershipField: membership,
			MatchKey:        matchKey,
		})
	}
	return pairs, skipped
}

// stringField returns field if ref's schema declares it as a string property,
// else "". Used to find the scalar edge identity (an account's `name`).
func stringField(spec *rest.Spec, ref, field string) string {
	sch, ok := spec.ResolveSchema(ref)
	if !ok {
		return ""
	}
	props, ok := sch["properties"].(map[string]any)
	if !ok {
		return ""
	}
	p, ok := props[field].(map[string]any)
	if !ok {
		return ""
	}
	if t, _ := p["type"].(string); t == typeString {
		return field
	}
	return ""
}

// findLoneScalarList returns the name of the response's single []string list,
// or "" when there are zero or several (ambiguous → defer). Sorted iteration
// keeps the choice deterministic.
func findLoneScalarList(props map[string]any) string {
	found := ""
	for _, name := range slices.Sorted(maps.Keys(props)) {
		m, ok := props[name].(map[string]any)
		if !ok {
			continue
		}
		if t, _ := m["type"].(string); t != typeArray {
			continue
		}
		items, ok := m["items"].(map[string]any)
		if !ok {
			continue
		}
		if _, hasRef := items["$ref"]; hasRef {
			continue
		}
		if it, _ := items["type"].(string); it == typeString {
			if found != "" {
				return "" // ambiguous
			}
			found = name
		}
	}
	return found
}

// pathParamClash returns the first edge field whose name matches one of the
// mutation op's path parameters, or "" when there's no collision.
func pathParamClash(op *rest.Operation, edgeFields []string) string {
	params := map[string]bool{}
	for _, pp := range pathParamsOf(op) {
		params[pp] = true
	}
	for _, f := range edgeFields {
		if params[f] {
			return f
		}
	}
	return ""
}

// attachmentStem returns the CamelCase remainder after prefix, requiring the
// next char to be uppercase so "addStack" matches but "address" doesn't.
func attachmentStem(field, prefix string) (string, bool) {
	if !strings.HasPrefix(field, prefix) {
		return "", false
	}
	rest := field[len(prefix):]
	if rest == "" || !unicode.IsUpper(rune(rest[0])) {
		return "", false
	}
	return rest, true
}

func refOf(prop any) string {
	m, ok := prop.(map[string]any)
	if !ok {
		return ""
	}
	r, _ := m["$ref"].(string)
	return r
}

// findListFieldByItemRef returns the name of the array property whose element
// $ref equals ref. Matched by type, not name (addPolicyPack → appliedPolicyPacks,
// not policyPacks). Sorted iteration keeps the choice deterministic.
func findListFieldByItemRef(props map[string]any, ref string) string {
	for _, name := range slices.Sorted(maps.Keys(props)) {
		m, ok := props[name].(map[string]any)
		if !ok {
			continue
		}
		if t, _ := m["type"].(string); t != typeArray {
			continue
		}
		items, ok := m["items"].(map[string]any)
		if !ok {
			continue
		}
		if r, _ := items["$ref"].(string); r == ref {
			return name
		}
	}
	return ""
}

// pureStringIdentity returns the sorted field names of ref's schema iff it is a
// flat object whose every property is a required string — i.e. the whole object
// is the edge's identity, so it round-trips cleanly through an idFormat and a
// membership match. Anything with optional or non-string fields (e.g. a policy
// pack carrying config) is rejected for auto-emit and left for curation.
func pureStringIdentity(spec *rest.Spec, ref string) ([]string, bool) {
	sch, ok := spec.ResolveSchema(ref)
	if !ok {
		return nil, false
	}
	if _, hasAllOf := sch["allOf"]; hasAllOf {
		return nil, false
	}
	props, ok := sch["properties"].(map[string]any)
	if !ok || len(props) == 0 {
		return nil, false
	}
	required := map[string]bool{}
	if rr, ok := sch["required"].([]any); ok {
		for _, r := range rr {
			if s, ok := r.(string); ok {
				required[s] = true
			}
		}
	}
	if len(required) != len(props) {
		return nil, false
	}
	keys := make([]string, 0, len(props))
	for k, v := range props {
		if !required[k] {
			return nil, false
		}
		vm, ok := v.(map[string]any)
		if !ok {
			return nil, false
		}
		if t, _ := vm["type"].(string); t != typeString {
			return nil, false
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys, true
}

// attachmentIDFormat builds the composite ID template: the mutation op's path
// params (parent identity) followed by the edge's MatchKey fields. Attachments
// carry no renames, so wire names are used verbatim.
func attachmentIDFormat(updOp *rest.Operation, matchKey []string) string {
	var parts []string
	for _, m := range pathParamPattern.FindAllStringSubmatch(updOp.Path, -1) {
		parts = append(parts, "{"+m[1]+"}")
	}
	for _, k := range matchKey {
		parts = append(parts, "{"+k+"}")
	}
	return strings.Join(parts, "/")
}

// mergeAttachment layers a derived attachment descriptor onto an existing
// entry: the attachment block is replaced wholesale (derived), while token,
// idFormat, and deleteBeforeReplace follow write-if-absent so hand edits stick.
func mergeAttachment(existing json.RawMessage, d attachmentDerivation) (json.RawMessage, bool, error) {
	var entry map[string]any
	if len(existing) > 0 {
		if err := json.Unmarshal(existing, &entry); err != nil {
			return nil, false, err
		}
	}
	if entry == nil {
		entry = map[string]any{}
	}

	matchKey := make([]any, len(d.MatchKey))
	for i, k := range d.MatchKey {
		matchKey[i] = k
	}
	newAtt := map[string]any{
		"mutationOp":      d.MutationOp,
		"readOp":          d.ReadOp,
		"addField":        d.AddField,
		"removeField":     d.RemoveField,
		"membershipField": d.MembershipField,
		"matchKey":        matchKey,
	}
	prevAtt, _ := json.Marshal(entry["attachment"])
	curAtt, _ := json.Marshal(newAtt)
	changed := string(prevAtt) != string(curAtt)
	entry["attachment"] = newAtt

	if _, has := entry["token"]; !has {
		entry["token"] = d.Token
	}
	if d.IDFormat != "" {
		if existing, has := entry["idFormat"].(string); has {
			if existing != d.IDFormat {
				fmt.Fprintf(os.Stderr,
					"  INFO: idFormat pinned at %q; heuristic now suggests %q (preserving pin)\n",
					existing, d.IDFormat)
			}
		} else {
			entry["idFormat"] = d.IDFormat
		}
	}
	if _, has := entry["deleteBeforeReplace"]; !has {
		entry["deleteBeforeReplace"] = true
	}

	encoded, err := encodeStable(entry)
	if err != nil {
		return nil, false, err
	}
	return encoded, changed, nil
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
			if _, claimed := out[nameFieldKey]; !claimed && requestHasField(spec, createOp, nameFieldKey) {
				out[nameFieldKey] = p
				seen[p] = true
				continue
			}
		}
	}

	// Rule (4): body field that duplicates a path param under a verbose alias.
	// Pairs must come from verboseBodyAliases — a HasPrefix relation alone
	// would also map "organizationalUnitName" → "orgName", which is wrong.
	if createOp != nil {
		bodyProps := flattenedProps(spec, createOp.RequestRef)
		for _, p := range pathParamsOf(createOp) {
			pPrefix, pSuffix, pOK := splitTrailingSuffix(p)
			if !pOK || pPrefix == "" {
				continue
			}
			allowed := verboseBodyAliases[strings.ToLower(pPrefix)]
			if len(allowed) == 0 {
				continue
			}
			for body := range bodyProps {
				if body == p {
					continue
				}
				bPrefix, bSuffix, bOK := splitTrailingSuffix(body)
				if !bOK || bSuffix != pSuffix {
					continue
				}
				if !slices.Contains(allowed, strings.ToLower(bPrefix)) {
					continue
				}
				if _, exists := out[body]; !exists {
					out[body] = p
				}
			}
		}
	}

	// Rule (5): case-insensitive path-param ↔ response body alias. When the
	// OpenAPI spec is inconsistent — path uses {changeRequestID} but the
	// response body returns "changeRequestId" — the same logical field
	// surfaces under two casings and chained references break. Pick the
	// response form as the Pulumi-side: response casing tends to be the more
	// idiomatic JSON-camelCase shape and is what users see when reading state.
	for _, op := range []*rest.Operation{createOp, readOp, updateOp, deleteOp} {
		if op == nil || op.ResponseRef == "" {
			continue
		}
		responseFields := flattenedProps(spec, op.ResponseRef)
		if len(responseFields) == 0 {
			continue
		}
		for _, p := range pathParamsOf(op) {
			if seen[p] {
				continue
			}
			if _, exact := responseFields[p]; exact {
				continue
			}
			pLower := strings.ToLower(p)
			for f := range responseFields {
				if f == p {
					continue
				}
				if strings.ToLower(f) != pLower {
					continue
				}
				if _, claimed := out[f]; claimed {
					continue
				}
				out[f] = p
				seen[p] = true
				break
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

// encodeStable marshals an entry to JSON. Key order is json.Marshal's
// default (alphabetical), matching what indentJSON's downstream
// MarshalIndent re-parses-then-emits anyway — making this single source
// of order produce the same result as the file on disk.
func encodeStable(entry map[string]any) (json.RawMessage, error) {
	return json.Marshal(entry)
}

func writeMetadata(path string, doc *metadataDoc) error {
	var b strings.Builder
	b.WriteString("{\n")
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

	return atomicWriteFile(path, []byte(b.String()), 0o600)
}

// atomicWriteFile writes via "<path>.tmp" + os.Rename so a Ctrl-C, OOM, or
// disk-full mid-write can't leave a truncated metadata.json on disk
// (which would break //go:embed until git checkout).
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
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
// prefix as a tiebreaker for ambiguous POSTs. Ops marked deprecated
// upstream still contribute to candidates — tracked in stats for reporting.
func derive(paths map[string]map[string]any) (map[string]derivedOps, deriveStats) {
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
			if _, hasCreate := byNoun[shorter][opCreate]; hasCreate {
				continue
			}
			c, ok := byNoun[longer][opCreate]
			if !ok {
				continue
			}
			byNoun[shorter][opCreate] = c
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
			if slot == opRead {
				continue
			}
			if _, set := byNoun[singular][slot]; !set {
				byNoun[singular][slot] = opID
			}
		}
		hasOnlyReadLeft := true
		for slot := range byNoun[plural] {
			if slot != opRead {
				hasOnlyReadLeft = false
				break
			}
		}
		otherOps[singular] = append(otherOps[singular], otherOps[plural]...)
		if hasOnlyReadLeft {
			otherOps[singular] = append(otherOps[singular], byNoun[plural][opRead])
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
		if _, hasOwnCreate := byNoun[bare][opCreate]; hasOwnCreate {
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
		if _, hasCreate := final[opCreate]; !hasCreate {
			if u, hasUpdate := final[opUpdate]; hasUpdate {
				final[opCreate] = u
			}
		}
		_, hasCreate := final[opCreate]
		_, hasRead := final[opRead]
		_, hasDelete := final[opDelete]
		if !hasCreate || (!hasRead && !hasDelete) {
			slots := make([]string, 0, len(o))
			for k := range o {
				slots = append(slots, k)
			}
			sort.Strings(slots)
			stats.skipped = append(stats.skipped, fmt.Sprintf("%s [%s]", noun, strings.Join(slots, ",")))
			continue
		}
		candidates["pulumiservice:api:"+noun] = derivedOps{
			Create: final[opCreate],
			Read:   final[opRead],
			Update: final[opUpdate],
			Delete: final[opDelete],
		}
	}

	return candidates, stats
}

// isDeprecated returns true when an operation is marked deprecated, either via
// the OpenAPI standard `deprecated: true` boolean or the legacy Pulumi-custom
// `x-pulumi-route-property.Visibility = "Deprecated"` extension. Used for
// reporting only — deprecated ops still scaffold as api resources.
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
// module ends up in the user-facing token (pulumiservice:api/<module>:Type).
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

// deriveToken constructs the user-facing Pulumi token. For singletons (no
// shared module), it returns the metadata-key form so the token is locked
// from first appearance — this prevents a later sibling URL silently
// promoting the resource into a module and shifting its user-facing token.
// For multi-resource modules, type name is shortened by dropping a
// redundant module prefix (AgentPool in module agents → Pool) and a
// trailing route-mangled suffix (Environment_esc_environments in module
// esc → Environment).
func deriveToken(pkg, key string, ma moduleAssignment) string {
	if ma.Module == "" {
		return key
	}
	typ := key
	if i := strings.LastIndex(key, ":"); i >= 0 {
		typ = key[i+1:]
	}
	typ = stripRouteSuffix(typ, ma.OriginalPrefix)
	typ = stripModulePrefix(typ, ma.Module)
	return fmt.Sprintf("%s:api/%s:%s", pkg, ma.Module, typ)
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
		capped := strings.ToUpper(c[:1]) + c[1:]
		if !strings.HasPrefix(typ, capped) || len(typ) <= len(capped) {
			continue
		}
		if r := rune(typ[len(capped)]); r >= 'A' && r <= 'Z' {
			return typ[len(capped):]
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
