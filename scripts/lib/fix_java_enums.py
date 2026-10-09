#!/usr/bin/env python3
# Copyright 2026, Pulumi Corporation.
#
# `pulumi convert --language java` writes enum-typed values as string
# literals, which don't compile against the Java SDK. This rewrites the scope
# and RoleTagOperator arguments of a generated Java program into enum
# constants, using the names from the provider schema. A rule's `scopes`
# enum depends on the rule type, taken from the nearest preceding
# `Role<Type>RuleArgs.builder()`.
#
# Usage: fix_java_enums.py <schema.json> <App.java>

import json
import re
import sys

STRING_LITERAL = re.compile(r'"([^"]*)"')

RULE_BUILDER = re.compile(r"Role(Stack|Environment|InsightsAccount)RuleArgs\.builder\(\)")
ENUM_TYPES = ("RbacOrganizationScope", "RbacStackScope", "RbacEnvironmentScope",
              "RbacInsightsAccountScope", "RoleTagOperator")


def enum_for(method, src, pos):
    if method == "organizationScopes":
        return "RbacOrganizationScope"
    if method == "operator":
        return "RoleTagOperator"
    rules = list(RULE_BUILDER.finditer(src, 0, pos))
    if not rules:
        sys.exit(f"`.scopes(` at offset {pos} is not inside a rule builder")
    return f"Rbac{rules[-1].group(1)}Scope"


def main():
    schema_path, target = sys.argv[1:3]
    types = json.load(open(schema_path))["types"]
    names = {
        enum: {v["value"]: v["name"] for v in types[f"pulumiservice:index:{enum}"]["enum"]}
        for enum in ENUM_TYPES
    }
    src = open(target).read()
    used = set()

    def rewrite(m):
        method, args = m.group(1), m.group(2)
        enum = enum_for(method, src, m.start())

        def constant(lit):
            value = lit.group(1)
            if value not in names[enum]:
                sys.exit(f"{target}: {value!r} is not a {enum} value")
            used.add(enum)
            return f"{enum}.{names[enum][value]}"

        return f".{method}({STRING_LITERAL.sub(constant, args)})"

    src = re.sub(r"\.(scopes|organizationScopes|operator)\(([^()]*)\)", rewrite, src)
    imports = "".join(f"import com.pulumi.pulumiservice.enums.{e};\n" for e in sorted(used))
    src = src.replace("import com.pulumi.core.Output;\n", "import com.pulumi.core.Output;\n" + imports, 1)
    open(target, "w").write(src)


if __name__ == "__main__":
    main()
