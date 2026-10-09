#!/usr/bin/env python3
# Copyright 2026, Pulumi Corporation.
#
# Re-inserts the per-entity-rule comments of a YAML Pulumi program into the
# programs `pulumi convert` generated from it, which drops comments.
#
# Usage: inject_rule_comments.py <Pulumi.yaml> <lang> <generated file>
# where <lang> is one of typescript, python, go, csharp, java. The file is
# rewritten in place. Fails if the number of rules found in the generated
# program differs from the number of comments in the YAML.

import re
import sys

RULE_START = {
    "go": re.compile(r"^(\s*)&pulumiservice\.Role\w+RuleArgs\{$"),
    "csharp": re.compile(r"^(\s*)new PulumiService\.Inputs\.Role\w+Rule\w*Args$"),
    # A single-rule list starts on the setter's line: `.stackRules(RoleStackRuleArgs.builder()`.
    "java": re.compile(r"^(\s*)(\.\w+Rules\()?Role\w+RuleArgs\.builder\(\)$"),
}
BLOCK_OPEN = {
    "typescript": re.compile(r"^\s*(stackRules|environmentRules|insightsAccountRules): \[(\{)?$"),
    "python": re.compile(r"^\s*(stack_rules|environment_rules|insights_account_rules)=\[(\{)?$"),
}
RULE_LISTS = ("stackRules:", "environmentRules:", "insightsAccountRules:")
# Rules are list items directly under a `*Rules:` key in a
# `fn::invoke` argument, which puts their dash 10 spaces in.
RULE_INDENT = 10
COMMENT = {"python": "#", "typescript": "//", "go": "//", "csharp": "//", "java": "//"}


def yaml_rule_comments(path):
    """The comment above each entity rule, in program order."""
    comments, in_rules, pending = [], False, None
    for line in open(path):
        stripped = line.strip()
        if stripped in RULE_LISTS:
            in_rules, pending = True, None
            continue
        if not in_rules:
            continue
        if stripped.startswith("#"):
            pending = stripped[1:].strip()
        elif stripped.startswith("- ") and line.index("-") == RULE_INDENT:
            if pending is None:
                sys.exit(f"{path}: entity rule without a comment: {stripped}")
            comments.append(pending)
            pending = None
        elif stripped and not line.startswith(" " * RULE_INDENT):
            in_rules = False
    return comments


def rule_starts(lang, lines):
    """Indices of the lines that open each entity rule."""
    if lang in RULE_START:
        return [i for i, l in enumerate(lines) if RULE_START[lang].match(l)]
    starts = []
    for i, l in enumerate(lines):
        m = BLOCK_OPEN[lang].match(l)
        if not m:
            continue
        if m.group(2):  # a single-rule list, `stackRules: [{`: comment the list itself
            starts.append(i)
            continue
        indent = None
        for j in range(i + 1, len(lines)):
            cur = lines[j]
            if indent is None:
                indent = len(cur) - len(cur.lstrip())
            if cur.strip() == "{" and len(cur) - len(cur.lstrip()) == indent:
                starts.append(j)
            elif cur.strip().startswith("]") and len(cur) - len(cur.lstrip()) < indent:
                break
    return starts


def main():
    yaml_path, lang, target = sys.argv[1:4]
    comments = yaml_rule_comments(yaml_path)
    lines = open(target).read().split("\n")
    starts = rule_starts(lang, lines)
    if len(starts) != len(comments):
        sys.exit(f"{target}: found {len(starts)} entity rules but {len(comments)} comments in {yaml_path}")
    for idx, comment in reversed(list(zip(starts, comments))):
        indent = lines[idx][: len(lines[idx]) - len(lines[idx].lstrip())]
        lines.insert(idx, f"{indent}{COMMENT[lang]} {comment}")
    open(target, "w").write("\n".join(lines))


if __name__ == "__main__":
    main()
