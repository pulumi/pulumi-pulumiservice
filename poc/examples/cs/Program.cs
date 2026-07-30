using System;
using System.Collections.Generic;
using Pulumi;
using PulumiService = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var org = Environment.GetEnvironmentVariable("POC_ORG") ?? "poc-org";

    // The union slot is object? (6 members exceeds Union<T0,T1>), but the
    // typed variant args classes exist and nest: Group -> entries[Allow].
    var role = new PulumiService.Api.Role("poc-union-role", new()
    {
        OrgName = org,
        Name = "poc-union-role",
        UxPurpose = "set",
        ResourceType = "global",
        Description = "POC: discriminated union permission tree",
        Details = new PulumiService.Api.Inputs.PermissionDescriptorGroupArgs
        {
            __type = "PermissionDescriptorGroup",
            // InputList<object> collection initializers are ambiguous for T=object
            // (Add(params Input<object>[]) vs Add(InputList<object>)); assign an
            // object[] instead.
            Entries = new object[]
            {
                new PulumiService.Api.Inputs.PermissionDescriptorAllowArgs
                {
                    __type = "PermissionDescriptorAllow",
                    Permissions = new[] { "organization:read_usage" },
                },
            },
        },
    });

    return new Dictionary<string, object?>
    {
        ["roleId"] = role.Id,
        ["detailsOut"] = role.Details,
    };
});
