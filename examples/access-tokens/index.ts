import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const accessToken = new service.AccessToken("token", {
    description: "example-accesstoken"
});

export const token = accessToken.value;