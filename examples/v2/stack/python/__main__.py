import pulumi
import pulumi_pulumiservice.v2 as ps_v2

config = pulumi.Config()
service_org = config.get("serviceOrg") or "service-provider-test-org"
project_name = config.get("projectName") or "pulumi-service-stack-example"
stack_name = config.get("stackName") or "dev"
stack_purpose = config.get("stackPurpose") or "demo"

example_stack = ps_v2.Stack(
    "exampleStack",
    org_name=service_org,
    project_name=project_name,
    stack_name=stack_name,
    tags={
        "owner": "pulumicloud-v2-example",
        "purpose": stack_purpose,
    },
)

pulumi.export("stackName", example_stack.stack_name)
