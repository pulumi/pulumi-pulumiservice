#!/usr/bin/env bash
# Regenerates provider/pkg/resources/docs/rbac_example.md, the Example Usage
# section of the RbacPermissionSet and RbacRole schema docs, from
# examples/yaml-rbac-roles/Pulumi.yaml.
#
# The YAML example is the source of truth (it runs as an integration test).
# This script strips its test-only `digits` config, then uses `pulumi convert`
# with the locally built provider to produce the other languages.
#
# Usage: make provider && scripts/gen-rbac-example-docs.sh

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out="${repo_root}/provider/pkg/resources/docs/rbac_example.md"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

if [ ! -x "${repo_root}/bin/pulumi-resource-pulumiservice" ]; then
  echo "error: bin/pulumi-resource-pulumiservice not found; run 'make provider' first" >&2
  exit 1
fi

mkdir -p "$work/yaml"
python3 - "${repo_root}/examples/yaml-rbac-roles/Pulumi.yaml" "$work/yaml/Pulumi.yaml" <<'EOF'
import re, sys
s = open(sys.argv[1]).read()
s = s.replace("-${digits}", "")
s = re.sub(r"\n  digits:\n    type: string\n", "\n", s)
s = re.sub(r"# Everything below `resources:`.*?enforces it\.\n", "", s, flags=re.S)
s = s.replace("name: yaml-rbac-roles", "name: rbac-roles")
s = s.replace("    default: service-provider-test-org\n", "    default: my-org\n")
open(sys.argv[2], "w").write(s)
EOF

for lang in typescript python go csharp java; do
  (cd "$work/yaml" && PATH="${repo_root}/bin:$PATH" \
    pulumi convert --from yaml --language "$lang" --out "$work/$lang" --generate-only >/dev/null 2>&1)
done

block() { # block <fence language> <file>
  printf '```%s\n' "$1"
  cat "$2"
  printf '\n```\n\n'
}

{
  printf '{{%% examples %%}}\n## Example Usage\n\n{{%% example %%}}\n'
  printf '### A role with organization-level access and entity rules\n\n'
  block typescript "$work/typescript/index.ts"
  block python "$work/python/__main__.py"
  block go "$work/go/main.go"
  block csharp "$work/csharp/Program.cs"
  block java "$work/java/src/main/java/generated_program/App.java"
  # The resources/variables/outputs of the YAML program, without the project header.
  printf '```yaml\n'
  sed -n '/^config:/,$p' "$work/yaml/Pulumi.yaml"
  printf '```\n\n{{%% /example %%}}\n{{%% /examples %%}}\n'
} >"$out"

echo "wrote ${out#"$repo_root"/}"
