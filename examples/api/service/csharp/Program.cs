using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var serviceSuffix = config.Get("serviceSuffix") ?? "dev";

    new Ps.Api.Services.Service("catalogService", new()
    {
        OrgName = organizationName,
        Name = $"api-service-{serviceSuffix}",
        Description = "An example api service catalog entry.",
        OwnerType = "team",
        OwnerName = "platform",
        Items =
        {
            new Ps.Api.Inputs.AddServiceItemArgs { Type = "stack", Name = "service-provider-test-org/example-app/dev" },
        },
        Properties =
        {
            new Ps.Api.Inputs.ServicePropertyArgs { Key = "tier", Value = "gold", Type = "string", Order = 1 },
            new Ps.Api.Inputs.ServicePropertyArgs { Key = "oncall", Value = "platform-ops", Type = "string", Order = 2 },
        },
    });
});
