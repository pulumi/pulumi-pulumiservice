using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var secretValue = config.Get("secretValue") ?? "shhh";
    var hookSuffix = config.Get("hookSuffix") ?? "dev";

    var orgWebhookAll = new Ps.V1.OrganizationWebhook("orgWebhookAll", new()
    {
        OrganizationName = organizationName,
        Name = $"org-webhook-all-{hookSuffix}",
        DisplayName = "webhook-from-provider",
        PayloadUrl = "https://google.com",
        Active = true,
        Secret = secretValue,
    });

    var orgWebhookGroups = new Ps.V1.OrganizationWebhook("orgWebhookGroups", new()
    {
        OrganizationName = organizationName,
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
