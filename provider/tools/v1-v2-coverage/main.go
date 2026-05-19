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

// Command v1-v2-coverage regenerates docs/v1-v2-coverage.md from the
// provider's compiled schema.json and the inline v1→v2 mapping below.
//
// The mapping captures intent (which v1 resource a v2 resource supersedes)
// and is impossible to derive from token names alone, so it lives here as
// hand-curated data. Every v1 resource that ships in schema.json must have
// an entry; every v2 token without a v1 mapping is rendered as a v2-only row.
//
// Run after `make provider` so schema.json reflects the latest changes:
//
//	go run ./provider/tools/v1-v2-coverage
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

const (
	v1Prefix       = "pulumiservice:index:"
	v2Prefix       = "pulumiservice:v2:"  // root-level v2 (no module)
	v2ModulePrefix = "pulumiservice:v2/"  // module-namespaced v2
)

type mapping struct {
	V1   string   // unprefixed v1 type name, e.g. "AccessToken"
	V2   []string // full v2 tokens, e.g. "pulumiservice:v2/tokens:PersonalToken"
	Note string   // optional inline qualifier, e.g. "partial"
}

// mappings is the v1→v2 coverage matrix. Keep alphabetical by V1.
var mappings = []mapping{
	{V1: "AccessToken", V2: []string{"pulumiservice:v2/tokens:PersonalToken"}},
	{V1: "AgentPool", V2: []string{"pulumiservice:v2/agents:Pool"}},
	{V1: "ApprovalRule", V2: []string{"pulumiservice:v2:Gate"}},
	{V1: "DeploymentSchedule", V2: []string{"pulumiservice:v2/deployments:ScheduledDeployment"}},
	{V1: "DeploymentSettings", V2: []string{"pulumiservice:v2/deployments:Settings"}},
	{V1: "DriftSchedule", V2: []string{"pulumiservice:v2/deployments:ScheduledDeployment"}, Note: "partial"},
	{V1: "Environment", V2: []string{"pulumiservice:v2/esc:Environment"}},
	{V1: "EnvironmentRotationSchedule", V2: []string{"pulumiservice:v2/esc:EnvironmentSchedule"}},
	{V1: "EnvironmentVersionTag", V2: []string{"pulumiservice:v2/esc:RevisionTag"}},
	{V1: "InsightsAccount", V2: []string{"pulumiservice:v2/insights:Account"}},
	{V1: "OidcIssuer", V2: []string{"pulumiservice:v2/auth:OidcIssuer"}},
	{V1: "OrgAccessToken", V2: []string{"pulumiservice:v2/tokens:OrgToken"}},
	{V1: "OrganizationMember", V2: []string{"pulumiservice:v2:OrganizationMember"}},
	{V1: "OrganizationRole", V2: []string{"pulumiservice:v2:Role"}},
	{V1: "PolicyGroup", V2: []string{"pulumiservice:v2:PolicyGroup"}},
	{V1: "Stack", V2: []string{"pulumiservice:v2/stacks:Stack"}},
	{V1: "StackTag", V2: []string{"pulumiservice:v2/stacks:Tag"}},
	{V1: "StackTags", V2: []string{"pulumiservice:v2/stacks:Tag"}, Note: "singular only"},
	{V1: "Team", V2: []string{"pulumiservice:v2/teams:Team"}},
	{V1: "TeamAccessToken", V2: []string{"pulumiservice:v2/tokens:TeamToken"}},
	{V1: "TeamEnvironmentPermission", V2: []string{"pulumiservice:v2/auth:Policy"}},
	{V1: "TeamRoleAssignment", V2: []string{"pulumiservice:v2/teams:Role"}},
	{V1: "TeamStackPermission", V2: []string{"pulumiservice:v2/auth:Policy"}},
	{V1: "TemplateSource", V2: []string{"pulumiservice:v2:OrgTemplateCollection"}},
	{V1: "TtlSchedule", V2: []string{"pulumiservice:v2/deployments:ScheduledDeployment"}, Note: "partial"},
	{V1: "Webhook", V2: []string{"pulumiservice:v2:OrganizationWebhook", "pulumiservice:v2/stacks:Webhook", "pulumiservice:v2/esc:Webhook"}},
}

func main() {
	schemaPath := flag.String("schema", "provider/cmd/pulumi-resource-pulumiservice/schema.json", "Path to compiled schema.json")
	outPath := flag.String("out", "docs/v1-v2-coverage.md", "Path to write the coverage doc")
	flag.Parse()

	v1Tokens, v2Tokens, err := loadTokens(*schemaPath)
	if err != nil {
		fail("load schema %s: %v", *schemaPath, err)
	}

	if err := validate(v1Tokens, v2Tokens); err != nil {
		fail("validate mapping: %v", err)
	}

	out := render(v2Tokens)
	if err := os.WriteFile(*outPath, []byte(out), 0o644); err != nil {
		fail("write %s: %v", *outPath, err)
	}
	fmt.Fprintf(os.Stderr, "v1-v2-coverage: wrote %s (%d v1, %d v2)\n", *outPath, len(v1Tokens), len(v2Tokens))
}

func loadTokens(path string) (v1, v2 map[string]bool, err error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	var doc struct {
		Resources map[string]json.RawMessage `json:"resources"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, nil, err
	}
	v1, v2 = map[string]bool{}, map[string]bool{}
	for tok := range doc.Resources {
		switch {
		case strings.HasPrefix(tok, v1Prefix):
			v1[strings.TrimPrefix(tok, v1Prefix)] = true
		case strings.HasPrefix(tok, v2ModulePrefix), strings.HasPrefix(tok, v2Prefix):
			v2[tok] = true
		}
	}
	return v1, v2, nil
}

// validate ensures every v1 resource in the schema has exactly one mapping
// entry, and every v2 target referenced in mappings exists in the schema.
func validate(v1Schema, v2Schema map[string]bool) error {
	mapped := map[string]bool{}
	for _, m := range mappings {
		if mapped[m.V1] {
			return fmt.Errorf("duplicate mapping for v1 %q", m.V1)
		}
		mapped[m.V1] = true
		if !v1Schema[m.V1] {
			return fmt.Errorf("mapping references v1 %q which is not in schema.json", m.V1)
		}
		for _, t := range m.V2 {
			if !v2Schema[t] {
				return fmt.Errorf("mapping for v1 %q references v2 %q which is not in schema.json", m.V1, t)
			}
		}
	}
	var missing []string
	for v1 := range v1Schema {
		if !mapped[v1] {
			missing = append(missing, v1)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("v1 resources missing from mapping: %s", strings.Join(missing, ", "))
	}
	return nil
}

func render(v2Schema map[string]bool) string {
	var b strings.Builder
	b.WriteString("<!-- DO NOT EDIT. Generated by provider/tools/v1-v2-coverage. -->\n")
	b.WriteString("# v1 / v2 Resource Coverage\n\n")
	b.WriteString("> Regenerate after `make provider`:\n>\n")
	b.WriteString("> ```sh\n> go run ./provider/tools/v1-v2-coverage\n> ```\n>\n")
	b.WriteString("> The v1→v2 mapping is hand-curated in [provider/tools/v1-v2-coverage/main.go](../provider/tools/v1-v2-coverage/main.go); the v2-only rows come from `schema.json`.\n\n")
	b.WriteString("| v1 (`pulumiservice:index:`) | v2 (`pulumiservice:v2/`) |\n")
	b.WriteString("|---|---|\n")

	rows := append([]mapping(nil), mappings...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].V1 < rows[j].V1 })
	for _, m := range rows {
		parts := make([]string, len(m.V2))
		for i, t := range m.V2 {
			parts[i] = "`" + formatV2(t) + "`"
		}
		v2Cell := strings.Join(parts, ", ")
		if m.Note != "" {
			v2Cell += " (" + m.Note + ")"
		}
		fmt.Fprintf(&b, "| `%s` | %s |\n", m.V1, v2Cell)
	}

	mapped := map[string]bool{}
	for _, m := range mappings {
		for _, t := range m.V2 {
			mapped[t] = true
		}
	}
	var v2Only []string
	for tok := range v2Schema {
		if !mapped[tok] {
			v2Only = append(v2Only, tok)
		}
	}
	sort.Slice(v2Only, func(i, j int) bool { return v2OnlySortKey(v2Only[i]) < v2OnlySortKey(v2Only[j]) })
	for _, tok := range v2Only {
		fmt.Fprintf(&b, "| — | `%s` |\n", formatV2(tok))
	}
	return b.String()
}

// formatV2 collapses the v2 token to the form used in the doc:
//   - "pulumiservice:v2/<module>:<Name>" → "<module>:<Name>"
//   - "pulumiservice:v2:<Name>"          → unchanged (root-level v2 has no module)
func formatV2(tok string) string {
	if rest, ok := strings.CutPrefix(tok, v2ModulePrefix); ok {
		return rest
	}
	return tok
}

// v2OnlySortKey groups module-namespaced tokens above root-level ones, each
// group sorted alphabetically by display form.
func v2OnlySortKey(tok string) string {
	if strings.HasPrefix(tok, v2ModulePrefix) {
		return "0:" + formatV2(tok)
	}
	return "1:" + formatV2(tok)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "v1-v2-coverage: "+format+"\n", args...)
	os.Exit(1)
}
