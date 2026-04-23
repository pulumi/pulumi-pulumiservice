// C# variant of canonical/07-tiered-team-access.
// Three teams × two stacks × three permission tiers. Behavioral twin of
// the sibling YAML program.

using System.Collections.Generic;
using Pulumi;
using PulumiService = Pulumi.PulumiService;

return await Pulumi.Deployment.RunAsync(() =>
{
    var cfg = new Config();
    var organizationName = cfg.Get("organizationName") ?? "service-provider-test-org";
    var digits = cfg.Get("digits") ?? "00000";

    // Team-stack permission levels: 0=none, 101=read, 102=edit, 103=admin.
    const int permRead = 101;
    const int permEdit = 102;
    const int permAdmin = 103;

    var platformAdmins = new PulumiService.Orgs.Teams.Team("platformAdmins", new()
    {
        OrganizationName = organizationName,
        TeamType = "pulumi",
        Name = $"platform-admins-{digits}",
        DisplayName = "Platform Admins",
        Description = "Break-glass access to everything. Keep small.",
    });
    var billingOwners = new PulumiService.Orgs.Teams.Team("billingOwners", new()
    {
        OrganizationName = organizationName,
        TeamType = "pulumi",
        Name = $"billing-owners-{digits}",
        DisplayName = "Billing Service Owners",
        Description = "Owns the billing service stacks end-to-end.",
    });
    var developers = new PulumiService.Orgs.Teams.Team("developers", new()
    {
        OrganizationName = organizationName,
        TeamType = "pulumi",
        Name = $"developers-{digits}",
        DisplayName = "Developers (all)",
        Description = "Read everything; deploy nothing without an explicit grant.",
    });

    void Grant(string name, string project, Output<string> team, int perm) =>
        _ = new PulumiService.Stacks.Permissions.TeamStackPermission(name, new()
        {
            Organization = organizationName,
            Project = project,
            Stack = "prod",
            Team = team,
            Permission = perm,
        });

    Grant("platformAdminPerm", "platform", platformAdmins.Name!, permAdmin);
    Grant("platformDevRead", "platform", developers.Name!, permRead);
    Grant("billingAdminPerm", "billing-service", platformAdmins.Name!, permAdmin);
    Grant("billingOwnerPerm", "billing-service", billingOwners.Name!, permEdit);
    Grant("billingDevRead", "billing-service", developers.Name!, permRead);

    return new Dictionary<string, object?>();
});
