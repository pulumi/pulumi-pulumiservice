import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
project_name = config.get("projectName") or "v2-stack-config-example"
stack_name = config.get("stackName") or "dev"
hook_url = config.get("hookUrl") or "https://example.invalid/hooks/example"
env_ref = config.get("envRef") or "organization/credentials"

parent_stack = ps_v2.Stack(
    "parentStack",
    org_name=service_org,
    project_name=project_name,
    stack_name=stack_name,
)

ps_v2.StackConfig(
    "config",
    org_name=service_org,
    project_name=parent_stack.project_name,
    stack_name=parent_stack.stack_name,
    environment=env_ref,
)

ps_v2.StackWebhook(
    "hook",
    organization_name=service_org,
    project_name=parent_stack.project_name,
    stack_name=parent_stack.stack_name,
    name="v2-stackhook",
    display_name="Stack hook example",
    payload_url=hook_url,
    active=True,
    format="pulumi",
)

pulumi.export("stack", parent_stack.id)
