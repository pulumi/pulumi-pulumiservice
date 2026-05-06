import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";

const def = new ps.v2.DefaultOrganization("default", {
    orgName: serviceOrg,
});

export const defaultOrg = def.orgName;
