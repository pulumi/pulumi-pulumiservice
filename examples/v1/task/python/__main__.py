import pulumi
import pulumi_pulumiservice.v1 as ps_v1

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
task_suffix = config.get("taskSuffix") or "dev"
task_id = config.get("taskID") or "example-task-id"

pool = ps_v1.agents.Pool(
    "pool",
    org_name=organization_name,
    name=f"v1-task-pool-{task_suffix}",
    description="Pool used by the v1 task example",
)

ps_v1.agents.Task(
    "task",
    org_name=organization_name,
    task_id=task_id,
    approval_mode="manual",
    permission_mode="default",
    source="api",
    plan_mode=False,
)

pulumi.export("poolName", pool.name)
