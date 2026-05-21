import pulumi
import pulumi_pulumiservice.api as ps_api

config = pulumi.Config()
organization_name = config.get("organizationName") or "service-provider-test-org"
project_name = config.get("projectName") or "pulumi-service-stack-example"
stack_name = config.get("stackName") or "dev"
stack_purpose = config.get("stackPurpose") or "demo"

example_stack = ps_api.stacks.Stack(
    "exampleStack",
    org_name=organization_name,
    project_name=project_name,
    stack_name=stack_name,
    tags={
        "owner": "pulumicloud-api-example",
        "purpose": stack_purpose,
    },
)

pulumi.export("stackName", example_stack.stack_name)
