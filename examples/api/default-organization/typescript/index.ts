import * as pulumi from "@pulumi/pulumi";
import * as ps from "@pulumi/pulumiservice";

const config = new pulumi.Config();
const organizationName = config.get("organizationName") ?? "service-provider-test-org";

const def = new ps.api.DefaultOrganization("default", {
    orgName: organizationName,
});

// orgName is an input (program-owned); reference the source value.
// def's outputs are GitHubLogin/Messages — surface those instead.
export const defaultOrg = organizationName;
export const defaultOrgGitHubLogin = def.GitHubLogin;
