using System.Collections.Generic;
using System.Collections.Immutable;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var nameSuffix = config.Get("nameSuffix") ?? "manual";
    var roleDescription = config.Get("roleDescription") ?? "Read-only access to stacks, created by the v2 rbac example.";

    var readOnlyRole = new Ps.V2.Role("readOnlyRole", new()
    {
        OrgName = serviceOrg,
        Name = $"v2-rbac-read-only-{nameSuffix}",
        Description = roleDescription,
        UxPurpose = "role",
        Details = ImmutableDictionary.CreateRange(new[]
        {
            new KeyValuePair<string, object>("__type", "PermissionDescriptorAllow"),
            new KeyValuePair<string, object>("permissions", new[] { "stack:read" }),
        }),
    });

    var rbacTeam = new Ps.V2.Team("rbacTeam", new()
    {
        OrgName = serviceOrg,
        Name = $"v2-rbac-team-{nameSuffix}",
        DisplayName = $"v2 RBAC Team {nameSuffix}",
        Description = "Team scaffold used by the v2 rbac example.",
    });

    return new Dictionary<string, object?>
    {
        ["roleName"] = readOnlyRole.Name,
        ["teamName"] = rbacTeam.Name,
    };
});
