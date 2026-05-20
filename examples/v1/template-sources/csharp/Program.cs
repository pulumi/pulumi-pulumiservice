using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var templateSuffix = config.Get("templateSuffix") ?? "dev";
    var sourceUrl = config.Get("sourceUrl") ?? "https://github.com/pulumi/examples";

    var source = new Ps.V1.OrgTemplateCollection("source", new()
    {
        OrgName = organizationName,
        Name = $"v1-templates-{templateSuffix}",
        SourceURL = sourceUrl,
    });

    return new Dictionary<string, object?>
    {
        ["collectionName"] = source.Name,
    };
});
