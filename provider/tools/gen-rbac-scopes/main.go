// Copyright 2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command gen-rbac-scopes turns the Pulumi Cloud RBAC scope catalog (the
// normalized response of GET /api/orgs/{org}/roles/scopes) into the
// provider's per-entity-type scope enums, and the organization's built-in
// permission sets (from GET /api/orgs/{org}/roles?uxPurpose=set) into the
// BuiltInPermissionSet enum. Run it through scripts/gen-rbac-scopes.sh.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"github.com/pulumi/pulumi-cloud-sdk/go/apitype"
)

// enumTypes maps each catalog bucket (the entity type a scope applies to) to
// the Go enum type generated for it. A bucket missing here fails generation,
// so a new entity type is a deliberate change.
var enumTypes = map[string]string{
	"global":           "RbacOrganizationScope",
	"stack":            "RbacStackScope",
	"environment":      "RbacEnvironmentScope",
	"insights-account": "RbacInsightsAccountScope",
}

type scopeCatalog map[string][]struct {
	Name   string `json:"name"`
	Scopes []struct {
		Name     string `json:"name"`
		Metadata struct {
			Description string `json:"description"`
		} `json:"metadata"`
	} `json:"scopes"`
}

// permissionSetCatalog is the normalized list of built-in permission sets.
type permissionSetCatalog []struct {
	DefaultIdentifier string `json:"defaultIdentifier"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	ResourceType      string `json:"resourceType"`
}

// permissionSetEnum is the Go enum type generated for built-in permission sets.
const permissionSetEnum = "BuiltInPermissionSet"

// scope is one enum member: a scope, or a built-in permission set.
type scope struct {
	Value       string
	Name        string
	Description string
}

func main() {
	in := flag.String("catalog", "rbac-scopes.json", "Normalized /roles/scopes response")
	out := flag.String("out", "zz_generated_rbac_scopes.go", "Go file to write")
	setsIn := flag.String("permission-sets", "rbac-permission-sets.json", "Normalized built-in permission sets")
	setsOut := flag.String("permission-sets-out", "zz_generated_permission_sets.go", "Go file to write")
	allowRemovals := flag.Bool("allow-removals", false,
		"Allow dropping values the current enums have. Removing an enum member breaks SDK users.")
	flag.Parse()

	generate(*out, scopeEnums(*in), *allowRemovals)
	generate(*setsOut, map[string][]scope{permissionSetEnum: permissionSets(*setsIn)}, *allowRemovals)
}

// generate writes the enums to out, failing if that would drop a value the
// previously generated file has.
func generate(out string, enums map[string][]scope, allowRemovals bool) {
	if removed := removedValues(out, enums); len(removed) > 0 && !allowRemovals {
		fail("these values are in the current enums but not in the API's: %s. Removing an enum member "+
			"is a breaking change for SDK users; rerun with -allow-removals if that is intended",
			strings.Join(removed, ", "))
	}
	src, err := render(enums)
	if err != nil {
		fail("%v", err)
	}
	if err := os.WriteFile(out, src, 0o600); err != nil {
		fail("write %s: %v", out, err)
	}
	for _, typ := range sortedKeys(enums) {
		fmt.Fprintf(os.Stderr, "gen-rbac-scopes: %s: %d values\n", typ, len(enums[typ]))
	}
}

func scopeEnums(in string) map[string][]scope {
	raw, err := os.ReadFile(in)
	if err != nil {
		fail("read catalog: %v", err)
	}
	var catalog scopeCatalog
	if err := json.Unmarshal(raw, &catalog); err != nil {
		fail("parse catalog: %v", err)
	}
	enums := map[string][]scope{}
	for bucket := range catalog {
		typ, ok := enumTypes[bucket]
		if !ok {
			fail("catalog has unknown entity type %q; add it to enumTypes", bucket)
		}
		if enums[typ], err = bucketScopes(catalog, bucket); err != nil {
			fail("%s: %v", bucket, err)
		}
	}
	for bucket, typ := range enumTypes {
		if len(enums[typ]) == 0 {
			fail("catalog has no scopes for entity type %q", bucket)
		}
	}
	return enums
}

// permissionSets reads the built-in permission sets, sorted by identifier.
// Each member's description leads with the set's display name, as the
// Pulumi Cloud console shows it.
func permissionSets(in string) []scope {
	raw, err := os.ReadFile(in)
	if err != nil {
		fail("read permission sets: %v", err)
	}
	var catalog permissionSetCatalog
	if err := json.Unmarshal(raw, &catalog); err != nil {
		fail("parse permission sets: %v", err)
	}
	if len(catalog) == 0 {
		fail("%s has no built-in permission sets", in)
	}
	sets := make([]scope, 0, len(catalog))
	for _, s := range catalog {
		if _, ok := enumTypes[s.ResourceType]; !ok {
			fail("permission set %q has unknown entity type %q; add it to enumTypes",
				s.DefaultIdentifier, s.ResourceType)
		}
		sets = append(sets, scope{
			Value:       s.DefaultIdentifier,
			Name:        pascal(s.DefaultIdentifier),
			Description: fmt.Sprintf("%s (%s): %s", s.Name, s.ResourceType, s.Description),
		})
	}
	sort.Slice(sets, func(i, j int) bool { return sets[i].Value < sets[j].Value })
	return sets
}

// generatedConst matches an enum constant in a previously generated file,
// e.g. `RbacStackScopeStackRead RbacStackScope = "stack:read"`.
var generatedConst = regexp.MustCompile(`(?m)^\s*\w+\s+(\w+)\s+=\s+"([^"]+)"$`)

// removedValues lists the values, as "Enum value", that the previously
// generated file at path has but enums doesn't.
func removedValues(path string, enums map[string][]scope) []string {
	prev, err := os.ReadFile(path)
	if err != nil {
		return nil // first generation
	}
	current := map[string]bool{}
	for typ, scopes := range enums {
		for _, s := range scopes {
			current[typ+" "+s.Value] = true
		}
	}
	var removed []string
	for _, m := range generatedConst.FindAllStringSubmatch(string(prev), -1) {
		if key := m[1] + " " + m[2]; !current[key] {
			removed = append(removed, fmt.Sprintf("%s %q", m[1], m[2]))
		}
	}
	sort.Strings(removed)
	return removed
}

// bucketScopes flattens one catalog bucket into one entry per scope, sorted
// by value. Scopes the Pulumi Cloud SDK doesn't know yet are skipped, because
// the provider's helpers validate against the SDK and would reject them; bump
// the SDK and regenerate to pick them up.
func bucketScopes(catalog scopeCatalog, bucket string) ([]scope, error) {
	byValue := map[string]scope{}
	names := map[string]string{}
	for _, g := range catalog[bucket] {
		for _, s := range g.Scopes {
			if _, ok := byValue[s.Name]; ok || s.Name == "" {
				continue
			}
			if !apitype.RbacPermission(s.Name).IsValid() {
				fmt.Fprintf(os.Stderr, "gen-rbac-scopes: %s: skipping %q: not in the Pulumi Cloud SDK yet\n",
					bucket, s.Name)
				continue
			}
			name := pascal(s.Name)
			if other, ok := names[name]; ok {
				return nil, fmt.Errorf("scopes %q and %q both map to the name %q", other, s.Name, name)
			}
			names[name] = s.Name
			byValue[s.Name] = scope{Value: s.Name, Name: name, Description: s.Metadata.Description}
		}
	}
	scopes := make([]scope, 0, len(byValue))
	for _, s := range byValue {
		scopes = append(scopes, s)
	}
	sort.Slice(scopes, func(i, j int) bool { return scopes[i].Value < scopes[j].Value })
	return scopes, nil
}

// pascal turns a scope value such as "stack_deployment:create" into
// "StackDeploymentCreate".
func pascal(s string) string {
	var b strings.Builder
	upper := true
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			upper = true
			continue
		}
		if upper {
			r = unicode.ToUpper(r)
			upper = false
		}
		b.WriteRune(r)
	}
	return b.String()
}

func render(enums map[string][]scope) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(`// Code generated by provider/tools/gen-rbac-scopes; DO NOT EDIT.
// Regenerate with scripts/gen-rbac-scopes.sh.

package functions

import "github.com/pulumi/pulumi-go-provider/infer"
`)
	for _, typ := range sortedKeys(enums) {
		lower := strings.ToLower(typ[:1]) + typ[1:]
		b.WriteString("\nconst (\n")
		for _, s := range enums[typ] {
			fmt.Fprintf(&b, "\t%s%s %s = %q\n", typ, s.Name, typ, s.Value)
		}
		fmt.Fprintf(&b, ")\n\nvar %sValues = []infer.EnumValue[%s]{\n", lower, typ)
		for _, s := range enums[typ] {
			fmt.Fprintf(&b, "\t{Name: %q, Value: %s%s, Description: %q},\n", s.Name, typ, s.Name, s.Description)
		}
		b.WriteString("}\n")
	}
	return format.Source(b.Bytes())
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func fail(f string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen-rbac-scopes: "+f+"\n", args...)
	os.Exit(1)
}
