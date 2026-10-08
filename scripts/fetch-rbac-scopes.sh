#!/usr/bin/env bash
# Fetches the RBAC scope catalog from the Pulumi Cloud API and writes it,
# normalized, to provider/pkg/cloud/rbac-scopes.json. The generator in
# provider/tools/gen-rbac-scopes turns that file into the RbacScope and
# RbacResourceType enums.
#
# The endpoint is org-scoped and requires authentication, but every org gets
# the same catalog back. Use an org with all RBAC features enabled so no
# feature-gated scopes are missing.
#
# Usage: PULUMI_ACCESS_TOKEN=... scripts/fetch-rbac-scopes.sh

set -euo pipefail

: "${PULUMI_ACCESS_TOKEN:?PULUMI_ACCESS_TOKEN must be set}"
PULUMI_API="${PULUMI_API:-https://api.pulumi.com}"
PULUMI_ORG="${PULUMI_ORG:-service-provider-test-org}"

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out="${repo_root}/provider/pkg/cloud/rbac-scopes.json"
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

curl -sf --max-time 60 \
  -H "Authorization: token ${PULUMI_ACCESS_TOKEN}" \
  -H "Accept: application/vnd.pulumi+8" \
  "${PULUMI_API}/api/orgs/${PULUMI_ORG}/roles/scopes" \
  | jq --sort-keys 'map_values(sort_by(.name) | map(.scopes |= sort_by(.name)))' >"$tmp"

if [ "$(jq '[.[][] | .scopes[]] | length' "$tmp")" -eq 0 ]; then
  echo "error: scope catalog is empty" >&2
  exit 1
fi

mv "$tmp" "$out"
echo "wrote $(jq '[.[][] | .scopes[]] | length' "$out") scopes to ${out#"$repo_root"/}"
