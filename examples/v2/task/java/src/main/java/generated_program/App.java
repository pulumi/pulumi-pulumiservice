package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2.AgentPool;
import com.pulumi.pulumiservice.v2.AgentPoolArgs;
import com.pulumi.pulumiservice.v2.Task;
import com.pulumi.pulumiservice.v2.TaskArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var taskSuffix = config.get("taskSuffix").orElse("dev");
            var taskID = config.get("taskID").orElse("example-task-id");

            var pool = new AgentPool("pool",
                AgentPoolArgs.builder()
                    .orgName(serviceOrg)
                    .name("v2-task-pool-" + taskSuffix)
                    .description("Pool used by the v2 task example")
                    .build());

            new Task("task",
                TaskArgs.builder()
                    .orgName(serviceOrg)
                    .taskID(taskID)
                    .name("v2-task-" + taskSuffix)
                    .approvalMode("manual")
                    .permissionMode("maintainer")
                    .isShared(false)
                    .build());

            ctx.export("poolName", pool.name());
        });
    }
}
