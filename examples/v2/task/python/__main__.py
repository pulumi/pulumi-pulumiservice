import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
task_suffix = config.get("taskSuffix") or "dev"
task_id = config.get("taskID") or "example-task-id"

pool = ps_v2.agents.Pool(
    "pool",
    org_name=service_org,
    name=f"v2-task-pool-{task_suffix}",
    description="Pool used by the v2 task example",
)

ps_v2.agents.Task(
    "task",
    org_name=service_org,
    task_id=task_id,
    approval_mode="manual",
    permission_mode="default",
    source="api",
    plan_mode=False,
)

pulumi.export("poolName", pool.name)
