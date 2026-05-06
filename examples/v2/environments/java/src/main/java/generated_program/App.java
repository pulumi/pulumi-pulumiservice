package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2_esc.Environment;
import com.pulumi.pulumiservice.v2_esc.EnvironmentArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var projectName = config.get("projectName").orElse("test-project");
            var envSuffix = config.get("envSuffix").orElse("dev");

            var environment = new Environment("environment",
                EnvironmentArgs.builder()
                    .orgName(serviceOrg)
                    .project(projectName)
                    .name("testing-environment-" + envSuffix)
                    .build());

            ctx.export("envName", environment.name());
        });
    }
}
