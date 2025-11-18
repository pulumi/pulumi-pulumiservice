import * as service from "@pulumi/pulumiservice";
import * as pulumi from "@pulumi/pulumi";
import * as path from "path";

const config = new pulumi.Config();

const environment = new service.Environment("testing-environment", {
  organization: "service-provider-test-org",
  project: "my-project",
  name: "testing-environment-ts-file-asset"+config.require("digits"),
  yaml: new pulumi.asset.FileAsset(path.join(pulumi.runtime.getRootDirectory(), "env.yaml"))
})
