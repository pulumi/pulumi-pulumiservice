using System.Collections.Immutable;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var serviceOrg = config.Get("serviceOrg") ?? "service-provider-test-org";
    var policyId = config.Get("policyId") ?? "org";

    var allowPolicy = ImmutableDictionary<string, object>.Empty
        .Add("decision", "allow")
        .Add("permission", "read")
        .Add("tokenType", "organization");
    var denyPolicy = ImmutableDictionary<string, object>.Empty
        .Add("decision", "deny")
        .Add("permission", "admin")
        .Add("tokenType", "organization");

    new Ps.V2.Auth.Policy("policy", new()
    {
        OrgName = serviceOrg,
        PolicyId = policyId,
        Policies = new object[] { allowPolicy, denyPolicy },
    });
});
