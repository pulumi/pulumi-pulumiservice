import * as service from "@pulumi/pulumiservice";
import * as pulumi from "@pulumi/pulumi";

const config = new pulumi.Config();

const environment = new service.Environment("testing-environment", {
  organization: "service-provider-test-org",
  name: "testing-environment-ts-file-asset"+config.require("digits"),
  yaml: new pulumi.asset.FileAsset("env.yaml")
})
