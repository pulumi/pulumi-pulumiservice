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

package rest

import (
	"strings"
	"testing"
)

const synthCreateBodyRef = "#/components/schemas/CreateBody"

const yamlBodySpec = `{
  "openapi": "3.0.0",
  "components": {"schemas": {
    "CreateBody": {"type": "object", "properties": {"name": {"type": "string"}}}
  }},
  "paths": {
    "/api/json-body": {
      "post": {
        "operationId": "JsonOnly",
        "requestBody": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/CreateBody"}}}},
        "responses": {"200": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/CreateBody"}}}}}
      }
    },
    "/api/yaml-body": {
      "patch": {
        "operationId": "YamlBody",
        "requestBody": {"content": {"application/x-yaml": {"schema": {"type": "string"}}}},
        "responses": {"204": {"description": "no content"}}
      }
    },
    "/api/yaml-response": {
      "get": {
        "operationId": "YamlResponse",
        "responses": {"200": {"content": {"application/x-yaml": {"schema": {"type": "string"}}}}}
      }
    },
    "/api/json-preferred-over-yaml": {
      "get": {
        "operationId": "JsonPreferred",
        "responses": {
          "200": {"content": {"application/x-yaml": {"schema": {"type": "string"}}}},
          "201": {"content": {"application/json": {"schema": {"$ref": "#/components/schemas/CreateBody"}}}}
        }
      }
    }
  }
}`

// TestRealSpecRecognizesEnvironmentYamlBody pins yaml-body classification
// against the embedded Pulumi Cloud spec.
func TestRealSpecRecognizesEnvironmentYamlBody(t *testing.T) {
	spec, _ := loadFixtures(t)
	op, ok := spec.Op("UpdateEnvironment_esc_environments")
	if !ok {
		t.Fatal("UpdateEnvironment_esc_environments missing from real spec")
	}
	if op.RequestContentType != contentYAML {
		t.Errorf("UpdateEnvironment_esc_environments RequestContentType: got %q, want application/x-yaml",
			op.RequestContentType)
	}
	create, ok := spec.Op("CreateEnvironment_esc_environments")
	if !ok {
		t.Fatal("CreateEnvironment_esc_environments missing from real spec")
	}
	if create.RequestContentType != contentJSON {
		t.Errorf("CreateEnvironment_esc_environments RequestContentType: got %q, want application/json",
			create.RequestContentType)
	}
}

// TestParseSpecDuplicateOpIDDeterministic pins the duplicate-operationId
// error to a stable (lexicographically first) "existing" path so the
// message doesn't shuffle on map-iteration order.
func TestParseSpecDuplicateOpIDDeterministic(t *testing.T) {
	const dup = `{
	  "openapi": "3.0.0",
	  "paths": {
	    "/api/zeta": {"get": {"operationId": "Dup", "responses": {"200": {"description": "ok"}}}},
	    "/api/alpha": {"get": {"operationId": "Dup", "responses": {"200": {"description": "ok"}}}}
	  }
	}`
	// Run repeatedly to defeat any one-shot map-order coincidence.
	for range 32 {
		_, err := ParseSpec([]byte(dup))
		if err == nil {
			t.Fatalf("ParseSpec: expected duplicate-operationId error, got nil")
		}
		const want = `duplicate operationId "Dup" (paths /api/alpha and /api/zeta)`
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("ParseSpec error: got %q, want substring %q", err.Error(), want)
		}
	}
}

func TestParseSpecRecognizesContentTypes(t *testing.T) {
	spec, err := ParseSpec([]byte(yamlBodySpec))
	if err != nil {
		t.Fatalf("ParseSpec: %v", err)
	}

	t.Run("json request body still sets RequestRef and content type", func(t *testing.T) {
		op, ok := spec.Op("JsonOnly")
		if !ok {
			t.Fatal("op JsonOnly not found")
		}
		if op.RequestRef != synthCreateBodyRef {
			t.Errorf("RequestRef: got %q, want CreateBody $ref", op.RequestRef)
		}
		if op.RequestContentType != contentJSON {
			t.Errorf("RequestContentType: got %q, want application/json", op.RequestContentType)
		}
		if op.ResponseRef != synthCreateBodyRef {
			t.Errorf("ResponseRef: got %q, want CreateBody $ref", op.ResponseRef)
		}
		if op.ResponseContentType != contentJSON {
			t.Errorf("ResponseContentType: got %q, want application/json", op.ResponseContentType)
		}
	})

	t.Run("yaml request body sets content type without $ref", func(t *testing.T) {
		op, ok := spec.Op("YamlBody")
		if !ok {
			t.Fatal("op YamlBody not found")
		}
		if op.RequestRef != "" {
			t.Errorf("RequestRef: got %q, want empty (yaml has no $ref)", op.RequestRef)
		}
		if op.RequestContentType != contentYAML {
			t.Errorf("RequestContentType: got %q, want application/x-yaml", op.RequestContentType)
		}
	})

	t.Run("yaml-only response surfaces content type", func(t *testing.T) {
		op, ok := spec.Op("YamlResponse")
		if !ok {
			t.Fatal("op YamlResponse not found")
		}
		if op.ResponseRef != "" {
			t.Errorf("ResponseRef: got %q, want empty", op.ResponseRef)
		}
		if op.ResponseContentType != contentYAML {
			t.Errorf("ResponseContentType: got %q, want application/x-yaml", op.ResponseContentType)
		}
	})

	t.Run("json wins over yaml across 2xx codes", func(t *testing.T) {
		op, ok := spec.Op("JsonPreferred")
		if !ok {
			t.Fatal("op JsonPreferred not found")
		}
		if op.ResponseContentType != contentJSON {
			t.Errorf("ResponseContentType: got %q, want application/json "+
				"(JSON should win even at later code)", op.ResponseContentType)
		}
		if op.ResponseRef != synthCreateBodyRef {
			t.Errorf("ResponseRef: got %q, want CreateBody $ref", op.ResponseRef)
		}
	})
}
