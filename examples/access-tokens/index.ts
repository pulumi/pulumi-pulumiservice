import * as pulumi from "@pulumi/pulumi";
import * as service from "@pulumi/pulumiservice";

const accesstoken = new service.AccessToken("token", {
    description: "example-accesstoken"
})