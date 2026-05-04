import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
issue_id = config.get("issueId") or "example-issue-id"

ps_v2.PolicyIssue(
    "issue",
    org_name=service_org,
    issue_id=issue_id,
    priority="high",
    status="in_progress",
    assigned_to="pulumi-bot",
)
