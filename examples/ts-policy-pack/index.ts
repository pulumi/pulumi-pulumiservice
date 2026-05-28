import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";
import * as path from "path";

const config = new pulumi.Config();
const organizationName = config.require("organizationName");
const digits = config.require("digits");
const versionTag = config.get("versionTag") ?? "1.0.0";

// `policies` is omitted: the provider extracts metadata by running the policy
// analyzer plugin against `sourcePath` at publish time.
const pack = new service.PolicyPack("example-pack", {
    organization: organizationName,
    name: `example-pack-${digits}`,
    displayName: "Example Pack",
    versionTag,
    sourcePath: path.join(__dirname, "policy-pack"),
});

// The PolicyGroup applies the pack and supplies per-policy configuration. The
// shape of each entry must satisfy the matching policy's configSchema.
const group = new service.PolicyGroup("example-group", {
    organizationName,
    name: `example-group-${digits}`,
    policyPacks: [
        {
            name: pack.name,
            versionTag: pack.versionTag,
        },
    ],
});

export const policyPackVersion = pack.version;
export const policyPackHash = pack.contentHash;
export const policyGroupName = group.name;
