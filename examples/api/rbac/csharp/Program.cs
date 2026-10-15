using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var nameSuffix = config.Get("nameSuffix") ?? "manual";
    var roleDescription = config.Get("roleDescription") ?? "Stack read everywhere, plus write on stacks tagged team=platform.";

    // Pulumi Cloud models a role in three layers: a set grants scopes, a policy
    // composes sets (optionally gated on a condition), and a role composes policies.
    var stackReadSet = new Ps.Api.Role("stackReadSet", new()
    {
        OrgName = organizationName,
        Name = $"api-rbac-stack-read-{nameSuffix}",
        UxPurpose = "set",
        ResourceType = "stack",
        Details = Ps.BuildAllowPermissions.Invoke(new()
        {
            Permissions = { "stack:read" },
        }).Apply(r => (object)r.Permissions),
    });

    var stackWriteSet = new Ps.Api.Role("stackWriteSet", new()
    {
        OrgName = organizationName,
        Name = $"api-rbac-stack-write-{nameSuffix}",
        UxPurpose = "set",
        ResourceType = "stack",
        Details = Ps.BuildAllowPermissions.Invoke(new()
        {
            Permissions = { "stack:read", "stack:write" },
        }).Apply(r => (object)r.Permissions),
    });

    var platformPolicy = new Ps.Api.Role("platformPolicy", new()
    {
        OrgName = organizationName,
        Name = $"api-rbac-policy-{nameSuffix}",
        UxPurpose = "policy",
        Details = Ps.BuildGroupPermissions.Invoke(new()
        {
            Entries =
            {
                Ps.BuildComposePermissions.Invoke(new()
                {
                    PermissionDescriptorIds = { stackReadSet.RoleID },
                }).Apply(r => r.Permissions),
                Ps.BuildTagConditionalPermissions.Invoke(new()
                {
                    EntityType = Ps.RbacEntityType.Stack,
                    TagKey = "team",
                    TagValue = "platform",
                    SetIds = { stackWriteSet.RoleID },
                }).Apply(r => r.Permissions),
            },
        }).Apply(r => (object)r.Permissions),
    });

    var platformRole = new Ps.Api.Role("platformRole", new()
    {
        OrgName = organizationName,
        Name = $"api-rbac-role-{nameSuffix}",
        Description = roleDescription,
        UxPurpose = "role",
        Details = Ps.BuildComposePermissions.Invoke(new()
        {
            PermissionDescriptorIds = { platformPolicy.RoleID },
        }).Apply(r => (object)r.Permissions),
    });

    var rbacTeam = new Ps.Api.Teams.Team("rbacTeam", new()
    {
        OrgName = organizationName,
        Name = $"api-rbac-team-{nameSuffix}",
        DisplayName = $"api RBAC Team {nameSuffix}",
        Description = "Team scaffold used by the api rbac example.",
    });

    return new Dictionary<string, object?>
    {
        ["roleName"] = platformRole.Name,
        ["teamName"] = rbacTeam.Name,
    };
});
