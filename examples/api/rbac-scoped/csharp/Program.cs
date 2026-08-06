using System.Collections.Generic;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var nameSuffix = config.Get("nameSuffix") ?? "manual";
    var roleDescription = config.Get("roleDescription") ?? "Environment-scoped read access, created by the api rbac-scoped example.";
    var allowedEnvironmentId = config.Get("allowedEnvironmentId") ?? "c5549aa1-87db-4d67-a195-455b56772900";
    var deniedEnvironmentId = config.Get("deniedEnvironmentId") ?? "3cb9b7ad-0848-4e0d-aeff-8e9f093fd2d9";

    // Every level is a typed discriminated-union variant class.
    Ps.Api.Inputs.PermissionExpressionEqualArgs EnvMatch(string identity) => new()
    {
        Type__ = "PermissionExpressionEqual",
        Left = new Ps.Api.Inputs.PermissionExpressionEnvironmentArgs
        {
            Type__ = "PermissionExpressionEnvironment",
        },
        Right = new Ps.Api.Inputs.PermissionLiteralExpressionEnvironmentArgs
        {
            Type__ = "PermissionLiteralExpressionEnvironment",
            Identity = identity,
        },
    };

    var scopedRole = new Ps.Api.Role("scopedRole", new()
    {
        OrgName = organizationName,
        Name = $"api-rbac-scoped-{nameSuffix}",
        Description = roleDescription,
        UxPurpose = "role",
        Details = new Ps.Api.Inputs.PermissionDescriptorGroupArgs
        {
            Type__ = "PermissionDescriptorGroup",
            Entries = new object[]
            {
                new Ps.Api.Inputs.PermissionDescriptorConditionArgs
                {
                    Type__ = "PermissionDescriptorCondition",
                    Condition = new Ps.Api.Inputs.PermissionExpressionAndArgs
                    {
                        Type__ = "PermissionExpressionAnd",
                        Left = EnvMatch(allowedEnvironmentId),
                        Right = new Ps.Api.Inputs.PermissionExpressionNotArgs
                        {
                            Type__ = "PermissionExpressionNot",
                            Node = EnvMatch(deniedEnvironmentId),
                        },
                    },
                    SubNode = new Ps.Api.Inputs.PermissionDescriptorAllowArgs
                    {
                        Type__ = "PermissionDescriptorAllow",
                        Permissions = new[] { "environment:read", "environment:open" },
                    },
                },
            },
        },
    });

    return new Dictionary<string, object?>
    {
        ["roleName"] = scopedRole.Name,
    };
});
