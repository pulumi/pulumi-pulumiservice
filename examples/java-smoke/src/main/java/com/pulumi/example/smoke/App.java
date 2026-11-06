package com.pulumi.example.smoke;

import com.pulumi.Pulumi;
import com.pulumi.asset.StringAsset;
import com.pulumi.pulumiservice.Environment;
import com.pulumi.pulumiservice.EnvironmentArgs;
import com.pulumi.pulumiservice.OrgAccessToken;
import com.pulumi.pulumiservice.OrgAccessTokenArgs;
import com.pulumi.pulumiservice.PulumiserviceFunctions;
import com.pulumi.pulumiservice.inputs.GetEnvironmentArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var org = ctx.config().require("organizationName");
            var name = "test-smoke-java-" + ctx.config().require("digits");

            var environment = new Environment("environment", EnvironmentArgs.builder()
                    .organization(org)
                    .project("test-smoke")
                    .name(name)
                    .yaml(new StringAsset("values:\n  greeting: hello\n"))
                    .build());

            var token = new OrgAccessToken("token", OrgAccessTokenArgs.builder()
                    .organizationName(org)
                    .name(name)
                    .description("Created by the java-smoke example.")
                    .admin(false)
                    .build());

            var lookedUp = PulumiserviceFunctions.getEnvironment(GetEnvironmentArgs.builder()
                    .organizationName(environment.organization())
                    .projectName(environment.project())
                    .name(environment.name())
                    .build());

            ctx.export("environmentName", lookedUp.applyValue(result -> result.name()));
            ctx.export("tokenValue", token.value());
        });
    }
}
