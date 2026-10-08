// Copyright 2026, Pulumi Corporation.
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

// Command gen-rbac-scopes generates the RbacScope and RbacResourceType enums
// from the Pulumi Cloud RBAC scope catalog.
//
// Inputs:
//   - rbac-scopes.json: the normalized response of
//     GET /api/orgs/{org}/roles/scopes (see scripts/fetch-rbac-scopes.sh).
//     It is the source of truth for which scopes exist, their descriptions,
//     and which resource type (global, stack, environment, insights-account)
//     each scope belongs to.
//   - spec.json: the Pulumi Cloud OpenAPI spec. Only the enumFieldNames of
//     RbacScope.name are read, so enum members get the same names the
//     service uses. Scopes missing from the spec get a derived name.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"os"
	"slices"
	"sort"
	"strings"
	"unicode"
)

// scopeCatalog mirrors the /roles/scopes response: resource type -> groups.
type scopeCatalog map[string][]scopeGroup

type scopeGroup struct {
	Name   string `json:"name"`
	Scopes []struct {
		Name     string `json:"name"`
		Metadata struct {
			Description string `json:"description"`
		} `json:"metadata"`
	} `json:"scopes"`
}

type scope struct {
	Value         string
	Name          string
	Description   string
	ResourceTypes []string
	Groups        []string
}

type resourceType struct {
	Value string
	Name  string
}

type model struct {
	Scopes        []scope
	ResourceTypes []resourceType
}

func main() {
	scopesPath := flag.String("scopes", "rbac-scopes.json", "Normalized /roles/scopes response")
	specPath := flag.String("spec", "spec.json", "Pulumi Cloud OpenAPI spec")
	outPath := flag.String("out", "zz_generated_rbac.go", "Go file to write")
	flag.Parse()

	catalogBytes, err := os.ReadFile(*scopesPath)
	if err != nil {
		fail("read scopes: %v", err)
	}
	specBytes, err := os.ReadFile(*specPath)
	if err != nil {
		fail("read spec: %v", err)
	}

	var catalog scopeCatalog
	if err := json.Unmarshal(catalogBytes, &catalog); err != nil {
		fail("parse scopes: %v", err)
	}
	specNames, err := specFieldNames(specBytes)
	if err != nil {
		fail("%v", err)
	}

	m, err := buildModel(catalog, specNames)
	if err != nil {
		fail("%v", err)
	}
	reportSpecDrift(m, specNames)

	src, err := render(m)
	if err != nil {
		fail("%v", err)
	}
	if err := os.WriteFile(*outPath, src, 0o600); err != nil {
		fail("write %s: %v", *outPath, err)
	}
	fmt.Fprintf(os.Stderr, "gen-rbac-scopes: wrote %d scopes, %d resource types to %s\n",
		len(m.Scopes), len(m.ResourceTypes), *outPath)
}

// specFieldNames returns scope value -> enum member name from the
// x-pulumi-model-property of RbacScope.name in the OpenAPI spec.
func specFieldNames(specBytes []byte) (map[string]string, error) {
	var spec struct {
		Components struct {
			// Other schemas have non-string enums, so only RbacScope is decoded.
			Schemas map[string]json.RawMessage `json:"schemas"`
		} `json:"components"`
	}
	if err := json.Unmarshal(specBytes, &spec); err != nil {
		return nil, fmt.Errorf("parse spec: %w", err)
	}
	raw, ok := spec.Components.Schemas["RbacScope"]
	if !ok {
		return nil, fmt.Errorf("spec: RbacScope schema not found")
	}
	var rbacScope struct {
		Properties map[string]struct {
			Enum  []string `json:"enum"`
			Model struct {
				EnumFieldNames []string `json:"enumFieldNames"`
			} `json:"x-pulumi-model-property"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(raw, &rbacScope); err != nil {
		return nil, fmt.Errorf("parse spec RbacScope: %w", err)
	}
	prop, ok := rbacScope.Properties["name"]
	if !ok {
		return nil, fmt.Errorf("spec: RbacScope.name not found")
	}
	if len(prop.Enum) != len(prop.Model.EnumFieldNames) {
		return nil, fmt.Errorf("spec: RbacScope.name has %d enum values but %d enumFieldNames",
			len(prop.Enum), len(prop.Model.EnumFieldNames))
	}
	names := make(map[string]string, len(prop.Enum))
	for i, v := range prop.Enum {
		names[v] = prop.Model.EnumFieldNames[i]
	}
	return names, nil
}

func buildModel(catalog scopeCatalog, specNames map[string]string) (model, error) {
	byValue := map[string]*scope{}
	var m model

	rtKeys := make([]string, 0, len(catalog))
	for k := range catalog {
		rtKeys = append(rtKeys, k)
	}
	sort.Strings(rtKeys)

	for _, rt := range rtKeys {
		m.ResourceTypes = append(m.ResourceTypes, resourceType{Value: rt, Name: pascal(rt)})
		for _, g := range catalog[rt] {
			for _, s := range g.Scopes {
				if s.Name == "" {
					continue
				}
				sc, ok := byValue[s.Name]
				if !ok {
					name := specNames[s.Name]
					if name == "" {
						name = pascal(s.Name)
					}
					sc = &scope{Value: s.Name, Name: name, Description: s.Metadata.Description}
					byValue[s.Name] = sc
				}
				if !slices.Contains(sc.ResourceTypes, rt) {
					sc.ResourceTypes = append(sc.ResourceTypes, rt)
				}
				if !slices.Contains(sc.Groups, g.Name) {
					sc.Groups = append(sc.Groups, g.Name)
				}
			}
		}
	}
	if len(byValue) == 0 {
		return model{}, fmt.Errorf("scope catalog is empty")
	}

	seenNames := map[string]string{}
	for _, sc := range byValue {
		if prev, ok := seenNames[sc.Name]; ok {
			return model{}, fmt.Errorf("enum member name %q used by both %q and %q", sc.Name, prev, sc.Value)
		}
		seenNames[sc.Name] = sc.Value
		m.Scopes = append(m.Scopes, *sc)
	}
	sort.Slice(m.Scopes, func(i, j int) bool { return m.Scopes[i].Value < m.Scopes[j].Value })
	return m, nil
}

// reportSpecDrift logs scopes the live catalog and the spec disagree on.
// Spec-only values are expected (some permissions can't be put in a custom
// role); live-only values mean spec.json is stale.
func reportSpecDrift(m model, specNames map[string]string) {
	live := map[string]bool{}
	for _, s := range m.Scopes {
		live[s.Value] = true
		if _, ok := specNames[s.Value]; !ok {
			fmt.Fprintf(os.Stderr, "gen-rbac-scopes: %q is not in spec.json; using derived name %q\n", s.Value, s.Name)
		}
	}
	var specOnly int
	for v := range specNames {
		if v != "" && !live[v] {
			specOnly++
		}
	}
	if specOnly > 0 {
		fmt.Fprintf(os.Stderr,
			"gen-rbac-scopes: %d spec permissions are not assignable to custom roles (skipped)\n", specOnly)
	}
}

// pascal converts "stack_deployment:read" or "insights-account" to
// "StackDeploymentRead" / "InsightsAccount".
func pascal(s string) string {
	var b strings.Builder
	upper := true
	for _, r := range s {
		if r == ':' || r == '_' || r == '-' || r == '.' || r == ' ' {
			upper = true
			continue
		}
		if upper {
			b.WriteRune(unicode.ToUpper(r))
			upper = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func describe(s scope) string {
	desc := strings.TrimSpace(s.Description)
	if desc != "" && !strings.HasSuffix(desc, ".") {
		desc += "."
	}
	return fmt.Sprintf("%s Applies to: %s (%s).", desc,
		strings.Join(s.ResourceTypes, ", "), strings.Join(s.Groups, ", "))
}

func render(m model) ([]byte, error) {
	var b bytes.Buffer
	p := func(format string, args ...any) { fmt.Fprintf(&b, format, args...) }

	p("// Copyright 2026, Pulumi Corporation.\n//\n")
	p("// Code generated by provider/tools/gen-rbac-scopes. DO NOT EDIT.\n")
	p("// Source: GET /api/orgs/{org}/roles/scopes (provider/pkg/cloud/rbac-scopes.json).\n")
	p("// Regenerate with scripts/fetch-rbac-scopes.sh && go generate ./provider/pkg/resources/.\n\n")
	p("package resources\n\n")
	p("import \"github.com/pulumi/pulumi-go-provider/infer\"\n\n")

	p("// RbacResourceType is the kind of entity a permission set applies to.\n")
	p("type RbacResourceType string\n\nconst (\n")
	for _, rt := range m.ResourceTypes {
		p("\tRbacResourceType%s RbacResourceType = %q\n", rt.Name, rt.Value)
	}
	p(")\n\n")
	p("func (RbacResourceType) Values() []infer.EnumValue[RbacResourceType] {\n")
	p("\treturn []infer.EnumValue[RbacResourceType]{\n")
	for _, rt := range m.ResourceTypes {
		p("\t\t{Name: %q, Value: RbacResourceType%s, Description: %q},\n",
			rt.Name, rt.Name, resourceTypeDescription(rt.Value))
	}
	p("\t}\n}\n\n")

	p("// RbacScope is a single Pulumi Cloud RBAC permission that can be granted by a permission set.\n")
	p("type RbacScope string\n\nconst (\n")
	for _, s := range m.Scopes {
		p("\tRbacScope%s RbacScope = %q\n", s.Name, s.Value)
	}
	p(")\n\n")
	p("func (RbacScope) Values() []infer.EnumValue[RbacScope] {\n")
	p("\treturn []infer.EnumValue[RbacScope]{\n")
	for _, s := range m.Scopes {
		p("\t\t{Name: %q, Value: RbacScope%s, Description: %q},\n", s.Name, s.Name, describe(s))
	}
	p("\t}\n}\n\n")

	p("// rbacScopesByResourceType lists the scopes a permission set of each resource type may grant.\n")
	p("// A scope can belong to more than one resource type.\n")
	p("var rbacScopesByResourceType = map[RbacResourceType][]RbacScope{\n")
	for _, rt := range m.ResourceTypes {
		p("\tRbacResourceType%s: {\n", rt.Name)
		for _, s := range m.Scopes {
			if slices.Contains(s.ResourceTypes, rt.Value) {
				p("\t\tRbacScope%s,\n", s.Name)
			}
		}
		p("\t},\n")
	}
	p("}\n")

	src, err := format.Source(b.Bytes())
	if err != nil {
		return nil, fmt.Errorf("gofmt generated source: %w", err)
	}
	return src, nil
}

func resourceTypeDescription(rt string) string {
	switch rt {
	case "global":
		return "Organization-wide permissions, granted through a role's organization-level access."
	case "stack":
		return "Permissions on stacks, granted through a role's stack entity rules."
	case "environment":
		return "Permissions on ESC environments, granted through a role's environment entity rules."
	case "insights-account":
		return "Permissions on Insights accounts, granted through a role's Insights account entity rules."
	default:
		return fmt.Sprintf("Permissions on %s entities.", rt)
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen-rbac-scopes: "+format+"\n", args...)
	os.Exit(1)
}
