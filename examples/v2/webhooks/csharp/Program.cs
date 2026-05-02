using System.Collections.Generic;
using System.Linq;
using Pulumi;
using PulumiService = Pulumi.PulumiService;

return await Deployment.RunAsync(() => 
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var secretValue = config.Get("secretValue") ?? "shhh";
    // Organization-scoped webhook subscribed to all events.
    var orgWebhookAll = new PulumiService.V2.OrganizationWebhook("orgWebhookAll", new()
    {
        OrgName = serviceOrg,
        OrganizationName = serviceOrg,
        Name = "org-webhook-all",
        DisplayName = "webhook-from-provider",
        PayloadUrl = "https://google.com",
        Active = true,
        Secret = secretValue,
    });

    // Organization-scoped webhook subscribed only to environments and stacks groups.
    var orgWebhookGroups = new PulumiService.V2.OrganizationWebhook("orgWebhookGroups", new()
    {
        OrgName = serviceOrg,
        OrganizationName = serviceOrg,
        Name = "org-webhook-groups",
        DisplayName = "webhook-from-provider",
        PayloadUrl = "https://google.com",
        Active = true,
        Groups = new[]
        {
            "environments",
            "stacks",
        },
        Secret = secretValue,
    });

    return new Dictionary<string, object?>
    {
        ["orgWebhookId"] = orgWebhookAll.Id,
    };
});

