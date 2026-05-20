using System.Collections.Generic;
using System.Collections.Immutable;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var nameSuffix = config.Get("nameSuffix") ?? "manual";
    var roleDescription = config.Get("roleDescription") ?? "Read-only access to stacks, created by the v1 rbac example.";

    var readOnlyRole = new Ps.V1.Role("readOnlyRole", new()
    {
        OrgName = organizationName,
        Name = $"v1-rbac-read-only-{nameSuffix}",
        Description = roleDescription,
        UxPurpose = "role",
        Details = ImmutableDictionary.CreateRange(new[]
        {
            new KeyValuePair<string, object>("__type", "PermissionDescriptorAllow"),
            new KeyValuePair<string, object>("permissions", new[] { "stack:read" }),
        }),
    });

    var rbacTeam = new Ps.V1.Teams.Team("rbacTeam", new()
    {
        OrgName = organizationName,
        Name = $"v1-rbac-team-{nameSuffix}",
        DisplayName = $"v1 RBAC Team {nameSuffix}",
        Description = "Team scaffold used by the v1 rbac example.",
    });

    return new Dictionary<string, object?>
    {
        ["roleName"] = readOnlyRole.Name,
        ["teamName"] = rbacTeam.Name,
    };
});
