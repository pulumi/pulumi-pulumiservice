"""A Python Pulumi Service Environments program"""

import pulumi
from pulumi_pulumiservice import Environment, EnvironmentVersionTag
from pulumi import ResourceOptions

config = pulumi.Config()

environment = Environment(
    "testing-environment",
    organization="service-provider-test-org",
    name="testing-environment-py-"+config.require('digits'),
    yaml=pulumi.StringAsset("""values:
  myKey1: "myValue1"
  myNestedKey:
    myKey2: "myValue2"
    myNumber: 1""")
)

# A tag that will always be placed on the latest revision of the environment
stableTag = EnvironmentVersionTag(
  "StableTag",
	organization=environment.organization,
	environment=environment.name,
	tag_name="stable",
	revision=environment.revision,
)

# A tag that will be placed on each new version, and remain on old revisions
versionTag = EnvironmentVersionTag(
  "VersionTag",
	organization=environment.organization,
	environment=environment.name,
	tag_name=environment.revision.apply(
    lambda revision: "v" + str(revision)
  ),
	revision=environment.revision,
  opts=ResourceOptions(
    retain_on_delete=True
  )
)
