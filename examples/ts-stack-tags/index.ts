import * as pulumiservice from "@pulumi/pulumiservice";
import * as pulumi from "@pulumi/pulumi";

const config = new pulumi.Config()

export const stackTags = {
    "test-tag": "test-value",
    "another-tag": "another-value",
    "a-third-tag": "just because"
}

Object.entries(stackTags).forEach(([k, v]) => {
    new pulumiservice.StackTag(k, {
        organization: "pulumi",
        project: pulumi.getProject(),
        stack: pulumi.getStack(),
        name: k,
        value: v,
    })
})