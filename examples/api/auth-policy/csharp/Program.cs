using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var policyId = config.Get("policyId") ?? "org";

    new Ps.Api.Auth.Policy("policy", new()
    {
        OrgName = organizationName,
        PolicyId = policyId,
        Policies =
        {
            new Ps.Api.Inputs.AuthPolicyDefinitionArgs
            {
                Decision = "allow",
                AuthorizedPermissions = { "read" },
                TokenType = "organization",
            },
            new Ps.Api.Inputs.AuthPolicyDefinitionArgs
            {
                Decision = "deny",
                AuthorizedPermissions = { "admin" },
                TokenType = "organization",
            },
        },
    });
});
