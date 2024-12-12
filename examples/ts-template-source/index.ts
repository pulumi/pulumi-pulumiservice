import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

let config = new pulumi.Config();
let digits = config.require("digits");

const source = new service.TemplateSource("source", {
    organizationName: "service-provider-test-org",
    sourceName: "bootstrap-"+digits,
    sourceURL: "https://github.com/pulumi/pulumi-pulumiservice",
    destination: {
      url: "https://github.com/pulumi/pulumi-pulumiservice"
    }
  });