import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";
import * as random from "@pulumi/random";

let config = new pulumi.Config();

// Create the stack resource first so it exists before we reference it
const stackSuffix = new random.RandomPet("stack-suffix", {
  prefix: pulumi.getStack(),
  separator: "-",
});
const stack = new service.Stack("test-stack", {
  organizationName: "service-provider-test-org",
  projectName: pulumi.getProject(),
  stackName: stackSuffix.id,
});

const team = new service.Team("team", {
  organizationName: "service-provider-test-org",
  name: "pulumi-service-team-"+config.require("digits"),
  teamType: "pulumi",
  members: ["pulumi-bot", "service-provider-example-user"]
});

new service.TeamStackPermission("team-permission", {
  organization: team.organizationName,
  project: stack.projectName,
  stack: stack.stackName,
  team: team.name as pulumi.Output<string>,
  permission: service.TeamStackPermissionScope.Admin,
});
