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

package util

import (
	"fmt"
	"strings"
	"unicode"
)

// ImportDocs renders the `## Import` section appended to a resource
// description. The registry docs generator splits descriptions on
// "\n\n## Import" and renders the rest as the page's Import section.
// Each note is rendered as its own paragraph after the example command.
func ImportDocs(token, idFormat, exampleID string, notes ...string) string {
	var b strings.Builder
	b.WriteString("\n\n## Import\n\n")
	fmt.Fprintf(&b, "This resource can be imported using its `id`, which has the format `%s`, e.g.,\n\n", idFormat)
	name := Kebab(token[strings.LastIndex(token, ":")+1:])
	fmt.Fprintf(&b, "```sh\n $ pulumi import %s my-%s %s\n```\n", token, name, exampleID)
	for _, n := range notes {
		b.WriteString("\n" + n + "\n")
	}
	return b.String()
}

// ConsoleIDHint tells users where to copy a server-assigned ID from the
// Pulumi Cloud console, with a sample page URL ending in that ID.
func ConsoleIDHint(what, sampleURL string) string {
	return fmt.Sprintf("The %s is the last segment of its page URL in the Pulumi Cloud console, for example "+
		"`%s`.", what, sampleURL)
}

// APIIDHint tells users how to list a server-assigned ID that the Pulumi
// Cloud console doesn't show. apiPath is relative to https://api.pulumi.com
// and jqFilter selects the fields to print.
func APIIDHint(what, apiPath, jqFilter string) string {
	return fmt.Sprintf("The Pulumi Cloud console doesn't show the %s. List it with the "+
		"[Pulumi Cloud REST API](https://www.pulumi.com/docs/reference/cloud-rest-api/):\n\n"+
		"```sh\ncurl -s -H \"Authorization: token $PULUMI_ACCESS_TOKEN\" \\\n  https://api.pulumi.com%s | jq '%s'\n```",
		what, apiPath, jqFilter)
}

// ImportUnsupportedDocs renders an `## Import` section for resources that
// can't be imported.
func ImportUnsupportedDocs() string {
	return "\n\n## Import\n\nImport is not supported for this resource.\n"
}

// Kebab converts a PascalCase name to kebab-case, keeping acronyms together
// ("OidcIssuer" -> "oidc-issuer", "CustomVCSIntegration" -> "custom-vcs-integration").
func Kebab(s string) string {
	rs := []rune(s)
	var b strings.Builder
	for i, r := range rs {
		if unicode.IsUpper(r) && i > 0 {
			prevLower := !unicode.IsUpper(rs[i-1])
			nextLower := i+1 < len(rs) && unicode.IsLower(rs[i+1])
			if prevLower || nextLower {
				b.WriteByte('-')
			}
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}
