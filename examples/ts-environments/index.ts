import * as service from "@pulumi/pulumiservice";
import * as pulumi from "@pulumi/pulumi";

let config = new pulumi.Config();

var environment = new service.Environment("testing-environment", {
  organization: "service-provider-test-org",
  project: "my-project",
  name: "testing-environment-ts-"+config.require("digits"),
  yaml: new pulumi.asset.StringAsset(
`values:
  myKey1: "myValue1"
  myNestedKey:
    myKey2: "myValue2"
    myNumber: 1`
  )
})

// A tag that will always be placed on the latest revision of the environment
var stableTag = new service.EnvironmentVersionTag("StableTag", {
  organization: environment.organization,
  environment: environment.name,
  project: environment.project,
  tagName: "stable",
  revision: environment.revision
})

// A tag that will be placed on each new version, and remain on old revisions
var versionTag = new service.EnvironmentVersionTag("VersionTag", {
  organization: environment.organization,
  environment: environment.name,
  project: environment.project,
  tagName: environment.revision.apply((rev: number) => "v"+rev),
  revision: environment.revision
}, {
  retainOnDelete: true
})

const team = new service.Team("team", {
  description: "This was created with Pulumi",
  name: "ts-team-needing-permissions",
  displayName: "PulumiUP Team",
  organizationName: environment.organization,
  members: ["pulumi-bot", "service-provider-example-user"],
  teamType: "pulumi"
});

const teamEnvironmentPermission = new service.TeamEnvironmentPermission("teamEnvironmentPermission", {
  organization: environment.organization,
  team: team.name.apply((name: any) => name!!),
  environment: environment.name,
  project: environment.project,
  permission: "admin"
});
