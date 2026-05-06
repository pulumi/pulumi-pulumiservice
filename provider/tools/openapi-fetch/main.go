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

// Command openapi-fetch downloads the Pulumi Cloud OpenAPI spec and writes it
// to disk in a stable, normalized form so diffs are reviewable.
//
// Source URL defaults to the public Pulumi Cloud spec; override via -url or
// the PULUMI_CLOUD_OPENAPI_URL environment variable (the flag wins). Output
// path comes from -out (default spec.json in the working directory).
//
// Invoked by `go generate ./provider/pkg/cloud/...` (see spec.go) and via
// `make generate`. CI runs the same command and fails if the result differs
// from what's committed.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// defaultSpecURL is the public Pulumi Cloud OpenAPI spec.
const defaultSpecURL = "https://api.pulumi.com/api/openapi/pulumi-spec.json"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	urlFlag := flag.String("url", "", "OpenAPI spec URL (overrides $PULUMI_CLOUD_OPENAPI_URL; falls back to the public default)")
	outFlag := flag.String("out", "spec.json", "Output file path")
	flag.Parse()

	source := *urlFlag
	if source == "" {
		source = os.Getenv("PULUMI_CLOUD_OPENAPI_URL")
	}
	if source == "" {
		source = defaultSpecURL
	}

	body, err := download(source)
	if err != nil {
		return err
	}

	pretty, err := normalize(body)
	if err != nil {
		return err
	}

	if err := os.WriteFile(*outFlag, pretty, 0o644); err != nil {
		return fmt.Errorf("openapi-fetch: write %s: %w", *outFlag, err)
	}
	return nil
}

func download(url string) ([]byte, error) {
	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("openapi-fetch: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openapi-fetch: fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openapi-fetch: %s returned %d", url, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openapi-fetch: read response: %w", err)
	}
	return body, nil
}

// normalize reformats raw JSON for stable diffs:
//   - parses, then re-serializes with sorted keys and 2-space indentation
//   - ensures a trailing newline
func normalize(raw []byte) ([]byte, error) {
	var doc any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("openapi-fetch: parse spec: %w", err)
	}

	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("openapi-fetch: re-encode spec: %w", err)
	}
	return out.Bytes(), nil
}
