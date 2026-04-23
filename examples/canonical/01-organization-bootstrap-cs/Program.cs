// C# variant of canonical/01-organization-bootstrap.
// Day-0 provisioning: three teams, a team-scoped CI token, a baseline
// policy group. Functionally equivalent to the sibling YAML program.

using Pulumi;
using PulumiService = Pulumi.PulumiService;

return await Pulumi.Deployment.RunAsync(() =>
{
    var cfg = new Config();
    var organizationName = cfg.Get("organizationName") ?? "service-provider-test-org";
    var digits = cfg.Get("digits") ?? "00000";

    _ = new PulumiService.Orgs.Teams.Team("admins", new()
    {
        OrganizationName = organizationName,
        TeamType = "pulumi",
        Name = $"admins-{digits}",
        DisplayName = "Organization Admins",
        Description = "Full org control; rotate this membership quarterly.",
    });

    var deployers = new PulumiService.Orgs.Teams.Team("deployers", new()
    {
        OrganizationName = organizationName,
        TeamType = "pulumi",
        Name = $"deployers-{digits}",
        DisplayName = "CI Deployers",
        Description = "Automation-only team. Human members discouraged — use the CI token.",
    });

    _ = new PulumiService.Orgs.Teams.Team("readers", new()
    {
        OrganizationName = organizationName,
        TeamType = "pulumi",
        Name = $"readers-{digits}",
        DisplayName = "Developers (read-only)",
        Description = "Default team for new org members; grants stack read access.",
    });

    var ciToken = new PulumiService.Orgs.Tokens.TeamAccessToken("ciToken", new()
    {
        OrganizationName = organizationName,
        TeamName = deployers.Name!,
        Description = "Used by GitHub Actions to deploy non-production stacks.",
    });

    _ = new PulumiService.Orgs.Policies.PolicyGroup("defaultGuardrails", new()
    {
        OrganizationName = organizationName,
        Name = $"baseline-guardrails-{digits}",
    });

    return new Dictionary<string, object?>
    {
        ["ciTokenValue"] = ciToken.Value,
    };
});
