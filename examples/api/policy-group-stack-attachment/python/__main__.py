import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
group_name = config.get("groupName") or "example-attachment-group"
project_name = config.get("projectName") or "pulumi-service-attachment-example"
stack_name = config.get("stackName") or "dev"

example_stack = ps_api.stacks.Stack(
    "exampleStack",
    org_name=organization_name,
    project_name=project_name,
    stack_name=stack_name,
)

group = ps_api.PolicyGroup(
    "group",
    org_name=organization_name,
    name=group_name,
    entity_type="stacks",
)

attachment = ps_api.PolicyGroupStackAttachment(
    "attachment",
    org_name=organization_name,
    policy_group=group.name,
    name=example_stack.stack_name,
    routing_project=project_name,
    opts=pulumi.ResourceOptions(depends_on=[group, example_stack]),
)

pulumi.export("attachedStack", attachment.name)
