import * as pulumiservice from "@pulumi/pulumiservice";
import * as pulumi from "@pulumi/pulumi";
import * as random from "@pulumi/random";

// Create the stack resource first so it exists before we reference it
const stackSuffix = new random.RandomPet("stack-suffix", {
  prefix: pulumi.getStack(),
  separator: "-",
});
const stack = new pulumiservice.Stack("test-stack", {
  organizationName: "service-provider-test-org",
  projectName: pulumi.getProject(),
  stackName: stackSuffix.id,
});

export const stackTags = {
  "test-tag": "test-value",
  "another-tag": "another-value",
  "a-third-tag": "just because",
};

Object.entries(stackTags).forEach(([k, v]) => {
  new pulumiservice.StackTag(k, {
    organization: stack.organizationName,
    project: stack.projectName,
    stack: stack.stackName,
    name: k,
    value: v,
  });
});
