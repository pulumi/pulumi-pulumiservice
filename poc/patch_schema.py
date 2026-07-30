#!/usr/bin/env python3
"""POC: inject discriminated unions into the committed Pulumi schema.

Rewrites Role.details (recursive PermissionDescriptor, 6 variants) and
deployments Settings.vcs (DeploymentSettingsVCS, 5 variants) from Any to
inline oneOf + discriminator over const-tagged flattened variant types,
following the azure-native shape. Reads variant shapes from spec.json.
"""
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SCHEMA = ROOT / "provider/cmd/pulumi-resource-pulumiservice/schema.json"
SPEC = ROOT / "provider/pkg/cloud/spec.json"

UNIONS = [
    {
        "base": "PermissionDescriptor",
        "module": "api",
        "resource": "pulumiservice:api:Role",
        "props": ["details"],
    },
    {
        "base": "DeploymentSettingsVCS",
        "module": "api/deployments",
        "resource": "pulumiservice:api/deployments:Settings",
        "props": ["vcs"],
    },
]


def refname(ref):
    return ref.rsplit("/", 1)[-1]


def load(path):
    return json.loads(path.read_text())


def flatten_openapi(spec, name):
    """Resolve one schema: merge allOf chain into (props, required)."""
    node = spec["components"]["schemas"][name]
    props, required = {}, []

    def walk(n):
        if "$ref" in n:
            walk(spec["components"]["schemas"][refname(n["$ref"])])
            return
        for part in n.get("allOf", []):
            walk(part)
        props.update(n.get("properties", {}))
        required.extend(n.get("required", []))

    walk(node)
    return props, required


def convert_prop(node, union_bases, module):
    """OpenAPI property -> Pulumi TypeSpec. $ref to a union base becomes the
    inline union; any other $ref degrades to Any (same as today)."""
    out = {}
    if "description" in node:
        out["description"] = node["description"]
    ref = node.get("$ref")
    if ref:
        base = refname(ref)
        if base in union_bases:
            out.update(union_typespec(base, union_bases[base], module))
        else:
            out["$ref"] = "pulumi.json#/Any"
        return out
    t = node.get("type")
    if t == "array":
        items = node.get("items", {})
        out["type"] = "array"
        out["items"] = convert_prop(items, union_bases, module)
        return out
    if t in ("string", "integer", "number", "boolean"):
        out["type"] = t
        return out
    if t == "object" or t is None:
        out["$ref"] = "pulumi.json#/Any"
        return out
    out["$ref"] = "pulumi.json#/Any"
    return out


def union_typespec(base, mapping, module):
    return {
        "oneOf": [
            {"type": "object", "$ref": f"#/types/pulumiservice:{module}:{refname(v)}"}
            for v in mapping.values()
        ],
        "discriminator": {
            "propertyName": disc_prop_cache[base],
            "mapping": {
                tag: f"#/types/pulumiservice:{module}:{refname(v)}"
                for tag, v in mapping.items()
            },
        },
    }


disc_prop_cache = {}


def build_variants(spec, base, module, union_bases):
    """Emit one Pulumi named type per variant, base props flattened in,
    tag property const-ed and required."""
    node = spec["components"]["schemas"][base]
    disc = node["discriminator"]
    tag_prop = disc["propertyName"]
    types = {}
    for tag, ref in sorted(disc["mapping"].items()):
        vname = refname(ref)
        props, required = flatten_openapi(spec, vname)
        pprops = {}
        for k, v in sorted(props.items()):
            ps = convert_prop(v, union_bases, module)
            if k == tag_prop:
                ps = {"type": "string", "const": tag}
                desc = v.get("description", "")
                hint = f"Expected value is '{tag}'."
                ps["description"] = f"{desc} {hint}".strip()
            pprops[k] = ps
        req = sorted(set(required) | {tag_prop})
        # only keep required entries that exist as properties
        req = [r for r in req if r in pprops]
        vdesc = ""
        for part in spec["components"]["schemas"][vname].get("allOf", []):
            if "description" in part:
                vdesc = part["description"]
        types[f"pulumiservice:{module}:{vname}"] = {
            "type": "object",
            "description": vdesc or f"{vname} variant of {base}.",
            "properties": pprops,
            "required": req,
        }
    return types


def main():
    schema = load(SCHEMA)
    spec = load(SPEC)
    schema.setdefault("types", {})

    union_bases = {}
    for u in UNIONS:
        node = spec["components"]["schemas"][u["base"]]
        disc_prop_cache[u["base"]] = node["discriminator"]["propertyName"]
        union_bases[u["base"]] = node["discriminator"]["mapping"]

    for u in UNIONS:
        module = u["module"]
        new_types = build_variants(spec, u["base"], module, union_bases)
        for k, v in new_types.items():
            schema["types"][k] = v
        res = schema["resources"][u["resource"]]
        for prop in u["props"]:
            for side in ("inputProperties", "properties"):
                if prop not in res.get(side, {}):
                    print(f"WARN: {u['resource']}.{side}.{prop} missing", file=sys.stderr)
                    continue
                desc = res[side][prop].get("description", "")
                ts = union_typespec(u["base"], union_bases[u["base"]], module)
                if desc:
                    ts["description"] = desc
                res[side][prop] = ts
        print(f"{u['resource']}: {len(new_types)} variant types, props {u['props']}")

    SCHEMA.write_text(json.dumps(schema, indent=2) + "\n")
    print(f"wrote {SCHEMA}")


if __name__ == "__main__":
    main()
