using System.Collections.Generic;
using System.Collections.Immutable;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";

    var pack = new Ps.V2.PolicyPack("pack", new()
    {
        OrgName = serviceOrg,
        Name = "v2-example-policy-pack",
        DisplayName = "v2 example policy pack",
        Description = "Demo policy pack created via v2 metadata-driven provider.",
        Policies = new[]
        {
            ImmutableDictionary.CreateRange(new[]
            {
                new KeyValuePair<string, object>("name", "no-public-buckets"),
                new KeyValuePair<string, object>("description", "Reject S3 buckets with public ACLs"),
                new KeyValuePair<string, object>("enforcementLevel", "advisory"),
            }),
        },
    });

    return new Dictionary<string, object?>
    {
        ["policyPackName"] = pack.Name,
        ["policyPackVersion"] = pack.Version,
    };
});
