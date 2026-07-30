using System.Collections.Immutable;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var policyId = config.Get("policyId") ?? "org";

    var allowPolicy = ImmutableDictionary<string, object>.Empty
        .Add("decision", "allow")
        .Add("tokenType", "organization")
        .Add("authorizedPermissions", new[] { "standard" })
        .Add("rules", ImmutableDictionary<string, object>.Empty);
    var denyPolicy = ImmutableDictionary<string, object>.Empty
        .Add("decision", "deny")
        .Add("tokenType", "organization")
        .Add("authorizedPermissions", new[] { "admin" })
        .Add("rules", ImmutableDictionary<string, object>.Empty);

    new Ps.Api.Auth.Policy("policy", new()
    {
        OrgName = organizationName,
        PolicyId = policyId,
        Policies = new object[] { allowPolicy, denyPolicy },
    });
});
