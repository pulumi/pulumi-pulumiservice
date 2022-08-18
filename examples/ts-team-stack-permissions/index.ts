import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";
import * as aws from "@pulumi/aws";

new aws.Ec2();

new service.TeamStackPermission("team-permission", {
  organization: "service-provider-test-org",
  project: pulumi.getProject(),
  stack: pulumi.getStack(),
  team: "pulumi-service-team",
  permission: service.TeamStackPermissionScope.Admin,
});
