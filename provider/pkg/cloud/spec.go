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

// Package cloud holds the embedded OpenAPI spec and metadata document used
// to derive the v2 resource surface at runtime.
//
// The schema served by GetSchema and the v2 CRUD dispatch both flow from
// these two files; nothing is generated at build time. Refresh the spec
// with `go generate ./provider/pkg/cloud/...`.
package cloud

import (
	_ "embed"
	"fmt"

	"github.com/pulumi/pulumi-pulumiservice/provider/pkg/rest"
)

// Spec source: https://api.pulumi.com/api/openapi/pulumi-spec.json
//
// `go generate` runs two steps in order:
//  1. openapi-fetch  — refresh spec.json from the upstream URL
//  2. scaffold-metadata — re-derive metadata.candidate.json from the spec
//     so newly introduced API operations surface in PR diffs. metadata.json
//     itself is hand-curated; copy entries from the candidate file when
//     adding resources.
//
//go:generate go run ../../tools/openapi-fetch -out spec.json
//go:generate go run ../../tools/scaffold-metadata -in spec.json -out metadata.candidate.json

//go:embed spec.json
var specJSON []byte

//go:embed metadata.json
var metadataJSON []byte

var (
	parsedSpec     *rest.Spec
	parsedMetadata *rest.Metadata
)

func init() {
	s, err := rest.ParseSpec(specJSON)
	if err != nil {
		panic(fmt.Errorf("cloud: parse embedded OpenAPI spec: %w", err))
	}
	parsedSpec = s

	m, err := rest.ParseMetadata(metadataJSON)
	if err != nil {
		panic(fmt.Errorf("cloud: parse embedded metadata: %w", err))
	}
	parsedMetadata = m
}

// Spec returns the parsed OpenAPI spec.
func Spec() *rest.Spec { return parsedSpec }

// Metadata returns the parsed Pulumi resource metadata.
func Metadata() *rest.Metadata { return parsedMetadata }
