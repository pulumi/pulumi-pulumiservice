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

// Command v0-v1-coverage regenerates docs/v0-v1-coverage.md from the
// provider's compiled schema.json and the inline v0→v1 mapping below.
//
// The mapping captures intent (which v0 resource a v1 resource supersedes)
// and is impossible to derive from token names alone, so it lives here as
// hand-curated data. Every v0 resource that ships in schema.json must have
// an entry; every v1 token without a v0 mapping is rendered as a v1-only row.
//
// Run after `make provider` so schema.json reflects the latest changes:
//
//	go run ./provider/tools/v0-v1-coverage
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
	v0Prefix       = "pulumiservice:index:"
	v1Prefix       = "pulumiservice:v1:" // root-level v1 (no module)
	v1ModulePrefix = "pulumiservice:v1/" // module-namespaced v1
)

type mapping struct {
	V0   string   // unprefixed v0 type name, e.g. "AccessToken"
	V1   []string // full v1 tokens, e.g. "pulumiservice:v1/tokens:PersonalToken"
	Note string   // optional inline qualifier, e.g. "partial"
}

// mappings is the v0→v1 coverage matrix. Keep alphabetical by V0.
var mappings = []mapping{
	{V0: "AccessToken", V1: []string{"pulumiservice:v1/tokens:PersonalToken"}},
	{V0: "AgentPool", V1: []string{"pulumiservice:v1/agents:Pool"}},
	{V0: "ApprovalRule", V1: []string{"pulumiservice:v1:Gate"}},
	{V0: "DeploymentSchedule", V1: []string{"pulumiservice:v1/deployments:ScheduledDeployment"}},
	{V0: "DeploymentSettings", V1: []string{"pulumiservice:v1/deployments:Settings"}},
	{V0: "DriftSchedule", V1: []string{"pulumiservice:v1/deployments:ScheduledDeployment"}, Note: "partial"},
	{V0: "Environment", V1: []string{"pulumiservice:v1/esc:Environment"}},
	{V0: "EnvironmentRotationSchedule", V1: []string{"pulumiservice:v1/esc:EnvironmentSchedule"}},
	{V0: "EnvironmentVersionTag", V1: []string{"pulumiservice:v1/esc:RevisionTag"}},
	{V0: "InsightsAccount", V1: []string{"pulumiservice:v1/insights:Account"}},
	{V0: "OidcIssuer", V1: []string{"pulumiservice:v1/auth:OidcIssuer"}},
	{V0: "OrgAccessToken", V1: []string{"pulumiservice:v1/tokens:OrgToken"}},
	{V0: "OrganizationMember", V1: []string{"pulumiservice:v1:OrganizationMember"}},
	{V0: "OrganizationRole", V1: []string{"pulumiservice:v1:Role"}},
	{V0: "PolicyGroup", V1: []string{"pulumiservice:v1:PolicyGroup"}},
	{V0: "Stack", V1: []string{"pulumiservice:v1/stacks:Stack"}},
	{V0: "StackTag", V1: []string{"pulumiservice:v1/stacks:Tag"}},
	{V0: "StackTags", V1: []string{"pulumiservice:v1/stacks:Tag"}, Note: "singular only"},
	{V0: "Team", V1: []string{"pulumiservice:v1/teams:Team"}},
	{V0: "TeamAccessToken", V1: []string{"pulumiservice:v1/tokens:TeamToken"}},
	{V0: "TeamEnvironmentPermission", V1: []string{"pulumiservice:v1/auth:Policy"}},
	{V0: "TeamRoleAssignment", V1: []string{"pulumiservice:v1/teams:Role"}},
	{V0: "TeamStackPermission", V1: []string{"pulumiservice:v1/auth:Policy"}},
	{V0: "TemplateSource", V1: []string{"pulumiservice:v1:OrgTemplateCollection"}},
	{V0: "TtlSchedule", V1: []string{"pulumiservice:v1/deployments:ScheduledDeployment"}, Note: "partial"},
	{V0: "Webhook", V1: []string{
		"pulumiservice:v1:OrganizationWebhook",
		"pulumiservice:v1/stacks:Webhook",
		"pulumiservice:v1/esc:Webhook",
	}},
}

func main() {
	schemaPath := flag.String("schema",
		"provider/cmd/pulumi-resource-pulumiservice/schema.json",
		"Path to compiled schema.json")
	outPath := flag.String("out", "docs/v0-v1-coverage.md", "Path to write the coverage doc")
	flag.Parse()

	v0Tokens, v1Tokens, err := loadTokens(*schemaPath)
	if err != nil {
		fail("load schema %s: %v", *schemaPath, err)
	}

	if err := validate(v0Tokens, v1Tokens); err != nil {
		fail("validate mapping: %v", err)
	}

	out := render(v1Tokens)
	if err := os.WriteFile(*outPath, []byte(out), 0o600); err != nil {
		fail("write %s: %v", *outPath, err)
	}
	fmt.Fprintf(os.Stderr, "v0-v1-coverage: wrote %s (%d v0, %d v1)\n", *outPath, len(v0Tokens), len(v1Tokens))
}

func loadTokens(path string) (v0, v1 map[string]bool, err error) {
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
	v0, v1 = map[string]bool{}, map[string]bool{}
	for tok := range doc.Resources {
		switch {
		case strings.HasPrefix(tok, v0Prefix):
			v0[strings.TrimPrefix(tok, v0Prefix)] = true
		case strings.HasPrefix(tok, v1ModulePrefix), strings.HasPrefix(tok, v1Prefix):
			v1[tok] = true
		}
	}
	return v0, v1, nil
}

// validate ensures every v0 resource in the schema has exactly one mapping
// entry, and every v1 target referenced in mappings exists in the schema.
func validate(v0Schema, v1Schema map[string]bool) error {
	mapped := map[string]bool{}
	for _, m := range mappings {
		if mapped[m.V0] {
			return fmt.Errorf("duplicate mapping for v0 %q", m.V0)
		}
		mapped[m.V0] = true
		if !v0Schema[m.V0] {
			return fmt.Errorf("mapping references v0 %q which is not in schema.json", m.V0)
		}
		for _, t := range m.V1 {
			if !v1Schema[t] {
				return fmt.Errorf("mapping for v0 %q references v1 %q which is not in schema.json", m.V0, t)
			}
		}
	}
	var missing []string
	for v0 := range v0Schema {
		if !mapped[v0] {
			missing = append(missing, v0)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return fmt.Errorf("v0 resources missing from mapping: %s", strings.Join(missing, ", "))
	}
	return nil
}

func render(v1Schema map[string]bool) string {
	var b strings.Builder
	b.WriteString("<!-- DO NOT EDIT. Generated by provider/tools/v0-v1-coverage. -->\n")
	b.WriteString("# v0 / v1 Resource Coverage\n\n")
	b.WriteString("> Regenerate after `make provider`:\n>\n")
	b.WriteString("> ```sh\n> go run ./provider/tools/v0-v1-coverage\n> ```\n>\n")
	b.WriteString("> The v0→v1 mapping is hand-curated in " +
		"[provider/tools/v0-v1-coverage/main.go](../provider/tools/v0-v1-coverage/main.go); " +
		"the v1-only rows come from `schema.json`.\n\n")
	b.WriteString("| v0 (`pulumiservice:index:`) | v1 (`pulumiservice:v1/`) |\n")
	b.WriteString("|---|---|\n")

	rows := append([]mapping(nil), mappings...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].V0 < rows[j].V0 })
	for _, m := range rows {
		parts := make([]string, len(m.V1))
		for i, t := range m.V1 {
			parts[i] = "`" + formatV1(t) + "`"
		}
		v1Cell := strings.Join(parts, ", ")
		if m.Note != "" {
			v1Cell += " (" + m.Note + ")"
		}
		fmt.Fprintf(&b, "| `%s` | %s |\n", m.V0, v1Cell)
	}

	mapped := map[string]bool{}
	for _, m := range mappings {
		for _, t := range m.V1 {
			mapped[t] = true
		}
	}
	var v1Only []string
	for tok := range v1Schema {
		if !mapped[tok] {
			v1Only = append(v1Only, tok)
		}
	}
	sort.Slice(v1Only, func(i, j int) bool { return v1OnlySortKey(v1Only[i]) < v1OnlySortKey(v1Only[j]) })
	for _, tok := range v1Only {
		fmt.Fprintf(&b, "| — | `%s` |\n", formatV1(tok))
	}
	return b.String()
}

// formatV1 collapses the v1 token to the form used in the doc:
//   - "pulumiservice:v1/<module>:<Name>" → "<module>:<Name>"
//   - "pulumiservice:v1:<Name>"          → unchanged (root-level v1 has no module)
func formatV1(tok string) string {
	if rest, ok := strings.CutPrefix(tok, v1ModulePrefix); ok {
		return rest
	}
	return tok
}

// v1OnlySortKey groups module-namespaced tokens above root-level ones, each
// group sorted alphabetically by display form.
func v1OnlySortKey(tok string) string {
	if strings.HasPrefix(tok, v1ModulePrefix) {
		return "0:" + formatV1(tok)
	}
	return "1:" + formatV1(tok)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "v0-v1-coverage: "+format+"\n", args...)
	os.Exit(1)
}
