import pulumi
import pulumi_pulumiservice as pulumiservice

config = pulumi.Config()
org = config.require("organizationName")
name = "test-smoke-py-" + config.require("digits")

environment = pulumiservice.Environment(
    "environment",
    organization=org,
    project="test-smoke",
    name=name,
    yaml=pulumi.StringAsset("values:\n  greeting: hello\n"),
)

token = pulumiservice.OrgAccessToken(
    "token",
    organization_name=org,
    name=name,
    description="Created by the py-smoke example.",
    admin=False,
)

looked_up = pulumiservice.get_environment_output(
    organization_name=environment.organization,
    project_name=environment.project,
    name=environment.name,
)

pulumi.export("environmentName", looked_up.name)
pulumi.export("tokenValue", token.value)
