using System.Collections.Immutable;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var policyId = config.Get("policyId") ?? "org";

    var allowPolicy = new Ps.Api.Auth.Inputs.AuthPolicyDefinitionArgs
    {
        Decision = "allow",
        TokenType = "organization",
        AuthorizedPermissions = new[] { "standard" },
        Rules = ImmutableDictionary<string, ImmutableDictionary<string, object>>.Empty,
    };
    var denyPolicy = new Ps.Api.Auth.Inputs.AuthPolicyDefinitionArgs
    {
        Decision = "deny",
        TokenType = "organization",
        AuthorizedPermissions = new[] { "admin" },
        Rules = ImmutableDictionary<string, ImmutableDictionary<string, object>>.Empty,
    };

    new Ps.Api.Auth.Policy("policy", new()
    {
        OrgName = organizationName,
        PolicyId = policyId,
        Policies = new[] { allowPolicy, denyPolicy },
    });
});
