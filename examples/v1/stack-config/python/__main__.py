import pulumi
import pulumi_pulumiservice.v1 as ps_v1

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
project_name = config.get("projectName") or "v1-stack-config-example"
stack_name = config.get("stackName") or "dev"
hook_url = config.get("hookUrl") or "https://example.invalid/hooks/example"
env_ref = config.get("envRef") or "organization/credentials"

parent_stack = ps_v1.stacks.Stack(
    "parentStack",
    org_name=organization_name,
    project_name=project_name,
    stack_name=stack_name,
)

ps_v1.stacks.Config(
    "config",
    org_name=organization_name,
    project_name=parent_stack.project_name,
    stack_name=parent_stack.stack_name,
    environment=env_ref,
)

ps_v1.stacks.Webhook(
    "hook",
    organization_name=organization_name,
    project_name=parent_stack.project_name,
    stack_name=parent_stack.stack_name,
    name="v1-stackhook",
    display_name="Stack hook example",
    payload_url=hook_url,
    active=True,
    format="pulumi",
)

pulumi.export("stack", parent_stack.id)
