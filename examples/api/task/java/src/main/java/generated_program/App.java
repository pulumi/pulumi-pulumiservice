package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.api_agents.Pool;
import com.pulumi.pulumiservice.api_agents.PoolArgs;
import com.pulumi.pulumiservice.api_agents.Task;
import com.pulumi.pulumiservice.api_agents.TaskArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var organizationName = config.get("organizationName").orElse("service-provider-test-org");
            var taskSuffix = config.get("taskSuffix").orElse("dev");
            var taskID = config.get("taskID").orElse("example-task-id");

            var pool = new Pool("pool",
                PoolArgs.builder()
                    .orgName(organizationName)
                    .name("api-task-pool-" + taskSuffix)
                    .description("Pool used by the api task example")
                    .build());

            new Task("task",
                TaskArgs.builder()
                    .orgName(organizationName)
                    .taskID(taskID)
                    .approvalMode("manual")
                    .permissionMode("default")
                    .source("api")
                    .planMode(false)
                    .build());

            ctx.export("poolName", pool.name());
        });
    }
}
