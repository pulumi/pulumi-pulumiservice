#!/usr/bin/env bash
# Regenerates the provider's per-entity-type scope enums (RbacOrganizationScope,
# RbacStackScope, RbacEnvironmentScope, RbacInsightsAccountScope in
# provider/pkg/functions/zz_generated_rbac_scopes.go) from the Pulumi Cloud
# RBAC scope catalog.
#
# By default it fetches the catalog from the API and stores it, normalized, in
# provider/pkg/cloud/rbac-scopes.json. The endpoint is org-scoped, but every
# org gets the same catalog back; use one with all RBAC features enabled so no
# feature-gated scopes are missing. Pass --no-fetch to regenerate from the
# stored catalog only.
#
# Generation fails if the catalog drops a scope the enums already have,
# because removing an enum member breaks SDK users. Pass --allow-removals when
# that is intended (and note it in the CHANGELOG as a breaking change).
#
# Review the resulting diff: the enums are part of the provider's public
# interface, so new or removed scopes are a deliberate change.
#
# Usage: PULUMI_ACCESS_TOKEN=... [PULUMI_ORG=...] scripts/gen-rbac-scopes.sh [--no-fetch] [--allow-removals]

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
catalog="${repo_root}/provider/pkg/cloud/rbac-scopes.json"

fetch=true
gen_flags=()
for arg in "$@"; do
  case "$arg" in
    --no-fetch) fetch=false ;;
    --allow-removals) gen_flags+=(-allow-removals) ;;
    *) echo "unknown argument: $arg" >&2; exit 1 ;;
  esac
done

if $fetch; then
  : "${PULUMI_ACCESS_TOKEN:?PULUMI_ACCESS_TOKEN must be set}"
  api="${PULUMI_API:-https://api.pulumi.com}"
  org="${PULUMI_ORG:-service-provider-test-org}"
  tmp="$(mktemp)"
  trap 'rm -f "$tmp"' EXIT
  curl -sf --max-time 60 \
    -H "Authorization: token ${PULUMI_ACCESS_TOKEN}" \
    -H "Accept: application/vnd.pulumi+9" \
    "${api}/api/orgs/${org}/roles/scopes" \
    | jq --sort-keys 'map_values(sort_by(.name) | map(.scopes |= sort_by(.name)))' >"$tmp"
  mv "$tmp" "$catalog"
fi

cd "${repo_root}/provider"
go run ./tools/gen-rbac-scopes "${gen_flags[@]+"${gen_flags[@]}"}" \
  -catalog "$catalog" \
  -out pkg/functions/zz_generated_rbac_scopes.go
