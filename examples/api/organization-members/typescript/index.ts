import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";
const memberLogin = config.get("memberLogin") ?? "pulumi-bot";
const memberRole = config.get("memberRole") ?? "member";

const member = new ps.api.OrganizationMember("member", {
    orgName: organizationName,
    userLogin: memberLogin,
    role: memberRole,
});

export const memberId = member.id;
