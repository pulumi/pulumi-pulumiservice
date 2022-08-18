import * as pulumiservice from "@pulumi/pulumiservice";
import * as pulumi from "@pulumi/pulumi";

export const stackTags = {
  "test-tag": "test-value",
  "another-tag": "another-value",
  "a-third-tag": "just because",
};

Object.entries(stackTags).forEach(([k, v]) => {
  new pulumiservice.StackTag(k, {
    organization: "service-provider-test-org",
    project: pulumi.getProject(),
    stack: pulumi.getStack(),
    name: k,
    value: v,
  });
});
