import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const issueId = config.get("issueId") ?? "example-issue-id";

new ps.v2.PolicyIssue("issue", {
    orgName: serviceOrg,
    issueId: issueId,
    priority: "high",
    status: "in_progress",
    assignedTo: "pulumi-bot",
});
