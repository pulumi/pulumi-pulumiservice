#!/bin/bash
# Copyright 2016-2025, Pulumi Corporation.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -euo pipefail

SCHEMA_FILE="${1:-provider/cmd/pulumi-resource-pulumiservice/schema.json}"

if [ ! -f "$SCHEMA_FILE" ]; then
    echo "Error: Schema file not found: $SCHEMA_FILE"
    exit 1
fi

echo "Validating import documentation in $SCHEMA_FILE..."

# Check if jq is available
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed"
    exit 1
fi

# Get all resources that are missing import documentation
MISSING=$(jq -r '.resources | to_entries[] | select(.value.description | contains("### Import") | not) | .key' "$SCHEMA_FILE")

if [ -n "$MISSING" ]; then
    echo "❌ Error: The following resources are missing import documentation:"
    echo "$MISSING" | while read -r resource; do
        echo "  - $resource"
    done
    echo ""
    echo "All resources must include import documentation in their description field."
    echo "Format example:"
    echo ""
    echo "  ### Import"
    echo ""
    echo "  [Resource] can be imported using the \`id\`, which for [resource] is \`{format}\` e.g.,"
    echo ""
    echo "  \`\`\`sh"
    echo "   $ pulumi import pulumiservice:index:[Resource] [name] [id-example]"
    echo "  \`\`\`"
    echo ""
    exit 1
fi

TOTAL=$(jq -r '.resources | length' "$SCHEMA_FILE")
echo "✅ All $TOTAL resources have import documentation"
exit 0
