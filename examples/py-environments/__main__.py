"""A Python Pulumi Service Environments program"""

import pulumi
from pulumi_pulumiservice import Environment

environment = Environment(
    "testing-environment",
    organization="service-provider-test-org",
    name="testing-environment-py",
    yaml=pulumi.StringAsset("""
        values:
          myKey1: "myValue1"
          myNestedKey:
            myKey2: "myValue2"
            myNumber: 1
    """)
)
