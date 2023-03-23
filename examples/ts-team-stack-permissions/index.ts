import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const team = new service.Team("team", {
  organizationName: "service-provider-test-org",
  name: "pulumi-service-team",
  teamType: "pulumi",
  members: ["pulumi-bot", "service-provider-example-user"]
});

new service.TeamStackPermission("team-permission", {
  organization: "service-provider-test-org",
  project: pulumi.getProject(),
  stack: pulumi.getStack(),
  team: team.name as pulumi.Output<string>,
  permission: service.TeamStackPermissionScope.Admin,
});
