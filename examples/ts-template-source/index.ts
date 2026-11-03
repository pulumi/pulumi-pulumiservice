import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

let config = new pulumi.Config();
let digits = config.require("digits");
let sourceURL = config.require("sourceURL");

const source = new service.TemplateSource("source", {
    organizationName: "service-provider-test-org",
    sourceName: "bootstrap-"+digits,
    sourceURL: sourceURL,
    destination: {
      url: "https://github.com/pulumi/pulumi-pulumiservice"
    }
  });