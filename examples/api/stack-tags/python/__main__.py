import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
project_name = config.get("projectName") or "api-stack-tags-example"
stack_name = config.get("stackName") or "dev"
tag_value = config.get("tagValue") or "api-tag-value"

parent_stack = ps_api.stacks.Stack(
    "parentStack",
    org_name=organization_name,
    project_name=project_name,
    stack_name=stack_name,
)

ps_api.stacks.Tag(
    "ownerTag",
    org_name=organization_name,
    project_name=parent_stack.project_name,
    stack_name=parent_stack.stack_name,
    name="owner",
    value="pulumicloud-api-example",
)

ps_api.stacks.Tag(
    "customTag",
    org_name=organization_name,
    project_name=parent_stack.project_name,
    stack_name=parent_stack.stack_name,
    name="purpose",
    value=tag_value,
)

pulumi.export("parent", parent_stack.id)
