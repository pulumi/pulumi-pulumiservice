import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
project_name = config.get("projectName") or "v2-stack-tags-example"
stack_name = config.get("stackName") or "dev"
tag_value = config.get("tagValue") or "v2-tag-value"

parent_stack = ps_v2.stacks.Stack(
    "parentStack",
    org_name=service_org,
    project_name=project_name,
    stack_name=stack_name,
)

ps_v2.stacks.Tag(
    "ownerTag",
    org_name=service_org,
    project_name=parent_stack.project_name,
    stack_name=parent_stack.stack_name,
    name="owner",
    value="pulumicloud-v2-example",
)

ps_v2.stacks.Tag(
    "customTag",
    org_name=service_org,
    project_name=parent_stack.project_name,
    stack_name=parent_stack.stack_name,
    name="purpose",
    value=tag_value,
)

pulumi.export("parent", parent_stack.id)
