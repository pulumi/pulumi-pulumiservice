import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const serviceOrg = config.get("serviceOrg") ?? "service-provider-test-org";

const def = new ps.v2.DefaultOrganization("default", {
    orgName: serviceOrg,
});

// orgName is an input (program-owned); reference the source value.
// def's outputs are GitHubLogin/Messages — surface those instead.
export const defaultOrg = serviceOrg;
export const defaultOrgGitHubLogin = def.GitHubLogin;
