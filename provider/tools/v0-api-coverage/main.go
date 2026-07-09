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

// Command v0-api-coverage regenerates docs/v0-api-coverage.md from the
// provider's compiled schema.json and the inline v0→api mapping below.
//
// The mapping captures intent (which v0 resource an api resource supersedes)
// and is impossible to derive from token names alone, so it lives here as
// hand-curated data. Every v0 resource that ships in schema.json must have
// an entry; every api token without a v0 mapping is rendered as an api-only row.
//
// Run after `make provider` so schema.json reflects the latest changes:
//
//	go run ./provider/tools/v0-api-coverage
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
	v0Prefix        = "pulumiservice:index:"
	apiPrefix       = "pulumiservice:api:" // root-level api (no module)
	apiModulePrefix = "pulumiservice:api/" // module-namespaced api

	// scheduledDeployment is the api token shared by the several v0
	// schedule resources that all map onto the same deployments resource.
	scheduledDeployment = "pulumiservice:api/deployments:ScheduledDeployment"
)

type mapping struct {
	V0   string   // unprefixed v0 type name, e.g. "AccessToken"
	API  []string // full api tokens, e.g. "pulumiservice:api/tokens:PersonalToken"
	Note string   // optional inline qualifier, e.g. "partial"
}

// mappings is the v0→api coverage matrix. Keep alphabetical by V0.
var mappings = []mapping{
	{V0: "AccessToken", API: []string{"pulumiservice:api/tokens:PersonalToken"}},
	{V0: "AgentPool", API: []string{"pulumiservice:api/agents:Pool"}},
	{V0: "ApprovalRule", API: []string{"pulumiservice:api:Gate"}},
	{V0: "DeploymentSchedule", API: []string{scheduledDeployment}},
	{V0: "DeploymentSettings", API: []string{"pulumiservice:api/deployments:Settings"}},
	{V0: "DriftSchedule", API: []string{scheduledDeployment}, Note: "partial"},
	{V0: "Environment", API: []string{"pulumiservice:api/esc:Environment"}},
	{V0: "EnvironmentRotationSchedule", API: []string{"pulumiservice:api/esc:EnvironmentSchedule"}},
	{V0: "EnvironmentVersionTag", API: []string{"pulumiservice:api/esc:RevisionTag"}},
	{V0: "InsightsAccount", API: []string{"pulumiservice:api/insights:Account"}},
	{V0: "OidcIssuer", API: []string{"pulumiservice:api/auth:OidcIssuer"}},
	{V0: "OrgAccessToken", API: []string{"pulumiservice:api/tokens:OrgToken"}},
	{V0: "OrganizationMember", API: []string{"pulumiservice:api:OrganizationMember"}},
	{V0: "OrganizationRole", API: []string{"pulumiservice:api:Role"}},
	{V0: "PolicyGroup", API: []string{"pulumiservice:api:PolicyGroup"}},
	{V0: "Stack", API: []string{"pulumiservice:api/stacks:Stack"}},
	{V0: "StackTag", API: []string{"pulumiservice:api/stacks:Tag"}},
	{V0: "StackTags", API: []string{"pulumiservice:api/stacks:Tag"}, Note: "singular only"},
	{V0: "Team", API: []string{"pulumiservice:api/teams:Team"}},
	{V0: "TeamAccessToken", API: []string{"pulumiservice:api/tokens:TeamToken"}},
	{V0: "TeamEnvironmentPermission", API: []string{"pulumiservice:api/auth:Policy"}},
	{V0: "TeamRoleAssignment", API: []string{"pulumiservice:api/teams:Role"}},
	{V0: "TeamStackPermission", API: []string{"pulumiservice:api/auth:Policy"}},
	{V0: "TemplateSource", API: []string{"pulumiservice:api:OrgTemplateCollection"}},
	{V0: "TtlSchedule", API: []string{scheduledDeployment}, Note: "partial"},
	{V0: "Webhook", API: []string{
		"pulumiservice:api:OrganizationWebhook",
		"pulumiservice:api/stacks:Webhook",
		"pulumiservice:api/esc:Webhook",
	}},
}

func main() {
	schemaPath := flag.String("schema",
		"provider/cmd/pulumi-resource-pulumiservice/schema.json",
		"Path to compiled schema.json")
	outPath := flag.String("out", "docs/v0-api-coverage.md", "Path to write the coverage doc")
	flag.Parse()

	v0Tokens, apiTokens, err := loadTokens(*schemaPath)
	if err != nil {
		fail("load schema %s: %v", *schemaPath, err)
	}

	if err := validate(v0Tokens, apiTokens); err != nil {
		fail("validate mapping: %v", err)
	}

	out := render(apiTokens)
	if err := os.WriteFile(*outPath, []byte(out), 0o600); err != nil {
		fail("write %s: %v", *outPath, err)
	}
	fmt.Fprintf(os.Stderr, "v0-api-coverage: wrote %s (%d v0, %d api)\n", *outPath, len(v0Tokens), len(apiTokens))
}

func loadTokens(path string) (v0, api map[string]bool, err error) {
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
	v0, api = map[string]bool{}, map[string]bool{}
	for tok := range doc.Resources {
		switch {
		case strings.HasPrefix(tok, v0Prefix):
			v0[strings.TrimPrefix(tok, v0Prefix)] = true
		case strings.HasPrefix(tok, apiModulePrefix), strings.HasPrefix(tok, apiPrefix):
			api[tok] = true
		}
	}
	return v0, api, nil
}

// validate ensures every v0 resource in the schema has exactly one mapping
// entry, and every api target referenced in mappings exists in the schema.
func validate(v0Schema, apiSchema map[string]bool) error {
	mapped := map[string]bool{}
	for _, m := range mappings {
		if mapped[m.V0] {
			return fmt.Errorf("duplicate mapping for v0 %q", m.V0)
		}
		mapped[m.V0] = true
		if !v0Schema[m.V0] {
			return fmt.Errorf("mapping references v0 %q which is not in schema.json", m.V0)
		}
		for _, t := range m.API {
			if !apiSchema[t] {
				return fmt.Errorf("mapping for v0 %q references api %q which is not in schema.json", m.V0, t)
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

func render(apiSchema map[string]bool) string {
	var b strings.Builder
	b.WriteString("<!-- DO NOT EDIT. Generated by provider/tools/v0-api-coverage. -->\n")
	b.WriteString("# v0 / api Resource Coverage\n\n")
	b.WriteString("> Regenerate after `make provider`:\n>\n")
	b.WriteString("> ```sh\n> go run ./provider/tools/v0-api-coverage\n> ```\n>\n")
	b.WriteString("> The v0→api mapping is hand-curated in " +
		"[provider/tools/v0-api-coverage/main.go](../provider/tools/v0-api-coverage/main.go); " +
		"the api-only rows come from `schema.json`.\n\n")
	b.WriteString("| v0 (`pulumiservice:index:`) | api (`pulumiservice:api/`) |\n")
	b.WriteString("|---|---|\n")

	rows := append([]mapping(nil), mappings...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].V0 < rows[j].V0 })
	for _, m := range rows {
		parts := make([]string, len(m.API))
		for i, t := range m.API {
			parts[i] = "`" + formatAPI(t) + "`"
		}
		apiCell := strings.Join(parts, ", ")
		if m.Note != "" {
			apiCell += " (" + m.Note + ")"
		}
		fmt.Fprintf(&b, "| `%s` | %s |\n", m.V0, apiCell)
	}

	mapped := map[string]bool{}
	for _, m := range mappings {
		for _, t := range m.API {
			mapped[t] = true
		}
	}
	var apiOnly []string
	for tok := range apiSchema {
		if !mapped[tok] {
			apiOnly = append(apiOnly, tok)
		}
	}
	sort.Slice(apiOnly, func(i, j int) bool { return apiOnlySortKey(apiOnly[i]) < apiOnlySortKey(apiOnly[j]) })
	for _, tok := range apiOnly {
		fmt.Fprintf(&b, "| — | `%s` |\n", formatAPI(tok))
	}
	return b.String()
}

// formatAPI collapses the api token to the form used in the doc:
//   - "pulumiservice:api/<module>:<Name>" → "<module>:<Name>"
//   - "pulumiservice:api:<Name>"          → unchanged (root-level api has no module)
func formatAPI(tok string) string {
	if rest, ok := strings.CutPrefix(tok, apiModulePrefix); ok {
		return rest
	}
	return tok
}

// apiOnlySortKey groups module-namespaced tokens above root-level ones, each
// group sorted alphabetically by display form.
func apiOnlySortKey(tok string) string {
	if strings.HasPrefix(tok, apiModulePrefix) {
		return "0:" + formatAPI(tok)
	}
	return "1:" + formatAPI(tok)
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "v0-api-coverage: "+format+"\n", args...)
	os.Exit(1)
}
