package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api_esc.Environment;
import com.pulumi.pulumiservice.api_esc.EnvironmentArgs;
import com.pulumi.pulumiservice.api_esc.EnvironmentTag;
import com.pulumi.pulumiservice.api_esc.EnvironmentTagArgs;
import com.pulumi.resources.CustomResourceOptions;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var projectName = config.get("projectName").orElse("test-project");
            var envSuffix = config.get("envSuffix").orElse("dev");
            var tagValue = config.get("tagValue").orElse("env-tag-initial");

            var environment = new Environment("environment",
                EnvironmentArgs.builder()
                    .orgName(organizationName)
                    .project(projectName)
                    .name("testing-environment-" + envSuffix)
                    .build());

            var environmentTag = new EnvironmentTag("environmentTag",
                EnvironmentTagArgs.builder()
                    .orgName(organizationName)
                    .projectName(projectName)
                    .envName("testing-environment-" + envSuffix)
                    .name("purpose")
                    .value(tagValue)
                    .build(),
                CustomResourceOptions.builder().dependsOn(environment).build());

            ctx.export("environmentId", environment.id());
            ctx.export("environmentTagValue", environmentTag.value());
        });
    }
}
