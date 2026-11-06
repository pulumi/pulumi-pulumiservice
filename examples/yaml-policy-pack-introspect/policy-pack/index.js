"use strict";

const { PolicyPack, validateResourceOfType } = require("@pulumi/policy");

new PolicyPack("example-pack", {
    policies: [
        {
            name: "no-public-buckets",
            description: "Disallow public S3 buckets.",
            enforcementLevel: "mandatory",
            validateResource: validateResourceOfType("aws:s3/bucket:Bucket", (bucket, _args, reportViolation) => {
                if (bucket.acl === "public-read" || bucket.acl === "public-read-write") {
                    reportViolation("Public S3 buckets are not allowed.");
                }
            }),
        },
        {
            name: "required-bucket-tag",
            description: "S3 buckets must have a configurable tag key.",
            enforcementLevel: "advisory",
            configSchema: {
                properties: {
                    tagKey: { type: "string", default: "owner" },
                },
                required: ["tagKey"],
            },
            validateResource: validateResourceOfType("aws:s3/bucket:Bucket", (bucket, args, reportViolation) => {
                const { tagKey } = args.getConfig() || {};
                if (!tagKey) return;
                const tags = bucket.tags || {};
                if (!tags[tagKey]) {
                    reportViolation(`S3 bucket must have a '${tagKey}' tag.`);
                }
            }),
        },
    ],
});
