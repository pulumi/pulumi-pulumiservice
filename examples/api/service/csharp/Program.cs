using System.Collections.Immutable;
using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var serviceSuffix = config.Get("serviceSuffix") ?? "dev";

    var stackItem = ImmutableDictionary<string, object>.Empty
        .Add("kind", "stack")
        .Add("ref", "service-provider-test-org/example-app/dev");
    var tierProp = ImmutableDictionary<string, object>.Empty
        .Add("key", "tier")
        .Add("value", "gold");
    var oncallProp = ImmutableDictionary<string, object>.Empty
        .Add("key", "oncall")
        .Add("value", "platform-ops");

    new Ps.Api.Services.Service("catalogService", new()
    {
        OrgName = organizationName,
        Name = $"api-service-{serviceSuffix}",
        Description = "An example api service catalog entry.",
        OwnerType = "team",
        OwnerName = "platform",
        Items = new object[] { stackItem },
        Properties = new object[] { tierProp, oncallProp },
    });
});
