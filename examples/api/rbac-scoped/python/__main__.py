import pulumi
import pulumi_pulumiservice as pulumiservice

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
name_suffix = config.get("nameSuffix") or "manual"
role_description = config.get("roleDescription") or "Environment-scoped read access, created by the api rbac-scoped example."
allowed_environment_id = config.get("allowedEnvironmentId") or "c5549aa1-87db-4d67-a195-455b56772900"
denied_environment_id = config.get("deniedEnvironmentId") or "3cb9b7ad-0848-4e0d-aeff-8e9f093fd2d9"


# Dict form: the typed variant classes are pending an upstream fix for
# double-underscore property names in Python codegen.
def env_match(identity):
    return {
        "type__": "PermissionExpressionEqual",
        "left": {"type__": "PermissionExpressionEnvironment"},
        "right": {"type__": "PermissionLiteralExpressionEnvironment", "identity": identity},
    }


scoped_role = pulumiservice.api.Role(
    "scopedRole",
    org_name=organization_name,
    name=f"api-rbac-scoped-{name_suffix}",
    description=role_description,
    ux_purpose="role",
    details={
        "type__": "PermissionDescriptorGroup",
        "entries": [
            {
                "type__": "PermissionDescriptorCondition",
                "condition": {
                    "type__": "PermissionExpressionAnd",
                    "left": env_match(allowed_environment_id),
                    "right": {
                        "type__": "PermissionExpressionNot",
                        "node": env_match(denied_environment_id),
                    },
                },
                "subNode": {
                    "type__": "PermissionDescriptorAllow",
                    "permissions": ["environment:read", "environment:open"],
                },
            }
        ],
    },
)

pulumi.export("roleName", scoped_role.name)
