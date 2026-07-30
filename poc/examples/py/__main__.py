import os

import pulumi
import pulumi_pulumiservice as pulumiservice

org = os.environ.get("POC_ORG", "poc-org")

# The typed args classes are unusable for __type-tagged variants (Python name
# mangling turns the parameter into _PermissionDescriptorAllowArgs__type), so
# this program uses the dict form - which is also what today's Any users write.
role = pulumiservice.api.Role(
    "poc-union-role",
    org_name=org,
    name="poc-union-role",
    ux_purpose="set",
    resource_type="global",
    description="POC: discriminated union permission tree",
    details={
        "__type": "PermissionDescriptorGroup",
        "entries": [
            {
                "__type": "PermissionDescriptorAllow",
                "permissions": ["organization:read_usage"],
            }
        ],
    },
)

pulumi.export("roleId", role.id)
pulumi.export("detailsOut", role.details)
