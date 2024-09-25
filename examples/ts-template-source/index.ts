import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const source = new service.TemplateSource("source", {
    organizationName: "service-provider-test-org",
    sourceName: "bootstrap-ts",
    sourceURL: "https://github.com/pulumi/pulumi-pulumiservice",
    destination: {
      url: "https://github.com/pulumi/pulumi-pulumiservice"
    }
  });