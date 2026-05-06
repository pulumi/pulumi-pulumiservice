import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";
const memberLogin = config.get("memberLogin") ?? "pulumi-bot";
const memberRole = config.get("memberRole") ?? "member";

const member = new ps.v2.OrganizationMember("member", {
    orgName: serviceOrg,
    userLogin: memberLogin,
    role: memberRole,
});

export const memberId = member.id;
