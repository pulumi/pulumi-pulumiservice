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
