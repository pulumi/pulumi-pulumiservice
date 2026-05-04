using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var secretValue = config.Get("secretValue") ?? "shhh";
    var hookSuffix = config.Get("hookSuffix") ?? "dev";

    var orgWebhookAll = new Ps.V2.OrganizationWebhook("orgWebhookAll", new()
    {
        OrganizationName = serviceOrg,
        Name = $"org-webhook-all-{hookSuffix}",
        DisplayName = "webhook-from-provider",
        PayloadUrl = "https://google.com",
        Active = true,
        Secret = secretValue,
    });

    var orgWebhookGroups = new Ps.V2.OrganizationWebhook("orgWebhookGroups", new()
    {
        OrganizationName = serviceOrg,
        Name = $"org-webhook-groups-{hookSuffix}",
        DisplayName = "webhook-from-provider",
        PayloadUrl = "https://google.com",
        Active = true,
        Groups = new[] { "environments", "stacks" },
        Secret = secretValue,
    });

    return new Dictionary<string, object?>
    {
        ["orgWebhookId"] = orgWebhookAll.Id,
        ["orgWebhookGroupsId"] = orgWebhookGroups.Id,
    };
});
