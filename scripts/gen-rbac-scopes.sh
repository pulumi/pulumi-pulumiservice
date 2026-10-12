#!/usr/bin/env bash
# Regenerates the provider's RBAC enums from Pulumi Cloud:
#
# - the per-entity-type scope enums (RbacOrganizationScope, RbacStackScope,
#   RbacEnvironmentScope, RbacInsightsAccountScope in
#   provider/pkg/functions/zz_generated_rbac_scopes.go), from the RBAC scope
#   catalog, and
# - the BuiltInPermissionSet enum (in
#   provider/pkg/functions/zz_generated_permission_sets.go), from the
#   organization's built-in permission sets.
#
# By default it fetches both from the API and stores them, normalized, in
# provider/pkg/cloud/rbac-scopes.json and
# provider/pkg/cloud/rbac-permission-sets.json. The endpoints are org-scoped,
# but every org gets the same catalog and built-in sets back; use one with all
# RBAC features enabled so no feature-gated scopes are missing. Listing
# permission sets requires an org admin. Pass --no-fetch to regenerate from
# the stored files only.
#
# Generation fails if the API drops a value the enums already have, because
# removing an enum member breaks SDK users. Pass --allow-removals when that is
# intended (and note it in the CHANGELOG as a breaking change).
#
# Review the resulting diff: the enums are part of the provider's public
# interface, so new or removed values are a deliberate change.
#
# Usage: PULUMI_ACCESS_TOKEN=... [PULUMI_ORG=...] scripts/gen-rbac-scopes.sh [--no-fetch] [--allow-removals]

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
catalog="${repo_root}/provider/pkg/cloud/rbac-scopes.json"
permission_sets="${repo_root}/provider/pkg/cloud/rbac-permission-sets.json"

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
  get() {
    curl -sf --max-time 60 \
      -H "Authorization: token ${PULUMI_ACCESS_TOKEN}" \
      -H "Accept: application/vnd.pulumi+9" \
      "${api}/api/orgs/${org}/$1"
  }
  get roles/scopes \
    | jq --sort-keys 'map_values(sort_by(.name) | map(.scopes |= sort_by(.name)))' >"$tmp"
  mv "$tmp" "$catalog"
  # Built-in sets are the ones with a defaultIdentifier; custom sets are
  # per-org and have none.
  get 'roles?uxPurpose=set' \
    | jq --sort-keys '[.roles[]
        | select((.defaultIdentifier // "") != "")
        | {defaultIdentifier, name, description, resourceType}]
      | sort_by(.defaultIdentifier)' >"$tmp"
  mv "$tmp" "$permission_sets"
fi

cd "${repo_root}/provider"
go run ./tools/gen-rbac-scopes "${gen_flags[@]+"${gen_flags[@]}"}" \
  -catalog "$catalog" \
  -out pkg/functions/zz_generated_rbac_scopes.go \
  -permission-sets "$permission_sets" \
  -permission-sets-out pkg/functions/zz_generated_permission_sets.go
