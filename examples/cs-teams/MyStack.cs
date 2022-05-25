using Pulumi;
using Pulumi.PulumiService;
using Pulumi.Random;

class MyStack : Stack
{
    public MyStack()
    {
        var random = new RandomString("rand", new RandomStringArgs
        {
            Length = 5,
            Special = false,
        });
        var team = new Team("team", new TeamArgs
        {
            // Suffix random string to avoid collisions if multiple tests are running at once
            Name = random.Result.Apply(res => "brand-new-dotnet-team-" + res),
            OrganizationName = "service-provider-test-org",
            DisplayName = "PulumiUP Dotnet Team",
            TeamType = "pulumi",
            Members = {
                "pulumi-bot",
                "service-provider-example-user",
            }
        });
    }
}
