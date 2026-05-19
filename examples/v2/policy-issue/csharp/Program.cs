using Pulumi;
using Ps = Pulumi.PulumiService;

return await Deployment.RunAsync(() =>
{
    var config = new Config();
    var organizationName = config.Get("organizationName") ?? "service-provider-test-org";
    var issueId = config.Get("issueId") ?? "example-issue-id";

    new Ps.V2.PolicyIssue("issue", new()
    {
        OrgName = organizationName,
        IssueId = issueId,
        Priority = "high",
        Status = "in_progress",
        AssignedTo = "pulumi-bot",
    });
});
