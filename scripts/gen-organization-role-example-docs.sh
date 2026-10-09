#!/usr/bin/env bash
# Regenerates provider/pkg/resources/docs/organization_role_example.md, the
# Example Usage section of the OrganizationRole schema docs, from the YAML
# program in provider/pkg/resources/docs/organization_role_example/.
#
# `pulumi convert`, run with the locally built provider, produces the other
# languages; it type-checks the program against the schema but drops
# comments, so scripts/lib/inject_rule_comments.py puts each entity rule's
# comment back, and writes Java enum values as strings, which
# scripts/lib/fix_java_enums.py turns into enum constants.
#
# The docs program looks up existing entities rather than creating them, so
# it is not run as a test; examples/yaml-organization-role is the end-to-end
# test for the same features.
#
# Usage: make provider && scripts/gen-organization-role-example-docs.sh

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
src="${repo_root}/provider/pkg/resources/docs/organization_role_example/Pulumi.yaml"
out="${repo_root}/provider/pkg/resources/docs/organization_role_example.md"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

if [ ! -x "${repo_root}/bin/pulumi-resource-pulumiservice" ]; then
  echo "error: bin/pulumi-resource-pulumiservice not found; run 'make provider' first" >&2
  exit 1
fi

mkdir -p "$work/yaml"
cp "$src" "$work/yaml/Pulumi.yaml"

convert() { # convert <language> <generated file, relative to the output dir>
  (cd "$work/yaml" && PATH="${repo_root}/bin:$PATH" \
    pulumi convert --from yaml --language "$1" --out "$work/$1" --generate-only \
      >"$work/$1.log" 2>&1) || { cat "$work/$1.log" >&2; exit 1; }
  python3 "${repo_root}/scripts/lib/inject_rule_comments.py" "$src" "$1" "$work/$1/$2"
}

convert typescript index.ts
convert python __main__.py
convert go main.go
convert csharp Program.cs
convert java src/main/java/generated_program/App.java
python3 "${repo_root}/scripts/lib/fix_java_enums.py" \
  "${repo_root}/provider/cmd/pulumi-resource-pulumiservice/schema.json" \
  "$work/java/src/main/java/generated_program/App.java"

block() { # block <fence language> <file>
  printf '```%s\n' "$1"
  cat "$2"
  printf '\n```\n\n'
}

{
  printf '{{%% examples %%}}\n## Example Usage\n\n{{%% example %%}}\n'
  printf '### A role with organization-level access and entity rules\n\n'
  printf 'The program looks up an existing stack, environment, and Insights account, then defines two '
  printf 'roles: one granting scopes directly, and one granting permission sets through a policy.\n\n'
  block typescript "$work/typescript/index.ts"
  block python "$work/python/__main__.py"
  block go "$work/go/main.go"
  block csharp "$work/csharp/Program.cs"
  block java "$work/java/src/main/java/generated_program/App.java"
  # The YAML program without its project header.
  printf '```yaml\n'
  sed -n '/^config:/,$p' "$src"
  printf '```\n\n{{%% /example %%}}\n{{%% /examples %%}}\n'
} >"$out"

echo "wrote ${out#"$repo_root"/}"
