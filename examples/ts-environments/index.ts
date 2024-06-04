import * as service from "@pulumi/pulumiservice";
import * as pulumi from "@pulumi/pulumi";

var environment = new service.Environment("testing-environment", {
  organization: "service-provider-test-org",
  name: "testing-environment-ts",
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
  tagName: "stable",
  revision: environment.revision
})

// A tag that will be placed on each new version, and remain on old revisions
var versionTag = new service.EnvironmentVersionTag("VersionTag", {
  organization: environment.organization,
  environment: environment.name,
  tagName: environment.revision.apply((rev: number) => "v"+rev),
  revision: environment.revision
}, {
  retainOnDelete: true
})
